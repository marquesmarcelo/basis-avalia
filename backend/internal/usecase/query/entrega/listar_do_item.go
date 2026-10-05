package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarDoItemInput struct {
	Ator        autorizacao.Ator
	ItemPlanoID uuid.UUID
	Filtro      port.FiltroListarEntregas
}

// ListarDoItemUseCase serve GET /api/v1/itens/{itemId}/entregas — o
// coordenador vê todas as entregas dos cursos dele, inclusive de
// antecessores (VI-03); o PI vê todas da instituição.
type ListarDoItemUseCase struct{ repo port.EntregaRepository }

func NovoListarDoItemUseCase(repo port.EntregaRepository) *ListarDoItemUseCase {
	return &ListarDoItemUseCase{repo: repo}
}

func (uc *ListarDoItemUseCase) Executar(ctx context.Context, in ListarDoItemInput) (port.ResultadoListaEntregas, error) {
	esc, err := uc.autorizar(in.Ator)
	if err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	return uc.repo.ListarDoItem(ctx, esc, in.ItemPlanoID, in.Filtro)
}

// autorizar tenta o alcance de instituição (PI) e, se negado, o de
// carteira (Coordenador) — a mesma rota serve os dois perfis (§9 da spec:
// quem acumula os dois vê pelos dois recortes).
func (uc *ListarDoItemUseCase) autorizar(ator autorizacao.Ator) (autorizacao.Escopo, error) {
	return autorizarEntregaLeitura(ator)
}
