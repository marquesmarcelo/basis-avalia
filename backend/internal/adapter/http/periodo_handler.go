package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
	periodocmd "github.com/basis-avalia/backend/internal/usecase/command/periodo"
	periodoquery "github.com/basis-avalia/backend/internal/usecase/query/periodo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoPeriodoAllowlist = map[string]bool{"nome": true, "data_inicio": true, "data_fim": true}

// PeriodoHandler serve /api/v1/periodos (specs/plano-acao/design.md §6).
type PeriodoHandler struct {
	criar     *periodocmd.CriarPeriodoUseCase
	atualizar *periodocmd.AtualizarPeriodoUseCase
	excluir   *periodocmd.ExcluirPeriodoUseCase
	listar    *periodoquery.ListarPeriodosUseCase
	buscar    *periodoquery.BuscarPeriodoUseCase
}

func NovoPeriodoHandler(
	criar *periodocmd.CriarPeriodoUseCase, atualizar *periodocmd.AtualizarPeriodoUseCase, excluir *periodocmd.ExcluirPeriodoUseCase,
	listar *periodoquery.ListarPeriodosUseCase, buscar *periodoquery.BuscarPeriodoUseCase,
) *PeriodoHandler {
	return &PeriodoHandler{criar: criar, atualizar: atualizar, excluir: excluir, listar: listar, buscar: buscar}
}

func periodoParaResposta(item port.ItemPeriodo, hoje time.Time) PeriodoResponse {
	var atualizadoEm *string
	if item.Periodo.AtualizadoEm != nil {
		s := item.Periodo.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	fimTexto := ""
	if fim := item.Periodo.Vigencia.Fim(); fim != nil {
		fimTexto = fim.String()
	}
	return PeriodoResponse{
		ID: item.Periodo.ID.String(), Nome: item.Periodo.Nome.String(),
		DataInicio: item.Periodo.Vigencia.Inicio().String(), DataFim: fimTexto,
		Planos: item.Planos, CriadoEm: item.Periodo.CriadoEm.Format(time.RFC3339),
		AtualizadoEm: atualizadoEm, Versao: item.Periodo.Versao,
	}
}

// Listar godoc
// @Summary      Listar períodos
// @Tags         periodos
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/periodos [get]
func (h *PeriodoHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoPeriodoAllowlist, "data_inicio", "desc")
	if err != nil {
		c.Error(err)
		return
	}
	resultado, err := h.listar.Executar(c.Request.Context(), periodoquery.ListarPeriodosInput{
		Ator: ator, Nome: c.Query("nome"), Situacao: c.DefaultQuery("situacao", "todas"),
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]PeriodoResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, periodoParaResposta(item, time.Now()))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Buscar godoc
// @Summary      Buscar período por ID
// @Tags         periodos
// @Produce      json
// @Success      200 {object} PeriodoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/periodos/{id} [get]
func (h *PeriodoHandler) Buscar(c *gin.Context) {
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
	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, periodoParaResposta(item, time.Now()))
}

// Criar godoc
// @Summary      Cadastrar período
// @Tags         periodos
// @Accept       json
// @Produce      json
// @Success      201 {object} PeriodoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/periodos [post]
func (h *PeriodoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req PeriodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	novo, err := h.criar.Executar(c.Request.Context(), periodocmd.CriarPeriodoInput{
		Ator: ator, Nome: req.Nome, DataInicio: req.DataInicio, DataFim: req.DataFim,
	})
	if err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, novo.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, periodoParaResposta(item, time.Now()))
}

// Atualizar godoc
// @Summary      Atualizar período
// @Tags         periodos
// @Accept       json
// @Produce      json
// @Success      200 {object} PeriodoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/periodos/{id} [put]
func (h *PeriodoHandler) Atualizar(c *gin.Context) {
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
	var req PeriodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	_, err = h.atualizar.Executar(c.Request.Context(), periodocmd.AtualizarPeriodoInput{
		Ator: ator, PeriodoID: id, Nome: req.Nome, DataInicio: req.DataInicio, DataFim: req.DataFim, Versao: req.Versao,
	})
	if err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, periodoParaResposta(item, time.Now()))
}

// Excluir godoc
// @Summary      Excluir período
// @Tags         periodos
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/periodos/{id} [delete]
func (h *PeriodoHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), periodocmd.ExcluirPeriodoInput{Ator: ator, PeriodoID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
