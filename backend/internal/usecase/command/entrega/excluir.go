package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirEntregaInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
}

// ExcluirEntregaUseCase cobre EN-10, EN-11: só quem enviou, só se não
// aceita, exclusão lógica (design.md §3.2, Entrega.PodeExcluir).
type ExcluirEntregaUseCase struct {
	repo  port.EntregaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirEntregaUseCase(repo port.EntregaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirEntregaUseCase {
	return &ExcluirEntregaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirEntregaUseCase) Executar(ctx context.Context, in ExcluirEntregaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoExcluir, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
		if err != nil {
			return err
		}
		atual := entregaDeDetalhe(detalhe)
		autorID := in.Ator.UsuarioID()
		if err := atual.PodeExcluir(autorID); err != nil {
			return err
		}
		if err := uc.repo.ExcluirLogicamente(ctx, esc, in.EntregaID); err != nil {
			return err
		}
		evento := auditoria.NovoEvento(auditoria.ExcluirEntrega, auditoria.ResultadoSucesso)
		evento.AtorID = &autorID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &in.EntregaID
		return uc.audit.Registrar(ctx, evento)
	})
}
