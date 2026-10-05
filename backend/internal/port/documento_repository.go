package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/documento"
	"github.com/google/uuid"
)

// DocumentoRepository grava o registro de cada geração do .docx
// (specs/plano-acao/design.md §7.4/§7.5). Gerar documento é sempre um
// INSERT aqui, nunca um UPDATE em plano (P-11).
type DocumentoRepository interface {
	Inserir(ctx context.Context, escopo autorizacao.Escopo, d *documento.Documento) error
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (documento.Documento, error)
}
