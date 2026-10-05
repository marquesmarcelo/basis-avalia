package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type FiltroListarMetas struct {
	Busca       string
	IndicadorID *uuid.UUID
	Origem      string // "todas" | "plataforma" | "instituicao"
	Situacao    string
	Page        int
	PageSize    int
	Sort        string
	Order       string
}

// IndicadorDaMeta é o indicador embutido na resposta de meta — código,
// nome, origem e referência, o suficiente para a coluna "Indicadores" do
// grid sem segunda chamada por linha (ux.md).
type IndicadorDaMeta struct {
	ID                    uuid.UUID
	Codigo                string
	Nome                  string
	Escopo                string
	ReferenciaInstrumento string
	Situacao              string
}

// ItemMeta.Planos é sempre 0 até specs/plano-acao criar item_plano
// — não há tabela para contar ainda. Ver testes-pendentes.md.
type ItemMeta struct {
	Meta        meta.Meta
	Indicadores []IndicadorDaMeta
	Planos      int
}

type ResultadoListaMetas struct {
	Itens []ItemMeta
	Total int
}

type ItemSugestaoMeta struct {
	ID          uuid.UUID
	Nome        string
	Indicadores []IndicadorDaMeta
	// QuantidadeSugerida — o item-do-plano-form lê isto para pré-preencher
	// a quantidade quando o PI escolhe esta meta (indicadores/design.md).
	QuantidadeSugerida *int
}

// MetaRepository serve /api/v1/metas (PI; Coordenador lê como contexto,
// embutido em outras rotas de specs/metas-coordenacao).
type MetaRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemMeta, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarMetas) (ResultadoListaMetas, error)
	Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]ItemSugestaoMeta, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, m *meta.Meta) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, m *meta.Meta, versaoEsperada int) error
	AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error
	// Excluir — MC-11 (bloqueio por uso em item_plano) fica pendente até
	// specs/plano-acao criar essa tabela; hoje não há nada para
	// bloquear contra. Ver testes-pendentes.md.
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
}
