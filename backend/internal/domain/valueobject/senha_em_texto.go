package valueobject

import (
	"unicode/utf8"

	"github.com/basis-avalia/backend/internal/domain"
)

// SenhaEmTexto — Value Object. A única validação é presença e o limite
// técnico de fronteira de confiança (P12) — nunca política de senha (3.3).
// Sem trim: espaço é caractere válido de senha.
type SenhaEmTexto struct {
	valor string
}

func NovaSenhaEmTexto(bruta string) (SenhaEmTexto, error) {
	if bruta == "" {
		return SenhaEmTexto{}, domain.ErrSenhaObrigatoria
	}
	if utf8.RuneCountInString(bruta) > 1024 {
		return SenhaEmTexto{}, domain.ErrSenhaAcimaDoLimite
	}
	return SenhaEmTexto{valor: bruta}, nil
}

// Revelar é o único acesso ao valor bruto — usado só pelo adapter de hash.
func (s SenhaEmTexto) Revelar() string { return s.valor }

// String nunca expõe o valor — log acidental é inofensivo.
func (s SenhaEmTexto) String() string { return "[senha omitida]" }
