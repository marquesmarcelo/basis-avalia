package valueobject

import (
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func TestSenhaEmTexto_S10_Vazia(t *testing.T) {
	_, err := NovaSenhaEmTexto("")
	if err != domain.ErrSenhaObrigatoria {
		t.Fatalf("esperava ErrSenhaObrigatoria, obtido %v", err)
	}
}

func TestSenhaEmTexto_S11_UmCaractereAceito(t *testing.T) {
	s, err := NovaSenhaEmTexto("a")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.Revelar() != "a" {
		t.Fatalf("esperado 'a', obtido %q", s.Revelar())
	}
}

func TestSenhaEmTexto_S12_1024AceitaE1025Recusada(t *testing.T) {
	limite := strings.Repeat("a", 1024)
	if _, err := NovaSenhaEmTexto(limite); err != nil {
		t.Fatalf("1024 caracteres deveria ser aceito: %v", err)
	}

	acima := strings.Repeat("a", 1025)
	if _, err := NovaSenhaEmTexto(acima); err != domain.ErrSenhaAcimaDoLimite {
		t.Fatalf("esperava ErrSenhaAcimaDoLimite, obtido %v", err)
	}
}

func TestSenhaEmTexto_S09_100CaracteresComAcentoEEspacoSemAlteracao(t *testing.T) {
	original := strings.Repeat("á é í ó ú ç ão ", 6) + "final-com-acento"
	if len([]rune(original)) < 100 {
		t.Fatalf("massa de teste precisa ter ao menos 100 runas, tem %d", len([]rune(original)))
	}
	s, err := NovaSenhaEmTexto(original)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if s.Revelar() != original {
		t.Fatalf("senha alterada: esperado %q, obtido %q", original, s.Revelar())
	}
}

func TestSenhaEmTexto_String_NuncaExpoeValor(t *testing.T) {
	s, _ := NovaSenhaEmTexto("segredo-123")
	if s.String() != "[senha omitida]" {
		t.Fatalf("String() vazou o valor: %q", s.String())
	}
}
