package http

import (
	"net/http"

	instituicaoquery "github.com/basis-avalia/backend/internal/usecase/query/instituicao"
	"github.com/gin-gonic/gin"
)

type InstituicaoPublicaHandler struct {
	listar *instituicaoquery.ListarInstituicoesPublicasUseCase
}

func NovoInstituicaoPublicaHandler(listar *instituicaoquery.ListarInstituicoesPublicasUseCase) *InstituicaoPublicaHandler {
	return &InstituicaoPublicaHandler{listar: listar}
}

// Listar godoc
// @Summary      Instituições disponíveis para login
// @Tags         publico
// @Produce      json
// @Success      200 {array} InstituicaoPublicaResponse
// @Router       /api/v1/publico/instituicoes [get]
func (h *InstituicaoPublicaHandler) Listar(c *gin.Context) {
	itens, err := h.listar.Executar(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]InstituicaoPublicaResponse, 0, len(itens))
	for _, item := range itens {
		resposta = append(resposta, InstituicaoPublicaResponse{ID: item.ID.String(), Nome: item.Nome, Sigla: item.Sigla})
	}
	c.JSON(http.StatusOK, resposta)
}
