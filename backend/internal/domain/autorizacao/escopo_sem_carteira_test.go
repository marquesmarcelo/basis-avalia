package autorizacao

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// TestEscopo_SemCarteira_RemoveSoACarteiraMantemInstituicao prova
// fundacao-metas.md §3.4/design.md de plano-acao: SemCarteira() tira só a
// dimensão de carteira — a instituição continua isolando.
func TestEscopo_SemCarteira_RemoveSoACarteiraMantemInstituicao(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	conjunto, err := valueobject.NovoConjunto(valueobject.CoordenadorCurso)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	esc, err := Autorizar(ator, PlanosDaCarteira, AcaoBuscar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	if esc.RestritoACarteiraDe() == nil {
		t.Fatal("pré-condição: o escopo original deveria ter carteira")
	}

	semCarteira := esc.SemCarteira()
	if semCarteira.RestritoACarteiraDe() != nil {
		t.Fatal("SemCarteira() deveria remover a restrição de carteira")
	}
	if semCarteira.InstituicaoID() == nil || *semCarteira.InstituicaoID() != instituicaoID {
		t.Fatal("SemCarteira() nunca deveria alterar a instituição")
	}
	// O escopo original não é mutado — Value Object imutável.
	if esc.RestritoACarteiraDe() == nil {
		t.Fatal("o escopo ORIGINAL não deveria ser afetado por SemCarteira()")
	}
}
