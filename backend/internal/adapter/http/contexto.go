package http

import (
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/gin-gonic/gin"
)

const (
	chaveAtor             = "basis_avalia.ator"
	chaveContextoDeSessao = "basis_avalia.contexto_de_sessao"
)

// AtorDoContexto recupera o Ator resolvido pelo middleware de sessão —
// nunca reconstruído a partir do token dentro do handler.
func AtorDoContexto(c *gin.Context) (autorizacao.Ator, bool) {
	v, existe := c.Get(chaveAtor)
	if !existe {
		return autorizacao.Ator{}, false
	}
	ator, ok := v.(autorizacao.Ator)
	return ator, ok
}

// ContextoDeSessaoDoContexto recupera o que o middleware já carregou do
// banco nesta requisição — evita uma segunda consulta idêntica em rotas
// como /auth/eu.
func ContextoDeSessaoDoContexto(c *gin.Context) (*port.ContextoDeSessao, bool) {
	v, existe := c.Get(chaveContextoDeSessao)
	if !existe {
		return nil, false
	}
	ctxSessao, ok := v.(*port.ContextoDeSessao)
	return ctxSessao, ok
}
