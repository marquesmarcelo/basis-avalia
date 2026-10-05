package valueobject

import (
	"net/mail"
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// Email — Value Object. Normalização (trim + minúsculas) acontece aqui,
// no construtor, nunca no handler (design.md §3.1, U-04).
type Email struct {
	valor string
}

func NovoEmail(bruto string) (Email, error) {
	normalizado := strings.ToLower(strings.TrimSpace(bruto))
	if normalizado == "" || len(normalizado) > 320 {
		return Email{}, &domain.ErrValidacao{Campo: "email", Mensagem: "Informe um e-mail válido."}
	}
	if _, err := mail.ParseAddress(normalizado); err != nil {
		return Email{}, &domain.ErrValidacao{Campo: "email", Mensagem: "Informe um e-mail válido."}
	}
	return Email{valor: normalizado}, nil
}

func (e Email) String() string          { return e.valor }
func (e Email) Equals(outro Email) bool { return e.valor == outro.valor }
