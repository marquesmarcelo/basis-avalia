package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	designacaodomain "github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirDesignacaoInput struct {
	Ator         autorizacao.Ator
	DesignacaoID uuid.UUID
}

// ExcluirDesignacaoUseCase cobre DG-08: só designação futura pode ser
// excluída — o histórico de quem coordenou não se apaga.
type ExcluirDesignacaoUseCase struct {
	repo  port.DesignacaoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirDesignacaoUseCase(repo port.DesignacaoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirDesignacaoUseCase {
	return &ExcluirDesignacaoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirDesignacaoUseCase) Executar(ctx context.Context, in ExcluirDesignacaoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.DesignacaoID)
		if err != nil {
			return err
		}
		if !item.Designacao.PodeExcluir(esc.DataDeReferencia()) {
			return designacaodomain.ErrSeNaoForFutura(item.Designacao.SituacaoEm(esc.DataDeReferencia()))
		}
		if err := uc.repo.Excluir(ctx, esc, in.DesignacaoID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirDesignacao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Designacao"
		evento.RecursoID = &in.DesignacaoID
		return uc.audit.Registrar(ctx, evento)
	})
}
