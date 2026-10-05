package periodo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarPeriodoUseCase struct{ repo port.PeriodoRepository }

func NovoBuscarPeriodoUseCase(repo port.PeriodoRepository) *BuscarPeriodoUseCase {
	return &BuscarPeriodoUseCase{repo: repo}
}

func (uc *BuscarPeriodoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemPeriodo, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemPeriodo{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
