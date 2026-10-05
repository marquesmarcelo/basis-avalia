package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarMetaInput struct {
	Ator               autorizacao.Ator
	MetaID             uuid.UUID
	Nome               string
	Descricao          string
	Indicadores        []uuid.UUID
	QuantidadeSugerida *int
	Versao             int
}

// AtualizarMetaUseCase cobre MC-13: troca de indicadores vale dali em
// diante; a auditoria registra a lista anterior e a nova, completas.
type AtualizarMetaUseCase struct {
	repo          port.MetaRepository
	indicadorRepo port.IndicadorRepository
	audit         port.AuditLogger
	uow           port.UnidadeDeTrabalho
}

func NovoAtualizarMetaUseCase(repo port.MetaRepository, indicadorRepo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarMetaUseCase {
	return &AtualizarMetaUseCase{repo: repo, indicadorRepo: indicadorRepo, audit: audit, uow: uow}
}

func (uc *AtualizarMetaUseCase) Executar(ctx context.Context, in AtualizarMetaInput) (meta.Meta, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return meta.Meta{}, err
	}

	var resultado meta.Meta
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.MetaID)
		if err != nil {
			return err
		}
		atual := item.Meta
		if err := atual.Atualizar(in.Nome, in.Descricao); err != nil {
			return err
		}
		if err := atual.DefinirQuantidadeSugerida(in.QuantidadeSugerida); err != nil {
			return err
		}
		anteriores, err := atual.DefinirIndicadores(in.Indicadores)
		if err != nil {
			return err
		}
		if err := validarIndicadoresAtivos(ctx, uc.indicadorRepo, esc, atual.Indicadores); err != nil {
			return err
		}
		if err := uc.repo.Atualizar(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarMeta, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Meta"
		evento.RecursoID = &atual.ID
		evento.Detalhes["indicadores_antes"] = uuidsParaString(anteriores)
		evento.Detalhes["indicadores_depois"] = uuidsParaString(atual.Indicadores)
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return meta.Meta{}, erro
	}
	return resultado, nil
}

func uuidsParaString(ids []uuid.UUID) []string {
	saida := make([]string, len(ids))
	for i, id := range ids {
		saida[i] = id.String()
	}
	return saida
}
