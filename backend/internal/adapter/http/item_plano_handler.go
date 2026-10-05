package http

import (
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	itemplanocmd "github.com/basis-avalia/backend/internal/usecase/command/item_plano"
	planoquery "github.com/basis-avalia/backend/internal/usecase/query/plano"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ItemPlanoHandler serve /api/v1/planos/{id}/itens
// (specs/plano-acao/design.md §6).
type ItemPlanoHandler struct {
	criar       *itemplanocmd.CriarItemUseCase
	atualizar   *itemplanocmd.AtualizarItemUseCase
	excluir     *itemplanocmd.ExcluirItemUseCase
	buscarPlano *planoquery.BuscarPlanoUseCase
}

func NovoItemPlanoHandler(
	criar *itemplanocmd.CriarItemUseCase, atualizar *itemplanocmd.AtualizarItemUseCase, excluir *itemplanocmd.ExcluirItemUseCase,
	buscarPlano *planoquery.BuscarPlanoUseCase,
) *ItemPlanoHandler {
	return &ItemPlanoHandler{criar: criar, atualizar: atualizar, excluir: excluir, buscarPlano: buscarPlano}
}

func (h *ItemPlanoHandler) respostaDoPlano(c *gin.Context, planoID uuid.UUID) {
	ator, _ := AtorDoContexto(c)
	item, err := h.buscarPlano.Executar(c.Request.Context(), ator, autorizacao.PlanosDaInstituicao, planoID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, detalhePlanoParaResposta(item))
}

// Criar godoc
// @Summary      Adicionar meta ao plano
// @Tags         itens-do-plano
// @Accept       json
// @Produce      json
// @Success      201 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/itens [post]
func (h *ItemPlanoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	planoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	var req ItemPlanoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	metaID, err := uuid.Parse(req.MetaID)
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	if _, err := h.criar.Executar(c.Request.Context(), itemplanocmd.CriarItemInput{
		Ator: ator, PlanoID: planoID, MetaID: metaID, Quantidade: req.Quantidade,
	}); err != nil {
		c.Error(err)
		return
	}
	h.respostaDoPlano(c, planoID)
}

// Atualizar godoc
// @Summary      Alterar a quantidade de um item do plano
// @Tags         itens-do-plano
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/itens/{item_id} [put]
func (h *ItemPlanoHandler) Atualizar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	planoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	var req ItemPlanoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if _, err := h.atualizar.Executar(c.Request.Context(), itemplanocmd.AtualizarItemInput{
		Ator: ator, PlanoID: planoID, ItemID: itemID, Quantidade: req.Quantidade, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	h.respostaDoPlano(c, planoID)
}

// Excluir godoc
// @Summary      Remover meta do plano
// @Tags         itens-do-plano
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/itens/{item_id} [delete]
func (h *ItemPlanoHandler) Excluir(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	planoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	if err := h.excluir.Executar(c.Request.Context(), itemplanocmd.ExcluirItemInput{Ator: ator, PlanoID: planoID, ItemID: itemID}); err != nil {
		c.Error(err)
		return
	}
	h.respostaDoPlano(c, planoID)
}
