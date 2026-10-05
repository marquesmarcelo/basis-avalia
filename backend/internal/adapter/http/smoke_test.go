package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/adapter/argon2"
	adapterauditoria "github.com/basis-avalia/backend/internal/adapter/auditoria"
	"github.com/basis-avalia/backend/internal/adapter/jwt"
	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/adapter/relogio"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	cursocmd "github.com/basis-avalia/backend/internal/usecase/command/curso"
	designacaocmd "github.com/basis-avalia/backend/internal/usecase/command/designacao"
	indicadorcmd "github.com/basis-avalia/backend/internal/usecase/command/indicador"
	indicadorplataformacmd "github.com/basis-avalia/backend/internal/usecase/command/indicador_plataforma"
	instituicaocmd "github.com/basis-avalia/backend/internal/usecase/command/instituicao"
	metacmd "github.com/basis-avalia/backend/internal/usecase/command/meta"
	sessaocmd "github.com/basis-avalia/backend/internal/usecase/command/sessao"
	usuariocmd "github.com/basis-avalia/backend/internal/usecase/command/usuario"
	cursoquery "github.com/basis-avalia/backend/internal/usecase/query/curso"
	designacaoquery "github.com/basis-avalia/backend/internal/usecase/query/designacao"
	indicadorquery "github.com/basis-avalia/backend/internal/usecase/query/indicador"
	indicadorplataformaquery "github.com/basis-avalia/backend/internal/usecase/query/indicador_plataforma"
	instituicaoquery "github.com/basis-avalia/backend/internal/usecase/query/instituicao"
	metaquery "github.com/basis-avalia/backend/internal/usecase/query/meta"
	sessaoquery "github.com/basis-avalia/backend/internal/usecase/query/sessao"
	usuarioquery "github.com/basis-avalia/backend/internal/usecase/query/usuario"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// hashParaTeste gera um hash argon2id real a partir de uma senha em texto
// — os testes de smoke autenticam de verdade, sem atalho.
func hashParaTeste(t *testing.T, senhaPlana string) string {
	t.Helper()
	senha, err := valueobject.NovaSenhaEmTexto(senhaPlana)
	if err != nil {
		t.Fatalf("senha de teste inválida: %v", err)
	}
	hash, err := argon2.NovoHashDeSenha(4).Gerar(context.Background(), senha)
	if err != nil {
		t.Fatalf("gerar hash de teste: %v", err)
	}
	return hash.Codificado()
}

