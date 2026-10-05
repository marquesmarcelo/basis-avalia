package usuario

import (
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func novoUsuarioTeste(t *testing.T) *Usuario {
	t.Helper()
	email, err := valueobject.NovoEmail("joao.ribeiro@ies.edu.br")
	if err != nil {
		t.Fatalf("email: %v", err)
	}
	hash, err := valueobject.NovaSenhaHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	instituicaoID := uuid.Must(uuid.NewV7())
	perfis, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	u, err := NovoUsuario("João Ribeiro", email, hash, perfis, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoUsuario: %v", err)
	}
	return u
}

func TestNovoUsuario_U01_NasceComInvariantes(t *testing.T) {
	u := novoUsuarioTeste(t)
	if u.Versao != 1 {
		t.Errorf("versao esperada 1, obtida %d", u.Versao)
	}
	if !u.SenhaProvisoria {
		t.Error("senha_provisoria deveria nascer true")
	}
	if u.ExcluidoEm != nil {
		t.Error("excluido_em deveria nascer nulo")
	}
	if u.ProvedorIdentidade != valueobject.CredencialLocal {
		t.Errorf("provedor_identidade esperado credencial_local, obtido %s", u.ProvedorIdentidade)
	}
	if u.IdentificadorExterno != nil {
		t.Error("identificador_externo deveria nascer nulo")
	}
	if !u.SessoesValidasAPartirDe.Equal(u.CriadoEm) {
		t.Error("sessoes_validas_a_partir_de deveria nascer igual a criado_em")
	}
}

func TestUsuario_E05_ExcluirLogicamenteAnulaOHash(t *testing.T) {
	u := novoUsuarioTeste(t)
	instante := time.Now()
	u.ExcluirLogicamente(instante)

	if u.ExcluidoEm == nil || !u.ExcluidoEm.Equal(instante) {
		t.Error("excluido_em deveria ser preenchido com o instante")
	}
	if u.SenhaHash != nil {
		t.Error("senha_hash deveria ser anulado na exclusão lógica (3.8)")
	}
}
