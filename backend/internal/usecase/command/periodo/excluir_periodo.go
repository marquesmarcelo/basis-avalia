package periodo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirPeriodoInput struct {
	Ator      autorizacao.Ator
	PeriodoID uuid.UUID
}

// ExcluirPeriodoUseCase cobre PE-05: período com plano não é excluído.
type ExcluirPeriodoUseCase struct {
	repo  port.PeriodoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirPeriodoUseCase(repo port.PeriodoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirPeriodoUseCase {
	return &ExcluirPeriodoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirPeriodoUseCase) Executar(ctx context.Context, in ExcluirPeriodoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		temPlano, err := uc.repo.TemPlano(ctx, esc, in.PeriodoID)
		if err != nil {
			return err
		}
		if temPlano {
			return domain.ErrPeriodoComPlano
		}
		if err := uc.repo.Excluir(ctx, esc, in.PeriodoID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirPeriodo, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Periodo"
		evento.RecursoID = &in.PeriodoID
		return uc.audit.Registrar(ctx, evento)
	})
}
