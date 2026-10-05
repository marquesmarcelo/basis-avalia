package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// Sigla — normalizada em MAIÚSCULAS: aparece no cabeçalho e no rodapé, e
// duas variações de caixa convivendo tornaria a tela ambígua (P15).
type Sigla struct {
	valor string
}

func NovaSigla(bruta string) (Sigla, error) {
	normalizada := strings.ToUpper(strings.TrimSpace(bruta))
	if normalizada == "" || len(normalizada) > 20 {
		return Sigla{}, &domain.ErrValidacao{Campo: "sigla", Mensagem: "A sigla é obrigatória."}
	}
	return Sigla{valor: normalizada}, nil
}

func (s Sigla) String() string { return s.valor }
