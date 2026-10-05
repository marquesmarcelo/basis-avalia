package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type FiltroListarInstituicoes struct {
	Busca    string
	Situacao string // "ativa" | "inativa" | "todas"
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// ItemInstituicao inclui a contagem de PIs ativos — computada na consulta,
// nunca persistida (design.md §5.6, evita dual write).
type ItemInstituicao struct {
	Instituicao          instituicao.Instituicao
	PesquisadoresAtivos  int
}

type ResultadoListaInstituicoes struct {
	Itens []ItemInstituicao
	Total int
}

// InstituicaoRepository exige Escopo.Plataforma() — é chamado apenas pelo
// Administrador do Sistema (design.md §4.2).
type InstituicaoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemInstituicao, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarInstituicoes) (ResultadoListaInstituicoes, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, i *instituicao.Instituicao) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *instituicao.Instituicao, versaoEsperada int) error
	AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoInstituicao, versaoEsperada int) error
}
