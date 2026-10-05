package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	sessaocmd "github.com/basis-avalia/backend/internal/usecase/command/sessao"
	sessaoquery "github.com/basis-avalia/backend/internal/usecase/query/sessao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	autenticar          *sessaocmd.AutenticarUseCase
	encerrarSessao      *sessaocmd.EncerrarSessaoUseCase
	alterarSenhaPropria *sessaocmd.AlterarSenhaPropriaUseCase
	obterContexto       *sessaoquery.ObterContextoDeSessaoUseCase
	tokenDeSessao       port.TokenDeSessao
}

func NovoAuthHandler(
	autenticar *sessaocmd.AutenticarUseCase,
	encerrarSessao *sessaocmd.EncerrarSessaoUseCase,
	alterarSenhaPropria *sessaocmd.AlterarSenhaPropriaUseCase,
	obterContexto *sessaoquery.ObterContextoDeSessaoUseCase,
	tokenDeSessao port.TokenDeSessao,
) *AuthHandler {
	return &AuthHandler{
		autenticar:          autenticar,
		encerrarSessao:      encerrarSessao,
		alterarSenhaPropria: alterarSenhaPropria,
		obterContexto:       obterContexto,
		tokenDeSessao:       tokenDeSessao,
	}
}

func contextoParaResposta(out sessaoquery.ContextoDeSessaoOutput) ContextoDeSessaoResponse {
	var instituicao *InstituicaoResumoResponse
	if out.Instituicao != nil {
		instituicao = &InstituicaoResumoResponse{
			ID: out.Instituicao.ID.String(), Nome: out.Instituicao.Nome, Sigla: out.Instituicao.Sigla,
		}
	}
	permissoes := make([]string, len(out.Permissoes))
	for i, p := range out.Permissoes {
		permissoes[i] = string(p)
	}
	perfis := make([]string, len(out.Perfis))
	for i, p := range out.Perfis {
		perfis[i] = string(p)
	}
	// nunca nil na resposta (json:"perfis_derivados") — [] quando ninguém
	// coordena hoje, não a ausência do campo.
	derivados := make([]string, len(out.PerfisDerivados))
	for i, p := range out.PerfisDerivados {
		derivados[i] = string(p)
	}
	return ContextoDeSessaoResponse{
		ID: out.UsuarioID.String(), Nome: out.Nome, Email: out.Email,
		Perfis: perfis, PerfisRotulos: out.PerfisRotulos,
		PerfisDerivados: derivados, CursosCoordenados: out.CursosCoordenados,
		SenhaProvisoria: out.SenhaProvisoria, Instituicao: instituicao, Permissoes: permissoes,
	}
}

// definirCookieSessao — HttpOnly, Secure e SameSite=Strict em qualquer
// ambiente, sem ramificar por APP_ENV (design.md §7.2).
func definirCookieSessao(c *gin.Context, token string, expiraEm time.Time) {
	maxAge := int(time.Until(expiraEm).Seconds())
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(NomeCookieSessao, token, maxAge, "/", "", true, true)
}

func apagarCookieSessao(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(NomeCookieSessao, "", -1, "/", "", true, true)
}

// Login godoc
// @Summary      Entrar
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Credenciais"
// @Success      200 {object} ContextoDeSessaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}

	var instituicaoID *uuid.UUID
	if req.InstituicaoID != nil && *req.InstituicaoID != "" {
		id, err := uuid.Parse(*req.InstituicaoID)
		if err != nil {
			c.Error(&domain.ErrValidacao{Campo: "instituicao_id", Mensagem: "Requisição inválida."})
			return
		}
		instituicaoID = &id
	}

	email, err := valueobject.NovoEmail(req.Email)
	if err != nil {
		c.Error(err)
		return
	}
	senha, err := valueobject.NovaSenhaEmTexto(req.Senha)
	if err != nil {
		c.Error(err)
		return
	}

	out, err := h.autenticar.Executar(c.Request.Context(), sessaocmd.AutenticarInput{
		InstituicaoID: instituicaoID, Email: email, Senha: senha,
	})
	if err != nil {
		c.Error(err)
		return
	}

	token, expiraEm, err := h.tokenDeSessao.Emitir(out.Ator, out.InstanteDeAutenticacao)
	if err != nil {
		c.Error(err)
		return
	}
	definirCookieSessao(c, token, expiraEm)

	contexto, err := h.obterContexto.Executar(c.Request.Context(), out.Ator.UsuarioID(), out.Ator.DataDeReferencia())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, contextoParaResposta(contexto))
}

// Logout godoc
// @Summary      Sair
// @Tags         auth
// @Produce      json
// @Success      204
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if ator, ok := AtorDoContexto(c); ok {
		_ = h.encerrarSessao.Executar(c.Request.Context(), ator)
	}
	apagarCookieSessao(c)
	c.Status(http.StatusNoContent)
}

// Eu godoc
// @Summary      Quem sou eu
// @Tags         auth
// @Produce      json
// @Success      200 {object} ContextoDeSessaoResponse
// @Failure      401 {object} erroResposta
// @Router       /api/v1/auth/eu [get]
func (h *AuthHandler) Eu(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	contexto, err := h.obterContexto.Executar(c.Request.Context(), ator.UsuarioID(), ator.DataDeReferencia())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, contextoParaResposta(contexto))
}

// AlterarSenha godoc
// @Summary      Alterar a própria senha
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body AlterarSenhaRequest true "Senha atual e nova"
// @Success      200 {object} ContextoDeSessaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Router       /api/v1/auth/senha [post]
func (h *AuthHandler) AlterarSenha(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}

	var req AlterarSenhaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}

	senhaAtual, err := valueobject.NovaSenhaEmTexto(req.SenhaAtual)
	if err != nil {
		c.Error(err)
		return
	}
	senhaNova, err := valueobject.NovaSenhaEmTexto(req.SenhaNova)
	if err != nil {
		if errors.Is(err, domain.ErrSenhaObrigatoria) {
			c.Error(&domain.ErrValidacao{Campo: "senha_nova", Mensagem: "Informe a senha."})
			return
		}
		c.Error(err)
		return
	}

	out, err := h.alterarSenhaPropria.Executar(c.Request.Context(), sessaocmd.AlterarSenhaPropriaInput{
		Ator: ator, SenhaAtual: senhaAtual, SenhaNova: senhaNova,
	})
	if err != nil {
		c.Error(err)
		return
	}

	token, expiraEm, err := h.tokenDeSessao.Emitir(ator, out.Instante)
	if err != nil {
		c.Error(err)
		return
	}
	definirCookieSessao(c, token, expiraEm)

	contexto, err := h.obterContexto.Executar(c.Request.Context(), ator.UsuarioID(), ator.DataDeReferencia())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, contextoParaResposta(contexto))
}
