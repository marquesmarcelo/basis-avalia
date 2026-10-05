package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarIndicadoresPlataformaInput struct {
	Ator     autorizacao.Ator
	Busca    string
	Situacao string
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ListarIndicadoresPlataformaUseCase struct {
	repo port.IndicadorPlataformaRepository
}

func NovoListarIndicadoresPlataformaUseCase(repo port.IndicadorPlataformaRepository) *ListarIndicadoresPlataformaUseCase {
	return &ListarIndicadoresPlataformaUseCase{repo: repo}
}

func (uc *ListarIndicadoresPlataformaUseCase) Executar(ctx context.Context, in ListarIndicadoresPlataformaInput) (port.ResultadoListaIndicadoresPlataforma, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaIndicadoresPlataforma{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarIndicadoresPlataforma{
		Busca: in.Busca, Situacao: in.Situacao, Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
