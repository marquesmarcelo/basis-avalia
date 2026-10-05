package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarCursoUseCase struct {
	repo port.CursoRepository
}

func NovoBuscarCursoUseCase(repo port.CursoRepository) *BuscarCursoUseCase {
	return &BuscarCursoUseCase{repo: repo}
}

func (uc *BuscarCursoUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemCurso, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemCurso{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
