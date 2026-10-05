package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/adapter/jwt"
	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/gin-gonic/gin"
)

type relogioFixoHTTP struct{ agora time.Time }

func (r relogioFixoHTTP) Agora() time.Time { return r.agora }

func conjuntoUnicoTeste(t *testing.T, p valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(p)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	return c
}

func montarRoteador(t *testing.T, tokenDeSessao *jwt.TokenDeSessao, repo *postgres.AutenticacaoRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mw := NovoMiddlewareSessao(tokenDeSessao, repo, relogioFixoHTTP{agora: time.Now()}, time.UTC)

	protegido := func(c *gin.Context) {
		ator, _ := AtorDoContexto(c)
		perfis := ator.Perfis().Ordenado()
		c.JSON(http.StatusOK, gin.H{"perfis": perfis})
	}
	router.GET("/api/v1/auth/eu", mw, protegido)
	router.GET("/api/v1/usuarios", mw, protegido)
	return router
}

func corpoDeErro(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var corpo map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("resposta não é JSON: %v (%s)", err, rec.Body.String())
	}
	return corpo
}

func codigoDeErro(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	corpo := corpoDeErro(t, rec)
	erroObj, ok := corpo["error"].(map[string]any)
	if !ok {
		t.Fatalf("resposta sem envelope de erro: %s", rec.Body.String())
	}
	codigo, _ := erroObj["code"].(string)
	return codigo
}

func TestMiddlewareSessao_SemCookie_SessaoExpirada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: time.Now()})
	router := montarRoteador(t, ts, repo)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d", rec.Code)
	}
	if codigoDeErro(t, rec) != "SESSAO_EXPIRADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_ContaExcluidaDuranteASessao_CONTA_EXCLUIDA(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA1"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	ator, err := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &fsaID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	// Exclui a conta DEPOIS de emitir o token — simula o PI excluindo
	// alguém que está navegando (SE-02).
	if _, err := db.Exec(`UPDATE usuario SET excluido_em = now(), senha_hash = NULL WHERE id = $1`, usuarioID); err != nil {
		t.Fatalf("excluir usuário: %v", err)
	}

	router := montarRoteador(t, ts, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "CONTA_EXCLUIDA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_InstituicaoInativada_INSTITUICAO_INATIVA(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA2"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "pesquisador_institucional"})

	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.PesquisadorInstitucional), &fsaID)
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	if _, err := db.Exec(`UPDATE instituicao SET situacao = 'inativa' WHERE id = $1`, fsaID); err != nil {
		t.Fatalf("inativar instituição: %v", err)
	}

	router := montarRoteador(t, ts, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "INSTITUICAO_INATIVA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_SenhaRedefinida_SessaoExpirada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA3"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &fsaID)
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	// redefine a senha => sessoes_validas_a_partir_de é gravado no futuro
	if _, err := db.Exec(`UPDATE usuario SET sessoes_validas_a_partir_de = $1 WHERE id = $2`, agora.Add(time.Minute), usuarioID); err != nil {
		t.Fatalf("atualizar sessoes_validas_a_partir_de: %v", err)
	}

	router := montarRoteador(t, ts, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "SESSAO_EXPIRADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_InsDivergenteDoVinculo_SessaoExpirada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA4"})
	outraID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA5"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	// Token forjado/obsoleto: emitido com uma instituição diferente da do
	// vínculo atual no banco.
	atorComOutraInstituicao, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &outraID)
	token, _, err := ts.Emitir(atorComOutraInstituicao, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	router := montarRoteador(t, ts, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "SESSAO_EXPIRADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_TokenComMaisDe8Horas_SessaoExpirada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	emissao := time.Now().Add(-9 * time.Hour)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA6"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	tsNaEmissao := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: emissao})
	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &fsaID)
	token, _, err := tsNaEmissao.Emitir(ator, emissao)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	tsAgora := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: time.Now()})
	router := montarRoteador(t, tsAgora, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "SESSAO_EXPIRADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestMiddlewareSessao_SE03_MudancaDePerfilNaoEncerraSessao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA7"})
	// pesquisador_institucional — não coordenador_curso: desde DC-4
	// (specs/cursos/design.md C-09), esse perfil nunca é atribuído
	// diretamente, é sempre derivado de designação. Qualquer perfil
	// atribuível serve para provar o mecanismo desta troca ao vivo.
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "pesquisador_institucional"})

	// Capturado depois de criar o usuário: sessoes_validas_a_partir_de
	// nasce em now() no INSERT, e emt precisa ser >= a esse instante.
	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})
	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.PesquisadorInstitucional), &fsaID)
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM usuario_perfil WHERE usuario_id = $1`, usuarioID); err != nil {
		t.Fatalf("retirar perfil anterior: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,'professor',$2)`, usuarioID, fsaID); err != nil {
		t.Fatalf("mudar perfil: %v", err)
	}

	router := montarRoteador(t, ts, repo)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("mudança de perfil não deveria encerrar a sessão: esperava 200, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	corpo := corpoDeErro(t, rec)
	perfis, _ := corpo["perfis"].([]any)
	if len(perfis) != 1 || perfis[0] != "professor" {
		t.Fatalf("o perfil novo deveria valer na requisição seguinte, obtido %v", corpo["perfis"])
	}
}

