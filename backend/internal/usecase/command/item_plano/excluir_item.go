package itemplano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirItemInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	ItemID  uuid.UUID
}

// ExcluirItemUseCase cobre IT-10: item com entrega não é removido — a
// saída é reduzir a quantidade.
type ExcluirItemUseCase struct {
	planoRepo port.PlanoRepository
	itemRepo  port.ItemPlanoRepository
	audit     port.AuditLogger
	uow       port.UnidadeDeTrabalho
}

func NovoExcluirItemUseCase(planoRepo port.PlanoRepository, itemRepo port.ItemPlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirItemUseCase {
	return &ExcluirItemUseCase{planoRepo: planoRepo, itemRepo: itemRepo, audit: audit, uow: uow}
}

func (uc *ExcluirItemUseCase) Executar(ctx context.Context, in ExcluirItemInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		temEntrega, err := uc.planoRepo.ExisteEntregaNoItem(ctx, esc, in.ItemID)
		if err != nil {
			return err
		}
		if temEntrega {
			return domain.ErrItemComEntrega
		}
		if err := uc.itemRepo.Excluir(ctx, esc, in.ItemID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirItemPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "ItemDoPlano"
		evento.RecursoID = &in.ItemID
		return uc.audit.Registrar(ctx, evento)
	})
}
