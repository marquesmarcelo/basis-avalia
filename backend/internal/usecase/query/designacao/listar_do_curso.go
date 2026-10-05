package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarDoCursoInput struct {
	Ator          autorizacao.Ator
	CursoID       uuid.UUID
	Situacao      string
	CoordenadorID *uuid.UUID
	Page          int
	PageSize      int
	Sort          string
	Order         string
}

type ListarDoCursoUseCase struct {
	repo port.DesignacaoRepository
}

func NovoListarDoCursoUseCase(repo port.DesignacaoRepository) *ListarDoCursoUseCase {
	return &ListarDoCursoUseCase{repo: repo}
}

func (uc *ListarDoCursoUseCase) Executar(ctx context.Context, in ListarDoCursoInput) (port.ResultadoListaDesignacoes, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaDesignacoes{}, err
	}
	return uc.repo.ListarDoCurso(ctx, esc, in.CursoID, port.FiltroListarDesignacoes{
		Situacao: in.Situacao, CoordenadorID: in.CoordenadorID,
		Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