func TestMiddlewareSessao_SE08_SemTokenParaUsuarioExcluido_NuncaContaExcluida(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: time.Now()})
	router := montarRoteador(t, ts, repo)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	router.ServeHTTP(rec, req)

	if codigoDeErro(t, rec) != "SESSAO_EXPIRADA" {
		t.Fatalf("sem token nunca deveria produzir CONTA_EXCLUIDA nem INSTITUICAO_INATIVA, obtido %s", rec.Body.String())
	}
}

// relogioContador conta quantas vezes Agora() é chamado — usado só para
// provar T-111 (fundacao-metas.md §5.2): exatamente UMA leitura de
// relógio por requisição no caminho de sessão, nunca duas independentes
// que poderiam divergir perto da meia-noite.
type relogioContador struct {
	agora    time.Time
	chamadas *int
}

func (r relogioContador) Agora() time.Time {
	*r.chamadas++
	return r.agora
}

func TestMiddlewareSessao_T111_UmaSoLeituraDeRelogioPorRequisicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA9"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	agora := time.Now()
	tsRelogio := relogioFixoHTTP{agora: agora}
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", tsRelogio)
	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &fsaID)
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	var chamadas int
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mw := NovoMiddlewareSessao(ts, repo, relogioContador{agora: agora, chamadas: &chamadas}, time.UTC)
	router.GET("/api/v1/auth/eu", mw, func(c *gin.Context) { c.Status(http.StatusOK) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	req.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if chamadas != 1 {
		t.Fatalf("esperava exatamente 1 chamada a Relogio.Agora() no caminho de sessão, obtido %d", chamadas)
	}
}

func TestMiddlewareSessao_SenhaProvisoria_RotaForaDaListaDevolve403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := postgres.NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "MSA8"})
	senhaProvisoria := true
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor", SenhaProvisoria: &senhaProvisoria})

	agora := time.Now()
	ts := jwt.NovoTokenDeSessao("segredo-de-teste-http", relogioFixoHTTP{agora: agora})
	ator, _ := autorizacao.NovoAtor(usuarioID, conjuntoUnicoTeste(t, valueobject.Professor), &fsaID)
	token, _, err := ts.Emitir(ator, agora)
	if err != nil {
		t.Fatalf("Emitir: %v", err)
	}

	router := montarRoteador(t, ts, repo)

	// Rota fora da lista permitida — deve responder 403.
	recForaDaLista := httptest.NewRecorder()
	reqForaDaLista := httptest.NewRequest(http.MethodGet, "/api/v1/usuarios", nil)
	reqForaDaLista.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(recForaDaLista, reqForaDaLista)
	if recForaDaLista.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obtido %d: %s", recForaDaLista.Code, recForaDaLista.Body.String())
	}
	if codigoDeErro(t, recForaDaLista) != "SENHA_PROVISORIA" {
		t.Fatalf("código incorreto: %s", recForaDaLista.Body.String())
	}

	// Rota permitida (/auth/eu) — deve funcionar normalmente.
	recPermitida := httptest.NewRecorder()
	reqPermitida := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	reqPermitida.AddCookie(&http.Cookie{Name: NomeCookieSessao, Value: token})
	router.ServeHTTP(recPermitida, reqPermitida)
	if recPermitida.Code != http.StatusOK {
		t.Fatalf("rota permitida com senha provisória deveria funcionar: %d (%s)", recPermitida.Code, recPermitida.Body.String())
	}

}
