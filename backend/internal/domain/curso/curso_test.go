package curso

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/google/uuid"
)

func TestNovoCurso_RecusaGrauForaDaLista(t *testing.T) {
	_, err := NovoCurso(uuid.Must(uuid.NewV7()), "Engenharia de Software", "", "mestrado", "presencial")
	if err != domain.ErrValorInvalido {
		t.Fatalf("esperava ErrValorInvalido, obtido %v", err)
	}
}

func TestNovoCurso_RecusaModalidadeForaDaLista(t *testing.T) {
	_, err := NovoCurso(uuid.Must(uuid.NewV7()), "Engenharia de Software", "", "bacharelado", "hibrida")
	if err != domain.ErrValorInvalido {
		t.Fatalf("esperava ErrValorInvalido, obtido %v", err)
	}
}

func TestNovoCurso_CodigoEMecVazioFicaNulo(t *testing.T) {
	c, err := NovoCurso(uuid.Must(uuid.NewV7()), "Pedagogia", "", "licenciatura", "a_distancia")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.CodigoEMec != nil {
		t.Fatal("esperava CodigoEMec nulo quando não informado")
	}
	if c.Situacao != "ativo" {
		t.Fatalf("esperava nascer ativo, obtido %v", c.Situacao)
	}
}

func TestNovoCurso_ComCodigoEMec(t *testing.T) {
	c, err := NovoCurso(uuid.Must(uuid.NewV7()), "Engenharia de Software", "1122334", "bacharelado", "presencial")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.CodigoEMec == nil || c.CodigoEMec.String() != "1122334" {
		t.Fatalf("esperava código e-MEC 1122334, obtido %v", c.CodigoEMec)
	}
}
