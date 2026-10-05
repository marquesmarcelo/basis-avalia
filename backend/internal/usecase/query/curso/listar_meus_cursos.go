package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/port"
)

// ListarMeusCursosUseCase cobre CV-01: cursos ativos com designação
// vigente do ator. O alcance CursosDaCarteira sempre liga
// RestritoACarteiraDe (fundacao-metas.md §3.4) — quem chama sem ser
// Coordenador de Curso não passa da autorização.
type ListarMeusCursosUseCase struct {
	repo port.CursoRepository
}

func NovoListarMeusCursosUseCase(repo port.CursoRepository) *ListarMeusCursosUseCase {
	return &ListarMeusCursosUseCase{repo: repo}
}

func (uc *ListarMeusCursosUseCase) Executar(ctx context.Context, ator autorizacao.Ator) ([]curso.Curso, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.CursosDaCarteira, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.ListarMeusCursos(ctx, esc)
}
