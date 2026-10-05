package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// EscopoIndicador — plataforma (catálogo do INEP, comum à instalação) ou
// instituicao (indicador próprio). Imutável após a criação da entidade
// (specs/indicadores/design.md §3.1).
type EscopoIndicador string

const (
	EscopoIndicadorPlataforma  EscopoIndicador = "plataforma"
	EscopoIndicadorInstituicao EscopoIndicador = "instituicao"
)

func NovoEscopoIndicador(bruto string) (EscopoIndicador, error) {
	e := EscopoIndicador(bruto)
	switch e {
	case EscopoIndicadorPlataforma, EscopoIndicadorInstituicao:
		return e, nil
	default:
		return "", domain.ErrValorInvalido
	}
}

// PertenceAInstituicao é o predicado que amarra o par (escopo, instituição)
// num lugar só (design.md §3.1).
func (e EscopoIndicador) PertenceAInstituicao() bool {
	return e == EscopoIndicadorInstituicao
}
