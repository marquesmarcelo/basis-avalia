package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/google/uuid"
)

// ItemDesignacao carrega o nome do coordenador (evita N+1 no grid) e
// OutrasDesignacoesVigentes — o campo único que substitui os dois que o
// ux.md propôs (design.md §5.5, C-08).
type ItemDesignacao struct {
	Designacao                designacao.Designacao
	CoordenadorNome           string
	OutrasDesignacoesVigentes int
}

type FiltroListarDesignacoes struct {
	Situacao      string
	CoordenadorID *uuid.UUID
	Page          int
	PageSize      int
	Sort          string
	Order         string
}

type ResultadoListaDesignacoes struct {
	Itens []ItemDesignacao
	Total int
}

// ItemCandidato é a linha do autocomplete de coordenador (CP-07) — com os
// perfis de cada um, para a tela avisar acúmulo de papel antes de
// confirmar.
type ItemCandidato struct {
	ID     uuid.UUID
	Nome   string
	Email  string
	Perfis []string
}

// DesignacaoRepository serve /api/v1/cursos/{id}/designacoes,
// /api/v1/designacoes/{id} e /api/v1/designacoes/candidatos.
type DesignacaoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemDesignacao, error)
	ListarDoCurso(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID, filtro FiltroListarDesignacoes) (ResultadoListaDesignacoes, error)
	// ListarCandidatos — CP-07: possui qualquer perfil além de aluno, na
	// instituição do escopo. busca filtra por nome/e-mail, limite de 20.
	ListarCandidatos(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]ItemCandidato, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao) error
	// AtualizarCompleta — só quando a situação ATUAL é futura (design.md
	// §5.3); o use case decide, o repositório só grava.
	AtualizarCompleta(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao, versaoEsperada int) error
	// AtualizarFimEPortaria — vigente ou encerrada (design.md §5.3).
	AtualizarFimEPortaria(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao, versaoEsperada int) error
	// Excluir — só futura (DG-08); o use case confere antes de chamar.
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
}
