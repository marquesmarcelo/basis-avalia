package entrega

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type MarcarPendenciaVistaInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
}

// MarcarPendenciaVistaUseCase — NT-04: idempotente, zera o badge do
// coordenador sem afetar o prazo. Nenhuma auditoria própria (é um "visto",
// não um ato de negócio) e nenhuma transação: é um único UPDATE
// condicional.
type MarcarPendenciaVistaUseCase struct {
	repo port.EntregaRepository
}

func NovoMarcarPendenciaVistaUseCase(repo port.EntregaRepository) *MarcarPendenciaVistaUseCase {
	return &MarcarPendenciaVistaUseCase{repo: repo}
}

func (uc *MarcarPendenciaVistaUseCase) Executar(ctx context.Context, in MarcarPendenciaVistaInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoEditar, nil)
	if err != nil {
		return err
	}
	return uc.repo.MarcarPendenciaVista(ctx, esc, in.EntregaID)
}
