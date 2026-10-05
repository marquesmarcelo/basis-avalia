package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AlterarSituacaoIndicadorPlataformaInput struct {
	Ator         autorizacao.Ator
	IndicadorID  uuid.UUID
	NovaSituacao valueobject.SituacaoCatalogo
	Versao       int
}

// AlterarSituacaoIndicadorPlataformaUseCase cobre IE-06: inativar não
// afeta metas em uso, some do autocomplete de todas as instituições.
type AlterarSituacaoIndicadorPlataformaUseCase struct {
	repo  port.IndicadorPlataformaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAlterarSituacaoIndicadorPlataformaUseCase(repo port.IndicadorPlataformaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AlterarSituacaoIndicadorPlataformaUseCase {
	return &AlterarSituacaoIndicadorPlataformaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AlterarSituacaoIndicadorPlataformaUseCase) Executar(ctx context.Context, in AlterarSituacaoIndicadorPlataformaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoInativar, nil)
	if err != nil {
		return err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.IndicadorID)
		if err != nil {
			return err
		}
		situacaoAntes := item.Indicador.Situacao

		if err := uc.repo.AlterarSituacao(ctx, esc, in.IndicadorID, in.NovaSituacao, in.Versao); err != nil {
			return err
		}

		acao := auditoria.InativarIndicadorInep
		if in.NovaSituacao == valueobject.CatalogoAtivo {
			acao = auditoria.ReativarIndicadorInep
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &in.IndicadorID
		evento.Detalhes["situacao_antes"] = string(situacaoAntes)
		evento.Detalhes["situacao_depois"] = string(in.NovaSituacao)
		evento.Detalhes["metas_total"] = item.MetasTotal
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return erro
	}
	return nil
}
