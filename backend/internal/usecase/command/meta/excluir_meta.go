package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirMetaInput struct {
	Ator   autorizacao.Ator
	MetaID uuid.UUID
}

// ExcluirMetaUseCase — MC-11 (bloqueio por uso em item_plano) fica
// pendente até specs/plano-acao existir: ver
// port/meta_repository.go e testes-pendentes.md.
type ExcluirMetaUseCase struct {
	repo  port.MetaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirMetaUseCase(repo port.MetaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirMetaUseCase {
	return &ExcluirMetaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirMetaUseCase) Executar(ctx context.Context, in ExcluirMetaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Excluir(ctx, esc, in.MetaID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirMeta, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Meta"
		evento.RecursoID = &in.MetaID
		return uc.audit.Registrar(ctx, evento)
	})
}
