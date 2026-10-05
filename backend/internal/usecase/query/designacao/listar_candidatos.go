package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// ListarCandidatosUseCase cobre CP-07 — o mesmo alcance de
// DesignacoesDaInstituicao autoriza (quem gerencia designação é quem
// escolhe candidato), a listagem em si mora em AlvoUsuario.
type ListarCandidatosUseCase struct {
	repo port.DesignacaoRepository
}

func NovoListarCandidatosUseCase(repo port.DesignacaoRepository) *ListarCandidatosUseCase {
	return &ListarCandidatosUseCase{repo: repo}
}

func (uc *ListarCandidatosUseCase) Executar(ctx context.Context, ator autorizacao.Ator, busca string) ([]port.ItemCandidato, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.ListarCandidatos(ctx, esc, busca)
}