// montarAppDeTeste monta o mesmo grafo de dependências do cmd/api/main.go,
// contra o banco de teste — é o smoke de integração exigido pelo design.md
// §13.4, compensação à ausência de E2E (§15 da spec).
func montarAppDeTeste(t *testing.T, db *sqlx.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	relogioReal := relogio.Novo()
	fusoDeExibicao, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	autenticacaoRepo := postgres.NovoAutenticacaoRepository(db)
	usuarioRepo := postgres.NovoUsuarioRepository(db)
	instituicaoRepo := postgres.NovoInstituicaoRepository(db)
	instituicaoPublicaQuery := postgres.NovoInstituicaoPublicaQuery(db)
	indicadorPlataformaRepo := postgres.NovoIndicadorPlataformaRepository(db)
	indicadorRepo := postgres.NovoIndicadorRepository(db)
	metaRepo := postgres.NovoMetaRepository(db)
	cursoRepo := postgres.NovoCursoRepository(db)
	designacaoRepo := postgres.NovoDesignacaoRepository(db)
	periodoRepo := postgres.NovoPeriodoRepository(db)
	planoRepo := postgres.NovoPlanoRepository(db)
	hashDeSenha := argon2.NovoHashDeSenha(4)
	tokenDeSessao := jwt.NovoTokenDeSessao("segredo-de-teste-smoke", relogioReal)
	uow := postgres.NovaUnidadeDeTrabalho(db)

	syslogAdapter, _ := adapterauditoria.NovoSyslog("", "tcp", "basis-avalia-teste")
	auditoriaRepo := postgres.NovoAuditoriaRepository(db)
	auditLogger := adapterauditoria.NovoComposto(auditoriaRepo, syslogAdapter)

	autenticarUC := sessaocmd.NovoAutenticarUseCase(autenticacaoRepo, hashDeSenha, relogioReal, fusoDeExibicao)
	encerrarSessaoUC := sessaocmd.NovoEncerrarSessaoUseCase(autenticacaoRepo, relogioReal)
	alterarSenhaPropriaUC := sessaocmd.NovoAlterarSenhaPropriaUseCase(autenticacaoRepo, hashDeSenha, auditLogger, relogioReal, uow)
	obterContextoUC := sessaoquery.NovoObterContextoDeSessaoUseCase(autenticacaoRepo)
	listarInstituicoesPublicasUC := instituicaoquery.NovoListarInstituicoesPublicasUseCase(instituicaoPublicaQuery)

	criarInstituicaoUC := instituicaocmd.NovoCriarInstituicaoUseCase(instituicaoRepo, auditLogger, uow)
	atualizarInstituicaoUC := instituicaocmd.NovoAtualizarInstituicaoUseCase(instituicaoRepo, auditLogger, uow)
	alterarSituacaoInstituicaoUC := instituicaocmd.NovoAlterarSituacaoInstituicaoUseCase(instituicaoRepo, auditLogger, uow)
	listarInstituicoesUC := instituicaoquery.NovoListarInstituicoesUseCase(instituicaoRepo)
	buscarInstituicaoUC := instituicaoquery.NovoBuscarInstituicaoUseCase(instituicaoRepo)

	criarUsuarioUC := usuariocmd.NovoCriarUsuarioUseCase(usuarioRepo, hashDeSenha, auditLogger, uow)
	atualizarUsuarioUC := usuariocmd.NovoAtualizarUsuarioUseCase(usuarioRepo, auditLogger, uow)
	excluirUsuarioUC := usuariocmd.NovoExcluirUsuarioUseCase(usuarioRepo, auditLogger, uow)
	redefinirSenhaUsuarioUC := usuariocmd.NovoRedefinirSenhaUsuarioUseCase(usuarioRepo, hashDeSenha, auditLogger, uow)
	listarUsuariosUC := usuarioquery.NovoListarUsuariosUseCase(usuarioRepo)
	buscarUsuarioUC := usuarioquery.NovoBuscarUsuarioUseCase(usuarioRepo)

	criarIndicadorPlataformaUC := indicadorplataformacmd.NovoCriarIndicadorPlataformaUseCase(indicadorPlataformaRepo, auditLogger, uow)
	atualizarIndicadorPlataformaUC := indicadorplataformacmd.NovoAtualizarIndicadorPlataformaUseCase(indicadorPlataformaRepo, auditLogger, uow)
	alterarSituacaoIndicadorPlataformaUC := indicadorplataformacmd.NovoAlterarSituacaoIndicadorPlataformaUseCase(indicadorPlataformaRepo, auditLogger, uow)
	excluirIndicadorPlataformaUC := indicadorplataformacmd.NovoExcluirIndicadorPlataformaUseCase(indicadorPlataformaRepo, auditLogger, uow)
	listarIndicadoresPlataformaUC := indicadorplataformaquery.NovoListarIndicadoresPlataformaUseCase(indicadorPlataformaRepo)
	buscarIndicadorPlataformaUC := indicadorplataformaquery.NovoBuscarIndicadorPlataformaUseCase(indicadorPlataformaRepo)

	criarIndicadorUC := indicadorcmd.NovoCriarIndicadorUseCase(indicadorRepo, auditLogger, uow)
	atualizarIndicadorUC := indicadorcmd.NovoAtualizarIndicadorUseCase(indicadorRepo, auditLogger, uow)
	alterarSituacaoIndicadorUC := indicadorcmd.NovoAlterarSituacaoIndicadorUseCase(indicadorRepo, auditLogger, uow)
	excluirIndicadorUC := indicadorcmd.NovoExcluirIndicadorUseCase(indicadorRepo, auditLogger, uow)
	listarDoCatalogoUC := indicadorquery.NovoListarDoCatalogoUseCase(indicadorRepo)
	buscarIndicadorUC := indicadorquery.NovoBuscarIndicadorUseCase(indicadorRepo)
	sugerirIndicadorUC := indicadorquery.NovoSugerirIndicadorUseCase(indicadorRepo)

	criarMetaUC := metacmd.NovoCriarMetaUseCase(metaRepo, indicadorRepo, auditLogger, uow)
	atualizarMetaUC := metacmd.NovoAtualizarMetaUseCase(metaRepo, indicadorRepo, auditLogger, uow)
	alterarSituacaoMetaUC := metacmd.NovoAlterarSituacaoMetaUseCase(metaRepo, auditLogger, uow)
	excluirMetaUC := metacmd.NovoExcluirMetaUseCase(metaRepo, auditLogger, uow)
	listarMetasUC := metaquery.NovoListarMetasUseCase(metaRepo)
	buscarMetaUC := metaquery.NovoBuscarMetaUseCase(metaRepo)
	sugerirMetaUC := metaquery.NovoSugerirMetaUseCase(metaRepo)

	criarCursoUC := cursocmd.NovoCriarCursoUseCase(cursoRepo, auditLogger, uow)
	atualizarCursoUC := cursocmd.NovoAtualizarCursoUseCase(cursoRepo, auditLogger, uow)
	alterarSituacaoCursoUC := cursocmd.NovoAlterarSituacaoCursoUseCase(cursoRepo, auditLogger, uow)
	excluirCursoUC := cursocmd.NovoExcluirCursoUseCase(cursoRepo, auditLogger, uow)
	listarCursosUC := cursoquery.NovoListarCursosUseCase(cursoRepo, periodoRepo, planoRepo)
	buscarCursoUC := cursoquery.NovoBuscarCursoUseCase(cursoRepo)
	listarMeusCursosUC := cursoquery.NovoListarMeusCursosUseCase(cursoRepo)

	criarDesignacaoUC := designacaocmd.NovoCriarDesignacaoUseCase(designacaoRepo, cursoRepo, usuarioRepo, auditLogger, uow)
	atualizarDesignacaoUC := designacaocmd.NovoAtualizarDesignacaoUseCase(designacaoRepo, auditLogger, uow)
	excluirDesignacaoUC := designacaocmd.NovoExcluirDesignacaoUseCase(designacaoRepo, auditLogger, uow)
	listarDesignacoesDoCursoUC := designacaoquery.NovoListarDoCursoUseCase(designacaoRepo)
	buscarDesignacaoUC := designacaoquery.NovoBuscarDesignacaoUseCase(designacaoRepo)
	listarCandidatosUC := designacaoquery.NovoListarCandidatosUseCase(designacaoRepo)

	authHandler := NovoAuthHandler(autenticarUC, encerrarSessaoUC, alterarSenhaPropriaUC, obterContextoUC, tokenDeSessao)
	instituicaoPublicaHandler := NovoInstituicaoPublicaHandler(listarInstituicoesPublicasUC)
	instituicaoHandler := NovoInstituicaoHandler(criarInstituicaoUC, atualizarInstituicaoUC, alterarSituacaoInstituicaoUC, listarInstituicoesUC, buscarInstituicaoUC)
	usuarioHandler := NovoUsuarioHandler(criarUsuarioUC, atualizarUsuarioUC, excluirUsuarioUC, redefinirSenhaUsuarioUC, listarUsuariosUC, buscarUsuarioUC)
	indicadorPlataformaHandler := NovoIndicadorPlataformaHandler(
		criarIndicadorPlataformaUC, atualizarIndicadorPlataformaUC, alterarSituacaoIndicadorPlataformaUC, excluirIndicadorPlataformaUC,
		listarIndicadoresPlataformaUC, buscarIndicadorPlataformaUC)
	indicadorHandler := NovoIndicadorHandler(
		criarIndicadorUC, atualizarIndicadorUC, alterarSituacaoIndicadorUC, excluirIndicadorUC,
		listarDoCatalogoUC, buscarIndicadorUC, sugerirIndicadorUC)
	metaHandler := NovoMetaHandler(
		criarMetaUC, atualizarMetaUC, alterarSituacaoMetaUC, excluirMetaUC, listarMetasUC, buscarMetaUC, sugerirMetaUC)
	cursoHandler := NovoCursoHandler(
		criarCursoUC, atualizarCursoUC, alterarSituacaoCursoUC, excluirCursoUC, listarCursosUC, buscarCursoUC, listarMeusCursosUC)
	designacaoHandler := NovoDesignacaoHandler(
		criarDesignacaoUC, atualizarDesignacaoUC, excluirDesignacaoUC, listarDesignacoesDoCursoUC, buscarDesignacaoUC, listarCandidatosUC)

	router := gin.New()
	router.Use(MiddlewareErro())

	grupoAPI := router.Group("/api/v1")
	middlewareSessao := NovoMiddlewareSessao(tokenDeSessao, autenticacaoRepo, relogioReal, fusoDeExibicao)
	middlewareAutorizacao := NovoMiddlewareAutorizacao(auditLogger)
	registro := NovoRegistro(grupoAPI, middlewareSessao, middlewareAutorizacao, NovoMiddlewareEscopoProprio())

	registro.Publica(http.MethodGet, "/publico/instituicoes", instituicaoPublicaHandler.Listar)
	registro.AutenticadaSemPermissao(http.MethodPost, "/auth/logout", authHandler.Logout)
	registro.AutenticadaSemPermissao(http.MethodGet, "/auth/eu", authHandler.Eu)
	registro.AutenticadaSemPermissao(http.MethodPost, "/auth/senha", authHandler.AlterarSenha)
	grupoAPI.POST("/auth/login", authHandler.Login)

	registro.Autenticada(http.MethodGet, "/instituicoes", autorizacao.InstituicaoListar, instituicaoHandler.Listar)
	registro.Autenticada(http.MethodGet, "/instituicoes/:id", autorizacao.InstituicaoListar, instituicaoHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/instituicoes", autorizacao.InstituicaoCriar, instituicaoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/instituicoes/:id", autorizacao.InstituicaoEditar, instituicaoHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/instituicoes/:id/situacao", autorizacao.InstituicaoInativar, instituicaoHandler.AlterarSituacao)

	registro.Autenticada(http.MethodGet, "/usuarios", autorizacao.UsuarioListar, usuarioHandler.Listar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodGet, "/usuarios/:usuario_id", autorizacao.UsuarioListar, usuarioHandler.Buscar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodPost, "/usuarios", autorizacao.UsuarioCriar, usuarioHandler.Criar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodPut, "/usuarios/:usuario_id", autorizacao.UsuarioEditar, usuarioHandler.Atualizar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodDelete, "/usuarios/:usuario_id", autorizacao.UsuarioExcluir, usuarioHandler.Excluir(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodPost, "/usuarios/:usuario_id/senha", autorizacao.UsuarioRedefinirSenha, usuarioHandler.RedefinirSenha(autorizacao.UsuariosDaPropriaInstituicao))

	registro.Autenticada(http.MethodGet, "/instituicoes/:id/pesquisadores", autorizacao.PIGerenciar, usuarioHandler.Listar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodGet, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.Buscar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodPost, "/instituicoes/:id/pesquisadores", autorizacao.PIGerenciar, usuarioHandler.Criar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodPut, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.Atualizar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodDelete, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.Excluir(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodPost, "/instituicoes/:id/pesquisadores/:usuario_id/senha", autorizacao.PIGerenciar, usuarioHandler.RedefinirSenha(autorizacao.PesquisadoresDeUmaInstituicao))

	registro.Autenticada(http.MethodGet, "/administradores", autorizacao.AdministradorGerenciar, usuarioHandler.Listar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodGet, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.Buscar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodPost, "/administradores", autorizacao.AdministradorGerenciar, usuarioHandler.Criar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodPut, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.Atualizar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodDelete, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.Excluir(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodPost, "/administradores/:usuario_id/senha", autorizacao.AdministradorGerenciar, usuarioHandler.RedefinirSenha(autorizacao.AdministradoresDaPlataforma))

	registro.Autenticada(http.MethodGet, "/plataforma/indicadores", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Listar)
	registro.Autenticada(http.MethodGet, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/plataforma/indicadores", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Criar)
	registro.Autenticada(http.MethodPut, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/plataforma/indicadores/:id/situacao", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Excluir)

	registro.Autenticada(http.MethodGet, "/indicadores", autorizacao.IndicadorListar, indicadorHandler.Listar)
	registro.Autenticada(http.MethodGet, "/indicadores/sugestoes", autorizacao.IndicadorListar, indicadorHandler.Sugerir)
	registro.Autenticada(http.MethodGet, "/indicadores/:id", autorizacao.IndicadorListar, indicadorHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/indicadores", autorizacao.IndicadorGerenciar, indicadorHandler.Criar)
	registro.Autenticada(http.MethodPut, "/indicadores/:id", autorizacao.IndicadorGerenciar, indicadorHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/indicadores/:id/situacao", autorizacao.IndicadorGerenciar, indicadorHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/indicadores/:id", autorizacao.IndicadorGerenciar, indicadorHandler.Excluir)

	registro.Autenticada(http.MethodGet, "/metas", autorizacao.MetaListar, metaHandler.Listar)
	registro.Autenticada(http.MethodGet, "/metas/sugestoes", autorizacao.MetaListar, metaHandler.Sugerir)
	registro.Autenticada(http.MethodGet, "/metas/:id", autorizacao.MetaListar, metaHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/metas", autorizacao.MetaGerenciar, metaHandler.Criar)
	registro.Autenticada(http.MethodPut, "/metas/:id", autorizacao.MetaGerenciar, metaHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/metas/:id/situacao", autorizacao.MetaGerenciar, metaHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/metas/:id", autorizacao.MetaGerenciar, metaHandler.Excluir)

	registro.Autenticada(http.MethodGet, "/cursos", autorizacao.CursoListar, cursoHandler.Listar)
	registro.Autenticada(http.MethodGet, "/cursos/:id", autorizacao.CursoListar, cursoHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/cursos", autorizacao.CursoGerenciar, cursoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/cursos/:id", autorizacao.CursoGerenciar, cursoHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/cursos/:id/situacao", autorizacao.CursoGerenciar, cursoHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/cursos/:id", autorizacao.CursoGerenciar, cursoHandler.Excluir)
	registro.Autenticada(http.MethodGet, "/cursos/:id/designacoes", autorizacao.DesignacaoGerenciar, designacaoHandler.ListarDoCurso)
	registro.Autenticada(http.MethodPost, "/cursos/:id/designacoes", autorizacao.DesignacaoGerenciar, designacaoHandler.Criar)
	registro.Autenticada(http.MethodGet, "/designacoes/candidatos", autorizacao.DesignacaoGerenciar, designacaoHandler.Candidatos)
	registro.Autenticada(http.MethodGet, "/designacoes/:id", autorizacao.DesignacaoGerenciar, designacaoHandler.Buscar)
	registro.Autenticada(http.MethodPut, "/designacoes/:id", autorizacao.DesignacaoGerenciar, designacaoHandler.Atualizar)
	registro.Autenticada(http.MethodDelete, "/designacoes/:id", autorizacao.DesignacaoGerenciar, designacaoHandler.Excluir)
	registro.Autenticada(http.MethodGet, "/meus-cursos", autorizacao.CursoLerProprio, cursoHandler.MeusCursos)

	return router
}

func fazerLogin(t *testing.T, router *gin.Engine, instituicaoID *string, email, senha string) *httptest.ResponseRecorder {
	t.Helper()
	corpo, _ := json.Marshal(LoginRequest{InstituicaoID: instituicaoID, Email: email, Senha: senha})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(corpo))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestSmoke_L01_LoginValidoDevolveCookieComAsTresMarcas(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK1"})
	senhaProvisoriaFalsa := false
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Maria Souza", Email: "maria.souza.smk1@fsa.edu.br",
		Perfil: "pesquisador_institucional", SenhaProvisoria: &senhaProvisoriaFalsa,
		SenhaHash: hashParaTeste(t, "reuniao-nde-2026"),
	})
	_ = usuarioID

	fsaIDStr := fsaID.String()
	rec := fazerLogin(t, router, &fsaIDStr, "maria.souza.smk1@fsa.edu.br", "reuniao-nde-2026")

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obtido %d: %s", rec.Code, rec.Body.String())
	}

	setCookie := rec.Header().Get("Set-Cookie")
	for _, marca := range []string{"HttpOnly", "Secure", "SameSite=Strict"} {
		if !containsCaseSensitive(setCookie, marca) {
			t.Fatalf("Set-Cookie sem a marca %q: %s", marca, setCookie)
		}
	}

	var corpoLogin map[string]any
	json.Unmarshal(rec.Body.Bytes(), &corpoLogin)

	cookies := rec.Result().Cookies()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	recEu := httptest.NewRecorder()
	router.ServeHTTP(recEu, req)
	var corpoEu map[string]any
	json.Unmarshal(recEu.Body.Bytes(), &corpoEu)

	if corpoLogin["nome"] != corpoEu["nome"] || corpoLogin["id"] != corpoEu["id"] {
		t.Fatalf("corpo de login deveria ser igual ao de /auth/eu: login=%v eu=%v", corpoLogin, corpoEu)
	}
}

func containsCaseSensitive(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func TestSmoke_LoginInvalido_401ComMensagemGenerica(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK2"})
	fsaIDStr := fsaID.String()

	rec := fazerLogin(t, router, &fsaIDStr, "ninguem@fsa.edu.br", "qualquercoisa")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d: %s", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "CREDENCIAIS_INVALIDAS" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

func TestSmoke_L05_20TentativasErradasNenhum429SemAuditoria(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK3"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Maria Souza", Email: "maria.souza.smk3@fsa.edu.br",
		Perfil: "pesquisador_institucional", SenhaProvisoria: &senhaProvisoriaFalsa,
		SenhaHash: hashParaTeste(t, "reuniao-nde-2026"),
	})

	fsaIDStr := fsaID.String()
	for i := 0; i < 20; i++ {
		rec := fazerLogin(t, router, &fsaIDStr, "maria.souza.smk3@fsa.edu.br", "senha-errada")
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("tentativa %d: nunca deveria responder 429", i+1)
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("tentativa %d: esperava 401, obtido %d", i+1, rec.Code)
		}
	}

	recFinal := fazerLogin(t, router, &fsaIDStr, "maria.souza.smk3@fsa.edu.br", "reuniao-nde-2026")
	if recFinal.Code != http.StatusOK {
		t.Fatalf("21ª tentativa (senha correta) deveria entrar: %d (%s)", recFinal.Code, recFinal.Body.String())
	}

	// Restrito à instituição deste teste — nunca a tabela inteira: outros
	// pacotes de teste podem inserir em auditoria ao mesmo tempo (go test
	// ./... roda pacotes em paralelo contra o mesmo banco de teste).
	var total int
	if err := db.Get(&total, `SELECT count(*) FROM auditoria WHERE instituicao_id = $1`, fsaID); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if total != 0 {
		t.Fatalf("nenhuma tentativa de login deveria gerar auditoria, encontrado %d registro(s)", total)
	}
}

func TestSmoke_SE01_LogoutApagaCookieECookieAntigoDevolve401(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "SMK4"})
	senhaProvisoriaFalsa := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Ana Lima", Email: "ana.lima.smk4@fsa.edu.br",
		// professor, não coordenador_curso: desde DC-4 (specs/cursos/
		// design.md C-09) esse perfil nunca é atribuído diretamente. O
		// teste só precisa de alguém logado para testar logout.
		Perfil: "professor", SenhaProvisoria: &senhaProvisoriaFalsa,
		SenhaHash: hashParaTeste(t, "colegiado-terca-14h"),
	})

	fsaIDStr := fsaID.String()
	recLogin := fazerLogin(t, router, &fsaIDStr, "ana.lima.smk4@fsa.edu.br", "colegiado-terca-14h")
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login deveria funcionar: %d (%s)", recLogin.Code, recLogin.Body.String())
	}
	cookieAntigo := recLogin.Result().Cookies()[0]

	reqLogout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	reqLogout.AddCookie(cookieAntigo)
	recLogout := httptest.NewRecorder()
	router.ServeHTTP(recLogout, reqLogout)
	if recLogout.Code != http.StatusNoContent {
		t.Fatalf("esperava 204, obtido %d", recLogout.Code)
	}

	reqSeguinte := httptest.NewRequest(http.MethodGet, "/api/v1/auth/eu", nil)
	reqSeguinte.AddCookie(cookieAntigo)
	recSeguinte := httptest.NewRecorder()
	router.ServeHTTP(recSeguinte, reqSeguinte)
	if recSeguinte.Code != http.StatusUnauthorized {
		t.Fatalf("cookie antigo após logout deveria devolver 401, obtido %d", recSeguinte.Code)
	}
}

func TestSmoke_RotaPublicaRespondeSemCookie(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/publico/instituicoes", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rota pública deveria responder 200 sem cookie, obtido %d: %s", rec.Code, rec.Body.String())
	}
}
