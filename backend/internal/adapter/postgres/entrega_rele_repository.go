package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

var _ port.EntregaReleRepository = (*EntregaRepository)(nil)

// tentativasMaximasParaBackoff — teto do expoente do backoff (design.md
// §8.4, T-127): 2^10 minutos (~17h) é intervalo grande o bastante para
// não martelar um destino permanentemente quebrado, sem crescer sem
// limite linha a linha.
const tentativasMaximasParaBackoff = 10

// ReivindicarNotificacoesPendentes — T-124/T-127 (design.md §8.4, revisão
// pós code-review). Dois statements, nunca um SELECT ... FOR UPDATE SKIP
// LOCKED isolado:
//
//  1. UPDATE ... WHERE id IN (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING
//     id — atômico mesmo em autocommit porque é UM statement: o SKIP
//     LOCKED evita que duas réplicas reivindiquem a mesma linha na
//     corrida, e a ESCRITA (notificacao_reivindicada_em) sobrevive ao
//     envio de e-mail que vem depois, ao contrário de um lock (que o
//     autocommit já teria liberado antes do SMTP rodar).
//  2. Um SELECT comum (sem lock) que busca os dados das linhas já
//     reivindicadas — não precisa de proteção própria, a proteção já
//     aconteceu no passo 1.
//
// O envio de e-mail em si roda inteiramente FORA de qualquer transação
// (rele.go), entre os dois passos: reivindicar (rápido, no banco) → enviar
// (lento, rede) → marcar enviada (rápido, no banco).
//
// `hoje` decide quem é o destinatário (designação vigente, DG-06) — vem do
// fuso de exibição, nunca de UTC (T-124: o dia de exibição diverge do dia
// UTC perto da meia-noite de Brasília).
func (r *EntregaRepository) ReivindicarNotificacoesPendentes(ctx context.Context, hoje valueobject.DataLocal, limite int) ([]port.EntregaParaNotificar, error) {
	reivindicacao := `
		UPDATE entrega
		   SET notificacao_reivindicada_em = now()
		 WHERE id IN (
		         SELECT entrega.id
		           FROM entrega
		          WHERE entrega.notificacao_enviada_em IS NULL
		            AND entrega.notificacao_evento IS NOT NULL
		            AND EXISTS (
		                  SELECT 1 FROM designacao d
		                   WHERE d.curso_id = entrega.curso_id AND d.excluido_em IS NULL
		                     AND ` + FragmentoDesignacaoVigente("d", 2) + `
		                )
		            AND (
		                  entrega.notificacao_reivindicada_em IS NULL
		               OR entrega.notificacao_reivindicada_em < now() - (
		                    interval '1 minute' * power(2, LEAST(entrega.notificacao_tentativas, $3))
		                  )
		                )
		          ORDER BY entrega.notificacao_gerada_em
		          FOR UPDATE OF entrega SKIP LOCKED
		          LIMIT $1
		       )
		 RETURNING id`
	var idsReivindicados []uuid.UUID
	if err := r.db.SelectContext(ctx, &idsReivindicados, reivindicacao, limite, hoje.String(), tentativasMaximasParaBackoff); err != nil {
		return nil, err
	}
	if len(idsReivindicados) == 0 {
		return nil, nil
	}

	// IN com placeholders numerados, nunca sqlx.In/array: os ids já
	// reivindicados são no máximo `limite` (o lote do relê, ~20) — uma
	// lista de $N é mais simples e não depende de suporte a array do
	// driver para o tipo uuid.UUID.
	marcadores := make([]string, len(idsReivindicados))
	args := make([]any, 0, len(idsReivindicados)+1)
	for i, id := range idsReivindicados {
		marcadores[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}
	nHoje := len(idsReivindicados) + 1
	args = append(args, hoje.String())

	consulta := `
		SELECT entrega.id AS entrega_id, entrega.notificacao_evento AS evento,
		       curso.nome AS curso_nome, meta.nome AS meta_nome, entrega.motivo,
		       entrega.prazo_correcao_ate, entrega.rodadas_de_recusa,
		       coord.email AS destinatario_email, coord.nome AS destinatario_nome,
		       entrega.notificacao_tentativas
		  FROM entrega
		  JOIN item_plano ON item_plano.id = entrega.item_plano_id
		  JOIN meta ON meta.id = item_plano.meta_id
		  JOIN curso ON curso.id = entrega.curso_id
		  JOIN designacao d ON d.curso_id = entrega.curso_id AND d.excluido_em IS NULL AND ` + FragmentoDesignacaoVigente("d", nHoje) + `
		  JOIN usuario coord ON coord.id = d.coordenador_id
		 WHERE entrega.id IN (` + strings.Join(marcadores, ",") + `)
		 ORDER BY entrega.notificacao_gerada_em`

	var linhas []struct {
		EntregaID             uuid.UUID  `db:"entrega_id"`
		Evento                string     `db:"evento"`
		CursoNome             string     `db:"curso_nome"`
		MetaNome              string     `db:"meta_nome"`
		Motivo                string     `db:"motivo"`
		PrazoCorrecao         *time.Time `db:"prazo_correcao_ate"`
		Rodadas               int        `db:"rodadas_de_recusa"`
		DestinatarioEmail     string     `db:"destinatario_email"`
		DestinatarioNome      string     `db:"destinatario_nome"`
		NotificacaoTentativas int        `db:"notificacao_tentativas"`
	}
	if err := r.db.SelectContext(ctx, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	resultado := make([]port.EntregaParaNotificar, 0, len(linhas))
	for _, l := range linhas {
		resultado = append(resultado, port.EntregaParaNotificar{
			EntregaID: l.EntregaID, Evento: l.Evento, CursoNome: l.CursoNome, MetaNome: l.MetaNome,
			Motivo: l.Motivo, PrazoCorrecao: l.PrazoCorrecao, Rodadas: l.Rodadas,
			DestinatarioEmail: l.DestinatarioEmail, DestinatarioNome: l.DestinatarioNome,
			NotificacaoTentativas: l.NotificacaoTentativas,
		})
	}
	return resultado, nil
}

func (r *EntregaRepository) ContarNotificacoesPendentes(ctx context.Context) (int, error) {
	var total int
	err := r.db.GetContext(ctx, &total,
		`SELECT count(*) FROM entrega WHERE notificacao_enviada_em IS NULL AND notificacao_evento IS NOT NULL`)
	return total, err
}

func (r *EntregaRepository) MarcarNotificacaoEnviada(ctx context.Context, entregaID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE entrega SET notificacao_enviada_em = now() WHERE id = $1`, entregaID)
	return err
}

// RegistrarFalhaDeNotificacao — nunca apaga a linha pendente (design.md
// §8.4): incrementa tentativas e guarda o último erro, para o backoff
// exponencial do relê decidir a próxima tentativa.
func (r *EntregaRepository) RegistrarFalhaDeNotificacao(ctx context.Context, entregaID uuid.UUID, erro string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE entrega SET notificacao_tentativas = notificacao_tentativas + 1, notificacao_ultimo_erro = $2 WHERE id = $1`,
		entregaID, erro)
	return err
}

// ListarCandidatosARestauracaoDePrazo — PM-4/VG-06 (design.md §10): o
// segundo predicado (prazo expirou DEPOIS de o curso ficar vago) é o que
// separa esta restauração do caso que não se restaura. "Início da
// vacância" é data_fim da última designação encerrada + 1 dia.
func (r *EntregaRepository) ListarCandidatosARestauracaoDePrazo(ctx context.Context, limite int) ([]port.CandidatoRestauracaoDePrazo, error) {
	consulta := `
		SELECT entrega.id AS entrega_id, entrega.curso_id
		  FROM entrega
		  JOIN LATERAL (
		    SELECT d.data_inicio
		      FROM designacao d
		     WHERE d.curso_id = entrega.curso_id AND d.excluido_em IS NULL
		       AND d.data_inicio > coalesce((
		             SELECT max(d2.data_fim) FROM designacao d2
		              WHERE d2.curso_id = entrega.curso_id AND d2.excluido_em IS NULL AND d2.data_fim IS NOT NULL
		           ), '0001-01-01')
		     ORDER BY d.data_inicio DESC LIMIT 1
		  ) proxima ON TRUE
		  JOIN LATERAL (
		    SELECT max(d3.data_fim) AS fim_da_vacancia
		      FROM designacao d3
		     WHERE d3.curso_id = entrega.curso_id AND d3.excluido_em IS NULL AND d3.data_fim IS NOT NULL
		       AND d3.data_fim < proxima.data_inicio
		  ) vacancia ON TRUE
		 WHERE entrega.excluido_em IS NULL AND entrega.situacao = 'recusada'
		   AND entrega.prazo_correcao_ate < (proxima.data_inicio::timestamptz)
		   AND (vacancia.fim_da_vacancia IS NULL OR entrega.prazo_correcao_ate >= (vacancia.fim_da_vacancia::timestamptz))
		   AND coalesce(entrega.notificacao_evento, '') <> 'prazo_restaurado'
		 FOR UPDATE OF entrega SKIP LOCKED
		 LIMIT $1`
	var linhas []struct {
		EntregaID uuid.UUID `db:"entrega_id"`
		CursoID   uuid.UUID `db:"curso_id"`
	}
	if err := r.db.SelectContext(ctx, &linhas, consulta, limite); err != nil {
		return nil, err
	}
	resultado := make([]port.CandidatoRestauracaoDePrazo, 0, len(linhas))
	for _, l := range linhas {
		resultado = append(resultado, port.CandidatoRestauracaoDePrazo{EntregaID: l.EntregaID, CursoID: l.CursoID})
	}
	return resultado, nil
}

func (r *EntregaRepository) RestaurarPrazoPorVacancia(ctx context.Context, entregaID uuid.UUID, novoPrazo time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE entrega SET prazo_correcao_ate = $2, notificacao_evento = 'prazo_restaurado',
		        notificacao_gerada_em = now(), notificacao_enviada_em = NULL,
		        notificacao_tentativas = 0, notificacao_ultimo_erro = NULL, atualizado_em = now()
		  WHERE id = $1`,
		entregaID, novoPrazo)
	return err
}

func (r *EntregaRepository) ExpurgarIdempotenciaAntesDe(ctx context.Context, antes time.Time) (int64, error) {
	resultado, err := r.db.ExecContext(ctx, `DELETE FROM idempotencia WHERE criado_em < $1`, antes)
	if err != nil {
		return 0, err
	}
	return resultado.RowsAffected()
}
