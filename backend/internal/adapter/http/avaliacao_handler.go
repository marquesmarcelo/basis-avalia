package http

import (
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
	avaliacaocmd "github.com/basis-avalia/backend/internal/usecase/command/avaliacao"
	entregaquery "github.com/basis-avalia/backend/internal/usecase/query/entrega"
	avaliacaoquery "github.com/basis-avalia/backend/internal/usecase/query/avaliacao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AvaliacaoHandler serve a fila de avaliação e os dois comandos do PI
// (avaliar, desfazer aceitação) — specs/metas-coordenacao/design.md §11.
type AvaliacaoHandler struct {
	avaliar          *avaliacaocmd.AvaliarUseCase
	desfazer         *avaliacaocmd.DesfazerAceitacaoUseCase
	listarFila       *avaliacaoquery.ListarFilaUseCase
	buscar           *entregaquery.BuscarUseCase
}

func NovoAvaliacaoHandler(
	avaliar *avaliacaocmd.AvaliarUseCase, desfazer *avaliacaocmd.DesfazerAceitacaoUseCase,
	listarFila *avaliacaoquery.ListarFilaUseCase, buscar *entregaquery.BuscarUseCase,
) *AvaliacaoHandler {
	return &AvaliacaoHandler{avaliar: avaliar, desfazer: desfazer, listarFila: listarFila, buscar: buscar}
}

// ListarFila godoc
// @Summary      Listar a fila de avaliação de entregas (Pesquisador Institucional)
// @Tags         avaliacoes
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/avaliacoes [get]
func (h *AvaliacaoHandler) ListarFila(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoEntregaAllowlist, "criado_em", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	filtro := port.FiltroFilaAvaliacao{
		Situacao: c.Query("situacao"), Page: paginacao.Page, PageSize: paginacao.PageSize,
		Sort: paginacao.Sort, Order: paginacao.Order,
	}
	if bruto := c.Query("periodo_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "periodo_id"})
			return
		}
		filtro.PeriodoID = &id
	}
	if bruto := c.Query("curso_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "curso_id"})
			return
		}
		filtro.CursoID = &id
	}
	if bruto := c.Query("meta_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "meta_id"})
			return
		}
		filtro.MetaID = &id
	}
	resultado, err := h.listarFila.Executar(c.Request.Context(), avaliacaoquery.ListarFilaInput{Ator: ator, Filtro: filtro})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]EntregaResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, entregaParaResposta(port.DetalheEntrega{LinhaEntrega: item}))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Avaliar godoc
// @Summary      Aceitar ou recusar uma entrega
// @Tags         avaliacoes
// @Accept       json
// @Produce      json
// @Success      200 {object} EntregaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/entregas/{id}/avaliacao [post]
func (h *AvaliacaoHandler) Avaliar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	var req AvaliarRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.avaliar.Executar(c.Request.Context(), avaliacaocmd.AvaliarInput{
		Ator: ator, EntregaID: id, Resultado: req.Resultado, Motivo: req.Motivo, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	detalhe, err := h.buscar.Executar(c.Request.Context(), entregaquery.BuscarInput{Ator: ator, EntregaID: id})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, entregaParaResposta(detalhe))
}

// DesfazerAceitacao godoc
// @Summary      Desfazer a aceitação de uma entrega
// @Tags         avaliacoes
// @Accept       json
// @Produce      json
// @Success      200 {object} EntregaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/entregas/{id}/desfazer-aceitacao [post]
func (h *AvaliacaoHandler) DesfazerAceitacao(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	var req DesfazerAceitacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.desfazer.Executar(c.Request.Context(), avaliacaocmd.DesfazerAceitacaoInput{
		Ator: ator, EntregaID: id, Motivo: req.Motivo, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	detalhe, err := h.buscar.Executar(c.Request.Context(), entregaquery.BuscarInput{Ator: ator, EntregaID: id})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, entregaParaResposta(detalhe))
}
