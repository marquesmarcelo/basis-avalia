package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// Portaria — texto livre, não vazio, ≤ 100 caracteres. Sem formato e sem
// unicidade (PC-6, specs/cursos/spec.md): a tela não pode recusar o que o
// papel administrativo já permite.
type Portaria struct {
	valor string
}

func NovaPortaria(bruta string) (Portaria, error) {
	normalizada := strings.TrimSpace(bruta)
	if normalizada == "" || len(normalizada) > 100 {
		return Portaria{}, &domain.ErrValidacao{Campo: "portaria", Mensagem: "A portaria é obrigatória."}
	}
	return Portaria{valor: normalizada}, nil
}

func (p Portaria) String() string { return p.valor }
