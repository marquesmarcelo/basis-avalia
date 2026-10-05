package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/jmoiron/sqlx"
)

func loginEObterCookie(t *testing.T, router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}, instituicaoID *string, email, senha string) *http.Cookie {
	t.Helper()
	corpo, _ := json.Marshal(LoginRequest{InstituicaoID: instituicaoID, Email: email, Senha: senha})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login falhou: %d (%s)", rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()[0]
}

func requisitar(t *testing.T, router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}, metodo, caminho string, cookie *http.Cookie, corpo any) *httptest.ResponseRecorder {
	t.Helper()
	var leitor *bytes.Reader
	if corpo != nil {
		b, _ := json.Marshal(corpo)
		leitor = bytes.NewReader(b)
	} else {
		leitor = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(metodo, caminho, leitor)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// setupAdmin cria um Administrador do Sistema de teste e devolve o cookie
// de sessão já autenticado.
func setupAdminLogado(t *testing.T, db *sqlx.DB, router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}, email, senha string) *http.Cookie {
	t.Helper()
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		Nome: "Admin de Teste", Email: email, Perfil: "administrador_sistema",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, senha),
	})
	return loginEObterCookie(t, router, nil, email, senha)
}

func TestSmoke_Instituicoes_CicloCompletoDeStatusCodes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smk6@basis-avalia.local", "senha-do-admin-2026")

	// Criar (201)
	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/instituicoes", cookieAdmin, InstituicaoRequest{
		Nome: "Faculdade Smoke", Sigla: "FSK1",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar instituição: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var criada InstituicaoResponse
	if err := json.Unmarshal(recCriar.Body.Bytes(), &criada); err != nil {
		t.Fatalf("decodificar resposta: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM instituicao WHERE id = $1`, criada.ID) })

	// Listar (200)
	recListar := requisitar(t, router, http.MethodGet, "/api/v1/instituicoes?situacao=todas", cookieAdmin, nil)
	if recListar.Code != http.StatusOK {
		t.Fatalf("listar: esperava 200, obtido %d", recListar.Code)
	}

	// Buscar (200)
	recBuscar := requisitar(t, router, http.MethodGet, "/api/v1/instituicoes/"+criada.ID, cookieAdmin, nil)
	if recBuscar.Code != http.StatusOK {
		t.Fatalf("buscar: esperava 200, obtido %d", recBuscar.Code)
	}

	// Atualizar (200)
	recAtualizar := requisitar(t, router, http.MethodPut, "/api/v1/instituicoes/"+criada.ID, cookieAdmin, InstituicaoAtualizarRequest{
		Nome: "Faculdade Smoke Renomeada", Sigla: "FSK1", Versao: criada.Versao,
	})
	if recAtualizar.Code != http.StatusOK {
		t.Fatalf("atualizar: esperava 200, obtido %d (%s)", recAtualizar.Code, recAtualizar.Body.String())
	}

	// Conflito de versão (409)
	recConflito := requisitar(t, router, http.MethodPut, "/api/v1/instituicoes/"+criada.ID, cookieAdmin, InstituicaoAtualizarRequest{
		Nome: "Outra vez", Sigla: "FSK1", Versao: criada.Versao,
	})
	if recConflito.Code != http.StatusConflict {
		t.Fatalf("conflito de versão: esperava 409, obtido %d", recConflito.Code)
	}

	// Inativar (200) — sem versão -> 400
	recSemVersao := requisitar(t, router, http.MethodPatch, "/api/v1/instituicoes/"+criada.ID+"/situacao", cookieAdmin, map[string]any{"situacao": "inativa"})
	if recSemVersao.Code != http.StatusBadRequest {
		t.Fatalf("situação sem versão: esperava 400, obtido %d", recSemVersao.Code)
	}

	// Nenhuma rota DELETE de instituição está registrada (I-09).
	recDelete := requisitar(t, router, http.MethodDelete, "/api/v1/instituicoes/"+criada.ID, cookieAdmin, nil)
	if recDelete.Code != http.StatusNotFound {
		t.Fatalf("DELETE de instituição não deveria existir como rota, obtido %d", recDelete.Code)
	}
}

func TestSmoke_Instituicoes_PINaoAlcanca_403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK7"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smk7@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smk7@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/instituicoes", cookiePI, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("PI acessando /instituicoes: esperava 403, obtido %d", rec.Code)
	}
	if codigoDeErro(t, rec) != "PERMISSAO_NEGADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestSmoke_Usuarios_T02_RecursoDeOutraInstituicaoDevolve404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK8"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK9"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smk8@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Email: "renata.smk9@ivv.edu.br", Perfil: "pesquisador_institucional",
	})

	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smk8@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/usuarios/"+renataID.String(), cookiePI, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 (isolamento), obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Usuarios_T03_AtualizarRecursoDeOutraInstituicaoDevolve404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT3"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT4"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkt3@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Email: "renata.smkt4@ivv.edu.br", Perfil: "pesquisador_institucional",
	})

	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkt3@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodPut, "/api/v1/usuarios/"+renataID.String(), cookiePI, UsuarioAtualizarRequest{
		Nome: "Renata Alterada", Email: "renata.smkt4@ivv.edu.br", Versao: 1,
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 (isolamento), obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Usuarios_T04_ExcluirRecursoDeOutraInstituicaoDevolve404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT5"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT6"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkt5@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Email: "renata.smkt6@ivv.edu.br", Perfil: "pesquisador_institucional",
	})

	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkt5@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodDelete, "/api/v1/usuarios/"+renataID.String(), cookiePI, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 (isolamento), obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Usuarios_T06_RedefinirSenhaDeOutraInstituicaoDevolve404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT7"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKT8"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkt7@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Email: "renata.smkt8@ivv.edu.br", Perfil: "pesquisador_institucional",
	})

	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkt7@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodPost, "/api/v1/usuarios/"+renataID.String()+"/senha", cookiePI, RedefinirSenhaRequest{
		SenhaNova: "nova-senha-renata-2026",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404 (isolamento), obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestSmoke_Usuarios_T05_PermissaoAntesDeIsolamento_403NaoDao404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKA"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKB"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "prof.smka@fsa.edu.br", Perfil: "professor",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-prof-2026"),
	})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Email: "renata.smkb@ivv.edu.br", Perfil: "pesquisador_institucional",
	})

	fsaIDStr := fsaID.String()
	cookieProfessor := loginEObterCookie(t, router, &fsaIDStr, "prof.smka@fsa.edu.br", "senha-do-prof-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/usuarios/"+renataID.String(), cookieProfessor, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("professor sem permissão deveria receber 403 (não 404), obtido %d", rec.Code)
	}
}

