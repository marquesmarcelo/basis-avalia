package valueobject

import "testing"

func TestSenhaHash_String_NuncaExpoeValor(t *testing.T) {
	h, err := NovaSenhaHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if h.String() != "[hash omitido]" {
		t.Fatalf("String() vazou o valor: %q", h.String())
	}
}

func TestSenhaHash_RecusaFormatoDiferenteDePHC(t *testing.T) {
	if _, err := NovaSenhaHash("qualquer-coisa"); err == nil {
		t.Fatal("esperava erro para hash fora do formato PHC argon2id")
	}
}
