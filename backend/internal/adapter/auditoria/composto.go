package auditoria

import (
	"context"
	"log"

	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var eventosTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "auditoria_eventos_total",
	Help: "Total de eventos de auditoria, por canal, ação e resultado.",
}, []string{"canal", "acao", "resultado"})

// Composto — local (transacional, autoritativo) + syslog (fire-and-forget
// pós-commit). Falha do syslog nunca falha a operação de negócio
// (design.md §8.1).
type Composto struct {
	local  port.AuditLogger
	syslog *Syslog
}

func NovoComposto(local port.AuditLogger, sys *Syslog) *Composto {
	return &Composto{local: local, syslog: sys}
}

var _ port.AuditLogger = (*Composto)(nil)

func (c *Composto) Registrar(ctx context.Context, e auditoria.Evento) error {
	if err := c.local.Registrar(ctx, e); err != nil {
		eventosTotal.WithLabelValues("local", string(e.Acao), "erro").Inc()
		return err
	}
	eventosTotal.WithLabelValues("local", string(e.Acao), string(e.Resultado)).Inc()

	if !c.deveIrParaSyslog(e) {
		return nil
	}

	postgres.AgendarAposCommit(ctx, func() {
		if err := c.syslog.Enviar(e); err != nil {
			eventosTotal.WithLabelValues("syslog", string(e.Acao), "erro").Inc()
			log.Printf("auditoria: falha ao enviar evento %s ao syslog: %v", e.ID, err)
			return
		}
		eventosTotal.WithLabelValues("syslog", string(e.Acao), string(e.Resultado)).Inc()
	})
	return nil
}

// deveIrParaSyslog cobre a exceção condicional da §11 da spec:
// atualizar_usuario (e atualizar_administrador) só vão ao syslog quando o
// conjunto de perfis mudou (design.md §8.1).
func (c *Composto) deveIrParaSyslog(e auditoria.Evento) bool {
	if e.Acao == auditoria.AtualizarUsuario || e.Acao == auditoria.AtualizarAdministrador {
		_, mudouPerfis := e.Detalhes["perfis_anterior"]
		return mudouPerfis
	}
	// specs/metas-coordenacao/spec.md §10: baixar_anexo só vai ao syslog
	// quando negado — download bem-sucedido é rotina, negado é o que
	// interessa a uma auditoria externa.
	if e.Acao == auditoria.BaixarAnexo {
		return e.Resultado == auditoria.ResultadoNegado
	}
	return e.Acao.VaiParaSyslog()
}
