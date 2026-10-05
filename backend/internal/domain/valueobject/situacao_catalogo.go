package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// SituacaoCatalogo — ativo ou inativo. Reaproveitada por Indicador e Meta
// (specs/indicadores/design.md §3.1).
type SituacaoCatalogo string

const (
	CatalogoAtivo   SituacaoCatalogo = "ativo"
	CatalogoInativo SituacaoCatalogo = "inativo"
)

func NovaSituacaoCatalogo(bruta string) (SituacaoCatalogo, error) {
	s := SituacaoCatalogo(bruta)
	switch s {
	case CatalogoAtivo, CatalogoInativo:
		return s, nil
	default:
		return "", domain.ErrValorInvalido
	}
}

func (s SituacaoCatalogo) Alternar() SituacaoCatalogo {
	if s == CatalogoAtivo {
		return CatalogoInativo
	}
	return CatalogoAtivo
}
