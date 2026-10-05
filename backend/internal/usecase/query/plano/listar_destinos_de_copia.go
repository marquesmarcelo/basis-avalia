package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarDestinosDeCopiaInput struct {
	Ator             autorizacao.Ator
	PlanoOrigemID    uuid.UUID
	PeriodoDestinoID uuid.UUID
	Busca            string
}

// ListarDestinosDeCopiaUseCase serve GET /planos/{id}/destinos-copia —
// UMA consulta, com ja_tem_plano e vago por curso (design.md V-5); a
// tela usa isso só como conforto, a garantia real é o índice único na
// hora de confirmar (CP-05).
type ListarDestinosDeCopiaUseCase struct{ repo port.PlanoRepository }

func NovoListarDestinosDeCopiaUseCase(repo port.PlanoRepository) *ListarDestinosDeCopiaUseCase {
	return &ListarDestinosDeCopiaUseCase{repo: repo}
}

func (uc *ListarDestinosDeCopiaUseCase) Executar(ctx context.Context, in ListarDestinosDeCopiaInput) ([]port.DestinoDeCopia, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.ListarDestinosDeCopia(ctx, esc, in.PlanoOrigemID, in.PeriodoDestinoID, in.Busca)
}
