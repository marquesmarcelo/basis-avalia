package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func executarComErro(t *testing.T, err error) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MiddlewareErro())
	router.GET("/x", func(c *gin.Context) { c.Error(err) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	router.ServeHTTP(rec, req)

	var corpo map[string]any
	if e := json.Unmarshal(rec.Body.Bytes(), &corpo); e != nil {
		t.Fatalf("resposta não é JSON válido: %v (%s)", e, rec.Body.String())
	}
	return rec, corpo
}

func TestMiddlewareErro_MapaDeErrosDeDominio(t *testing.T) {
	casos := []struct {
		err    error
		status int
		codigo string
	}{
		{domain.ErrSenhaObrigatoria, http.StatusBadRequest, "SENHA_OBRIGATORIA"},
		{domain.ErrSenhaAcimaDoLimite, http.StatusBadRequest, "SENHA_ACIMA_DO_LIMITE"},
		{&domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."}, http.StatusBadRequest, "PAYLOAD_INVALIDO"},
		{&domain.ErrParametro{Nome: "sort"}, http.StatusBadRequest, "PARAMETRO_INVALIDO"},
		{domain.ErrCredenciaisInvalidas, http.StatusUnauthorized, "CREDENCIAIS_INVALIDAS"},
		{domain.ErrSenhaAtualIncorreta, http.StatusUnauthorized, "SENHA_ATUAL_INCORRETA"},
		{domain.ErrPermissaoNegada, http.StatusForbidden, "PERMISSAO_NEGADA"},
		{domain.ErrPerfilNaoAtribuivel, http.StatusForbidden, "PERFIL_NAO_ATRIBUIVEL"},
		{domain.ErrAutoExclusaoNegada, http.StatusForbidden, "AUTO_EXCLUSAO_NEGADA"},
		{domain.ErrAlteracaoDosPropriosPerfisNegada, http.StatusForbidden, "ALTERACAO_DOS_PROPRIOS_PERFIS_NEGADA"},
		{domain.ErrNaoEncontrado, http.StatusNotFound, "NAO_ENCONTRADO"},
		{domain.ErrConflitoDeVersao, http.StatusConflict, "CONFLITO_DE_VERSAO"},
		{domain.ErrEmailDuplicado, http.StatusConflict, "EMAIL_DUPLICADO"},
		{domain.ErrSiglaDuplicada, http.StatusConflict, "SIGLA_DUPLICADA"},
		{domain.ErrCodigoEMecDuplicado, http.StatusConflict, "CODIGO_EMEC_DUPLICADO"},
		{domain.ErrUltimoPesquisadorInstitucional, http.StatusConflict, "ULTIMO_PESQUISADOR_INSTITUCIONAL"},
		{domain.ErrUltimoAdministradorSistema, http.StatusConflict, "ULTIMO_ADMINISTRADOR_SISTEMA"},
		{domain.ErrEscopoInvalido, http.StatusInternalServerError, "ERRO_INTERNO"},
	}

	for _, c := range casos {
		rec, corpo := executarComErro(t, c.err)
		if rec.Code != c.status {
			t.Errorf("erro %v: status esperado %d, obtido %d", c.err, c.status, rec.Code)
		}
		erroObj := corpo["error"].(map[string]any)
		if erroObj["code"] != c.codigo {
			t.Errorf("erro %v: code esperado %s, obtido %v", c.err, c.codigo, erroObj["code"])
		}
	}
}

func TestMiddlewareErro_ErroDesconhecidoNaoVazaDetalhe(t *testing.T) {
	original := errors.New("pq: relation \"usuario_secreta\" does not exist at /internal/adapter/postgres/x.go:42")
	rec, corpo := executarComErro(t, original)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500, obtido %d", rec.Code)
	}
	erroObj := corpo["error"].(map[string]any)
	if erroObj["code"] != "ERRO_INTERNO" {
		t.Fatalf("esperado ERRO_INTERNO, obtido %v", erroObj["code"])
	}
	if strings.Contains(rec.Body.String(), "usuario_secreta") || strings.Contains(rec.Body.String(), "pq:") {
		t.Fatalf("resposta vazou detalhe interno: %s", rec.Body.String())
	}
}
