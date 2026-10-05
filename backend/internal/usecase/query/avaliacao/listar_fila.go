package avaliacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ListarFilaInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroFilaAvaliacao
}

// ListarFilaUseCase serve GET /api/v1/avaliacoes (PI) — cada linha traz
// coordenado_pelo_avaliador, para a tela avisar antes de abrir (T-277).
type ListarFilaUseCase struct{ repo port.EntregaRepository }

func NovoListarFilaUseCase(repo port.EntregaRepository) *ListarFilaUseCase {
	return &ListarFilaUseCase{repo: repo}
}

func (uc *ListarFilaUseCase) Executar(ctx context.Context, in ListarFilaInput) (port.ResultadoListaEntregas, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	return uc.repo.ListarFilaDeAvaliacao(ctx, esc, in.Ator.Proprio(), in.Filtro)
}
