package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarDoCatalogoInput struct {
	Ator     autorizacao.Ator
	Busca    string
	Origem   string
	Situacao string
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ListarDoCatalogoUseCase cobre IN-06, IV-01, IV-02: um grid só, com os
// dois escopos misturados e a origem por linha.
type ListarDoCatalogoUseCase struct {
	repo port.IndicadorRepository
}

func NovoListarDoCatalogoUseCase(repo port.IndicadorRepository) *ListarDoCatalogoUseCase {
	return &ListarDoCatalogoUseCase{repo: repo}
}

func (uc *ListarDoCatalogoUseCase) Executar(ctx context.Context, in ListarDoCatalogoInput) (port.ResultadoListaIndicadores, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaIndicadores{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarIndicadores{
		Busca: in.Busca, Origem: in.Origem, Situacao: in.Situacao,
		Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
