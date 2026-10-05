package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarInstituicaoUseCase struct {
	repo port.InstituicaoRepository
}

func NovoBuscarInstituicaoUseCase(repo port.InstituicaoRepository) *BuscarInstituicaoUseCase {
	return &BuscarInstituicaoUseCase{repo: repo}
}

func (uc *BuscarInstituicaoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemInstituicao, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemInstituicao{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
