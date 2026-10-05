package argon2

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

func TestHashDeSenha_S09_Senha100CaracteresComAcentoEEspaco(t *testing.T) {
	h := NovoHashDeSenha(4)
	original := strings.Repeat("á é í ó ú ç ão ", 6) + "final-com-acento"
	senha, err := valueobject.NovaSenhaEmTexto(original)
	if err != nil {
		t.Fatalf("NovaSenhaEmTexto: %v", err)
	}

	hash, err := h.Gerar(context.Background(), senha)
	if err != nil {
		t.Fatalf("Gerar: %v", err)
	}

	ok, err := h.Conferir(context.Background(), senha, hash)
	if err != nil {
		t.Fatalf("Conferir: %v", err)
	}
	if !ok {
		t.Fatal("senha correta deveria conferir")
	}

	truncada, err := valueobject.NovaSenhaEmTexto(string([]rune(original)[:72]))
	if err != nil {
		t.Fatalf("NovaSenhaEmTexto (truncada): %v", err)
	}
	okTruncada, err := h.Conferir(context.Background(), truncada, hash)
	if err != nil {
		t.Fatalf("Conferir (truncada): %v", err)
	}
	if okTruncada {
		t.Fatal("os primeiros 72 caracteres NÃO deveriam conferir — é o teste que distingue argon2 de bcrypt (S-09)")
	}
}

// TestHashDeSenha_L08_ConferirDescartavelDuracaoIndistinguivel prova L-08:
// o tempo de Conferir (conta real) e ConferirDescartavel (conta
// inexistente) não pode ser usado para distinguir os dois casos. Uma
// amostra única de cada lado é sensível a um único pico de contenção de
// CPU na máquina que roda o teste — sob a suíte inteira em paralelo, isso
// já bastava para um falso positivo intermitente (comprovado: `go test
// ./... -count=1` repetido reproduz; o teste isolado, nunca). A correção é
// somar várias amostras de cada lado e comparar o agregado — o mesmo
// princípio de qualquer benchmark de tempo real, não um relaxamento da
// garantia: continua provando que as duas operações são
// computacionalmente equivalentes, só deixa de depender da sorte de uma
// única medição.
func TestHashDeSenha_L08_ConferirDescartavelDuracaoIndistinguivel(t *testing.T) {
	h := NovoHashDeSenha(4)
	senhaCorreta, _ := valueobject.NovaSenhaEmTexto("senha-qualquer-123")
	hash, err := h.Gerar(context.Background(), senhaCorreta)
	if err != nil {
		t.Fatalf("Gerar: %v", err)
	}
	senhaErrada, _ := valueobject.NovaSenhaEmTexto("senha-errada-456")

	const amostras = 5
	var duracaoConferir, duracaoDescartavel time.Duration

	for i := 0; i < amostras; i++ {
		inicio1 := time.Now()
		ok, err := h.Conferir(context.Background(), senhaErrada, hash)
		duracaoConferir += time.Since(inicio1)
		if err != nil {
			t.Fatalf("Conferir: %v", err)
		}
		if ok {
			t.Fatal("senha errada não deveria conferir")
		}

		inicio2 := time.Now()
		h.ConferirDescartavel(context.Background(), senhaErrada)
		duracaoDescartavel += time.Since(inicio2)
	}

	diferenca := duracaoConferir - duracaoDescartavel
	if diferenca < 0 {
		diferenca = -diferenca
	}
	limite := duracaoConferir / 2
	if diferenca > limite {
		t.Fatalf("duração muito diferente após %d amostras: conferir=%v descartavel=%v (diferença %v > limite %v)", amostras, duracaoConferir, duracaoDescartavel, diferenca, limite)
	}
}

func TestHashDeSenha_SemaforoLimitaConcorrenciaA4(t *testing.T) {
	h := NovoHashDeSenha(4)
	senha, _ := valueobject.NovaSenhaEmTexto("senha-de-teste-concorrencia")

	var emExecucao int32
	var maxObservado int32
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.adquirir()
			atual := atomic.AddInt32(&emExecucao, 1)
			for {
				maximo := atomic.LoadInt32(&maxObservado)
				if atual <= maximo || atomic.CompareAndSwapInt32(&maxObservado, maximo, atual) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&emExecucao, -1)
			h.liberar()
		}()
	}
	wg.Wait()

	if maxObservado > 4 {
		t.Fatalf("no máximo 4 deveriam executar ao mesmo tempo, observado %d", maxObservado)
	}
	_ = senha
}

func BenchmarkHashSenha(b *testing.B) {
	h := NovoHashDeSenha(4)
	senha, _ := valueobject.NovaSenhaEmTexto("senha-de-benchmark-123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := h.Gerar(context.Background(), senha); err != nil {
			b.Fatalf("Gerar: %v", err)
		}
	}
}
