package valueobject

import (
	"unicode"

	"github.com/basis-avalia/backend/internal/domain"
)

// ProibirCaractereDeControle — regra de domínio (fundacao-metas.md §4.7):
// validar na entrada não basta, porque quem injeta usa a gramática do
// destino. Um texto com \r\n é dado válido para o banco e para a tela, e
// vira INSTRUÇÃO ao atravessar a fronteira do SMTP (um cabeçalho novo) ou
// de outros destinos que interpretam controle. Esta regra vale mesmo que
// nenhum desses destinos existisse: um caractere de controle não é
// conteúdo de "uma linha" de texto livre — é comando para quem lê depois.
//
// Usada na construção de qualquer texto livre do usuário que se torna
// motivo/observação de um ato de negócio (o Value Object recusa; o
// adapter de saída, além disso, neutraliza — as duas camadas têm papéis
// distintos, e colapsá-las é como a regra volta a furar).
func ProibirCaractereDeControle(texto string) error {
	for _, r := range texto {
		if unicode.IsControl(r) {
			return domain.ErrCaractereDeControleNaoPermitido
		}
	}
	return nil
}
