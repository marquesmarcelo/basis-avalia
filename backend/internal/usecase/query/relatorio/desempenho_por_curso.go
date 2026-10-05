package relatorio

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type DesempenhoPorCursoInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroRelatorio
}

// DesempenhoPorCursoUseCase serve GET /api/v1/relatorios/desempenho/por-curso
// (ux.md, design.md §9.4) — o gráfico "Cumprimento de metas por curso".
// Mesma autorização de DesempenhoUseCase (união PI/Coordenador): o gráfico é
// passivo em relação aos mesmos filtros da tabela, nunca um recurso à parte.
type DesempenhoPorCursoUseCase struct{ repo port.EntregaRepository }

func NovoDesempenhoPorCursoUseCase(repo port.EntregaRepository) *DesempenhoPorCursoUseCase {
	return &DesempenhoPorCursoUseCase{repo: repo}
}

func (uc *DesempenhoPorCursoUseCase) Executar(ctx context.Context, in DesempenhoPorCursoInput) (port.ResultadoDesempenhoPorCurso, error) {
	esc, err := AutorizarRelatorio(in.Ator)
	if err != nil {
		return port.ResultadoDesempenhoPorCurso{}, err
	}
	return uc.repo.DesempenhoPorCurso(ctx, esc, in.Filtro)
}
