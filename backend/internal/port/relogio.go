package port

import "time"

// Relogio existe para que testes de TTL, emt e sessoes_validas_a_partir_de
// sejam determinísticos, sem time.Sleep (design.md §4.2).
type Relogio interface {
	Agora() time.Time
}
