package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
)

func TestSmoke_R5_CookieNovoFuncionaECookieAntigoExpira(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SEN1"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "João Ribeiro", Email: "joao.senha1@fsa.edu.br",
		Perfil: "professor", SenhaProvisoria: &senhaProvisoriaFalsa,
		SenhaHash: hashParaTeste(t, "senha-fsa-2026"),
	})

	fsaIDStr := fsaID.String()
	recLogin := fazerLogin(t, router, &fsaIDStr, "joao.senha1@fsa.edu.br", "senha-fsa-2026")
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login deveria funcionar: %d (%s)", recLogin.Code, recLogin.Body.String())
	}
	cookieAntigo := recLogin.Result().Cookies()[0]

	corpoSenha, _ := json.Marshal(map[string]string{"senha_atual": "senha-fsa-2026", "senha_nova": "colegiado-nova-2026"})
	reqSenha := httptest.NewRequest(http.MethodPost, "/api/v1/auth/senha", strings.NewReader(string(corpoSenha)))
	reqSenha.Header.Set("Content-Type", "application/json")
	reqSenha.AddCookie(cookieAntigo)
	recSenha := httptest.NewRecorder()
	router.ServeHTTP(recSenha, reqSenha)
	if recSenha.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtido %d (%s)", recSenha.Code, recSenha.Body.String())
	}
	cookieNovo := recSenha.Result().Cookies()[0]
	if cookieNovo.Value == cookieAntigo.Value {
		t.Fatal("o cookie deveria mudar após a troca de senha")
	}

	reqComNovo := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	reqComNovo.AddCookie(cookieNovo)
	recComNovo := httptest.NewRecorder()
	router.ServeHTTP(recComNovo, reqComNovo)
	if recComNovo.Code != http.StatusOK {
		t.Fatalf("cookie novo deveria funcionar: %d (%s)", recComNovo.Code, recComNovo.Body.String())
	}

	reqComAntigo := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	reqComAntigo.AddCookie(cookieAntigo)
	recComAntigo := httptest.NewRecorder()
	router.ServeHTTP(recComAntigo, reqComAntigo)
	if recComAntigo.Code != http.StatusUnauthorized {
		t.Fatalf("cookie antigo deveria devolver 401 SESSAO_EXPIRADA, obtido %d", recComAntigo.Code)
	}
	if codigoDeErro(t, recComAntigo) != "SESSAO_EXPIRADA" {
		t.Fatalf("código incorreto: %s", recComAntigo.Body.String())
	}
}

func TestSmoke_S09_TrocarSenhaELogarComSenhaExataDe100Caracteres(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SEN2"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Ávila Gomes", Email: "avila.senha2@fsa.edu.br",
		Perfil: "professor", SenhaProvisoria: &senhaProvisoriaFalsa,
		SenhaHash: hashParaTeste(t, "senha-inicial-avila"),
	})

	senha100 := strings.Repeat("á é í ó ú ç ão ", 6) + "final-com-acento"
	if len([]rune(senha100)) < 100 {
		t.Fatalf("massa de teste precisa ter ao menos 100 runas, tem %d", len([]rune(senha100)))
	}

	fsaIDStr := fsaID.String()
	recLogin := fazerLogin(t, router, &fsaIDStr, "avila.senha2@fsa.edu.br", "senha-inicial-avila")
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login deveria funcionar: %d (%s)", recLogin.Code, recLogin.Body.String())
	}
	cookie := recLogin.Result().Cookies()[0]

	corpoSenha, _ := json.Marshal(map[string]string{"senha_atual": "senha-inicial-avila", "senha_nova": senha100})
	reqSenha := httptest.NewRequest(http.MethodPost, "/api/v1/auth/senha", strings.NewReader(string(corpoSenha)))
	reqSenha.Header.Set("Content-Type", "application/json")
	reqSenha.AddCookie(cookie)
	recSenha := httptest.NewRecorder()
	router.ServeHTTP(recSenha, reqSenha)
	if recSenha.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtido %d (%s)", recSenha.Code, recSenha.Body.String())
	}

	recLoginNovo := fazerLogin(t, router, &fsaIDStr, "avila.senha2@fsa.edu.br", senha100)
	if recLoginNovo.Code != http.StatusOK {
		t.Fatalf("login com a senha nova de 100 caracteres deveria funcionar: %d (%s)", recLoginNovo.Code, recLoginNovo.Body.String())
	}
}
