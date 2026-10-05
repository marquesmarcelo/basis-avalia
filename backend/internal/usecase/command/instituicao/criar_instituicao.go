package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

type CriarInstituicaoInput struct {
	Ator       autorizacao.Ator
	Nome       string
	Sigla      valueobject.Sigla
	CodigoEMec valueobject.CodigoEMec
}

// CriarInstituicaoUseCase cobre I-01 a I-05 (design.md §4.1).
type CriarInstituicaoUseCase struct {
	repo  port.InstituicaoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarInstituicaoUseCase(repo port.InstituicaoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarInstituicaoUseCase {
	return &CriarInstituicaoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarInstituicaoUseCase) Executar(ctx context.Context, in CriarInstituicaoInput) (instituicao.Instituicao, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoCriar, nil)
	if err != nil {
		return instituicao.Instituicao{}, err
	}

	nova, err := instituicao.NovaInstituicao(in.Nome, in.Sigla, in.CodigoEMec)
	if err != nil {
		return instituicao.Instituicao{}, err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, nova); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarInstituicao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Instituicao"
		evento.RecursoID = &nova.ID
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return instituicao.Instituicao{}, erro
	}
	return *nova, nil
}
