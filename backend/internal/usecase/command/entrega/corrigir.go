package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type CorrigirEntregaInput struct {
	Ator       autorizacao.Ator
	EntregaID  uuid.UUID
	Observacao string
	Versao     int
}

// CorrigirEntregaUseCase cobre AV-04, AV-05, AV-10: o ramo passa por cima
// do período encerrado (X6) e usa Entrega.PodeCorrigir, a única função que
// decide se a correção ainda é possível (design.md §5.1). Relógio
// injetado — testável sem depender da hora real (design.md §12.4).
type CorrigirEntregaUseCase struct {
	repo    port.EntregaRepository
	audit   port.AuditLogger
	uow     port.UnidadeDeTrabalho
	relogio port.Relogio
}

func NovoCorrigirEntregaUseCase(repo port.EntregaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho, relogio port.Relogio) *CorrigirEntregaUseCase {
	return &CorrigirEntregaUseCase{repo: repo, audit: audit, uow: uow, relogio: relogio}
}

func (uc *CorrigirEntregaUseCase) Executar(ctx context.Context, in CorrigirEntregaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoEditar, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
		if err != nil {
			return err
		}
		atual := entregaDeDetalhe(detalhe)
		if err := atual.PodeCorrigir(uc.relogio.Agora()); err != nil {
			return err
		}
		autorID := in.Ator.UsuarioID()
		if err := atual.Corrigir(in.Observacao, autorID); err != nil {
			return err
		}
		if err := uc.repo.AtualizarCorrecao(ctx, esc, atual, in.Versao); err != nil {
			return err
		}
		evento := auditoria.NovoEvento(auditoria.CorrigirEntrega, auditoria.ResultadoSucesso)
		evento.AtorID = &autorID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &in.EntregaID
		return uc.audit.Registrar(ctx, evento)
	})
}
