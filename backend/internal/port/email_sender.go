package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

// EmailSender — adapter de saída SMTP (fundacao-metas.md §9). Não recebe
// uuid.UUID: por isso não precisa de autorizacao.Escopo, e a lista fechada
// das duas portas estreitas não cresce (TestNenhumaPortaNovaSemEscopo).
// Implementado com circuit breaker no adapter, nunca no use case — o
// relê é o único chamador sancionado (design.md §8.4).
type EmailSender interface {
	Enviar(ctx context.Context, destino valueobject.Email, assunto, corpo string) error
}
