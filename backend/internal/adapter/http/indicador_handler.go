package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	indicadorcmd "github.com/basis-avalia/backend/internal/usecase/command/indicador"
	indicadorquery "github.com/basis-avalia/backend/internal/usecase/query/indicador"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoIndicadorAllowlist = map[string]bool{"codigo": true, "nome": true, "escopo": true, "criado_em": true}

// IndicadorHandler serve /api/v1/indicadores — lê os dois escopos, escreve
// só no próprio (design.md §6, I-07).
type IndicadorHandler struct {
	criar           *indicadorcmd.CriarIndicadorUseCase
	atualizar       *indicadorcmd.AtualizarIndicadorUseCase
	alterarSituacao *indicadorcmd.AlterarSituacaoIndicadorUseCase
	excluir         *indicadorcmd.ExcluirIndicadorUseCase
	listar          *indicadorquery.ListarDoCatalogoUseCase
	buscar          *indicadorquery.BuscarIndicadorUseCase
	sugerir         *indicadorquery.SugerirIndicadorUseCase
}

func NovoIndicadorHandler(
	criar *indicadorcmd.CriarIndicadorUseCase,
	atualizar *indicadorcmd.AtualizarIndicadorUseCase,
	alterarSituacao *indicadorcmd.AlterarSituacaoIndicadorUseCase,
	excluir *indicadorcmd.ExcluirIndicadorUseCase,
	listar *indicadorquery.ListarDoCatalogoUseCase,
	buscar *indicadorquery.BuscarIndicadorUseCase,
	sugerir *indicadorquery.SugerirIndicadorUseCase,
) *IndicadorHandler {
	return &IndicadorHandler{criar: criar, atualizar: atualizar, alterarSituacao: alterarSituacao, excluir: excluir, listar: listar, buscar: buscar, sugerir: sugerir}
}

func indicadorParaResposta(item port.ItemIndicador) IndicadorResponse {
	var atualizadoEm *string
	if item.Indicador.AtualizadoEm != nil {
		s := item.Indicador.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	var referencia *string
	if item.Indicador.ReferenciaInstrumento != nil {
		s := item.Indicador.ReferenciaInstrumento.String()
		referencia = &s
	}
	return IndicadorResponse{
		ID: item.Indicador.ID.String(), Escopo: string(item.Indicador.Escopo), Codigo: item.Indicador.Codigo.String(),
		Nome: item.Indicador.Nome.String(), Descricao: item.Indicador.Descricao, ReferenciaInstrumento: referencia,
		Situacao: string(item.Indicador.Situacao), Metas: item.MetasDaInstituicao,
		CriadoEm: item.Indicador.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: item.Indicador.Versao,
	}
}

// Listar godoc
// @Summary      Listar indicadores (catálogo do INEP + próprios)
// @Tags         indicadores
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/indicadores [get]
func (h *IndicadorHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoIndicadorAllowlist, "codigo", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	origem := c.DefaultQuery("origem", "todos")
	if origem != "todos" && origem != "plataforma" && origem != "instituicao" {
		c.Error(&domain.ErrParametro{Nome: "origem"})
		return
	}
	situacao := c.DefaultQuery("situacao", "ativo")

	resultado, err := h.listar.Executar(c.Request.Context(), indicadorquery.ListarDoCatalogoInput{
		Ator: ator, Busca: c.Query("busca"), Origem: origem, Situacao: situacao,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]IndicadorResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, indicadorParaResposta(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Sugerir godoc
// @Summary      Autocomplete de indicador
// @Description  Mistura os dois escopos, só ativos, limitado a 20. Sem criação inline.
// @Tags         indicadores
// @Produce      json
// @Success      200 {array} IndicadorSugestaoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/indicadores/sugestoes [get]
func (h *IndicadorHandler) Sugerir(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	itens, err := h.sugerir.Executar(c.Request.Context(), ator, c.Query("busca"))
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]IndicadorSugestaoResponse, 0, len(itens))
	for _, item := range itens {
		resposta = append(resposta, IndicadorSugestaoResponse{
			ID: item.ID.String(), Codigo: item.Codigo, Nome: item.Nome, Escopo: item.Escopo, ReferenciaInstrumento: item.ReferenciaInstrumento,
		})
	}
	c.JSON(http.StatusOK, resposta)
}

// Buscar godoc
// @Summary      Buscar indicador por ID
// @Tags         indicadores
// @Produce      json
// @Success      200 {object} IndicadorResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/indicadores/{id} [get]
func (h *IndicadorHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, indicadorParaResposta(item))
}

// Criar godoc
// @Summary      Cadastrar indicador próprio
// @Tags         indicadores
// @Accept       json
// @Produce      json
// @Success      201 {object} IndicadorResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/indicadores [post]
func (h *IndicadorHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req IndicadorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	novo, err := h.criar.Executar(c.Request.Context(), indicadorcmd.CriarIndicadorInput{
		Ator: ator, Codigo: req.Codigo, Nome: req.Nome, Descricao: req.Descricao,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, indicadorParaResposta(port.ItemIndicador{Indicador: novo}))
}

// Atualizar godoc
// @Summary      Atualizar indicador próprio
// @Tags         indicadores
// @Accept       json
// @Produce      json
// @Success      200 {object} IndicadorResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/indicadores/{id} [put]
func (h *IndicadorHandler) Atualizar(c *gin.Context) {
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
	var req IndicadorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	atualizado, err := h.atualizar.Executar(c.Request.Context(), indicadorcmd.AtualizarIndicadorInput{
		Ator: ator, IndicadorID: id, Codigo: req.Codigo, Nome: req.Nome, Descricao: req.Descricao, Versao: req.Versao,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, indicadorParaResposta(port.ItemIndicador{Indicador: atualizado}))
}

// AlterarSituacao godoc
// @Summary      Inativar ou reativar indicador próprio
// @Tags         indicadores
// @Accept       json
// @Produce      json
// @Success      200 {object} IndicadorResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/indicadores/{id}/situacao [patch]
func (h *IndicadorHandler) AlterarSituacao(c *gin.Context) {
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
	var req AlterarSituacaoCatalogoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	nova, err := valueobject.NovaSituacaoCatalogo(req.Situacao)
	if err != nil {
		c.Error(err)
		return
	}
	if err := h.alterarSituacao.Executar(c.Request.Context(), indicadorcmd.AlterarSituacaoIndicadorInput{
		Ator: ator, IndicadorID: id, NovaSituacao: nova, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, indicadorParaResposta(item))
}

// Excluir godoc
// @Summary      Excluir indicador próprio
// @Tags         indicadores
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/indicadores/{id} [delete]
func (h *IndicadorHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), indicadorcmd.ExcluirIndicadorInput{Ator: ator, IndicadorID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
