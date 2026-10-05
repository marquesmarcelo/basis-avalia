package instituicao

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

func TestNovaInstituicao_I01_NasceAtiva(t *testing.T) {
	sigla, _ := valueobject.NovaSigla("FSA")
	codigo, _ := valueobject.NovoCodigoEMec("12345")
	inst, err := NovaInstituicao("Faculdade Serra Azul", sigla, codigo)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if inst.Situacao != valueobject.Ativa {
		t.Errorf("instituicao deveria nascer ativa, nasceu %s", inst.Situacao)
	}
	if inst.Versao != 1 {
		t.Errorf("versao esperada 1, obtida %d", inst.Versao)
	}
	if inst.ExcluidoEm != nil {
		t.Error("excluido_em deveria nascer nulo")
	}
}