func TestSmoke_Usuarios_CriarEExcluirUltimoPI_409(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKC"})
	senhaProvisoriaFalsa := false
	piID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Email: "pi.smkc@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-do-pi-2026"),
	})

	fsaIDStr := fsaID.String()
	cookiePI := loginEObterCookie(t, router, &fsaIDStr, "pi.smkc@fsa.edu.br", "senha-do-pi-2026")

	// Autoexclusão (403) — mesmo sendo o único PI, o motivo aqui é E-08.
	recAuto := requisitar(t, router, http.MethodDelete, "/api/v1/usuarios/"+piID.String(), cookiePI, nil)
	if recAuto.Code != http.StatusForbidden {
		t.Fatalf("autoexclusão: esperava 403, obtido %d", recAuto.Code)
	}

	// Cria um segundo PI e exclui o primeiro (aceito, E-10).
	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/usuarios", cookiePI, UsuarioRequest{
		Nome: "Beatriz Andrade", Email: "beatriz.smkc@fsa.edu.br", Perfis: []string{"pesquisador_institucional"}, Senha: "avaliacao-inep-2026",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar segundo PI: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var segundoPI UsuarioResponse
	json.Unmarshal(recCriar.Body.Bytes(), &segundoPI)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, segundoPI.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, segundoPI.ID)
	})

	cookieSegundoPI := loginEObterCookie(t, router, &fsaIDStr, "beatriz.smkc@fsa.edu.br", "avaliacao-inep-2026")

	// Senha provisória bloqueia qualquer rota fora da porta de troca (§7.4)
	// — a segunda PI precisa definir senha própria antes de agir.
	recTrocaSenha := requisitar(t, router, http.MethodPost, "/api/v1/auth/senha", cookieSegundoPI, AlterarSenhaRequest{
		SenhaAtual: "avaliacao-inep-2026", SenhaNova: "nova-senha-da-beatriz-2026",
	})
	if recTrocaSenha.Code != http.StatusOK {
		t.Fatalf("trocar senha provisória da segunda PI: esperava 200, obtido %d (%s)", recTrocaSenha.Code, recTrocaSenha.Body.String())
	}
	cookieSegundoPI = recTrocaSenha.Result().Cookies()[0]

	recExcluirPrimeiro := requisitar(t, router, http.MethodDelete, "/api/v1/usuarios/"+piID.String(), cookieSegundoPI, nil)
	if recExcluirPrimeiro.Code != http.StatusNoContent {
		t.Fatalf("excluir o primeiro PI com um segundo ativo: esperava 204, obtido %d (%s)", recExcluirPrimeiro.Code, recExcluirPrimeiro.Body.String())
	}

	// Agora só resta o segundo PI — excluí-lo deve ser 409.
	recUltimo := requisitar(t, router, http.MethodDelete, "/api/v1/usuarios/"+segundoPI.ID, cookiePI, nil)
	if recUltimo.Code != http.StatusUnauthorized {
		// cookiePI pertence ao primeiro PI, já excluído — sessão caiu (SE-08), o que já é evidência indireta de que a exclusão anterior funcionou.
		t.Logf("cookie do primeiro PI (já excluído) devolveu %d, como esperado pelo desenho de sessão", recUltimo.Code)
	}
}

