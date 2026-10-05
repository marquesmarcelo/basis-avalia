package valueobject

import "testing"

func TestNovoCodigoIndicador_RecusaVazio(t *testing.T) {
	if _, err := NovoCodigoIndicador("   "); err == nil {
		t.Fatal("esperava erro para código vazio")
	}
}

func TestNovoCodigoIndicador_RecusaAcimaDoLimite(t *testing.T) {
	grande := make([]byte, 51)
	for i := range grande {
		grande[i] = 'a'
	}
	if _, err := NovoCodigoIndicador(string(grande)); err == nil {
		t.Fatal("esperava erro para código acima de 50 caracteres")
	}
}

func TestNovoCodigoIndicador_RecusaQuebraDeLinha(t *testing.T) {
	if _, err := NovoCodigoIndicador("1.4\noutralinha"); err == nil {
		t.Fatal("esperava erro para código com quebra de linha")
	}
}

func TestNovoCodigoIndicador_AceitaNumericoENaoNormaliza(t *testing.T) {
	c, err := NovoCodigoIndicador(" GEST-01 ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.String() != "GEST-01" {
		t.Fatalf("esperava 'GEST-01' (sem normalizar caixa), obtido %q", c.String())
	}
}
