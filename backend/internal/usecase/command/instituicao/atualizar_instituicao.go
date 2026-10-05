package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarInstituicaoInput struct {
	Ator          autorizacao.Ator
	InstituicaoID uuid.UUID
	Nome          string
	Sigla         valueobject.Sigla
	CodigoEMec    valueobject.CodigoEMec
	Versao        int
}

// AtualizarInstituicaoUseCase cobre I-12.
type AtualizarInstituicaoUseCase struct {
	repo  port.InstituicaoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarInstituicaoUseCase(repo port.InstituicaoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarInstituicaoUseCase {
	return &AtualizarInstituicaoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarInstituicaoUseCase) Executar(ctx context.Context, in AtualizarInstituicaoInput) (instituicao.Instituicao, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoEditar, nil)
	if err != nil {
		return instituicao.Instituicao{}, err
	}

	var resultado instituicao.Instituicao
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.InstituicaoID)
		if err != nil {
			return err
		}
		atual := item.Instituicao
		if err := atual.Renomear(in.Nome, in.Sigla, in.CodigoEMec); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarInstituicao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Instituicao"
		evento.RecursoID = &atual.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return instituicao.Instituicao{}, erro
	}
	return resultado, nil
}
