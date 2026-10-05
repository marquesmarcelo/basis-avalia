package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirIndicadorInput struct {
	Ator        autorizacao.Ator
	IndicadorID uuid.UUID
}

// ExcluirIndicadorUseCase cobre IN-05 — mesma assimetria IE-04.
type ExcluirIndicadorUseCase struct {
	repo  port.IndicadorRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirIndicadorUseCase(repo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirIndicadorUseCase {
	return &ExcluirIndicadorUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirIndicadorUseCase) Executar(ctx context.Context, in ExcluirIndicadorInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoExcluir, nil)
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
		if err := uc.repo.ExcluirSeSemUso(ctx, esc, in.IndicadorID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirIndicador, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &in.IndicadorID
		return uc.audit.Registrar(ctx, evento)
	})
}
