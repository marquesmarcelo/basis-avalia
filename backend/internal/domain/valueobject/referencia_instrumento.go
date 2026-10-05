package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// ReferenciaInstrumento — texto livre, no máximo 500 caracteres. A regra de
// obrigatoriedade (obrigatória em plataforma, proibida em instituição) não
// mora aqui — mora na entidade Indicador, porque depende do escopo
// (specs/indicadores/design.md §3.1).
type ReferenciaInstrumento struct {
	valor string
}

func NovaReferenciaInstrumento(bruta string) (ReferenciaInstrumento, error) {
	normalizada := strings.TrimSpace(bruta)
	if len(normalizada) > 500 {
		return ReferenciaInstrumento{}, &domain.ErrValidacao{Campo: "referencia_instrumento", Mensagem: "A referência do instrumento é grande demais."}
	}
	return ReferenciaInstrumento{valor: normalizada}, nil
}

func (r ReferenciaInstrumento) Vazia() bool    { return r.valor == "" }
func (r ReferenciaInstrumento) String() string { return r.valor }
