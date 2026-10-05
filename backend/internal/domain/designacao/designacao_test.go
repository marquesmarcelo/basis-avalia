package designacao

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func data(t *testing.T, iso string) valueobject.DataLocal {
	t.Helper()
	d, err := valueobject.DataLocalTexto(iso)
	if err != nil {
		t.Fatalf("data inválida: %v", err)
	}
	return d
}

func novaDesignacaoTeste(t *testing.T, inicio valueobject.DataLocal, fim *valueobject.DataLocal) *Designacao {
	t.Helper()
	d, err := NovaDesignacao(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), "47/2026", inicio, fim, false)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return d
}

func TestNovaDesignacao_AutodesignacaoGravadaNoAto(t *testing.T) {
	coordenadorID := uuid.Must(uuid.NewV7())
	d, err := NovaDesignacao(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), coordenadorID, "70/2026", data(t, "2026-03-01"), nil, true)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !d.Autodesignacao {
		t.Fatal("esperava autodesignacao verdadeiro")
	}
}

// TestDesignacao_AtualizarCompleta_SoQuandoFutura prova §5.3: futura
// aceita mudar coordenador e início livremente.
func TestDesignacao_AtualizarCompleta_PermiteTrocarCoordenador(t *testing.T) {
	inicio := data(t, "2026-05-01")
	d := novaDesignacaoTeste(t, inicio, nil)
	novoCoordenador := uuid.Must(uuid.NewV7())
	if err := d.AtualizarCompleta(novoCoordenador, "88/2026", data(t, "2026-06-01"), nil); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if d.CoordenadorID != novoCoordenador {
		t.Fatal("esperava coordenador atualizado")
	}
}

// TestErrSeNaoForFutura_RecusaVigenteEEncerrada prova a assimetria de
// §5.3: 409, não 400 — é o estado do recurso que recusa, não o formato.
func TestErrSeNaoForFutura_RecusaVigenteEEncerrada(t *testing.T) {
	if err := ErrSeNaoForFutura(valueobject.Vigente); err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito para vigente, obtido %v", err)
	}
	if err := ErrSeNaoForFutura(valueobject.Encerrada); err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito para encerrada, obtido %v", err)
	}
	if err := ErrSeNaoForFutura(valueobject.Futura); err != nil {
		t.Fatalf("futura não deveria recusar, obtido %v", err)
	}
}

func TestDesignacao_AtualizarFimEPortaria_NaoTocaCoordenadorNemInicio(t *testing.T) {
	inicio := data(t, "2026-01-01")
	d := novaDesignacaoTeste(t, inicio, nil)
	coordenadorAntes := d.CoordenadorID
	novoFim := data(t, "2026-04-30")
	if err := d.AtualizarFimEPortaria("47/2026 (prorrogada)", &novoFim); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if d.CoordenadorID != coordenadorAntes {
		t.Fatal("AtualizarFimEPortaria não deveria tocar o coordenador")
	}
	if d.Vigencia.Inicio() != inicio {
		t.Fatal("AtualizarFimEPortaria não deveria tocar o início")
	}
	if d.Vigencia.Fim() == nil || *d.Vigencia.Fim() != novoFim {
		t.Fatal("esperava o novo fim aplicado")
	}
}

func TestDesignacao_PodeExcluir_SoFutura(t *testing.T) {
	d := novaDesignacaoTeste(t, data(t, "2026-05-01"), nil)
	if !d.PodeExcluir(data(t, "2026-03-15")) {
		t.Fatal("designação futura deveria poder ser excluída")
	}
	if d.PodeExcluir(data(t, "2026-05-01")) {
		t.Fatal("designação vigente não deveria poder ser excluída")
	}
}
