package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarPlanosInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroListarPlanos
}

type ListarPlanosUseCase struct{ repo port.PlanoRepository }

func NovoListarPlanosUseCase(repo port.PlanoRepository) *ListarPlanosUseCase {
	return &ListarPlanosUseCase{repo: repo}
}

func (uc *ListarPlanosUseCase) Executar(ctx context.Context, in ListarPlanosInput) (port.ResultadoListaPlanos, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaPlanos{}, err
	}
	return uc.repo.Listar(ctx, esc, in.Filtro)
}
