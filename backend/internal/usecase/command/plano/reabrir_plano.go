package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ReabrirPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	Versao  int
}

// ReabrirPlanoUseCase cobre SI-13: limpa o encerramento antecipado; se o
// período já encerrou, a situação efetiva permanece Encerrada — não é
// erro, a resposta 200 é a correta (Plano.Reabrir não recusa nada).
type ReabrirPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoReabrirPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ReabrirPlanoUseCase {
	return &ReabrirPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ReabrirPlanoUseCase) Executar(ctx context.Context, in ReabrirPlanoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		atual := detalhe.Plano
		atual.Reabrir()
		if err := uc.repo.AtualizarSituacao(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.ReabrirPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &atual.ID
		return uc.audit.Registrar(ctx, evento)
	})
}
