package pendencia

import (
	"context"
	"errors"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// ContarBadgesUseCase serve GET /api/v1/metas/pendencias — os dois badges
// do menu, conforme os PERFIS EFETIVOS do ator (fundacao-metas.md).
// Quem acumula os dois perfis vê os dois números, sem um esconder o
// outro — por isso os dois alcances são tentados independentemente,
// nunca com fallback um do outro.
type ContarBadgesUseCase struct{ repo port.EntregaRepository }

func NovoContarBadgesUseCase(repo port.EntregaRepository) *ContarBadgesUseCase {
	return &ContarBadgesUseCase{repo: repo}
}

func (uc *ContarBadgesUseCase) Executar(ctx context.Context, ator autorizacao.Ator) (port.PendenciasResposta, error) {
	resposta := port.PendenciasResposta{}
	temAlgum := false

	if esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoListar, nil); err == nil {
		n, err := uc.repo.ContarPendentesDeAvaliacao(ctx, esc)
		if err != nil {
			return port.PendenciasResposta{}, err
		}
		resposta.PendentesDeAvaliacao = n
		temAlgum = true
	} else if !errors.Is(err, domain.ErrPermissaoNegada) {
		return port.PendenciasResposta{}, err
	}

	if esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoListar, nil); err == nil {
		n, err := uc.repo.ContarPendenciasNaoVistas(ctx, esc)
		if err != nil {
			return port.PendenciasResposta{}, err
		}
		resposta.PendenciasNaoVistas = n
		temAlgum = true
	} else if !errors.Is(err, domain.ErrPermissaoNegada) {
		return port.PendenciasResposta{}, err
	}

	if !temAlgum {
		return port.PendenciasResposta{}, domain.ErrPermissaoNegada
	}
	return resposta, nil
}
