package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type DespublicarPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	Versao  int
}

// DespublicarPlanoUseCase cobre SI-10, PP-3: só permitido sem nenhuma
// entrega. A checagem por FOR UPDATE contra corrida (design.md §5.1)
// fica pendente até specs/metas-coordenacao criar a tabela entrega — ver
// testes-pendentes.md; ExisteEntregaNoPlano devolve sempre false por ora,
// então a corrida ainda não é observável.
type DespublicarPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoDespublicarPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *DespublicarPlanoUseCase {
	return &DespublicarPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *DespublicarPlanoUseCase) Executar(ctx context.Context, in DespublicarPlanoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return err
	}
	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		temEntrega, err := uc.repo.ExisteEntregaNoPlano(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		if temEntrega {
			return domain.ErrPlanoComEntrega
		}
		atual := detalhe.Plano
		atual.Despublicar()
		if err := uc.repo.AtualizarSituacao(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.DespublicarPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &atual.ID
		return uc.audit.Registrar(ctx, evento)
	})
}
