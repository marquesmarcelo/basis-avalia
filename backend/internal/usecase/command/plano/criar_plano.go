package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type CriarPlanoInput struct {
	Ator      autorizacao.Ator
	CursoID   uuid.UUID
	PeriodoID uuid.UUID
	Dados     plano.DadosDoPlano
}

// CriarPlanoUseCase cobre PL-01, PL-02, PL-03, PL-05 (specs/plano-acao/spec.md).
type CriarPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarPlanoUseCase {
	return &CriarPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarPlanoUseCase) Executar(ctx context.Context, in CriarPlanoInput) (plano.Plano, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return plano.Plano{}, err
	}

	var resultado plano.Plano
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		ativo, err := uc.repo.CursoValidoParaPlano(ctx, esc, in.CursoID)
		if err != nil {
			return err
		}
		if !ativo {
			return domain.ErrCursoInativo
		}
		novo, err := plano.NovoPlano(*esc.InstituicaoID(), in.CursoID, in.PeriodoID, in.Dados)
		if err != nil {
			return err
		}
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &novo.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *novo
		return nil
	})
	if erro != nil {
		return plano.Plano{}, erro
	}
	return resultado, nil
}
