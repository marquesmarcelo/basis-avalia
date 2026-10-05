package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// NomeCatalogo — não vazio, no máximo 300 caracteres. Reaproveitado por
// Indicador e Meta (specs/indicadores/design.md §3.1).
type NomeCatalogo struct {
	valor string
}

func NovoNomeCatalogo(bruto string) (NomeCatalogo, error) {
	normalizado := strings.TrimSpace(bruto)
	if normalizado == "" || len(normalizado) > 300 {
		return NomeCatalogo{}, &domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."}
	}
	return NomeCatalogo{valor: normalizado}, nil
}

func (n NomeCatalogo) String() string { return n.valor }
