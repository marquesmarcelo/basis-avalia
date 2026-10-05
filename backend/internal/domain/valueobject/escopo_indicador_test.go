package valueobject

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func TestNovoEscopoIndicador_AceitaOsDoisValores(t *testing.T) {
	if _, err := NovoEscopoIndicador("plataforma"); err != nil {
		t.Fatalf("plataforma: %v", err)
	}
	if _, err := NovoEscopoIndicador("instituicao"); err != nil {
		t.Fatalf("instituicao: %v", err)
	}
}

func TestNovoEscopoIndicador_RecusaValorFora(t *testing.T) {
	_, err := NovoEscopoIndicador("nacional")
	if err != domain.ErrValorInvalido {
		t.Fatalf("esperava ErrValorInvalido, obtido %v", err)
	}
}

func TestEscopoIndicador_PertenceAInstituicao(t *testing.T) {
	if EscopoIndicadorPlataforma.PertenceAInstituicao() {
		t.Fatal("plataforma não pertence a instituição")
	}
	if !EscopoIndicadorInstituicao.PertenceAInstituicao() {
		t.Fatal("instituicao pertence a instituição")
	}
}
