package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
)

// AuditLogger — canal único do ponto de vista do use case; a implementação
// combina canal local (transacional) e syslog (fire-and-forget pós-commit)
// (design.md §4.2, §8.1).
type AuditLogger interface {
	Registrar(ctx context.Context, e auditoria.Evento) error
}
