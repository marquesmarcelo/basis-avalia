package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarMeusPlanosInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroListarPlanos
}

// ListarMeusPlanosUseCase serve GET /api/v1/meus-planos — o coordenador
// vê os planos dos cursos que coordena, em TODAS as situações, somente
// leitura (P-10, design.md §5.4). O filtro de "somente vigentes" NÃO
// existe aqui — ele vive na consulta de obrigações, em
// specs/metas-coordenacao.
type ListarMeusPlanosUseCase struct{ repo port.PlanoRepository }

func NovoListarMeusPlanosUseCase(repo port.PlanoRepository) *ListarMeusPlanosUseCase {
	return &ListarMeusPlanosUseCase{repo: repo}
}

func (uc *ListarMeusPlanosUseCase) Executar(ctx context.Context, in ListarMeusPlanosInput) (port.ResultadoListaPlanos, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaCarteira, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaPlanos{}, err
	}
	return uc.repo.ListarMeusPlanos(ctx, esc, in.Filtro)
}
