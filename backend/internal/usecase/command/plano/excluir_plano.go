package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
}

// ExcluirPlanoUseCase cobre SI-14: só rascunho e sem entrega.
type ExcluirPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirPlanoUseCase {
	return &ExcluirPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirPlanoUseCase) Executar(ctx context.Context, in ExcluirPlanoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		if detalhe.Plano.SituacaoPublicacao != valueobject.PlanoRascunho {
			return domain.ErrPlanoVigenteNaoExcluivel
		}
		temEntrega, err := uc.repo.ExisteEntregaNoPlano(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		if temEntrega {
			return domain.ErrPlanoComEntrega
		}
		if err := uc.repo.Excluir(ctx, esc, in.PlanoID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &in.PlanoID
		return uc.audit.Registrar(ctx, evento)
	})
}
