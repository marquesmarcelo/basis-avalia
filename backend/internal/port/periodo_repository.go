package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/google/uuid"
)

type FiltroListarPeriodos struct {
	Nome     string
	Situacao string // "todas" | "nao_iniciado" | "aberto" | "encerrado"
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ItemPeriodo — período mais o número de planos vinculados (coluna
// "Planos" do grid, ux.md).
type ItemPeriodo struct {
	Periodo periodo.Periodo
	Planos  int
}

type ResultadoListaPeriodos struct {
	Itens []ItemPeriodo
	Total int
}

// PeriodoRepository serve /api/v1/periodos (specs/plano-acao/design.md §5).
type PeriodoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemPeriodo, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarPeriodos) (ResultadoListaPeriodos, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo, versaoEsperada int) error
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
	// TemPlano sustenta PE-05 (409 PERIODO_COM_PLANO).
	TemPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (bool, error)
}
