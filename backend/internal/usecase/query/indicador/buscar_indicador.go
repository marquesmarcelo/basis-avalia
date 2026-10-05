package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarIndicadorUseCase struct {
	repo port.IndicadorRepository
}

func NovoBuscarIndicadorUseCase(repo port.IndicadorRepository) *BuscarIndicadorUseCase {
	return &BuscarIndicadorUseCase{repo: repo}
}

func (uc *BuscarIndicadorUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemIndicador, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemIndicador{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
