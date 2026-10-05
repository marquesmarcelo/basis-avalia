package valueobject

import "testing"

func TestNovaReferenciaInstrumento_AceitaVazia(t *testing.T) {
	r, err := NovaReferenciaInstrumento("   ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !r.Vazia() {
		t.Fatal("esperava referência vazia")
	}
}

func TestNovaReferenciaInstrumento_RecusaAcimaDoLimite(t *testing.T) {
	grande := make([]byte, 501)
	for i := range grande {
		grande[i] = 'a'
	}
	if _, err := NovaReferenciaInstrumento(string(grande)); err == nil {
		t.Fatal("esperava erro para referência acima de 500 caracteres")
	}
}
