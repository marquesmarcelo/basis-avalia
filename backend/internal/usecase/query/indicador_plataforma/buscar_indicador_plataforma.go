package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type BuscarIndicadorPlataformaUseCase struct {
	repo port.IndicadorPlataformaRepository
}

func NovoBuscarIndicadorPlataformaUseCase(repo port.IndicadorPlataformaRepository) *BuscarIndicadorPlataformaUseCase {
	return &BuscarIndicadorPlataformaUseCase{repo: repo}
}

// Executar cobre IE-09: Administrador pedindo indicador institucional ou
// meta recebe 404 — aqui ele só nunca alcança nada fora do escopo
// plataforma, o que já é garantido por AplicarEscopo.
func (uc *BuscarIndicadorPlataformaUseCase) Executar(ctx context.Context, ator autorizacao.Ator, id uuid.UUID) (port.ItemIndicadorPlataforma, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoBuscar, nil)
	if err != nil {
		return port.ItemIndicadorPlataforma{}, err
	}
	return uc.repo.BuscarPorID(ctx, esc, id)
}
