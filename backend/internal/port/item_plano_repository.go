package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/itemplano"
	"github.com/google/uuid"
)

// ItemPlanoRepository serve os comandos de item do plano
// (specs/plano-acao/design.md §5.2).
type ItemPlanoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (itemplano.ItemDoPlano, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano) error
	AtualizarQuantidade(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano, versaoEsperada int) error
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
	// ExisteMetaNoPlano sustenta META_DUPLICADA_NO_PLANO (IT-05).
	ExisteMetaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID, metaID uuid.UUID) (bool, error)
}
