package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
}

// BuscarUseCase serve GET /api/v1/entregas/{id} — usado tanto pela tela de
// correção do coordenador quanto pela tela de avaliação do PI.
type BuscarUseCase struct{ repo port.EntregaRepository }

func NovoBuscarUseCase(repo port.EntregaRepository) *BuscarUseCase {
	return &BuscarUseCase{repo: repo}
}

func (uc *BuscarUseCase) Executar(ctx context.Context, in BuscarInput) (port.DetalheEntrega, error) {
	esc, err := autorizarEntregaLeitura(in.Ator)
	if err != nil {
		return port.DetalheEntrega{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
}
