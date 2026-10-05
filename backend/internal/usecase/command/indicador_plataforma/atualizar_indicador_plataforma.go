package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarIndicadorPlataformaInput struct {
	Ator                  autorizacao.Ator
	IndicadorID           uuid.UUID
	Codigo                string
	Nome                  string
	Descricao             string
	ReferenciaInstrumento string
	Versao                int
}

// AtualizarIndicadorPlataformaUseCase cobre IE-05: correção vale para
// todas as instituições imediatamente, auditada com antes/depois.
type AtualizarIndicadorPlataformaUseCase struct {
	repo  port.IndicadorPlataformaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarIndicadorPlataformaUseCase(repo port.IndicadorPlataformaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarIndicadorPlataformaUseCase {
	return &AtualizarIndicadorPlataformaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarIndicadorPlataformaUseCase) Executar(ctx context.Context, in AtualizarIndicadorPlataformaInput) (indicador.Indicador, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoEditar, nil)
	if err != nil {
		return indicador.Indicador{}, err
	}

	var resultado indicador.Indicador
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.IndicadorID)
		if err != nil {
			return err
		}
		atual := item.Indicador
		codigoAntes := atual.Codigo.String()
		referenciaAntes := atual.ReferenciaInstrumento.String()

		if err := atual.AtualizarDaPlataforma(in.Codigo, in.Nome, in.Descricao, in.ReferenciaInstrumento); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}

		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarIndicadorInep, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &atual.ID
		evento.Detalhes["codigo_antes"] = codigoAntes
		evento.Detalhes["codigo_depois"] = atual.Codigo.String()
		evento.Detalhes["referencia_instrumento_antes"] = referenciaAntes
		evento.Detalhes["referencia_instrumento_depois"] = atual.ReferenciaInstrumento.String()
		evento.Detalhes["metas_total"] = item.MetasTotal
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return indicador.Indicador{}, erro
	}
	return resultado, nil
}
