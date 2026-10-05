package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// BuscarPlanoUseCase serve tanto /planos/{id} quanto /meus-planos/{id} —
// o alcance que o handler escolhe já resolve a carteira (P-10): o
// coordenador enxerga o plano do curso dele em qualquer situação, em
// somente leitura; fora do recorte, 404.
type BuscarPlanoUseCase struct{ repo port.PlanoRepository }

func NovoBuscarPlanoUseCase(repo port.PlanoRepository) *BuscarPlanoUseCase {
	return &BuscarPlanoUseCase{repo: repo}
}

func (uc *BuscarPlanoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, alcance autorizacao.Alcance, id uuid.UUID) (port.DetalhePlano, error) {
	esc, err := autorizacao.Autorizar(ator, alcance, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.DetalhePlano{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
