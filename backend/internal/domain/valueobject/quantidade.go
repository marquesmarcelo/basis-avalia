package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// Quantidade — inteiro >= 1 (specs/plano-acao/design.md §3.1). Existe
// como tipo próprio porque é o número que a feature inteira existe para
// proteger: nunca dividida entre cursos, nunca multiplicada pelo número
// de indicadores da meta — um tipo próprio impede que ela seja somada,
// dividida ou multiplicada por acidente em qualquer camada.
type Quantidade struct {
	valor int
}

func NovaQuantidade(bruta int) (Quantidade, error) {
	if bruta < 1 {
		return Quantidade{}, domain.ErrQuantidadeInvalida
	}
	return Quantidade{valor: bruta}, nil
}

func (q Quantidade) Int() int { return q.valor }
