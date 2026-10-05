package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	_ "github.com/basis-avalia/backend/docs"
	adapterhttp "github.com/basis-avalia/backend/internal/adapter/http"

	"github.com/basis-avalia/backend/internal/adapter/argon2"
	adapterauditoria "github.com/basis-avalia/backend/internal/adapter/auditoria"
	"github.com/basis-avalia/backend/internal/adapter/docx"
	adapterjwt "github.com/basis-avalia/backend/internal/adapter/jwt"
	adaptermetricas "github.com/basis-avalia/backend/internal/adapter/metricas"
	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/adapter/relogio"
	"github.com/basis-avalia/backend/internal/adapter/s3"
	adaptersmtp "github.com/basis-avalia/backend/internal/adapter/smtp"
	anexocmd "github.com/basis-avalia/backend/internal/usecase/command/anexo"
	avaliacaocmd "github.com/basis-avalia/backend/internal/usecase/command/avaliacao"
	cursocmd "github.com/basis-avalia/backend/internal/usecase/command/curso"
	designacaocmd "github.com/basis-avalia/backend/internal/usecase/command/designacao"
	documentocmd "github.com/basis-avalia/backend/internal/usecase/command/documento"
	entregacmd "github.com/basis-avalia/backend/internal/usecase/command/entrega"
	indicadorcmd "github.com/basis-avalia/backend/internal/usecase/command/indicador"
	indicadorplataformacmd "github.com/basis-avalia/backend/internal/usecase/command/indicador_plataforma"
	instituicaocmd "github.com/basis-avalia/backend/internal/usecase/command/instituicao"
	itemplanocmd "github.com/basis-avalia/backend/internal/usecase/command/item_plano"
	metacmd "github.com/basis-avalia/backend/internal/usecase/command/meta"
	periodocmd "github.com/basis-avalia/backend/internal/usecase/command/periodo"
	planocmd "github.com/basis-avalia/backend/internal/usecase/command/plano"
	sessaocmd "github.com/basis-avalia/backend/internal/usecase/command/sessao"
	usuariocmd "github.com/basis-avalia/backend/internal/usecase/command/usuario"
	anexoquery "github.com/basis-avalia/backend/internal/usecase/query/anexo"
	avaliacaoquery "github.com/basis-avalia/backend/internal/usecase/query/avaliacao"
	cursoquery "github.com/basis-avalia/backend/internal/usecase/query/curso"
	designacaoquery "github.com/basis-avalia/backend/internal/usecase/query/designacao"
	documentoquery "github.com/basis-avalia/backend/internal/usecase/query/documento"
	entregaquery "github.com/basis-avalia/backend/internal/usecase/query/entrega"
	indicadorquery "github.com/basis-avalia/backend/internal/usecase/query/indicador"
	indicadorplataformaquery "github.com/basis-avalia/backend/internal/usecase/query/indicador_plataforma"
	instituicaoquery "github.com/basis-avalia/backend/internal/usecase/query/instituicao"
	metaquery "github.com/basis-avalia/backend/internal/usecase/query/meta"
	pendenciaquery "github.com/basis-avalia/backend/internal/usecase/query/pendencia"
	periodoquery "github.com/basis-avalia/backend/internal/usecase/query/periodo"
	planoquery "github.com/basis-avalia/backend/internal/usecase/query/plano"
	relatorioquery "github.com/basis-avalia/backend/internal/usecase/query/relatorio"
	sessaoquery "github.com/basis-avalia/backend/internal/usecase/query/sessao"
	usuarioquery "github.com/basis-avalia/backend/internal/usecase/query/usuario"
	"github.com/basis-avalia/backend/internal/usecase/servico/rele"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
)

