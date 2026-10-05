package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type FiltroListarIndicadoresPlataforma struct {
	Busca    string
	Situacao string // "ativo" | "inativo" | "todos"
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ItemIndicadorPlataforma inclui a contagem TOTAL de metas que referenciam
// o indicador, somadas todas as instituições (PI-5) — nunca por
// instituição (specs/indicadores/design.md §7.2).
type ItemIndicadorPlataforma struct {
	Indicador  indicador.Indicador
	MetasTotal int
}

type ResultadoListaIndicadoresPlataforma struct {
	Itens []ItemIndicadorPlataforma
	Total int
}

// IndicadorPlataformaRepository serve /api/v1/plataforma/indicadores —
// exige Escopo.Plataforma() (design.md §6, família de propósito do
// Administrador do Sistema).
type IndicadorPlataformaRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemIndicadorPlataforma, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarIndicadoresPlataforma) (ResultadoListaIndicadoresPlataforma, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error
	AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error
	// ExcluirSeSemUso — IE-07: bloqueia dentro da transação com FOR UPDATE
	// na linha do indicador; devolve *domain.ErrIndicadorComMeta{Total} com
	// a contagem TOTAL (PI-5) quando há uso.
	ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
}
