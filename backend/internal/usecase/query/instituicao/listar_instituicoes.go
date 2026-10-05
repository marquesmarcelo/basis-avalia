package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarInstituicoesInput struct {
	Ator     autorizacao.Ator
	Busca    string
	Situacao string
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ListarInstituicoesUseCase cobre I-08, I-10, I-11.
type ListarInstituicoesUseCase struct {
	repo port.InstituicaoRepository
}

func NovoListarInstituicoesUseCase(repo port.InstituicaoRepository) *ListarInstituicoesUseCase {
	return &ListarInstituicoesUseCase{repo: repo}
}

func (uc *ListarInstituicoesUseCase) Executar(ctx context.Context, in ListarInstituicoesInput) (port.ResultadoListaInstituicoes, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaInstituicoes{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarInstituicoes{
		Busca: in.Busca, Situacao: in.Situacao, Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