// @title        basis-avalia API
// @version      1.0
// @description  Autenticação, instituições e usuários com múltiplos perfis.
// @BasePath     /
func main() {
	databaseURL := requireEnv("DATABASE_URL")
	redisURL := requireEnv("REDIS_URL")
	jwtSecret := requireEnv("JWT_SECRET")
	porta := getEnv("PORT", "3001")
	adminPorta := getEnv("ADMIN_PORT", "9090")
	appEnv := getEnv("APP_ENV", "production")
	appVersao := getEnv("APP_VERSION", "0.0.0-dev")
	origensPermitidas := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", ""), ",")
	argon2Concorrencia := getEnvInt("ARGON2_CONCORRENCIA", 4)
	s3Endpoint := requireEnv("S3_ENDPOINT")
	s3Bucket := requireEnv("S3_BUCKET")
	s3AccessKey := requireEnv("S3_ACCESS_KEY")
	s3SecretKey := requireEnv("S3_SECRET_KEY")
	smtpHost := requireEnv("SMTP_HOST")
	smtpPorta := requireEnv("SMTP_PORT")
	smtpUsuario := getEnv("SMTP_USER", "")
	smtpSenha := getEnv("SMTP_PASSWORD", "")
	smtpRemetente := requireEnv("SMTP_FROM")
	smtpTLS := getEnv("SMTP_TLS", "false") == "true"
	releHabilitado := getEnv("RELE_HABILITADO", "true") == "true"

	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		log.Fatalf("conectando ao postgres: %v", err)
	}
	defer db.Close()

	opcoesRedis, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("parseando REDIS_URL: %v", err)
	}
	redisClient := redis.NewClient(opcoesRedis)
	defer redisClient.Close()

	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	relogioReal := relogio.Novo()

	fusoDeExibicao, err := time.LoadLocation(getEnv("APP_TIMEZONE", "America/Sao_Paulo"))
	if err != nil {
		log.Fatalf("APP_TIMEZONE inválido: %v", err)
	}

	// Adapters de saída.
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
	itemPlanoRepo := postgres.NovoItemPlanoRepository(db)
	documentoRepo := postgres.NovoDocumentoRepository(db)
	entregaRepo := postgres.NovoEntregaRepository(db)
	armazenamento := s3.Novo(s3Endpoint, "us-east-1", s3AccessKey, s3SecretKey, s3Bucket)
	geradorDeDocumento := docx.Novo()
	emailSender := adaptersmtp.Novo(smtpHost, smtpPorta, smtpUsuario, smtpSenha, smtpRemetente, smtpTLS)
	hashDeSenha := argon2.NovoHashDeSenha(argon2Concorrencia)
	tokenDeSessao := adapterjwt.NovoTokenDeSessao(jwtSecret, relogioReal)
	uow := postgres.NovaUnidadeDeTrabalho(db)

	syslogServer := getEnv("SYSLOG_SERVER", "")
	syslogProtocolo := getEnv("SYSLOG_PROTOCOL", "tcp")
	syslogAppName := getEnv("SYSLOG_APP_NAME", "basis-avalia")
	syslogAdapter, err := adapterauditoria.NovoSyslog(syslogServer, syslogProtocolo, syslogAppName)
	if err != nil {
		log.Printf("aviso: syslog não conectou (%v) — canal 2 de auditoria em modo no-op", err)
	}
	auditoriaRepo := postgres.NovoAuditoriaRepository(db)
	auditLogger := adapterauditoria.NovoComposto(auditoriaRepo, syslogAdapter)

	// Use cases.
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

	// specs/cursos — curso e designação de coordenação, com o perfil de
	// Coordenador derivado de designação vigente (fundacao-metas.md §4).
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

	// specs/plano-acao — período, plano de ação curso/coordenador, itens
	// e documento.
	criarPeriodoUC := periodocmd.NovoCriarPeriodoUseCase(periodoRepo, auditLogger, uow)
	atualizarPeriodoUC := periodocmd.NovoAtualizarPeriodoUseCase(periodoRepo, auditLogger, uow)
	excluirPeriodoUC := periodocmd.NovoExcluirPeriodoUseCase(periodoRepo, auditLogger, uow)
	listarPeriodosUC := periodoquery.NovoListarPeriodosUseCase(periodoRepo)
	buscarPeriodoUC := periodoquery.NovoBuscarPeriodoUseCase(periodoRepo)

	criarPlanoUC := planocmd.NovoCriarPlanoUseCase(planoRepo, auditLogger, uow)
	atualizarPlanoUC := planocmd.NovoAtualizarPlanoUseCase(planoRepo, auditLogger, uow)
	publicarPlanoUC := planocmd.NovoPublicarPlanoUseCase(planoRepo, periodoRepo, auditLogger, uow)
	despublicarPlanoUC := planocmd.NovoDespublicarPlanoUseCase(planoRepo, auditLogger, uow)
	encerrarPlanoUC := planocmd.NovoEncerrarPlanoUseCase(planoRepo, auditLogger, uow)
	reabrirPlanoUC := planocmd.NovoReabrirPlanoUseCase(planoRepo, auditLogger, uow)
	excluirPlanoUC := planocmd.NovoExcluirPlanoUseCase(planoRepo, auditLogger, uow)
	copiarEmLoteUC := planocmd.NovoCopiarEmLoteUseCase(planoRepo, auditLogger, uow)
	listarPlanosUC := planoquery.NovoListarPlanosUseCase(planoRepo)
	listarMeusPlanosUC := planoquery.NovoListarMeusPlanosUseCase(planoRepo)
	buscarPlanoUC := planoquery.NovoBuscarPlanoUseCase(planoRepo)
	listarDestinosDeCopiaUC := planoquery.NovoListarDestinosDeCopiaUseCase(planoRepo)

	criarItemUC := itemplanocmd.NovoCriarItemUseCase(planoRepo, itemPlanoRepo, metaRepo, auditLogger, uow)
	atualizarItemUC := itemplanocmd.NovoAtualizarItemUseCase(planoRepo, itemPlanoRepo, auditLogger, uow)
	excluirItemUC := itemplanocmd.NovoExcluirItemUseCase(planoRepo, itemPlanoRepo, auditLogger, uow)

	gerarDocumentoUC := documentocmd.NovoGerarDocumentoUseCase(planoRepo, documentoRepo, armazenamento, geradorDeDocumento, auditLogger, uow)
	baixarDocumentoUC := documentoquery.NovoBaixarDocumentoUseCase(documentoRepo, armazenamento, auditLogger)

	// specs/metas-coordenacao — entrega, anexo, avaliação, notificação e
	// relatório de desempenho. Última feature da cadeia de metas.
	registrarEntregaUC := entregacmd.NovoRegistrarEntregaUseCase(entregaRepo, itemPlanoRepo, planoRepo, armazenamento, auditLogger, uow)
	corrigirEntregaUC := entregacmd.NovoCorrigirEntregaUseCase(entregaRepo, auditLogger, uow, relogioReal)
	excluirEntregaUC := entregacmd.NovoExcluirEntregaUseCase(entregaRepo, auditLogger, uow)
	marcarPendenciaVistaUC := entregacmd.NovoMarcarPendenciaVistaUseCase(entregaRepo)
	adicionarAnexoUC := anexocmd.NovoAdicionarAnexoUseCase(entregaRepo, armazenamento)
	removerAnexoUC := anexocmd.NovoRemoverAnexoUseCase(entregaRepo)
	listarDoItemUC := entregaquery.NovoListarDoItemUseCase(entregaRepo)
	buscarEntregaUC := entregaquery.NovoBuscarUseCase(entregaRepo)
	minhasMetasUC := entregaquery.NovoMinhasMetasUseCase(entregaRepo)
	listarPeriodosParaMinhasMetasUC := entregaquery.NovoListarPeriodosUseCase(entregaRepo)
	baixarAnexoUC := anexoquery.NovoBaixarUseCase(entregaRepo, armazenamento, auditLogger)

	avaliarUC := avaliacaocmd.NovoAvaliarUseCase(entregaRepo, auditLogger, uow, relogioReal, fusoDeExibicao)
	desfazerAceitacaoUC := avaliacaocmd.NovoDesfazerAceitacaoUseCase(entregaRepo, auditLogger, uow, relogioReal, fusoDeExibicao)
	listarFilaUC := avaliacaoquery.NovoListarFilaUseCase(entregaRepo)

	contarBadgesUC := pendenciaquery.NovoContarBadgesUseCase(entregaRepo)
	desempenhoUC := relatorioquery.NovoDesempenhoUseCase(entregaRepo)
	desempenhoPorCursoUC := relatorioquery.NovoDesempenhoPorCursoUseCase(entregaRepo)
	exportarDesempenhoUC := relatorioquery.NovoExportarDesempenhoUseCase(entregaRepo, auditLogger)

	// Handlers.
	authHandler := adapterhttp.NovoAuthHandler(autenticarUC, encerrarSessaoUC, alterarSenhaPropriaUC, obterContextoUC, tokenDeSessao)
	instituicaoPublicaHandler := adapterhttp.NovoInstituicaoPublicaHandler(listarInstituicoesPublicasUC)
	instituicaoHandler := adapterhttp.NovoInstituicaoHandler(criarInstituicaoUC, atualizarInstituicaoUC, alterarSituacaoInstituicaoUC, listarInstituicoesUC, buscarInstituicaoUC)
	indicadorPlataformaHandler := adapterhttp.NovoIndicadorPlataformaHandler(
		criarIndicadorPlataformaUC, atualizarIndicadorPlataformaUC, alterarSituacaoIndicadorPlataformaUC, excluirIndicadorPlataformaUC,
		listarIndicadoresPlataformaUC, buscarIndicadorPlataformaUC)
	indicadorHandler := adapterhttp.NovoIndicadorHandler(
		criarIndicadorUC, atualizarIndicadorUC, alterarSituacaoIndicadorUC, excluirIndicadorUC,
		listarDoCatalogoUC, buscarIndicadorUC, sugerirIndicadorUC)
	metaHandler := adapterhttp.NovoMetaHandler(
		criarMetaUC, atualizarMetaUC, alterarSituacaoMetaUC, excluirMetaUC, listarMetasUC, buscarMetaUC, sugerirMetaUC)
	cursoHandler := adapterhttp.NovoCursoHandler(
		criarCursoUC, atualizarCursoUC, alterarSituacaoCursoUC, excluirCursoUC, listarCursosUC, buscarCursoUC, listarMeusCursosUC)
	designacaoHandler := adapterhttp.NovoDesignacaoHandler(
		criarDesignacaoUC, atualizarDesignacaoUC, excluirDesignacaoUC, listarDesignacoesDoCursoUC, buscarDesignacaoUC, listarCandidatosUC)
	usuarioHandler := adapterhttp.NovoUsuarioHandler(criarUsuarioUC, atualizarUsuarioUC, excluirUsuarioUC, redefinirSenhaUsuarioUC, listarUsuariosUC, buscarUsuarioUC)
	periodoHandler := adapterhttp.NovoPeriodoHandler(criarPeriodoUC, atualizarPeriodoUC, excluirPeriodoUC, listarPeriodosUC, buscarPeriodoUC)
	planoHandler := adapterhttp.NovoPlanoHandler(
		criarPlanoUC, atualizarPlanoUC, publicarPlanoUC, despublicarPlanoUC, encerrarPlanoUC, reabrirPlanoUC, excluirPlanoUC,
		copiarEmLoteUC, listarPlanosUC, listarMeusPlanosUC, buscarPlanoUC, listarDestinosDeCopiaUC)
	itemPlanoHandler := adapterhttp.NovoItemPlanoHandler(criarItemUC, atualizarItemUC, excluirItemUC, buscarPlanoUC)
	documentoHandler := adapterhttp.NovoDocumentoHandler(gerarDocumentoUC, baixarDocumentoUC)

	entregaHandler := adapterhttp.NovoEntregaHandler(
		registrarEntregaUC, corrigirEntregaUC, excluirEntregaUC, marcarPendenciaVistaUC,
		adicionarAnexoUC, removerAnexoUC, listarDoItemUC, buscarEntregaUC, minhasMetasUC, listarPeriodosParaMinhasMetasUC, baixarAnexoUC)
	avaliacaoHandler := adapterhttp.NovoAvaliacaoHandler(avaliarUC, desfazerAceitacaoUC, listarFilaUC, buscarEntregaUC)
	relatorioHandler := adapterhttp.NovoRelatorioHandler(desempenhoUC, desempenhoPorCursoUC, exportarDesempenhoUC)
	pendenciaHandler := adapterhttp.NovoPendenciaHandler(contarBadgesUC)

	// Duas portas com propósitos distintos (design.md §4.5, §16 T-100): a
	// pública serve /api/v1 e /version, porque o navegador precisa delas;
	// a administrativa serve /metrics, /healthz e /readyz, e não é
	// publicada no host nem no ingress — scrape e probes vêm da rede
	// interna. Mesmo processo, mesmo registro Prometheus.
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(adapterhttp.MetricasMiddleware())
	router.Use(corsMiddleware(origensPermitidas))
	router.Use(adapterhttp.MiddlewareSeguranca(appEnv))
	router.Use(adapterhttp.MiddlewareErro())

	healthHandler := adapterhttp.NewHealthHandler(db, redisClient, appVersao).
		ComVerificacaoDeArmazenamento(func(ctx context.Context) error { return armazenamento.VerificarDisponibilidade(ctx) }).
		ComVerificacaoDeSMTP(func(ctx context.Context) error { return emailSender.VerificarDisponibilidade(ctx) })
	router.GET("/version", healthHandler.Versao)
	// Não registrar é mais forte que proteger: rota que não existe não
	// pode ser mal configurada (design.md §4.5, achado A-2).
	if appEnv != "production" {
		router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	adminRouter := gin.New()
	adminRouter.Use(gin.Recovery())
	adminRouter.GET("/healthz", healthHandler.Liveness)
	adminRouter.GET("/readyz", healthHandler.Readiness)
	adminRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

	grupoAPI := router.Group("/api/v1")
	middlewareSessao := adapterhttp.NovoMiddlewareSessao(tokenDeSessao, autenticacaoRepo, relogioReal, fusoDeExibicao)
	middlewareAutorizacao := adapterhttp.NovoMiddlewareAutorizacao(auditLogger)
	registro := adapterhttp.NovoRegistro(grupoAPI, middlewareSessao, middlewareAutorizacao, adapterhttp.NovoMiddlewareEscopoProprio())

	registro.Publica(http.MethodGet, "/publico/instituicoes", instituicaoPublicaHandler.Listar)
	registro.AutenticadaSemPermissao(http.MethodPost, "/auth/logout", authHandler.Logout)
	registro.AutenticadaSemPermissao(http.MethodGet, "/auth/eu", authHandler.Eu)
	registro.AutenticadaSemPermissao(http.MethodPost, "/auth/senha", authHandler.AlterarSenha)

	// POST /auth/login é registrada direto no grupo, sem o middleware de
	// sessão (não existe sessão a validar antes de logar) nem o de
	// permissão: nem Publica() (que a spec reserva à rota do combo, 3.17)
	// nem AutenticadaSemPermissao (que pressupõe sessão já estabelecida).
	// É por isso que ela fica de fora das duas contagens de T-006.
	grupoAPI.POST("/auth/login", authHandler.Login)

	// R12–R16 — Instituições (Administrador do Sistema).
	registro.Autenticada(http.MethodGet, "/instituicoes", autorizacao.InstituicaoListar, instituicaoHandler.Listar)
	registro.Autenticada(http.MethodGet, "/instituicoes/:id", autorizacao.InstituicaoListar, instituicaoHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/instituicoes", autorizacao.InstituicaoCriar, instituicaoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/instituicoes/:id", autorizacao.InstituicaoEditar, instituicaoHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/instituicoes/:id/situacao", autorizacao.InstituicaoInativar, instituicaoHandler.AlterarSituacao)

	// R6–R11 — Usuários da própria instituição (Pesquisador Institucional).
	registro.Autenticada(http.MethodGet, "/usuarios", autorizacao.UsuarioListar, usuarioHandler.Listar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodGet, "/usuarios/:usuario_id", autorizacao.UsuarioListar, usuarioHandler.Buscar(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodPost, "/usuarios", autorizacao.UsuarioCriar, usuarioHandler.CriarUsuario())
	registro.Autenticada(http.MethodPut, "/usuarios/:usuario_id", autorizacao.UsuarioEditar, usuarioHandler.AtualizarUsuario())
	registro.Autenticada(http.MethodDelete, "/usuarios/:usuario_id", autorizacao.UsuarioExcluir, usuarioHandler.Excluir(autorizacao.UsuariosDaPropriaInstituicao))
	registro.Autenticada(http.MethodPost, "/usuarios/:usuario_id/senha", autorizacao.UsuarioRedefinirSenha, usuarioHandler.RedefinirSenha(autorizacao.UsuariosDaPropriaInstituicao))

	// R17–R22 — Pesquisadores Institucionais de uma instituição (Administrador).
	registro.Autenticada(http.MethodGet, "/instituicoes/:id/pesquisadores", autorizacao.PIGerenciar, usuarioHandler.Listar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodGet, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.Buscar(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodPost, "/instituicoes/:id/pesquisadores", autorizacao.PIGerenciar, usuarioHandler.CriarPesquisador())
	registro.Autenticada(http.MethodPut, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.AtualizarPesquisador())
	registro.Autenticada(http.MethodDelete, "/instituicoes/:id/pesquisadores/:usuario_id", autorizacao.PIGerenciar, usuarioHandler.Excluir(autorizacao.PesquisadoresDeUmaInstituicao))
	registro.Autenticada(http.MethodPost, "/instituicoes/:id/pesquisadores/:usuario_id/senha", autorizacao.PIGerenciar, usuarioHandler.RedefinirSenha(autorizacao.PesquisadoresDeUmaInstituicao))

	// R23–R28 — Administradores do Sistema (Administrador).
	registro.Autenticada(http.MethodGet, "/administradores", autorizacao.AdministradorGerenciar, usuarioHandler.Listar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodGet, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.Buscar(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodPost, "/administradores", autorizacao.AdministradorGerenciar, usuarioHandler.CriarAdministrador())
	registro.Autenticada(http.MethodPut, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.AtualizarAdministrador())
	registro.Autenticada(http.MethodDelete, "/administradores/:usuario_id", autorizacao.AdministradorGerenciar, usuarioHandler.Excluir(autorizacao.AdministradoresDaPlataforma))
	registro.Autenticada(http.MethodPost, "/administradores/:usuario_id/senha", autorizacao.AdministradorGerenciar, usuarioHandler.RedefinirSenha(autorizacao.AdministradoresDaPlataforma))

	// Catálogo do INEP (Administrador do Sistema) — specs/indicadores.
	registro.Autenticada(http.MethodGet, "/plataforma/indicadores", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Listar)
	registro.Autenticada(http.MethodGet, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/plataforma/indicadores", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Criar)
	registro.Autenticada(http.MethodPut, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/plataforma/indicadores/:id/situacao", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/plataforma/indicadores/:id", autorizacao.IndicadorPlataformaGerenciar, indicadorPlataformaHandler.Excluir)

	// Indicadores da instituição — lê os dois escopos, escreve só o
	// próprio (PI, Coordenador como contexto). specs/indicadores.
	registro.Autenticada(http.MethodGet, "/indicadores", autorizacao.IndicadorListar, indicadorHandler.Listar)
	registro.Autenticada(http.MethodGet, "/indicadores/sugestoes", autorizacao.IndicadorListar, indicadorHandler.Sugerir)
	registro.Autenticada(http.MethodGet, "/indicadores/:id", autorizacao.IndicadorListar, indicadorHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/indicadores", autorizacao.IndicadorGerenciar, indicadorHandler.Criar)
	registro.Autenticada(http.MethodPut, "/indicadores/:id", autorizacao.IndicadorGerenciar, indicadorHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/indicadores/:id/situacao", autorizacao.IndicadorGerenciar, indicadorHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/indicadores/:id", autorizacao.IndicadorGerenciar, indicadorHandler.Excluir)

	// Catálogo de metas da instituição (PI). specs/indicadores.
	registro.Autenticada(http.MethodGet, "/metas", autorizacao.MetaListar, metaHandler.Listar)
	registro.Autenticada(http.MethodGet, "/metas/sugestoes", autorizacao.MetaListar, metaHandler.Sugerir)
	registro.Autenticada(http.MethodGet, "/metas/:id", autorizacao.MetaListar, metaHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/metas", autorizacao.MetaGerenciar, metaHandler.Criar)
	registro.Autenticada(http.MethodPut, "/metas/:id", autorizacao.MetaGerenciar, metaHandler.Atualizar)
	registro.Autenticada(http.MethodPatch, "/metas/:id/situacao", autorizacao.MetaGerenciar, metaHandler.AlterarSituacao)
	registro.Autenticada(http.MethodDelete, "/metas/:id", autorizacao.MetaGerenciar, metaHandler.Excluir)

	// Cursos e designações de coordenação (PI). specs/cursos. O perfil de
	// Coordenador nunca é aceito em /usuarios (DC-4) — é sempre derivado
	// daqui.
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

	// Cursos da carteira do coordenador — somente leitura (CV-01).
	registro.Autenticada(http.MethodGet, "/meus-cursos", autorizacao.CursoLerProprio, cursoHandler.MeusCursos)

	// Períodos da instituição (PI). specs/plano-acao.
	registro.Autenticada(http.MethodGet, "/periodos", autorizacao.PeriodoGerenciar, periodoHandler.Listar)
	registro.Autenticada(http.MethodGet, "/periodos/:id", autorizacao.PeriodoGerenciar, periodoHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/periodos", autorizacao.PeriodoGerenciar, periodoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/periodos/:id", autorizacao.PeriodoGerenciar, periodoHandler.Atualizar)
	registro.Autenticada(http.MethodDelete, "/periodos/:id", autorizacao.PeriodoGerenciar, periodoHandler.Excluir)

	// Planos de ação curso/coordenador da instituição (PI, leitura e
	// escrita). specs/plano-acao.
	registro.Autenticada(http.MethodGet, "/planos", autorizacao.PlanoListar, planoHandler.Listar)
	registro.Autenticada(http.MethodGet, "/planos/:id", autorizacao.PlanoListar, planoHandler.Buscar)
	registro.Autenticada(http.MethodPost, "/planos", autorizacao.PlanoGerenciar, planoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/planos/:id", autorizacao.PlanoGerenciar, planoHandler.Atualizar)
	registro.Autenticada(http.MethodDelete, "/planos/:id", autorizacao.PlanoGerenciar, planoHandler.Excluir)
	registro.Autenticada(http.MethodPost, "/planos/:id/publicar", autorizacao.PlanoGerenciar, planoHandler.Publicar)
	registro.Autenticada(http.MethodPost, "/planos/:id/despublicar", autorizacao.PlanoGerenciar, planoHandler.Despublicar)
	registro.Autenticada(http.MethodPost, "/planos/:id/encerrar", autorizacao.PlanoGerenciar, planoHandler.Encerrar)
	registro.Autenticada(http.MethodPost, "/planos/:id/reabrir", autorizacao.PlanoGerenciar, planoHandler.Reabrir)
	registro.Autenticada(http.MethodGet, "/planos/:id/destinos-copia", autorizacao.PlanoGerenciar, planoHandler.ListarDestinosDeCopia)
	registro.Autenticada(http.MethodPost, "/planos/:id/copias", autorizacao.PlanoGerenciar, planoHandler.CopiarEmLote)
	registro.Autenticada(http.MethodPost, "/planos/:id/itens", autorizacao.PlanoGerenciar, itemPlanoHandler.Criar)
	registro.Autenticada(http.MethodPut, "/planos/:id/itens/:item_id", autorizacao.PlanoGerenciar, itemPlanoHandler.Atualizar)
	registro.Autenticada(http.MethodDelete, "/planos/:id/itens/:item_id", autorizacao.PlanoGerenciar, itemPlanoHandler.Excluir)
	registro.Autenticada(http.MethodPost, "/planos/:id/documentos", autorizacao.PlanoListar, documentoHandler.Gerar)

	// Planos da carteira do coordenador — somente leitura, em qualquer
	// situação (P-10, design.md §5.4). O filtro de "somente vigentes" fica
	// na consulta de obrigações de specs/metas-coordenacao, nunca aqui.
	registro.Autenticada(http.MethodGet, "/meus-planos", autorizacao.PlanoLerProprio, planoHandler.ListarMeus)
	registro.Autenticada(http.MethodGet, "/meus-planos/:id", autorizacao.PlanoLerProprio, planoHandler.BuscarMeu)
	registro.Autenticada(http.MethodPost, "/meus-planos/:id/documentos", autorizacao.PlanoLerProprio, documentoHandler.GerarMeu)

	// Download do documento — alcançado tanto pelo PI quanto pelo
	// coordenador; a rota não escolhe entre os dois de antemão, o use
	// case tenta os dois alcances (design.md §7.3).
	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/documentos/:id/conteudo", documentoHandler.Baixar)

	// specs/metas-coordenacao — entrega, anexo, avaliação e relatório de
	// desempenho. Rotas alcançadas só pelo coordenador usam a permissão
	// fixa (entrega.registrar); as alcançadas pelos dois perfis usam
	// AutenticadaPorEscopoProprio e o use case tenta os dois alcances
	// (mesmo padrão do documento acima).
	registro.Autenticada(http.MethodPost, "/itens/:itemId/entregas", autorizacao.EntregaRegistrar, entregaHandler.Registrar)
	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/itens/:itemId/entregas", entregaHandler.ListarDoItem)
	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/entregas/:id", entregaHandler.Buscar)
	registro.Autenticada(http.MethodPut, "/entregas/:id", autorizacao.EntregaRegistrar, entregaHandler.Corrigir)
	registro.Autenticada(http.MethodDelete, "/entregas/:id", autorizacao.EntregaRegistrar, entregaHandler.Excluir)
	registro.Autenticada(http.MethodPost, "/entregas/:id/pendencia-vista", autorizacao.EntregaRegistrar, entregaHandler.MarcarPendenciaVista)
	registro.Autenticada(http.MethodGet, "/minhas-metas", autorizacao.EntregaRegistrar, entregaHandler.MinhasMetas)
	registro.Autenticada(http.MethodGet, "/minhas-metas/periodos", autorizacao.EntregaRegistrar, entregaHandler.ListarPeriodos)
	registro.Autenticada(http.MethodPost, "/entregas/:id/anexos", autorizacao.EntregaRegistrar, entregaHandler.AdicionarAnexo)
	registro.Autenticada(http.MethodDelete, "/anexos/:id", autorizacao.EntregaRegistrar, entregaHandler.RemoverAnexo)
	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/anexos/:id/conteudo", entregaHandler.BaixarAnexo)

	registro.Autenticada(http.MethodGet, "/avaliacoes", autorizacao.EntregaListar, avaliacaoHandler.ListarFila)
	registro.Autenticada(http.MethodPost, "/entregas/:id/avaliacao", autorizacao.EntregaAvaliar, avaliacaoHandler.Avaliar)
	registro.Autenticada(http.MethodPost, "/entregas/:id/desfazer-aceitacao", autorizacao.EntregaAvaliar, avaliacaoHandler.DesfazerAceitacao)

	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/relatorios/desempenho", relatorioHandler.Desempenho)
	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/relatorios/desempenho/por-curso", relatorioHandler.PorCurso)
	registro.Autenticada(http.MethodGet, "/relatorios/desempenho/exportacao", autorizacao.RelatorioExportar, relatorioHandler.Exportar)

	registro.AutenticadaPorEscopoProprio(http.MethodGet, "/metas/pendencias", pendenciaHandler.ContarBadges)

	// O relê em segundo plano (fundacao-metas.md §7): notificação de
	// recusa/desfazimento/restauração de prazo, e restauração de prazo por
	// vacância. Desligável por configuração, sem deploy (RELE_HABILITADO).
	if releHabilitado {
		releIntervalo := 30 * time.Second
		releServico := rele.Novo(entregaRepo, emailSender, auditLogger, relogioReal, adaptermetricas.NovoRele(), fusoDeExibicao, releIntervalo)
		releCtx, cancelarRele := context.WithCancel(context.Background())
		defer cancelarRele()
		go releServico.Iniciar(releCtx)
	} else {
		log.Print("rele de notificações desabilitado por configuração (RELE_HABILITADO=false)")
	}

	go func() {
		log.Printf("porta administrativa ouvindo em %s (/metrics, /healthz, /readyz)", adminPorta)
		if err := adminRouter.Run(":" + adminPorta); err != nil {
			log.Fatalf("servidor administrativo encerrado: %v", err)
		}
	}()

	log.Printf("basis-avalia backend ouvindo na porta %s (APP_ENV=%s)", porta, appEnv)
	if err := router.Run(":" + porta); err != nil {
		log.Fatalf("servidor encerrado: %v", err)
	}
}

func corsMiddleware(origensPermitidas []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Vary: Origin — sem isto, um cache compartilhado à frente da API
		// pode servir a resposta liberada para uma origem a outra
		// (design.md §4.3, achado O-2).
		c.Header("Vary", "Origin")
		origem := c.Request.Header.Get("Origin")
		for _, permitida := range origensPermitidas {
			if permitida != "" && permitida == origem {
				c.Header("Access-Control-Allow-Origin", origem)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
				break
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func requireEnv(chave string) string {
	valor := os.Getenv(chave)
	if valor == "" {
		log.Fatalf("variável de ambiente obrigatória ausente: %s", chave)
	}
	return valor
}

func getEnv(chave, padrao string) string {
	if valor := os.Getenv(chave); valor != "" {
		return valor
	}
	return padrao
}

func getEnvInt(chave string, padrao int) int {
	valor := os.Getenv(chave)
	if valor == "" {
		return padrao
	}
	n, err := strconv.Atoi(valor)
	if err != nil {
		return padrao
	}
	return n
}
