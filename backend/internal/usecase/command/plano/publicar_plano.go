package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// PublicarPlanoInput.Versao é a versão que o cliente tinha em mãos —
// conflito de concorrência otimista de qualquer forma (PL-07).
type PublicarPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	Versao  int
}

type PublicarPlanoResultado struct {
	Itens        int
	TotalExigido int
	CursoVago    bool
	SemAprovacao bool
}

// PublicarPlanoUseCase cobre SI-02, SI-03, SI-04, SI-08, SI-09.
type PublicarPlanoUseCase struct {
	repo        port.PlanoRepository
	periodoRepo port.PeriodoRepository
	audit       port.AuditLogger
	uow         port.UnidadeDeTrabalho
}

func NovoPublicarPlanoUseCase(repo port.PlanoRepository, periodoRepo port.PeriodoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *PublicarPlanoUseCase {
	return &PublicarPlanoUseCase{repo: repo, periodoRepo: periodoRepo, audit: audit, uow: uow}
}

func (uc *PublicarPlanoUseCase) Executar(ctx context.Context, in PublicarPlanoInput) (PublicarPlanoResultado, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return PublicarPlanoResultado{}, err
	}

	var resultado PublicarPlanoResultado
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		per, err := uc.periodoRepo.BuscarPorID(ctx, esc, detalhe.Plano.PeriodoID)
		if err != nil {
			return err
		}
		periodoEncerrado := per.Periodo.Vigencia.SituacaoEm(esc.DataDeReferencia()) == valueobject.Encerrada

		atual := detalhe.Plano
		if err := atual.Publicar(detalhe.TotalItens > 0, periodoEncerrado); err != nil {
			return err
		}
		if err := uc.repo.AtualizarSituacao(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.PublicarPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &atual.ID
		// tinha_aprovacao: exigência de auditoria explícita
		// (design.md §8) — a pergunta que alguém faz meses depois não é
		// dedutível da data de aprovação sozinha.
		evento.Detalhes["tinha_aprovacao"] = atual.Aprovacao != nil
		evento.Detalhes["curso_vago"] = detalhe.CursoVago
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = PublicarPlanoResultado{
			Itens: detalhe.TotalItens, TotalExigido: detalhe.TotalExigido,
			CursoVago: detalhe.CursoVago, SemAprovacao: atual.Aprovacao == nil,
		}
		return nil
	})
	if erro != nil {
		return PublicarPlanoResultado{}, erro
	}
	return resultado, nil
}
