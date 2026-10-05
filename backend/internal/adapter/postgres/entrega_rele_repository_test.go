package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// montarPendenteDeNotificacao cria curso + designação vigente + item de
// plano + entrega recusada, e marca a entrega como pendente de
// notificação (notificacao_evento/gerada_em) — o mínimo que
// ReivindicarNotificacoesPendentes precisa para encontrar um destinatário.
func montarPendenteDeNotificacao(t *testing.T, db *sqlx.DB, instituicaoID, coordenadorID uuid.UUID, motivo string) uuid.UUID {
	t.Helper()
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{
		InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente",
	})
	itemPlanoID := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{
		PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1,
	})
	entregaID := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{
		ItemPlanoID: itemPlanoID, CursoID: cursoID, InstituicaoID: instituicaoID,
		EnviadaPor: coordenadorID, Situacao: "recusada", Motivo: motivo,
	})
	if _, err := db.Exec(
		`UPDATE entrega SET notificacao_evento = 'recusa', notificacao_gerada_em = now() WHERE id = $1`, entregaID,
	); err != nil {
		t.Fatalf("marcar entrega como pendente de notificação: %v", err)
	}
	return entregaID
}

// T-127 (fundacao-metas.md §4.7, design.md §8.4): a reivindicação é uma
// ESCRITA — a segunda chamada, imediatamente em seguida (simula uma
// segunda réplica na mesma janela), não pode reivindicar a MESMA entrega
// de novo. O antigo SELECT...FOR UPDATE SKIP LOCKED isolado não garantia
// isto (o lock já tinha sido liberado ao fim do próprio SELECT).
func TestEntregaRepository_T127_SegundaReivindicacaoNaMesmaJanelaNaoRepete(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fusoBrasilia, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("carregar fuso: %v", err)
	}
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	entregaID := montarPendenteDeNotificacao(t, db, instituicaoID, coordenadorID, "Ata sem assinatura")

	repo := NovoEntregaRepository(db)
	hoje := valueobject.DataLocalDe(time.Now(), fusoBrasilia)

	primeira, err := repo.ReivindicarNotificacoesPendentes(context.Background(), hoje, 20)
	if err != nil {
		t.Fatalf("primeira reivindicação: %v", err)
	}
	if !contemEntrega(primeira, entregaID) {
		t.Fatalf("esperava a entrega %s na primeira reivindicação, obteve %v", entregaID, primeira)
	}

	segunda, err := repo.ReivindicarNotificacoesPendentes(context.Background(), hoje, 20)
	if err != nil {
		t.Fatalf("segunda reivindicação: %v", err)
	}
	if contemEntrega(segunda, entregaID) {
		t.Fatalf("segunda réplica reivindicou a MESMA entrega %s — dupla notificação (T-127)", entregaID)
	}
}

// T-127: se o processo cair entre reivindicar e marcar como enviada (ou
// entre reivindicar e o e-mail sair), a reivindicação não pode travar a
// entrega para sempre — ela precisa voltar a ficar elegível depois que a
// janela de backoff passa. Semântica honesta: ao menos uma vez.
func TestEntregaRepository_T127_ReivindicacaoExpiradaVoltaAFicarElegivel(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fusoBrasilia, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("carregar fuso: %v", err)
	}
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	entregaID := montarPendenteDeNotificacao(t, db, instituicaoID, coordenadorID, "Comprovante ilegível")

	// Simula uma reivindicação de uma volta anterior que nunca terminou
	// (processo caiu): tentativas = 0 → janela de backoff = 2^0 = 1
	// minuto; reivindicada 2 minutos atrás já está fora da janela.
	if _, err := db.Exec(
		`UPDATE entrega SET notificacao_reivindicada_em = now() - interval '2 minutes' WHERE id = $1`, entregaID,
	); err != nil {
		t.Fatalf("simular reivindicação expirada: %v", err)
	}

	repo := NovoEntregaRepository(db)
	hoje := valueobject.DataLocalDe(time.Now(), fusoBrasilia)

	reivindicadas, err := repo.ReivindicarNotificacoesPendentes(context.Background(), hoje, 20)
	if err != nil {
		t.Fatalf("reivindicar: %v", err)
	}
	if !contemEntrega(reivindicadas, entregaID) {
		t.Fatalf("esperava a entrega %s elegível de novo após a janela de backoff expirar, obteve %v", entregaID, reivindicadas)
	}
}

