package valueobject

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

// fundacao-metas.md §4.7 (T-123): motivo de recusa é texto livre que
// atravessa a fronteira do SMTP (corpo do e-mail de notificação) — um
// \r\n aqui é dado válido para o banco e vira início de um cabeçalho
// novo ao montar a mensagem. O Value Object recusa antes de qualquer
// escrita.
func TestNovoResultadoDeAvaliacao_MotivoComCRLF(t *testing.T) {
	_, err := NovoResultadoDeAvaliacao("recusada", "motivo qualquer\r\nBcc: atacante@example.com")
	if err != domain.ErrCaractereDeControleNaoPermitido {
		t.Fatalf("esperava ErrCaractereDeControleNaoPermitido, obteve %v", err)
	}
}

func TestNovoResultadoDeAvaliacao_MotivoSemCaractereDeControleAceito(t *testing.T) {
	r, err := NovoResultadoDeAvaliacao("recusada", "A lista de presença não corresponde à data da ata.")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !r.Recusada() {
		t.Fatal("esperava resultado recusada")
	}
}

func TestProibirCaractereDeControle(t *testing.T) {
	casos := []struct {
		texto     string
		permitido bool
	}{
		{"texto normal", true},
		{"texto com acentuação e travessão — assim", true},
		{"", true},
		{"linha 1\nlinha 2", false},
		{"cabeçalho\r\ninjetado", false},
		{"tabulação\tno meio", false},
	}
	for _, c := range casos {
		err := ProibirCaractereDeControle(c.texto)
		if c.permitido && err != nil {
			t.Errorf("%q: esperava nil, obteve %v", c.texto, err)
		}
		if !c.permitido && err != domain.ErrCaractereDeControleNaoPermitido {
			t.Errorf("%q: esperava ErrCaractereDeControleNaoPermitido, obteve %v", c.texto, err)
		}
	}
}
