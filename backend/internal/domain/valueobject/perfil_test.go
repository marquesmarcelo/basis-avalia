package valueobject

import "testing"

// TestPerfil_A01aA06_MatrizDePermissoes cobre os cinco perfis × as 11
// permissões da matriz 3.9 da spec (design.md §13.2).
func TestPerfil_A01aA06_MatrizDePermissoes(t *testing.T) {
	todasPermissoes := []Permissao{
		InstituicaoListar, InstituicaoCriar, InstituicaoEditar, InstituicaoInativar,
		PIGerenciar,
		UsuarioListar, UsuarioCriar, UsuarioEditar, UsuarioExcluir, UsuarioRedefinirSenha,
		AdministradorGerenciar,
	}
	esperado := map[Perfil]map[Permissao]bool{
		AdministradorSistema: {
			InstituicaoListar: true, InstituicaoCriar: true, InstituicaoEditar: true,
			InstituicaoInativar: true, PIGerenciar: true, AdministradorGerenciar: true,
		},
		PesquisadorInstitucional: {
			UsuarioListar: true, UsuarioCriar: true, UsuarioEditar: true,
			UsuarioExcluir: true, UsuarioRedefinirSenha: true,
		},
		CoordenadorCurso: {},
		Professor:        {},
		Aluno:            {},
	}

	for perfil, mapa := range esperado {
		for _, permissao := range todasPermissoes {
			quer := mapa[permissao]
			obtido := perfil.Pode(permissao)
			if obtido != quer {
				t.Errorf("perfil %s, permissão %s: esperado %v, obtido %v", perfil, permissao, quer, obtido)
			}
		}
	}
}

func TestPerfil_U10_PINaoAtribuiAdministradorSistema(t *testing.T) {
	if PesquisadorInstitucional.PodeAtribuir(AdministradorSistema) {
		t.Fatal("PI não pode atribuir administrador_sistema")
	}
}

func TestPerfil_36_AdministradorNaoAtribuiPerfisInstitucionaisComuns(t *testing.T) {
	for _, p := range []Perfil{CoordenadorCurso, Professor, Aluno} {
		if AdministradorSistema.PodeAtribuir(p) {
			t.Errorf("administrador não deveria atribuir %s", p)
		}
	}
}

func TestPerfil_AdministradorPodeAtribuirOutroAdministrador(t *testing.T) {
	if !AdministradorSistema.PodeAtribuir(AdministradorSistema) {
		t.Fatal("administrador deveria poder atribuir administrador_sistema (P17)")
	}
	if !AdministradorSistema.PodeAtribuir(PesquisadorInstitucional) {
		t.Fatal("administrador deveria poder atribuir pesquisador_institucional")
	}
}

func TestPerfil_PIPodeAtribuirPerfisInstitucionais(t *testing.T) {
	for _, p := range []Perfil{PesquisadorInstitucional, Professor, Aluno} {
		if !PesquisadorInstitucional.PodeAtribuir(p) {
			t.Errorf("PI deveria poder atribuir %s", p)
		}
	}
}

// TestPerfil_DC4_PINaoAtribuiCoordenadorCurso prova specs/cursos/design.md
// C-09: coordenador_curso saiu da lista do PI — é sempre derivado de
// designação, nunca atribuído (nem pelo PI, que antes podia).
func TestPerfil_DC4_PINaoAtribuiCoordenadorCurso(t *testing.T) {
	if PesquisadorInstitucional.PodeAtribuir(CoordenadorCurso) {
		t.Fatal("PI não deveria mais poder atribuir coordenador_curso (DC-4)")
	}
}
