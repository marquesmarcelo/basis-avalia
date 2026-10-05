package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarCursoInput struct {
	Ator       autorizacao.Ator
	CursoID    uuid.UUID
	Nome       string
	CodigoEMec string
	Grau       string
	Modalidade string
	Versao     int
}

type AtualizarCursoUseCase struct {
	repo  port.CursoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarCursoUseCase(repo port.CursoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarCursoUseCase {
	return &AtualizarCursoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarCursoUseCase) Executar(ctx context.Context, in AtualizarCursoInput) (curso.Curso, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return curso.Curso{}, err
	}

	var resultado curso.Curso
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.CursoID)
		if err != nil {
			return err
		}
		atual := item.Curso
		if err := atual.Atualizar(in.Nome, in.CodigoEMec, in.Grau, in.Modalidade); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarCurso, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Curso"
		evento.RecursoID = &atual.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return curso.Curso{}, erro
	}
	return resultado, nil
}
