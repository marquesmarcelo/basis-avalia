package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// ListarPeriodosUseCase alimenta o seletor de período de "Minhas metas"
// (ux.md) — mesmo alcance de carteira que autoriza a própria tela.
type ListarPeriodosUseCase struct{ repo port.EntregaRepository }

func NovoListarPeriodosUseCase(repo port.EntregaRepository) *ListarPeriodosUseCase {
	return &ListarPeriodosUseCase{repo: repo}
}

func (uc *ListarPeriodosUseCase) Executar(ctx context.Context, ator autorizacao.Ator) ([]port.PeriodoOpcao, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.ListarPeriodosParaMinhasMetas(ctx, esc)
}
