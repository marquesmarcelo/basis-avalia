package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// SituacaoEntrega — TRÊS valores, não quatro (specs/metas-coordenacao/
// design.md M-03, §3.1): "recusada_definitiva" não existe como valor
// armazenado. Uma entrega recusada está definitivamente recusada quando
// as rodadas se esgotaram OU o prazo de correção não está mais em curso —
// duas portas derivadas, decididas por entrega.Entrega.PodeCorrigir, nunca
// uma quarta coluna. Armazenar o estado terminal exigiria uma escrita no
// instante em que o prazo vence, e não há escrita naquele instante — é a
// mesma razão do perfil de Coordenador derivado (fundacao-metas.md §4.1).
type SituacaoEntrega string

const (
	PendenteAvaliacao SituacaoEntrega = "pendente_avaliacao"
	Aceita            SituacaoEntrega = "aceita"
	Recusada          SituacaoEntrega = "recusada"
)

func NovaSituacaoEntrega(bruta string) (SituacaoEntrega, error) {
	s := SituacaoEntrega(bruta)
	switch s {
	case PendenteAvaliacao, Aceita, Recusada:
		return s, nil
	default:
		return "", domain.ErrValorInvalido
	}
}
