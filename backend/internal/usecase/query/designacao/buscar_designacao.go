package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarDesignacaoUseCase struct {
	repo port.DesignacaoRepository
}

func NovoBuscarDesignacaoUseCase(repo port.DesignacaoRepository) *BuscarDesignacaoUseCase {
	return &BuscarDesignacaoUseCase{repo: repo}
}

func (uc *BuscarDesignacaoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemDesignacao, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemDesignacao{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
