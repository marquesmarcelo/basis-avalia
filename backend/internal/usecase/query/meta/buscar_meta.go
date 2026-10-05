package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarMetaUseCase struct {
	repo port.MetaRepository
}

func NovoBuscarMetaUseCase(repo port.MetaRepository) *BuscarMetaUseCase {
	return &BuscarMetaUseCase{repo: repo}
}

func (uc *BuscarMetaUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemMeta, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemMeta{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
