package valueobject

import "testing"

func TestNovoNomeCatalogo_RecusaVazio(t *testing.T) {
	if _, err := NovoNomeCatalogo("  "); err == nil {
		t.Fatal("esperava erro para nome vazio")
	}
}

func TestNovoNomeCatalogo_RecusaAcimaDoLimite(t *testing.T) {
	grande := make([]byte, 301)
	for i := range grande {
		grande[i] = 'a'
	}
	if _, err := NovoNomeCatalogo(string(grande)); err == nil {
		t.Fatal("esperava erro para nome acima de 300 caracteres")
	}
}
