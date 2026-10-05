package usuario

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ExcluirUsuarioInput struct {
	Ator                 autorizacao.Ator
	Alcance              autorizacao.Alcance
	InstituicaoDoCaminho *uuid.UUID
	UsuarioID            uuid.UUID
}

// ExcluirUsuarioUseCase cobre E-05, E-08, E-09, E-17, AS-06, T-04.
type ExcluirUsuarioUseCase struct {
	repo  port.UsuarioRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoExcluirUsuarioUseCase(repo port.UsuarioRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *ExcluirUsuarioUseCase {
	return &ExcluirUsuarioUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *ExcluirUsuarioUseCase) Executar(ctx context.Context, in ExcluirUsuarioInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoExcluir, in.InstituicaoDoCaminho)
	if err != nil {
		return err
	}

	if in.UsuarioID == in.Ator.UsuarioID() {
		return domain.ErrAutoExclusaoNegada // E-08
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		alvo, err := uc.repo.BuscarPorID(ctx, esc, in.UsuarioID) // 404 por escopo (T-04)
		if err != nil {
			return err
		}

		// E-09/E-17/AS-06: excluir o alvo remove TODOS os perfis dele de
		// uma vez — a trava e a contagem rodam para cada perfil crítico
		// que ele possui, antes de qualquer escrita (§5.4, §5.8). É o
		// teste de concorrência que prova a trava (T-090).
		if err := verificarInvarianteDeUltimoDetentor(ctx, uc.repo, esc, alvo.Perfis.Ordenado()); err != nil {
			return err
		}

		instante := time.Now()
		if err := uc.repo.ExcluirLogicamente(ctx, esc, in.UsuarioID, instante); err != nil {
			return err
		}

		acao := auditoria.ExcluirUsuario
		if in.Alcance == autorizacao.AdministradoresDaPlataforma {
			acao = auditoria.ExcluirAdministrador
		}
		atorID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &atorID
		evento.InstituicaoID = alvo.InstituicaoID
		evento.RecursoTipo = "Usuario"
		evento.RecursoID = &alvo.ID
		return uc.audit.Registrar(ctx, evento)
	})
}
