package plano

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type EncerrarPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	Motivo  string
	Versao  int
}

// EncerrarPlanoUseCase cobre SI-12: motivo obrigatório, auditado.
type EncerrarPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoEncerrarPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *EncerrarPlanoUseCase {
	return &EncerrarPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *EncerrarPlanoUseCase) Executar(ctx context.Context, in EncerrarPlanoInput) error {
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
		if err := atual.Encerrar(in.Motivo, time.Now()); err != nil {
			return err
		}
		if err := uc.repo.AtualizarSituacao(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.EncerrarPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &atual.ID
		evento.Detalhes["motivo"] = in.Motivo
		return uc.audit.Registrar(ctx, evento)
	})
}
