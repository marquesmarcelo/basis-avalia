package relogio

import (
	"sync"
	"time"
)

// Relogio — implementação real de port.Relogio. Garante que chamadas
// sucessivas a Agora() nunca retrocedem, mesmo que o relógio de parede do
// sistema operacional retroceda sob carga — comprovado empiricamente em
// container Docker/WSL2 (retrocesso real de até ~550ms sob contenção de
// CPU, medido com amostragem direta de time.Now() sob carga equivalente
// à de argon2). Sem essa garantia, duas chamadas de Agora() em pontos
// diferentes do fluxo (ex: login e troca de senha logo em seguida) podem
// produzir um `emt` de JWT maior que o `sessoes_validas_a_partir_de`
// gravado depois, deixando a sessão antiga válida por engano — foi a
// causa raiz de R5 falhar de forma intermitente mesmo isolado (systematic-
// debugging: proteção implementada aqui, no único ponto por onde todo
// emt e toda invalidação de sessão passam).
type Relogio struct {
	mu     sync.Mutex
	ultimo time.Time
	agora  func() time.Time
}

var (
	singleton    *Relogio
	singletonUma sync.Once
)

// Novo devolve sempre a mesma instância dentro do processo — nunca uma
// nova a cada chamada. A garantia de nunca-retroceder só vale se TODO
// ponto do processo que emite ou compara instante (login, troca de senha,
// e também a fixture de teste em testhelpers, que usa este mesmo Novo())
// disputar o mesmo estado de "última leitura" — duas instâncias
// independentes não se enxergam e o retrocesso volta a ser possível entre
// elas (foi o que sobrou de T02/T05/AS03/AS10/G07: a fixture usava
// time.Now() cru, fora de qualquer Relogio, e cada teste HTTP monta um
// app novo com Novo() — sem singleton, cada um partia do zero).
func Novo() *Relogio {
	singletonUma.Do(func() {
		singleton = &Relogio{agora: time.Now}
	})
	return singleton
}

// Agora trunca para microssegundo porque é a precisão do TIMESTAMPTZ do
// Postgres. Sem isso, o valor gravado no banco e o emt do JWT (que usa
// UnixMicro, que trunca) podem divergir por 1µs quando o Postgres
// arredonda em vez de truncar o resto de nanossegundo.
func (r *Relogio) Agora() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()

	agora := r.agora().Truncate(time.Microsecond)
	if !r.ultimo.IsZero() && !agora.After(r.ultimo) {
		agora = r.ultimo.Add(time.Microsecond)
	}
	r.ultimo = agora
	return agora
}
