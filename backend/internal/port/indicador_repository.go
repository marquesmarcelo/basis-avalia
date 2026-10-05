package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type FiltroListarIndicadores struct {
	Busca    string
	Origem   string // "todos" | "plataforma" | "instituicao"
	Situacao string // "ativo" | "inativo" | "todos"
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ItemIndicador inclui a contagem de uso DA INSTITUIÇÃO da sessão (IV-03)
// — nunca o total agregado, que é exclusivo da família de plataforma.
type ItemIndicador struct {
	Indicador          indicador.Indicador
	MetasDaInstituicao int
}

type ResultadoListaIndicadores struct {
	Itens []ItemIndicador
	Total int
}

// ItemSugestaoIndicador é o formato do autocomplete (design.md §6, ux.md):
// o suficiente para montar { value, label, badge, secundario } no cliente.
type ItemSugestaoIndicador struct {
	ID                    uuid.UUID
	Codigo                string
	Nome                  string
	Escopo                string
	ReferenciaInstrumento string
}

// IndicadorRepository serve /api/v1/indicadores (PI, Coordenador) — lê os
// dois escopos (catálogo comum + próprio), escreve só no próprio
// (design.md §6, I-07).
type IndicadorRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemIndicador, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarIndicadores) (ResultadoListaIndicadores, error)
	Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]ItemSugestaoIndicador, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error
	AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error
	// ExcluirSeSemUso — IN-05: bloqueia com a contagem DA INSTITUIÇÃO.
	ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
}
