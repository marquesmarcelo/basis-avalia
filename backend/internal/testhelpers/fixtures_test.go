package testhelpers

import "testing"

// TestFixtures_T012_CriarUsuarioLimpaSozinho prova que um teste que chame
// CriarUsuario e nada mais deixa o banco no estado anterior — o t.Cleanup
// do subteste roda ao final dele, sem nenhum defer db.Exec("DELETE...")
// a cargo de quem chama.
// A contagem é restrita à instituição criada neste teste (nunca a tabela
// inteira): go test ./... roda pacotes em paralelo, e outros pacotes
// inserem linhas em usuario/instituicao ao mesmo tempo contra o mesmo
// banco de teste compartilhado.
func TestFixtures_T012_CriarUsuarioLimpaSozinho(t *testing.T) {
	db := BancoDeTeste(t)
	instituicaoID := CriarInstituicao(t, db, OpcoesInstituicao{})

	antes := contarUsuariosDaInstituicao(t, db, instituicaoID)

	t.Run("subteste_isolado", func(t *testing.T) {
		CriarUsuario(t, db, OpcoesUsuario{InstituicaoID: &instituicaoID})
		durante := contarUsuariosDaInstituicao(t, db, instituicaoID)
		if durante != antes+1 {
			t.Fatalf("esperava %d linha(s) durante o subteste, obtido %d", antes+1, durante)
		}
	})

	depois := contarUsuariosDaInstituicao(t, db, instituicaoID)
	if depois != antes {
		t.Fatalf("cleanup não restaurou o estado: esperado %d, obtido %d", antes, depois)
	}
}

func TestFixtures_CriarUsuarioExcluido_NasceComHashNulo(t *testing.T) {
	db := BancoDeTeste(t)
	instituicaoID := CriarInstituicao(t, db, OpcoesInstituicao{})
	id := CriarUsuarioExcluido(t, db, OpcoesUsuario{InstituicaoID: &instituicaoID})

	var excluidoEm *string
	var senhaHash *string
	if err := db.QueryRow(`SELECT excluido_em::text, senha_hash FROM usuario WHERE id = $1`, id).Scan(&excluidoEm, &senhaHash); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if excluidoEm == nil {
		t.Fatal("excluido_em deveria estar preenchido")
	}
	if senhaHash != nil {
		t.Fatal("senha_hash deveria estar nulo")
	}
}
