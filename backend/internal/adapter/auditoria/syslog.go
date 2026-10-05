package auditoria

import (
	"fmt"
	"log/syslog"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
)

// Syslog — canal 2 (espelho para SIEM). Quando SYSLOG_SERVER está vazio o
// adapter é no-op (design.md §8.1) — estado padrão em desenvolvimento.
type Syslog struct {
	writer *syslog.Writer
}

// NovoSyslog nunca falha a inicialização do processo: servidor vazio ou
// inalcançável sempre devolvem um Syslog utilizável (no-op), mais o erro
// para quem chama apenas registrar no log — nunca abortar a subida do
// serviço por causa do canal 2 (design.md §8.1).
func NovoSyslog(servidor, protocolo, appName string) (*Syslog, error) {
	if servidor == "" {
		return &Syslog{}, nil
	}
	w, err := syslog.Dial(protocolo, servidor, syslog.LOG_INFO|syslog.LOG_LOCAL0, appName)
	if err != nil {
		return &Syslog{}, err
	}
	return &Syslog{writer: w}, nil
}

func (s *Syslog) Enviar(e auditoria.Evento) error {
	if s == nil || s.writer == nil {
		return nil
	}
	ator := "-"
	if e.AtorID != nil {
		ator = e.AtorID.String()
	}
	instituicao := "-"
	if e.InstituicaoID != nil {
		instituicao = e.InstituicaoID.String()
	}
	recurso := "-"
	if e.RecursoID != nil {
		recurso = fmt.Sprintf("%s/%s", e.RecursoTipo, e.RecursoID.String())
	}
	msg := fmt.Sprintf("acao=%s resultado=%s ator=%s instituicao=%s recurso=%s executado_em=%s",
		e.Acao, e.Resultado, ator, instituicao, recurso, e.ExecutadoEm.Format("2006-01-02T15:04:05Z07:00"))
	return s.writer.Info(msg)
}
