package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirIndicadorPlataformaInput struct {
	Ator        autorizacao.Ator
	IndicadorID uuid.UUID
}

// ExcluirIndicadorPlataformaUseCase cobre IE-07: bloqueado enquanto houver
// uso, com a contagem TOTAL (PI-5) — nunca por instituição.
type ExcluirIndicadorPlataformaUseCase struct {
	repo  port.IndicadorPlataformaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirIndicadorPlataformaUseCase(repo port.IndicadorPlataformaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirIndicadorPlataformaUseCase {
	return &ExcluirIndicadorPlataformaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirIndicadorPlataformaUseCase) Executar(ctx context.Context, in ExcluirIndicadorPlataformaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.IndicadorID)
		if err != nil {
			return err
		}
		if err := uc.repo.ExcluirSeSemUso(ctx, esc, in.IndicadorID); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ExcluirIndicadorInep, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &in.IndicadorID
		evento.Detalhes["codigo"] = item.Indicador.Codigo.String()
		evento.Detalhes["metas_total"] = item.MetasTotal
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return erro
	}
	return nil
}
