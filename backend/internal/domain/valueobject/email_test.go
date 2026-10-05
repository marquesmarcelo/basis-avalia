package valueobject

import "testing"

func TestEmail_U04_NormalizaTrimEMinusculas(t *testing.T) {
	e, err := NovoEmail("  Joao.Ribeiro@IES.edu.BR  ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if e.String() != "joao.ribeiro@ies.edu.br" {
		t.Fatalf("esperado joao.ribeiro@ies.edu.br, obtido %q", e.String())
	}
}

func TestEmail_FormatoInvalido(t *testing.T) {
	if _, err := NovoEmail("joao.ribeiro.ies.edu.br"); err == nil {
		t.Fatal("esperava erro para e-mail sem formato válido")
	}
}
