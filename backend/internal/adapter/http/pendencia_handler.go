package http

import (
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	pendenciaquery "github.com/basis-avalia/backend/internal/usecase/query/pendencia"
	"github.com/gin-gonic/gin"
)

// PendenciaHandler serve os dois badges do menu (fundacao-metas.md,
// specs/metas-coordenacao/design.md §11).
type PendenciaHandler struct {
	contarBadges *pendenciaquery.ContarBadgesUseCase
}

func NovoPendenciaHandler(contarBadges *pendenciaquery.ContarBadgesUseCase) *PendenciaHandler {
	return &PendenciaHandler{contarBadges: contarBadges}
}

// ContarBadges godoc
// @Summary      Contar as pendências de avaliação e de correção (os dois badges do menu)
// @Tags         metas
// @Produce      json
// @Success      200 {object} PendenciasResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/metas/pendencias [get]
func (h *PendenciaHandler) ContarBadges(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	resultado, err := h.contarBadges.Executar(c.Request.Context(), ator)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, PendenciasResponse{
		PendentesDeAvaliacao: resultado.PendentesDeAvaliacao, PendenciasNaoVistas: resultado.PendenciasNaoVistas,
	})
}
