package http

import (
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/gin-gonic/gin"
)

// Registro é a tabela de rotas: a permissão é obrigatória na assinatura de
// toda rota de negócio (design.md §6.1). Não existe forma de registrar uma
// rota autenticada sem declarar a permissão que ela exige.
//
// MiddlewareSessao e MiddlewareAutorizacao são injetados — este arquivo não
// depende da implementação concreta (que só existe a partir do Grupo 3),
// mantendo T-006 verificável isoladamente.
type Registro struct {
	grupo                   *gin.RouterGroup
	middlewareSessao        gin.HandlerFunc
	middlewareAutorizacao   func(autorizacao.Permissao) gin.HandlerFunc
	middlewareEscopoProprio gin.HandlerFunc
}

func NovoRegistro(grupo *gin.RouterGroup, middlewareSessao gin.HandlerFunc, middlewareAutorizacao func(autorizacao.Permissao) gin.HandlerFunc, middlewareEscopoProprio gin.HandlerFunc) *Registro {
	return &Registro{
		grupo:                   grupo,
		middlewareSessao:        middlewareSessao,
		middlewareAutorizacao:   middlewareAutorizacao,
		middlewareEscopoProprio: middlewareEscopoProprio,
	}
}

// Publica registra rota sem qualquer verificação de sessão. Existe uma
// única chamada em todo o sistema (o combo público de instituições, R1).
func (r *Registro) Publica(metodo, caminho string, h gin.HandlerFunc) {
	r.grupo.Handle(metodo, caminho, h)
}

// Autenticada registra rota que exige sessão válida e a permissão
// declarada — fonte da métrica e da auditoria de acesso_negado (§8.2).
func (r *Registro) Autenticada(metodo, caminho string, p autorizacao.Permissao, h gin.HandlerFunc) {
	r.grupo.Handle(metodo, caminho, r.middlewareSessao, r.middlewareAutorizacao(p), h)
}

// AutenticadaSemPermissao registra rota de sessão própria — sem permissão
// exigida porque age sobre a própria conta de quem chama. Só aceita
// caminho sob /auth/ ou /version — regra derivada do próprio caminho e
// verificada pelo guarda de rotas (cmd/api/guarda_rotas_test.go, T-128):
// rota de negócio registrada aqui quebra o build, porque perde a segunda
// declaração de intenção e a métrica de decisão que
// AutenticadaPorEscopoProprio (abaixo) preserva.
func (r *Registro) AutenticadaSemPermissao(metodo, caminho string, h gin.HandlerFunc) {
	r.grupo.Handle(metodo, caminho, r.middlewareSessao, h)
}

// AutenticadaPorEscopoProprio registra rota de recurso de negócio em que
// não existe permissão a conferir porque a autorização É o escopo: o
// alcance vem do vínculo do próprio ator (carteira do coordenador,
// indicadores da própria instituição), não de uma permissão concedida por
// perfil — design.md §6.1 (T-128). Distinta de AutenticadaSemPermissao,
// que cobre só a conta do próprio ator (auth/eu, auth/logout, auth/senha):
// esta cobre rota alcançada por mais de um alcance, em que o use case
// decide sozinho qual vale (ex.: /entregas/:id, alcançada tanto pelo PI
// quanto pelo coordenador — design.md §7.3). A distinção existe para que
// o revisor tenha, de novo, uma segunda declaração de intenção a comparar
// com o use case, e para que a decisão vire métrica
// (autorizacao_decisoes_total, rótulo escopo_proprio) em vez de
// desaparecer.
func (r *Registro) AutenticadaPorEscopoProprio(metodo, caminho string, h gin.HandlerFunc) {
	r.grupo.Handle(metodo, caminho, r.middlewareSessao, r.middlewareEscopoProprio, h)
}
