package valueobject

import "testing"

func TestSigla_NormalizaMaiusculas(t *testing.T) {
	s, err := NovaSigla("fsa ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.String() != "FSA" {
		t.Fatalf("esperado FSA, obtido %q", s.String())
	}
}
