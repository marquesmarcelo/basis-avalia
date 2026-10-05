package rele

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

// LoteMaximo — quantas entregas o relê processa por volta (design.md §8.4).
const LoteMaximo = 20

// Rele é o único processo de fundo do sistema (fundacao-metas.md §7): uma
// goroutine com ticker, sem agendador externo, sem mensageria — nenhum
// dos dois existe no projeto e nenhum se justifica só por isto.
//
// Duas tarefas por volta: enviar as notificações pendentes (§8.4) e
// restaurar prazo de correção após vacância (§10, PM-4/VG-06). A janela
// entre execuções já funciona como backoff natural para falhas de envio
// — uma falha só é retentada na volta seguinte, nunca imediatamente.
type Rele struct {
	repo           port.EntregaReleRepository
	email          port.EmailSender
	audit          port.AuditLogger
	relogio        port.Relogio
	metricas       port.MetricasDoRele
	fusoDeExibicao *time.Location
	intervalo      time.Duration
}

func Novo(repo port.EntregaReleRepository, email port.EmailSender, audit port.AuditLogger, relogio port.Relogio, metricas port.MetricasDoRele, fuso *time.Location, intervalo time.Duration) *Rele {
	return &Rele{repo: repo, email: email, audit: audit, relogio: relogio, metricas: metricas, fusoDeExibicao: fuso, intervalo: intervalo}
}

