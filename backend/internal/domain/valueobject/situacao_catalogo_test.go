package valueobject

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func TestNovaSituacaoCatalogo_RecusaValorFora(t *testing.T) {
	_, err := NovaSituacaoCatalogo("pendente")
	if err != domain.ErrValorInvalido {
		t.Fatalf("esperava ErrValorInvalido, obtido %v", err)
	}
}

func TestSituacaoCatalogo_Alternar(t *testing.T) {
	if CatalogoAtivo.Alternar() != CatalogoInativo {
		t.Fatal("ativo deveria alternar para inativo")
	}
	if CatalogoInativo.Alternar() != CatalogoAtivo {
		t.Fatal("inativo deveria alternar para ativo")
	}
}
