package periodo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarPeriodoInput struct {
	Ator                autorizacao.Ator
	PeriodoID           uuid.UUID
	Nome                string
	DataInicio, DataFim string
	Versao              int
}

// AtualizarPeriodoUseCase cobre PE-06 (prorrogar reabre o período).
type AtualizarPeriodoUseCase struct {
	repo  port.PeriodoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarPeriodoUseCase(repo port.PeriodoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarPeriodoUseCase {
	return &AtualizarPeriodoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarPeriodoUseCase) Executar(ctx context.Context, in AtualizarPeriodoInput) (periodo.Periodo, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoEditar, nil)
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

	var resultado periodo.Periodo
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.PeriodoID)
		if err != nil {
			return err
		}
		atual := item.Periodo
		fimAnterior := atual.Vigencia.Fim()
		if err := atual.Atualizar(in.Nome, inicio, fim); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarPeriodo, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Periodo"
		evento.RecursoID = &atual.ID
		// A data de fim entra antes/depois — ela muda retroativamente o
		// resultado de uma apuração (fundacao-metas.md §8, exceção da
		// regra "só quais campos mudaram").
		if fimAnterior != nil {
			evento.Detalhes["data_fim_antes"] = fimAnterior.String()
		}
		evento.Detalhes["data_fim_depois"] = fim.String()
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return periodo.Periodo{}, erro
	}
	return resultado, nil
}
