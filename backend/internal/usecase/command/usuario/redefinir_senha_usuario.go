package usuario

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type RedefinirSenhaUsuarioInput struct {
	Ator                 autorizacao.Ator
	Alcance              autorizacao.Alcance
	InstituicaoDoCaminho *uuid.UUID
	UsuarioID            uuid.UUID
	SenhaNova            valueobject.SenhaEmTexto
}

// RedefinirSenhaUsuarioUseCase cobre E-11, E-12, SE-04, T-06.
type RedefinirSenhaUsuarioUseCase struct {
	repo  port.UsuarioRepository
	hash  port.HashDeSenha
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoRedefinirSenhaUsuarioUseCase(repo port.UsuarioRepository, hash port.HashDeSenha, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *RedefinirSenhaUsuarioUseCase {
	return &RedefinirSenhaUsuarioUseCase{repo: repo, hash: hash, audit: audit, uow: uow}
}

func (uc *RedefinirSenhaUsuarioUseCase) Executar(ctx context.Context, in RedefinirSenhaUsuarioInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoRedefinirSenha, in.InstituicaoDoCaminho)
	if err != nil {
		return err
	}

	if in.UsuarioID == in.Ator.UsuarioID() {
		return domain.ErrRedefinirPropriaSenhaNegada // E-12
	}

	hash, err := uc.hash.Gerar(ctx, in.SenhaNova)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		alvo, err := uc.repo.BuscarPorID(ctx, esc, in.UsuarioID) // 404 por escopo (T-06)
		if err != nil {
			return err
		}

		instante := time.Now()
		// DefinirSenha grava sessoes_validas_a_partir_de = instante — é o
		// que derruba a sessão em andamento do afetado (E-11, SE-04).
		if err := uc.repo.DefinirSenha(ctx, esc, in.UsuarioID, hash, true, instante); err != nil {
			return err
		}

		atorID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.RedefinirSenhaUsuario, auditoria.ResultadoSucesso)
		evento.AtorID = &atorID
		evento.InstituicaoID = alvo.InstituicaoID
		evento.RecursoTipo = "Usuario"
		evento.RecursoID = &alvo.ID
		return uc.audit.Registrar(ctx, evento)
	})
}
