package http

import (
	"net"
	"net/http"
	"net/netip"

	"github.com/basis-avalia/backend/internal/adapter/metricas"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/gin-gonic/gin"
)

func enderecoDeOrigem(c *gin.Context) netip.Addr {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.Request.RemoteAddr
	}
	endereco, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return endereco
}

// NovoMiddlewareAutorizacao devolve uma fábrica de middleware por
// permissão — é o que a tabela de rotas (§6.1) injeta em Registro. Confere
// a permissão declarada contra o perfil já resolvido pelo middleware de
// sessão, registra a métrica de decisão e, quando negado, a auditoria de
// acesso_negado (design.md §8.2, A-04).
func NovoMiddlewareAutorizacao(audit port.AuditLogger) func(autorizacao.Permissao) gin.HandlerFunc {
	return func(permissao autorizacao.Permissao) gin.HandlerFunc {
		return func(c *gin.Context) {
			ator, ok := AtorDoContexto(c)
			if !ok {
				responderErroDeSessao(c, "SESSAO_EXPIRADA", "Sua sessão expirou. Entre novamente.")
				return
			}

			if !ator.Perfis().Pode(permissao) {
				metricas.RegistrarNegado(permissao)

				usuarioID := ator.UsuarioID()
				evento := auditoria.NovoEvento(auditoria.AcessoNegado, auditoria.ResultadoNegado)
				evento.AtorID = &usuarioID
				evento.InstituicaoID = ator.InstituicaoID()
				evento.Detalhes["permissao_exigida"] = string(permissao)
				evento.IPOrigem = enderecoDeOrigem(c)
				_ = audit.Registrar(c.Request.Context(), evento)

				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": gin.H{
					"code": "PERMISSAO_NEGADA", "message": "Você não tem permissão para executar esta ação.",
				}})
				return
			}

			metricas.RegistrarPermitido(permissao)
			c.Next()
		}
	}
}

// NovoMiddlewareEscopoProprio devolve o middleware das rotas em que não há
// permissão a conferir porque a autorização É o escopo (T-128, design.md
// §6.1): o alcance vem do vínculo do próprio ator — carteira do
// coordenador, indicadores da própria instituição — resolvido dentro do
// use case (Escopo/Proprio), nunca aqui. Só registra a decisão, sempre
// "permitido": quem chegou aqui já passou pelo middleware de sessão; a
// rejeição, se houver, é 404 por escopo dentro do use case, não 403 por
// permissão — não existe ramo "negado" para registrar.
func NovoMiddlewareEscopoProprio() gin.HandlerFunc {
	return func(c *gin.Context) {
		metricas.RegistrarPermitido(autorizacao.EscopoProprio)
		c.Next()
	}
}
