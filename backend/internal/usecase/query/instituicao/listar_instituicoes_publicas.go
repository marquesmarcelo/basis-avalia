package instituicao

import (
	"context"

	"github.com/basis-avalia/backend/internal/port"
)

// ListarInstituicoesPublicasUseCase alimenta a rota pública do combo de
// login (R1). Não recebe Ator nem Escopo — é público por desenho
// (design.md §4.1, 3.17).
type ListarInstituicoesPublicasUseCase struct {
	query port.InstituicaoPublicaQuery
}

func NovoListarInstituicoesPublicasUseCase(query port.InstituicaoPublicaQuery) *ListarInstituicoesPublicasUseCase {
	return &ListarInstituicoesPublicasUseCase{query: query}
}

func (uc *ListarInstituicoesPublicasUseCase) Executar(ctx context.Context) ([]port.InstituicaoPublica, error) {
	return uc.query.ListarParaCombo(ctx)
}
