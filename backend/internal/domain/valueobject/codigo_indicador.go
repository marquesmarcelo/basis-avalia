package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// CodigoIndicador — não vazio, no máximo 50 caracteres, sem quebra de
// linha. Não normaliza para maiúscula: "1.4" e "GEST-01" convivem, e o
// texto é o que as pessoas usam para se referir ao indicador
// (specs/indicadores/design.md §3.1).
type CodigoIndicador struct {
	valor string
}

func NovoCodigoIndicador(bruto string) (CodigoIndicador, error) {
	normalizado := strings.TrimSpace(bruto)
	if normalizado == "" || len(normalizado) > 50 {
		return CodigoIndicador{}, &domain.ErrValidacao{Campo: "codigo", Mensagem: "O código é obrigatório."}
	}
	if strings.ContainsAny(normalizado, "\n\r") {
		return CodigoIndicador{}, &domain.ErrValidacao{Campo: "codigo", Mensagem: "O código não pode ter quebra de linha."}
	}
	return CodigoIndicador{valor: normalizado}, nil
}

func (c CodigoIndicador) String() string { return c.valor }
