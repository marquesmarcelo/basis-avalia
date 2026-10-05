package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// SugerirIndicadorUseCase alimenta o autocomplete — sem criação inline
// (spec.md §9).
type SugerirIndicadorUseCase struct {
	repo port.IndicadorRepository
}

func NovoSugerirIndicadorUseCase(repo port.IndicadorRepository) *SugerirIndicadorUseCase {
	return &SugerirIndicadorUseCase{repo: repo}
}

func (uc *SugerirIndicadorUseCase) Executar(ctx context.Context, ator autorizacao.Ator, busca string) ([]port.ItemSugestaoIndicador, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.Sugerir(ctx, esc, busca)
}