// T-124 (fundacao-metas.md §5.1, DG-06): a data de referência que decide
// quem recebe a notificação é a do FUSO DE EXIBIÇÃO, nunca UTC. O teste
// roda com o container em UTC de propósito (como TestVigencia_
// RotuloEPredicadoConcordam) — se o adapter voltar a usar `time.Now().
// UTC()` em vez do `hoje` recebido, a designação que termina em 31/07
// aparenta já ter encerrado às 23h58 de Brasília (ainda 31/07 local, mas
// já 01/08 em UTC), e a notificação não seria reivindicada.
func TestEntregaRepository_T124_HojeDoFusoDecideDestinatario(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fusoBrasilia, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("carregar fuso: %v", err)
	}
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	fimEmJulho := "2026-07-31"
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID,
		DataInicio: "2026-01-01", DataFim: &fimEmJulho,
	})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{
		InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente",
	})
	itemPlanoID := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{
		PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1,
	})
	entregaID := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{
		ItemPlanoID: itemPlanoID, CursoID: cursoID, InstituicaoID: instituicaoID,
		EnviadaPor: coordenadorID, Situacao: "recusada", Motivo: "Prazo perdido",
	})
	if _, err := db.Exec(
		`UPDATE entrega SET notificacao_evento = 'recusa', notificacao_gerada_em = now() WHERE id = $1`, entregaID,
	); err != nil {
		t.Fatalf("marcar entrega como pendente de notificação: %v", err)
	}

	repo := NovoEntregaRepository(db)

	// 23h58 de 31/07 em Brasília: a designação AINDA é vigente no dia de
	// exibição — a notificação deve ser reivindicada.
	hojeAindaVigente := valueobject.DataLocalDe(time.Date(2026, 7, 31, 23, 58, 0, 0, fusoBrasilia), fusoBrasilia)
	reivindicadas, err := repo.ReivindicarNotificacoesPendentes(context.Background(), hojeAindaVigente, 20)
	if err != nil {
		t.Fatalf("reivindicar (ainda vigente): %v", err)
	}
	if !contemEntrega(reivindicadas, entregaID) {
		t.Fatalf("esperava a entrega %s reivindicada às 23h58 de 31/07 em Brasília (designação ainda vigente no dia local)", entregaID)
	}

	// Devolve a reivindicação para poder testar a segunda fronteira sem
	// depender da janela de backoff.
	if _, err := db.Exec(`UPDATE entrega SET notificacao_reivindicada_em = NULL WHERE id = $1`, entregaID); err != nil {
		t.Fatalf("devolver reivindicação: %v", err)
	}

	// 00h02 de 01/08 em Brasília: a designação já encerrou no dia de
	// exibição — a notificação NÃO deve ser reivindicada (sem destinatário
	// vigente para notificar).
	hojeJaEncerrada := valueobject.DataLocalDe(time.Date(2026, 8, 1, 0, 2, 0, 0, fusoBrasilia), fusoBrasilia)
	reivindicadas, err = repo.ReivindicarNotificacoesPendentes(context.Background(), hojeJaEncerrada, 20)
	if err != nil {
		t.Fatalf("reivindicar (já encerrada): %v", err)
	}
	if contemEntrega(reivindicadas, entregaID) {
		t.Fatalf("entrega %s reivindicada às 00h02 de 01/08 em Brasília, mas a designação já encerrou no dia local — bug de fuso (T-124)", entregaID)
	}
}

func contemEntrega(lista []port.EntregaParaNotificar, entregaID uuid.UUID) bool {
	for _, p := range lista {
		if p.EntregaID == entregaID {
			return true
		}
	}
	return false
}
