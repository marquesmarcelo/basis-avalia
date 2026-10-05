package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type MinhasMetasInput struct {
	Ator      autorizacao.Ator
	PeriodoID uuid.UUID
}

// MinhasMetasUseCase serve GET /api/v1/minhas-metas (T-270) — a única
// exceção ao padrão de CRUD (3.11 da spec): a tela chama esta rota assim
// que monta, sem esperar "Pesquisar".
type MinhasMetasUseCase struct{ repo port.EntregaRepository }

func NovoMinhasMetasUseCase(repo port.EntregaRepository) *MinhasMetasUseCase {
	return &MinhasMetasUseCase{repo: repo}
}

func (uc *MinhasMetasUseCase) Executar(ctx context.Context, in MinhasMetasInput) ([]port.GrupoMinhasMetas, error) {
	if in.PeriodoID == uuid.Nil {
		return nil, domain.ErrPeriodoObrigatorio
	}
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.MinhasMetas(ctx, esc, in.PeriodoID)
}
