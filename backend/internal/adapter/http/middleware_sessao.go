package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NomeCookieSessao é o nome do cookie de sessão (design.md §7.2).
const NomeCookieSessao = "basis_avalia_sessao"

// rotasPermitidasComSenhaProvisoria é o conjunto declarado ao lado da
// tabela de rotas — nunca um `if` dentro de cada handler (design.md §7.4).
var rotasPermitidasComSenhaProvisoria = map[string]bool{
	"GET /api/v1/auth/eu":      true,
	"POST /api/v1/auth/senha":  true,
	"POST /api/v1/auth/logout": true,
	"GET /version":             true,
}

func responderErroDeSessao(c *gin.Context, codigo, mensagem string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": codigo, "message": mensagem}})
}

// NovoMiddlewareSessao implementa a cadeia de decisão de design.md §2.2 e
// §7.3 — os três códigos de 401 (3.18) e a porta de senha provisória
// (3.4, §7.4). Este arquivo nunca consulta a variável de ambiente de execução (3.10): o bypass de
// desenvolvimento do CLAUDE.md não se aplica a esta feature.
// relogio e fusoDeExibicao existem para fixar dataDeReferencia (fundacao-
// metas.md §5.2) — uma única leitura de relógio por requisição, aqui, no
// caminho de sessão. Se o perfil derivado e o filtro de carteira lessem o
// relógio cada um por conta própria, uma requisição atravessando a meia-
// noite poderia produzir um ator COM o perfil de coordenador e SEM nenhum
// curso na carteira: um 403 que ninguém consegue reproduzir.
func NovoMiddlewareSessao(tokenDeSessao port.TokenDeSessao, repo port.AutenticacaoRepository, relogio port.Relogio, fusoDeExibicao *time.Location) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(NomeCookieSessao)
		if err != nil || cookie == "" {
			responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
			return
		}

		claims, err := tokenDeSessao.Validar(cookie)
		if err != nil {
			responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
			return
		}

		// Uma única leitura do relógio por requisição (fundacao-metas.md
		// §5.2), ANTES da consulta de sessão — ela agora recebe a data como
		// parâmetro para o EXISTS de designação vigente, e o Ator construído
		// abaixo usa a MESMA data. Ler o relógio de novo depois é exatamente
		// o que produziria o Ator com o perfil de coordenador e a carteira
		// de outro dia, perto da meia-noite de Brasília.
		dataDeReferencia := valueobject.DataLocalDe(relogio.Agora(), fusoDeExibicao)

		ctxSessao, err := repo.CarregarContextoDeSessao(c.Request.Context(), claims.UsuarioID, dataDeReferencia)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		// Linha ausente: mesmo código de token inválido — o específico só
		// é alcançável depois que a identidade foi provada (SE-08).
		if ctxSessao == nil {
			responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
			return
		}
		if ctxSessao.ExcluidoEm != nil {
			responderErroDeSessao(c, "CONTA_EXCLUIDA", "Sua conta foi removida. Entre em contato com quem administra o sistema.")
			return
		}

		if !mesmaInstituicao(claims.InstituicaoID, ctxSessao.InstituicaoID) {
			responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
			return
		}

		if ctxSessao.InstituicaoID != nil && ctxSessao.InstituicaoSituacao == "inativa" {
			responderErroDeSessao(c, "INSTITUICAO_INATIVA", "Sua instituição foi desativada. Entre em contato com o Administrador do Sistema.")
			return
		}

		emt := time.UnixMicro(claims.EmitidoEmMicro)
		if emt.Before(ctxSessao.SessoesValidasAPartirDe) {
			responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
			return
		}

		// O Ator usa o conjunto EFETIVO (atribuído + Coordenador de Curso
		// quando coordena hoje), nunca só o atribuído — CP-14, specs/cursos/
		// design.md, fundacao-metas.md §4.2: montarConjuntoEfetivo é a única
		// função que acrescenta CoordenadorCurso, e este é um dos dois
		// lugares sancionados a chamá-la (o outro é ObterContextoDeSessao).
		efetivo, err := autorizacao.MontarConjuntoEfetivo(ctxSessao.Perfis, ctxSessao.CoordenaHoje)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		// Conjunto vazio é estado impossível (design.md §5.6, 3.6): NovoAtor
		// falha e o erro sobe como 500 ERRO_INTERNO com log — nunca
		// SESSAO_EXPIRADA, que mandaria a pessoa para um laço de login sem
		// causa visível.
		ator, err := autorizacao.NovoAtor(ctxSessao.UsuarioID, efetivo, ctxSessao.InstituicaoID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		ator = ator.ComDataDeReferencia(dataDeReferencia)
		c.Set(chaveAtor, ator)
		c.Set(chaveContextoDeSessao, ctxSessao)

		if ctxSessao.SenhaProvisoria {
			chaveRota := c.Request.Method + " " + c.FullPath()
			if !rotasPermitidasComSenhaProvisoria[chaveRota] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
					"code": "SENHA_PROVISORIA", "message": "Defina uma senha própria para continuar.",
				}})
				return
			}
		}

		c.Next()
	}
}

// mesmaInstituicao confere a instituição do token contra o vínculo atual
// no banco (SE-07) — divergência (inclusive um dos dois nulo e o outro
// não) significa token obsoleto ou forjado.
func mesmaInstituicao(daClaim, doVinculo *uuid.UUID) bool {
	if daClaim == nil && doVinculo == nil {
		return true
	}
	if daClaim == nil || doVinculo == nil {
		return false
	}
	return *daClaim == *doVinculo
}
