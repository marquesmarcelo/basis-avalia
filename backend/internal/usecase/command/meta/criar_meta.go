package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type CriarMetaInput struct {
	Ator               autorizacao.Ator
	Nome               string
	Descricao          string
	Indicadores        []uuid.UUID
	QuantidadeSugerida *int
}

// CriarMetaUseCase cobre MC-01 a MC-07 — cada indicador precisa existir,
// estar visível à instituição (404 senão) e ativo (400 INDICADOR_INATIVO).
type CriarMetaUseCase struct {
	repo          port.MetaRepository
	indicadorRepo port.IndicadorRepository
	audit         port.AuditLogger
	uow           port.UnidadeDeTrabalho
}

func NovoCriarMetaUseCase(repo port.MetaRepository, indicadorRepo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarMetaUseCase {
	return &CriarMetaUseCase{repo: repo, indicadorRepo: indicadorRepo, audit: audit, uow: uow}
}

func (uc *CriarMetaUseCase) Executar(ctx context.Context, in CriarMetaInput) (meta.Meta, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return meta.Meta{}, err
	}

	nova, err := meta.NovaMeta(*esc.InstituicaoID(), in.Nome, in.Descricao, in.Indicadores)
	if err != nil {
		return meta.Meta{}, err
	}
	if err := nova.DefinirQuantidadeSugerida(in.QuantidadeSugerida); err != nil {
		return meta.Meta{}, err
	}

	var resultado meta.Meta
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := validarIndicadoresAtivos(ctx, uc.indicadorRepo, esc, nova.Indicadores); err != nil {
			return err
		}
		if err := uc.repo.Inserir(ctx, esc, nova); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarMeta, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Meta"
		evento.RecursoID = &nova.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *nova
		return nil
	})
	if erro != nil {
		return meta.Meta{}, erro
	}
	return resultado, nil
}

// validarIndicadoresAtivos confere, um a um, que cada indicador existe e é
// visível à instituição (404 por AplicarEscopo/BuscarPorID senão) e está
// ativo (MC-07, 400 INDICADOR_INATIVO).
func validarIndicadoresAtivos(ctx context.Context, repo port.IndicadorRepository, esc autorizacao.Escopo, ids []uuid.UUID) error {
	for _, id := range ids {
		item, err := repo.BuscarPorID(ctx, esc, id)
		if err != nil {
			return err
		}
		if item.Indicador.Situacao != valueobject.CatalogoAtivo {
			return domain.ErrIndicadorInativo
		}
	}
	return nil
}
