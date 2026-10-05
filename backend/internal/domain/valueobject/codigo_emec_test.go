package valueobject

import "testing"

func TestCodigoEMec_I02_AceitaNulo(t *testing.T) {
	c, err := NovoCodigoEMec("")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !c.Nulo() {
		t.Fatal("código e-MEC vazio deveria ser nulo")
	}
}

func TestCodigoEMec_NaoValidaFormato(t *testing.T) {
	c, err := NovoCodigoEMec("qualquer-coisa-123")
	if err != nil {
		t.Fatalf("não deveria validar formato: %v", err)
	}
	if c.Nulo() || c.String() != "qualquer-coisa-123" {
		t.Fatalf("valor não preservado: %q", c.String())
	}
}
