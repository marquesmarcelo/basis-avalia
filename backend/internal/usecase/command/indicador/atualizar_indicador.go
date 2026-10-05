package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarIndicadorInput struct {
	Ator        autorizacao.Ator
	IndicadorID uuid.UUID
	Codigo      string
	Nome        string
	Descricao   string
	Versao      int
}

// AtualizarIndicadorUseCase cobre IN-01. A assimetria de IE-04 mora aqui:
// o PI ENXERGA o indicador do INEP (a exceção do catálogo o traz na
// consulta) e é recusado ao ESCREVER, com 403 — não 404, porque o recurso
// é legitimamente visível para ele.
type AtualizarIndicadorUseCase struct {
	repo  port.IndicadorRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarIndicadorUseCase(repo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarIndicadorUseCase {
	return &AtualizarIndicadorUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarIndicadorUseCase) Executar(ctx context.Context, in AtualizarIndicadorInput) (indicador.Indicador, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoEditar, nil)
	if err != nil {
		return indicador.Indicador{}, err
	}

	var resultado indicador.Indicador
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.IndicadorID)
		if err != nil {
			return err
		}
		if !item.Indicador.Escopo.PertenceAInstituicao() {
			return domain.ErrPermissaoNegada // IE-04
		}
		atual := item.Indicador
		if err := atual.AtualizarDaInstituicao(in.Codigo, in.Nome, in.Descricao); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarIndicador, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &atual.ID
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
