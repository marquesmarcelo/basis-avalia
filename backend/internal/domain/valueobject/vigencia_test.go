package valueobject

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func data(t *testing.T, iso string) DataLocal {
	t.Helper()
	d, err := DataLocalTexto(iso)
	if err != nil {
		t.Fatalf("data inválida %q: %v", iso, err)
	}
	return d
}

func TestNovaVigencia_RecusaFimAntesDoInicio(t *testing.T) {
	inicio := data(t, "2026-03-01")
	fim := data(t, "2026-02-01")
	_, err := NovaVigencia(inicio, &fim)
	if err != domain.ErrDesignacaoDatasInvalidas {
		t.Fatalf("esperava ErrDesignacaoDatasInvalidas, obtido %v", err)
	}
}

// TestVigencia_SituacaoEm_DG06 prova as três fronteiras do dia inteiro
// (specs/cursos/spec.md DG-06): 31/07 ainda vigente, 01/08 já encerrada,
// data_fim nula sempre vigente.
func TestVigencia_SituacaoEm_DG06(t *testing.T) {
	inicio := data(t, "2026-01-01")
	fim := data(t, "2026-07-31")
	v, err := NovaVigencia(inicio, &fim)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if got := v.SituacaoEm(data(t, "2026-07-31")); got != Vigente {
		t.Fatalf("31/07 deveria ser vigente, obtido %v", got)
	}
	if got := v.SituacaoEm(data(t, "2026-08-01")); got != Encerrada {
		t.Fatalf("01/08 deveria ser encerrada, obtido %v", got)
	}

	semFim, err := NovaVigencia(inicio, nil)
	if err != nil {
		t.Fatalf("setup sem fim: %v", err)
	}
	if got := semFim.SituacaoEm(data(t, "2030-01-01")); got != Vigente {
		t.Fatalf("sem data de fim deveria ser sempre vigente, obtido %v", got)
	}
}

func TestVigencia_SituacaoEm_Futura(t *testing.T) {
	v, err := NovaVigencia(data(t, "2026-05-01"), nil)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if got := v.SituacaoEm(data(t, "2026-03-15")); got != Futura {
		t.Fatalf("esperava futura, obtido %v", got)
	}
}
