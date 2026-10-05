package valueobject

import "testing"

// TestNovaAprovacao_PL03_ParOuNenhum prova que não existe construtor que
// aceite data sem órgão nem órgão sem data (specs/plano-acao/spec.md PL-03).
func TestNovaAprovacao_PL03_ParOuNenhum(t *testing.T) {
	if _, _, err := NovaAprovacao("2026-02-10", ""); err == nil {
		t.Fatal("esperava erro com data sem órgão")
	}
	if _, _, err := NovaAprovacao("", "nde"); err == nil {
		t.Fatal("esperava erro com órgão sem data")
	}
	_, preenchida, err := NovaAprovacao("", "")
	if err != nil || preenchida {
		t.Fatalf("nenhum dos dois deveria ser válido e vazio: preenchida=%v err=%v", preenchida, err)
	}
	aprovacao, preenchida, err := NovaAprovacao("2026-02-10", "nde")
	if err != nil || !preenchida {
		t.Fatalf("os dois preenchidos deveriam ser válidos: %v", err)
	}
	if aprovacao.Orgao() != OrgaoNDE {
		t.Fatalf("esperava NDE, obtido %v", aprovacao.Orgao())
	}
}

func TestNovoOrgaoDeAprovacao_ListaFechada(t *testing.T) {
	if _, err := NovoOrgaoDeAprovacao("Reitoria"); err == nil {
		t.Fatal("esperava VALOR_INVALIDO para órgão fora da lista")
	}
	if _, err := NovoOrgaoDeAprovacao("colegiado_curso"); err != nil {
		t.Fatalf("colegiado_curso deveria ser aceito: %v", err)
	}
}
