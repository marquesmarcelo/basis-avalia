package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// RodadaDeRecusa — inteiro 0..3 (specs/metas-coordenacao/design.md §3.1,
// PM-2). O construtor recusa qualquer valor acima de 3 — nunca um `if`
// espalhado pelos use cases confere o limite.
type RodadaDeRecusa struct {
	valor int
}

func NovaRodadaDeRecusa(bruta int) (RodadaDeRecusa, error) {
	if bruta < 0 || bruta > 3 {
		return RodadaDeRecusa{}, domain.ErrValorInvalido
	}
	return RodadaDeRecusa{valor: bruta}, nil
}

func (r RodadaDeRecusa) Int() int { return r.valor }

// Esgotada — a terceira recusa já aconteceu (AV-10): nenhum novo prazo é
// aberto, e reenviar responde 409 LIMITE_DE_RODADAS_ATINGIDO.
func (r RodadaDeRecusa) Esgotada() bool { return r.valor >= 3 }

// MaisUma soma uma rodada — usada tanto na recusa direta quanto no
// desfazimento de aceitação (que também consome uma rodada, 3.4).
func (r RodadaDeRecusa) MaisUma() (RodadaDeRecusa, error) {
	return NovaRodadaDeRecusa(r.valor + 1)
}
