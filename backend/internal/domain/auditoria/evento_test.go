package auditoria

import "testing"

// TestAcao_VaiParaSyslog_BateComATabelaDaSpec cobre a coluna "Syslog" da
// §11 da spec para o caso incondicional (o condicional — atualizar_usuario
// só quando o perfil muda — é decidido no adapter Composto).
func TestAcao_VaiParaSyslog_BateComATabelaDaSpec(t *testing.T) {
	casos := map[Acao]bool{
		AlterarSenhaPropria:   true,
		RedefinirSenhaUsuario: true,
		CriarUsuario:          false,
		AtualizarUsuario:      false,
		ExcluirUsuario:        true,
		CriarInstituicao:      true,
		AtualizarInstituicao:  false,
		InativarInstituicao:   true,
		ReativarInstituicao:   true,
		AcessoNegado:          true,
	}
	for acao, esperado := range casos {
		if obtido := acao.VaiParaSyslog(); obtido != esperado {
			t.Errorf("%s: esperado %v, obtido %v", acao, esperado, obtido)
		}
	}
}
