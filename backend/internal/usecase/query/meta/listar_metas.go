package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarMetasInput struct {
	Ator        autorizacao.Ator
	Busca       string
	IndicadorID *uuid.UUID
	Origem      string
	Situacao    string
	Page        int
	PageSize    int
	Sort        string
	Order       string
}

// ListarMetasUseCase cobre MC-14: filtrar por indicador/origem nunca
// duplica a meta.
type ListarMetasUseCase struct {
	repo port.MetaRepository
}

func NovoListarMetasUseCase(repo port.MetaRepository) *ListarMetasUseCase {
	return &ListarMetasUseCase{repo: repo}
}

func (uc *ListarMetasUseCase) Executar(ctx context.Context, in ListarMetasInput) (port.ResultadoListaMetas, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaMetas{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarMetas{
		Busca: in.Busca, IndicadorID: in.IndicadorID, Origem: in.Origem, Situacao: in.Situacao,
		Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
