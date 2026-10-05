package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AlterarSituacaoIndicadorInput struct {
	Ator         autorizacao.Ator
	IndicadorID  uuid.UUID
	NovaSituacao valueobject.SituacaoCatalogo
	Versao       int
}

// AlterarSituacaoIndicadorUseCase cobre IN-04 — mesma assimetria IE-04.
type AlterarSituacaoIndicadorUseCase struct {
	repo  port.IndicadorRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAlterarSituacaoIndicadorUseCase(repo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AlterarSituacaoIndicadorUseCase {
	return &AlterarSituacaoIndicadorUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AlterarSituacaoIndicadorUseCase) Executar(ctx context.Context, in AlterarSituacaoIndicadorInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoInativar, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.IndicadorID)
		if err != nil {
			return err
		}
		if !item.Indicador.Escopo.PertenceAInstituicao() {
			return domain.ErrPermissaoNegada // IE-04
		}
		if err := uc.repo.AlterarSituacao(ctx, esc, in.IndicadorID, in.NovaSituacao, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AlterarSituacaoIndicador, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &in.IndicadorID
		return uc.audit.Registrar(ctx, evento)
	})
}
