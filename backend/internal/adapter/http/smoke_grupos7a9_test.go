package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
)

func TestSmoke_Usuarios_A04_ProfessorSemPermissaoDeListarRecebe403ComAuditoria(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKF"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "prof.smkf@fsa.edu.br", Perfil: "professor",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-prof-2026"),
	})
	fsaIDStr := fsaID.String()
	cookieProfessor := loginEObterCookie(t, router, &fsaIDStr, "prof.smkf@fsa.edu.br", "senha-do-prof-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/usuarios", cookieProfessor, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("professor listando usuários: esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "PERMISSAO_NEGADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM auditoria WHERE acao = 'acesso_negado' AND detalhes->>'permissao_exigida' = 'usuario.listar'`); err != nil {
		t.Fatalf("consultar auditoria: %v", err)
	}
	if total < 1 {
		t.Fatal("esperava um evento de auditoria acesso_negado para a tentativa de listar")
	}
}

func TestSmoke_Usuarios_A06_AlunoCriandoUsuarioRecebe403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKG"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "aluno.smkg@fsa.edu.br", Perfil: "aluno",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-aluno-2026"),
	})
	fsaIDStr := fsaID.String()
	cookieAluno := loginEObterCookie(t, router, &fsaIDStr, "aluno.smkg@fsa.edu.br", "senha-do-aluno-2026")

	rec := requisitar(t, router, http.MethodPost, "/api/v1/usuarios", cookieAluno, UsuarioRequest{
		Nome: "X", Email: "x.smkg@fsa.edu.br", Senha: "qualquer-senha-26",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("aluno criando usuário: esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Usuarios_AS03_AdministradorChamandoUsuariosRecebe403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkh@basis-avalia.local", "senha-do-admin-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/usuarios", cookieAdmin, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("administrador chamando /usuarios: esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Administradores_AS10_PINaoAlcancaNenhumaRota(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKI"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smki@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smki@fsa.edu.br", "senha-do-pi-2026")

	recListar := requisitar(t, router, http.MethodGet, "/api/v1/administradores", cookiePI, nil)
	if recListar.Code != http.StatusForbidden {
		t.Fatalf("PI listando administradores: esperava 403, obtido %d (%s)", recListar.Code, recListar.Body.String())
	}
	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/administradores", cookiePI, UsuarioRequest{
		Nome: "X", Email: "x.smki@basis-avalia.local", Senha: "qualquer-senha-26",
	})
	if recCriar.Code != http.StatusForbidden {
		t.Fatalf("PI criando administrador: esperava 403, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
}

func TestSmoke_Usuarios_G07_OrdenarPorPerfilDevolve400(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKJ"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkj@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkj@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/usuarios?sort=perfil", cookiePI, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("sort=perfil: esperava 400, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Pesquisadores_R22_RedefinirSenhaUsaVersaoEscopada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkk@basis-avalia.local", "senha-do-admin-2026")

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKK"})
	piID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkk@fsa.edu.br", Perfil: "pesquisador_institucional",
	})

	rec := requisitar(t, router, http.MethodPost, "/api/v1/instituicoes/"+fsaID.String()+"/pesquisadores/"+piID.String()+"/senha", cookieAdmin, RedefinirSenhaRequest{
		SenhaNova: "nova-senha-provisoria-26",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("redefinir senha do PI: esperava 204, obtido %d (%s)", rec.Code, rec.Body.String())
	}

	var senhaProvisoria bool
	if err := db.Get(&senhaProvisoria, `SELECT senha_provisoria FROM usuario WHERE id = $1`, piID); err != nil {
		t.Fatalf("consultar usuário: %v", err)
	}
	if !senhaProvisoria {
		t.Fatal("senha redefinida deveria voltar ao estado provisório (E-11)")
	}
}

func TestSmoke_Administradores_AS05_ListaTrazSoAdministradores(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkm@basis-avalia.local", "senha-do-admin-2026")

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKM"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkm@fsa.edu.br", Perfil: "pesquisador_institucional",
	})

	rec := requisitar(t, router, http.MethodGet, "/api/v1/administradores", cookieAdmin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar administradores: esperava 200, obtido %d", rec.Code)
	}
	var corpo map[string]any
	json.Unmarshal(rec.Body.Bytes(), &corpo)
	itens, _ := corpo["data"].([]any)
	for _, bruto := range itens {
		item, _ := bruto.(map[string]any)
		if item["email"] == "pi.smkm@fsa.edu.br" {
			t.Fatal("a listagem de administradores não deveria trazer um Pesquisador Institucional")
		}
	}
}

func TestSmoke_Administradores_PerfisNoCorpoEhIgnorado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smko@basis-avalia.local", "senha-do-admin-2026")

	rec := requisitar(t, router, http.MethodPost, "/api/v1/administradores", cookieAdmin, UsuarioRequest{
		Nome: "Terceiro", Email: "terceiro.smko@basis-avalia.local", Senha: "senha-do-terceiro-2026",
		Perfis: []string{"aluno", "professor"},
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar administrador: esperava 201, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	var resposta UsuarioResponse
	json.Unmarshal(rec.Body.Bytes(), &resposta)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, resposta.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, resposta.ID)
	})
	if len(resposta.Perfis) != 1 || resposta.Perfis[0] != "administrador_sistema" {
		t.Fatalf("perfis do corpo deveriam ser ignorados, esperava {administrador_sistema}, obtido %v", resposta.Perfis)
	}
}

func TestSmoke_Administradores_ArmadilhaEmailDuplicadoRecusadaPeloBanco(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkn@basis-avalia.local", "senha-do-admin-2026")

	recPrimeiro := requisitar(t, router, http.MethodPost, "/api/v1/administradores", cookieAdmin, UsuarioRequest{
		Nome: "Primeiro", Email: "duplicado.smkn@basis-avalia.local", Senha: "senha-do-primeiro-2026",
	})
	if recPrimeiro.Code != http.StatusCreated {
		t.Fatalf("criar primeiro administrador: esperava 201, obtido %d (%s)", recPrimeiro.Code, recPrimeiro.Body.String())
	}
	var primeiro UsuarioResponse
	json.Unmarshal(recPrimeiro.Body.Bytes(), &primeiro)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, primeiro.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, primeiro.ID)
	})

	recSegundo := requisitar(t, router, http.MethodPost, "/api/v1/administradores", cookieAdmin, UsuarioRequest{
		Nome: "Segundo", Email: "duplicado.smkn@basis-avalia.local", Senha: "senha-do-segundo-2026",
	})
	if recSegundo.Code != http.StatusConflict {
		t.Fatalf("segundo administrador com o mesmo e-mail: esperava 409, obtido %d (%s)", recSegundo.Code, recSegundo.Body.String())
	}
	if codigoDeErro(t, recSegundo) != "EMAIL_DUPLICADO" {
		t.Fatalf("código incorreto: %s", recSegundo.Body.String())
	}
}

func TestSmoke_Usuarios_PostSemPerfisGravaAlunoEDevolveNaResposta(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKL"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkl@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkl@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodPost, "/api/v1/usuarios", cookiePI, UsuarioRequest{
		Nome: "Letícia Moraes", Email: "leticia.smkl@fsa.edu.br", Senha: "primeiro-acesso-26",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar sem perfis: esperava 201, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	var resposta UsuarioResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resposta); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, resposta.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, resposta.ID)
	})
	if len(resposta.Perfis) != 1 || resposta.Perfis[0] != "aluno" {
		t.Fatalf("esperava {aluno} (U-12, E-16, D-18), obtido %v", resposta.Perfis)
	}
}
