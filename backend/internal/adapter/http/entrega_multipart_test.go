package http

import (
	"io"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

// leitorInfinito simula um cliente malicioso oferecendo um stream sem
// fim — nunca devolve io.EOF. Conta quantos bytes foram efetivamente
// pedidos pelo chamador, para provar que lerConteudoLimitado nunca lê
// além do limite combinado, não importa quantos bytes a origem ofereça.
type leitorInfinito struct {
	bytesPedidos int64
}

func (l *leitorInfinito) Read(p []byte) (int, error) {
	l.bytesPedidos += int64(len(p))
	for i := range p {
		p[i] = 'A'
	}
	return len(p), nil
}

// M-08 (design.md §6.3): nunca io.ReadAll sobre um reader ilimitado. Um
// envio de 2 GB tem de ser cortado ANTES de esgotar a origem — este teste
// prova isso sem precisar montar 2 GB de verdade: a origem é literalmente
// infinita, e mesmo assim a função devolve o erro e nunca lê mais do que
// limite+1 bytes.
func TestLerConteudoLimitado_NuncaLeAlemDoLimite(t *testing.T) {
	const limite = 1024 // 1 KB, para o teste rodar instantâneo
	origem := &leitorInfinito{}

	_, err := lerConteudoLimitado(origem, limite)
	if err != domain.ErrAnexoAcimaDoLimite {
		t.Fatalf("esperava ErrAnexoAcimaDoLimite, obteve %v", err)
	}
	if origem.bytesPedidos > limite+1 {
		t.Fatalf("lerConteudoLimitado leu %d bytes de uma origem infinita — esperava no máximo %d (limite+1)", origem.bytesPedidos, limite+1)
	}
}

func TestLerConteudoLimitado_AceitaExatamenteNoLimite(t *testing.T) {
	const limite = 1024
	origemFinita := io.LimitReader(&leitorInfinito{}, limite)

	conteudo, err := lerConteudoLimitado(origemFinita, limite)
	if err != nil {
		t.Fatalf("erro inesperado no limite exato: %v", err)
	}
	if int64(len(conteudo)) != limite {
		t.Fatalf("esperava %d bytes, obteve %d", limite, len(conteudo))
	}
}

func TestLerConteudoLimitado_RecusaUmByteAcimaDoLimite(t *testing.T) {
	const limite = 1024
	origemFinita := io.LimitReader(&leitorInfinito{}, limite+1)

	_, err := lerConteudoLimitado(origemFinita, limite)
	if err != domain.ErrAnexoAcimaDoLimite {
		t.Fatalf("esperava ErrAnexoAcimaDoLimite para um arquivo 1 byte acima do limite, obteve %v", err)
	}
}