// Iniciar bloqueia até ctx ser cancelado — chamado numa goroutine própria
// pelo main(). RELE_HABILITADO=false é o botão de desligar por
// configuração (CLAUDE.md, seção de incidente) — quem decide isso é o
// chamador, não este tipo: se o relê nunca é iniciado, ele nunca roda.
func (r *Rele) Iniciar(ctx context.Context) {
	ticker := time.NewTicker(r.intervalo)
	defer ticker.Stop()
	for {
		r.executarUmaVolta(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Rele) executarUmaVolta(ctx context.Context) {
	r.enviarNotificacoesPendentes(ctx)
	r.restaurarPrazosPorVacancia(ctx)

	if total, err := r.repo.ContarNotificacoesPendentes(ctx); err == nil {
		r.metricas.RegistrarNotificacoesPendentes(total)
	} else {
		log.Printf("rele: erro ao contar notificações pendentes: %v", err)
	}
}

// enviarNotificacoesPendentes — a reivindicação atômica (T-127, design.md
// §8.4) é o que protege contra N réplicas enviando N e-mails para a mesma
// recusa; o envio de e-mail que segue roda inteiramente fora de qualquer
// lock ou transação. `hoje` decide QUEM recebe (designação vigente,
// DG-06) e vem sempre do fuso de exibição — nunca de UTC (T-124).
func (r *Rele) enviarNotificacoesPendentes(ctx context.Context) {
	hoje := valueobject.DataLocalDe(r.relogio.Agora(), r.fusoDeExibicao)
	pendentes, err := r.repo.ReivindicarNotificacoesPendentes(ctx, hoje, LoteMaximo)
	if err != nil {
		log.Printf("rele: erro ao reivindicar notificações pendentes: %v", err)
		return
	}
	for _, p := range pendentes {
		destino, err := valueobject.NovoEmail(p.DestinatarioEmail)
		if err != nil {
			log.Printf("rele: e-mail de destino inválido para a entrega %s: %v", p.EntregaID, err)
			_ = r.repo.RegistrarFalhaDeNotificacao(ctx, p.EntregaID, "e-mail de destino inválido")
			continue
		}
		assunto, corpo := montarMensagem(p)
		// Falha de e-mail NUNCA desfaz a recusa e nunca devolve erro ao PI
		// (NT-03) — o relê só registra a tentativa e segue para a
		// próxima. O circuit breaker vive no adapter de SMTP, nunca aqui.
		if err := r.email.Enviar(ctx, destino, assunto, corpo); err != nil {
			log.Printf("rele: falha ao enviar notificação da entrega %s: %v", p.EntregaID, err)
			_ = r.repo.RegistrarFalhaDeNotificacao(ctx, p.EntregaID, err.Error())
			continue
		}
		// Semântica honesta: ao menos uma vez (at-least-once). Se o
		// processo cair EXATAMENTE aqui — depois do SMTP confirmar o envio
		// e antes deste UPDATE — a reivindicação expira e a mesma entrega
		// é reivindicada de novo na volta seguinte, duplicando o e-mail.
		// Não corrigimos isso com transação: segurar uma transação através
		// da chamada SMTP (latência imprevisível) trocaria "risco raro de
		// e-mail duplicado" por "conexão/lock preso enquanto o servidor de
		// e-mail está lento". Preferimos duplicar a perder o alerta — um
		// coordenador recebendo o mesmo e-mail duas vezes é aborrecimento;
		// um coordenador que nunca soube da recusa é o defeito que este
		// relê existe para evitar.
		if err := r.repo.MarcarNotificacaoEnviada(ctx, p.EntregaID); err != nil {
			log.Printf("rele: falha ao marcar notificação enviada da entrega %s: %v", p.EntregaID, err)
			continue
		}
		r.metricas.RegistrarNotificacaoEnviada(p.Evento)
	}
}

func montarMensagem(p port.EntregaParaNotificar) (assunto, corpo string) {
	switch p.Evento {
	case "desfazimento":
		assunto = "A aceitação de uma entrega foi desfeita — " + p.CursoNome
	case "prazo_restaurado":
		assunto = "Prazo de correção restaurado — " + p.CursoNome
	default:
		assunto = "Entrega recusada — " + p.CursoNome
	}
	prazoTxt := ""
	if p.PrazoCorrecao != nil {
		prazoTxt = p.PrazoCorrecao.Format("02/01/2006 15:04")
	}
	// Sem anexo, sem conteúdo de comprovante, sem dado de terceiro — o
	// e-mail é transferência de dado para fora do sistema e diz o mínimo
	// (design.md §8.4).
	corpo = "Curso: " + p.CursoNome + "\nMeta: " + p.MetaNome + "\nMotivo: " + p.Motivo +
		"\nPrazo: " + prazoTxt + "\nRodada: " + strconv.Itoa(p.Rodadas) + " de 3"
	return assunto, corpo
}

// restaurarPrazosPorVacancia — PM-4/VG-06 (design.md §10): entrega
// recusada cujo prazo expirou DURANTE a vacância do curso ganha 7 dias
// novos a partir do início da designação seguinte, sem consumir rodada.
// Auditado como ação do sistema, na mesma volta em que é restaurado.
func (r *Rele) restaurarPrazosPorVacancia(ctx context.Context) {
	candidatos, err := r.repo.ListarCandidatosARestauracaoDePrazo(ctx, LoteMaximo)
	if err != nil {
		log.Printf("rele: erro ao listar candidatos a restauração de prazo: %v", err)
		return
	}
	agora := r.relogio.Agora()
	novoPrazo := valueobject.NovoPrazoDeCorrecao(agora, 7, r.fusoDeExibicao).Limite()
	for _, cand := range candidatos {
		if err := r.repo.RestaurarPrazoPorVacancia(ctx, cand.EntregaID, novoPrazo); err != nil {
			log.Printf("rele: falha ao restaurar prazo da entrega %s: %v", cand.EntregaID, err)
			continue
		}
		evento := auditoria.NovoEvento(auditoria.RestaurarPrazoPorVacancia, auditoria.ResultadoSucesso)
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &cand.EntregaID
		evento.Detalhes["curso_id"] = cand.CursoID.String()
		if err := r.audit.Registrar(ctx, evento); err != nil {
			log.Printf("rele: falha ao auditar restauração de prazo da entrega %s: %v", cand.EntregaID, err)
		}
	}
}