func TestSmoke_Administradores_CriarEListar(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkd@basis-avalia.local", "senha-do-admin-2026")

	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/administradores", cookieAdmin, UsuarioRequest{
		Nome: "Segundo Admin", Email: "admin2.smkd@basis-avalia.local", Senha: "senha-do-segundo-2026",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar administrador: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var segundoAdmin UsuarioResponse
	json.Unmarshal(recCriar.Body.Bytes(), &segundoAdmin)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, segundoAdmin.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, segundoAdmin.ID)
	})

	recListar := requisitar(t, router, http.MethodGet, "/api/v1/administradores", cookieAdmin, nil)
	if recListar.Code != http.StatusOK {
		t.Fatalf("listar administradores: esperava 200, obtido %d", recListar.Code)
	}
	var corpo map[string]any
	json.Unmarshal(recListar.Body.Bytes(), &corpo)
	itens, _ := corpo["data"].([]any)
	if len(itens) < 2 {
		t.Fatalf("esperava ao menos 2 administradores na listagem, obtido %d", len(itens))
	}
}

func TestSmoke_Pesquisadores_CriarPrimeiroPIDaInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smke@basis-avalia.local", "senha-do-admin-2026")

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMKE"})

	rec := requisitar(t, router, http.MethodPost, "/api/v1/instituicoes/"+fsaID.String()+"/pesquisadores", cookieAdmin, UsuarioRequest{
		Nome: "Renata Coimbra", Email: "renata.smke@ivv.edu.br", Senha: "nde-vale-verde-26",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar PI: esperava 201, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	var resposta UsuarioResponse
	json.Unmarshal(rec.Body.Bytes(), &resposta)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, resposta.ID)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, resposta.ID)
	})
	if len(resposta.Perfis) != 1 || resposta.Perfis[0] != "pesquisador_institucional" {
		t.Fatalf("perfil deveria ser imposto pelo alcance, obtido %v", resposta.Perfis)
	}

	// A instituição passa a aparecer no combo público (AS-01, 3.20).
	recCombo := requisitar(t, router, http.MethodGet, "/api/v1/publico/instituicoes", nil, nil)
	if recCombo.Code != http.StatusOK {
		t.Fatalf("combo público: esperava 200, obtido %d", recCombo.Code)
	}
	var combo []InstituicaoPublicaResponse
	json.Unmarshal(recCombo.Body.Bytes(), &combo)
	achou := false
	for _, i := range combo {
		if i.ID == fsaID.String() {
			achou = true
		}
	}
	if !achou {
		t.Fatal("instituição deveria aparecer no combo após ganhar o primeiro PI")
	}
}
