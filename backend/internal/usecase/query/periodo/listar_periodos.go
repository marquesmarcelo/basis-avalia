package periodo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarPeriodosInput struct {
	Ator           autorizacao.Ator
	Nome, Situacao string
	Page, PageSize int
	Sort, Order    string
}

type ListarPeriodosUseCase struct{ repo port.PeriodoRepository }

func NovoListarPeriodosUseCase(repo port.PeriodoRepository) *ListarPeriodosUseCase {
	return &ListarPeriodosUseCase{repo: repo}
}

func (uc *ListarPeriodosUseCase) Executar(ctx context.Context, in ListarPeriodosInput) (port.ResultadoListaPeriodos, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaPeriodos{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarPeriodos{
		Nome: in.Nome, Situacao: in.Situacao, Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
