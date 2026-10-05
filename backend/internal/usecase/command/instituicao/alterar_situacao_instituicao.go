package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AlterarSituacaoInstituicaoInput struct {
	Ator          autorizacao.Ator
	InstituicaoID uuid.UUID
	NovaSituacao  valueobject.SituacaoInstituicao
	Versao        int
}

// AlterarSituacaoInstituicaoUseCase cobre I-06, I-07. Não escreve nada em
// usuario (I-06): a sessão cai porque o middleware relê a situação da
// instituição a cada requisição (design.md §5.2, R16).
type AlterarSituacaoInstituicaoUseCase struct {
	repo  port.InstituicaoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAlterarSituacaoInstituicaoUseCase(repo port.InstituicaoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AlterarSituacaoInstituicaoUseCase {
	return &AlterarSituacaoInstituicaoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AlterarSituacaoInstituicaoUseCase) Executar(ctx context.Context, in AlterarSituacaoInstituicaoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoInativar, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.AlterarSituacao(ctx, esc, in.InstituicaoID, in.NovaSituacao, in.Versao); err != nil {
			return err
		}
		acao := auditoria.InativarInstituicao
		if in.NovaSituacao == valueobject.Ativa {
			acao = auditoria.ReativarInstituicao
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Instituicao"
		evento.RecursoID = &in.InstituicaoID
		return uc.audit.Registrar(ctx, evento)
	})
}
