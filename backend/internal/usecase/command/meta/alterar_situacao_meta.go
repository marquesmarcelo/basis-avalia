package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AlterarSituacaoMetaInput struct {
	Ator         autorizacao.Ator
	MetaID       uuid.UUID
	NovaSituacao valueobject.SituacaoCatalogo
	Versao       int
}

// AlterarSituacaoMetaUseCase cobre MC-10: inativar não afeta plano
// vigente, só some do autocomplete de novo item.
type AlterarSituacaoMetaUseCase struct {
	repo  port.MetaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAlterarSituacaoMetaUseCase(repo port.MetaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AlterarSituacaoMetaUseCase {
	return &AlterarSituacaoMetaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AlterarSituacaoMetaUseCase) Executar(ctx context.Context, in AlterarSituacaoMetaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoInativar, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.AlterarSituacao(ctx, esc, in.MetaID, in.NovaSituacao, in.Versao); err != nil {
			return err
		}
		acao := auditoria.AlterarSituacaoMeta
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Meta"
		evento.RecursoID = &in.MetaID
		evento.Detalhes["situacao_depois"] = string(in.NovaSituacao)
		return uc.audit.Registrar(ctx, evento)
	})
}
