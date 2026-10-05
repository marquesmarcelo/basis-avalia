package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AlterarSituacaoCursoInput struct {
	Ator         autorizacao.Ator
	CursoID      uuid.UUID
	NovaSituacao valueobject.SituacaoCurso
	Versao       int
}

// AlterarSituacaoCursoUseCase cobre CU-06: não toca em designação, plano,
// entrega nem perfil — a confirmação SUGERE encerrar a portaria, sem
// fazê-lo (a ação em si, se o PI quiser, é uma AtualizarDesignacao à
// parte).
type AlterarSituacaoCursoUseCase struct {
	repo  port.CursoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAlterarSituacaoCursoUseCase(repo port.CursoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AlterarSituacaoCursoUseCase {
	return &AlterarSituacaoCursoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AlterarSituacaoCursoUseCase) Executar(ctx context.Context, in AlterarSituacaoCursoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoInativar, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.AlterarSituacao(ctx, esc, in.CursoID, in.NovaSituacao, in.Versao); err != nil {
			return err
		}
		acao := auditoria.InativarCurso
		if in.NovaSituacao == valueobject.CursoAtivo {
			acao = auditoria.ReativarCurso
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Curso"
		evento.RecursoID = &in.CursoID
		return uc.audit.Registrar(ctx, evento)
	})
}
