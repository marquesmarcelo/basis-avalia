package periodo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

type CriarPeriodoInput struct {
	Ator                autorizacao.Ator
	Nome                string
	DataInicio, DataFim string // "AAAA-MM-DD"
}

// CriarPeriodoUseCase cobre PE-01, PE-02 (specs/plano-acao/spec.md).
type CriarPeriodoUseCase struct {
	repo  port.PeriodoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarPeriodoUseCase(repo port.PeriodoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarPeriodoUseCase {
	return &CriarPeriodoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarPeriodoUseCase) Executar(ctx context.Context, in CriarPeriodoInput) (periodo.Periodo, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return periodo.Periodo{}, err
	}
	inicio, err := valueobject.DataLocalTexto(in.DataInicio)
	if err != nil {
		return periodo.Periodo{}, err
	}
	fim, err := valueobject.DataLocalTexto(in.DataFim)
	if err != nil {
		return periodo.Periodo{}, err
	}
	novo, err := periodo.NovoPeriodo(*esc.InstituicaoID(), in.Nome, inicio, fim)
	if err != nil {
		return periodo.Periodo{}, err
	}

	var resultado periodo.Periodo
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarPeriodo, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Periodo"
		evento.RecursoID = &novo.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *novo
		return nil
	})
	if erro != nil {
		return periodo.Periodo{}, erro
	}
	return resultado, nil
}
