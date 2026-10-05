package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirCursoInput struct {
	Ator    autorizacao.Ator
	CursoID uuid.UUID
}

// ExcluirCursoUseCase cobre CU-08 — 409 CURSO_COM_VINCULO decidido pelo
// repositório, sob SELECT ... FOR UPDATE (design.md §5.4).
type ExcluirCursoUseCase struct {
	repo  port.CursoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirCursoUseCase(repo port.CursoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirCursoUseCase {
	return &ExcluirCursoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirCursoUseCase) Executar(ctx context.Context, in ExcluirCursoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Excluir(ctx, esc, in.CursoID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirCurso, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Curso"
		evento.RecursoID = &in.CursoID
		return uc.audit.Registrar(ctx, evento)
	})
}
