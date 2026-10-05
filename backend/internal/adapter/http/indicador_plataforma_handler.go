package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	indicadorplataformacmd "github.com/basis-avalia/backend/internal/usecase/command/indicador_plataforma"
	indicadorplataformaquery "github.com/basis-avalia/backend/internal/usecase/query/indicador_plataforma"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoIndicadorPlataformaAllowlist = map[string]bool{"codigo": true, "nome": true, "escopo": true, "criado_em": true}

// IndicadorPlataformaHandler serve /api/v1/plataforma/indicadores —
// Administrador do Sistema, catálogo comum (design.md §6, I-07).
type IndicadorPlataformaHandler struct {
	criar           *indicadorplataformacmd.CriarIndicadorPlataformaUseCase
	atualizar       *indicadorplataformacmd.AtualizarIndicadorPlataformaUseCase
	alterarSituacao *indicadorplataformacmd.AlterarSituacaoIndicadorPlataformaUseCase
	excluir         *indicadorplataformacmd.ExcluirIndicadorPlataformaUseCase
	listar          *indicadorplataformaquery.ListarIndicadoresPlataformaUseCase
	buscar          *indicadorplataformaquery.BuscarIndicadorPlataformaUseCase
}

func NovoIndicadorPlataformaHandler(
	criar *indicadorplataformacmd.CriarIndicadorPlataformaUseCase,
	atualizar *indicadorplataformacmd.AtualizarIndicadorPlataformaUseCase,
	alterarSituacao *indicadorplataformacmd.AlterarSituacaoIndicadorPlataformaUseCase,
	excluir *indicadorplataformacmd.ExcluirIndicadorPlataformaUseCase,
	listar *indicadorplataformaquery.ListarIndicadoresPlataformaUseCase,
	buscar *indicadorplataformaquery.BuscarIndicadorPlataformaUseCase,
) *IndicadorPlataformaHandler {
	return &IndicadorPlataformaHandler{criar: criar, atualizar: atualizar, alterarSituacao: alterarSituacao, excluir: excluir, listar: listar, buscar: buscar}
}

// indicadorPlataformaParaResposta nunca inclui identificador nem nome de
// instituição — PI-5, IE-07: a resposta desta família não contém
// informação de instituição em campo nenhum.
func indicadorPlataformaParaResposta(item port.ItemIndicadorPlataforma) IndicadorPlataformaResponse {
	var atualizadoEm *string
	if item.Indicador.AtualizadoEm != nil {
		s := item.Indicador.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	referencia := ""
	if item.Indicador.ReferenciaInstrumento != nil {
		referencia = item.Indicador.ReferenciaInstrumento.String()
	}
	return IndicadorPlataformaResponse{
		ID: item.Indicador.ID.String(), Escopo: string(item.Indicador.Escopo), Codigo: item.Indicador.Codigo.String(),
		Nome: item.Indicador.Nome.String(), Descricao: item.Indicador.Descricao, ReferenciaInstrumento: referencia,
		Situacao: string(item.Indicador.Situacao), Metas: item.MetasTotal,
		CriadoEm: item.Indicador.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: item.Indicador.Versao,
	}
}

// Listar godoc
// @Summary      Listar o catálogo de indicadores do INEP
// @Tags         indicadores-plataforma
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores [get]
func (h *IndicadorPlataformaHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoIndicadorPlataformaAllowlist, "codigo", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	situacao := c.DefaultQuery("situacao", "ativo")

	resultado, err := h.listar.Executar(c.Request.Context(), indicadorplataformaquery.ListarIndicadoresPlataformaInput{
		Ator: ator, Busca: c.Query("busca"), Situacao: situacao,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]IndicadorPlataformaResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, indicadorPlataformaParaResposta(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Buscar godoc
// @Summary      Buscar indicador do INEP por ID
// @Tags         indicadores-plataforma
// @Produce      json
// @Success      200 {object} IndicadorPlataformaResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores/{id} [get]
func (h *IndicadorPlataformaHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, indicadorPlataformaParaResposta(item))
}

// Criar godoc
// @Summary      Cadastrar indicador do INEP
// @Description  Catálogo único da instalação — vale para todas as instituições imediatamente.
// @Tags         indicadores-plataforma
// @Accept       json
// @Produce      json
// @Success      201 {object} IndicadorPlataformaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores [post]
func (h *IndicadorPlataformaHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req IndicadorPlataformaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	novo, err := h.criar.Executar(c.Request.Context(), indicadorplataformacmd.CriarIndicadorPlataformaInput{
		Ator: ator, Codigo: req.Codigo, Nome: req.Nome, Descricao: req.Descricao, ReferenciaInstrumento: req.ReferenciaInstrumento,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, indicadorPlataformaParaResposta(port.ItemIndicadorPlataforma{Indicador: novo}))
}

// Atualizar godoc
// @Summary      Atualizar indicador do INEP
// @Tags         indicadores-plataforma
// @Accept       json
// @Produce      json
// @Success      200 {object} IndicadorPlataformaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores/{id} [put]
func (h *IndicadorPlataformaHandler) Atualizar(c *gin.Context) {
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
	var req IndicadorPlataformaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	atualizado, err := h.atualizar.Executar(c.Request.Context(), indicadorplataformacmd.AtualizarIndicadorPlataformaInput{
		Ator: ator, IndicadorID: id, Codigo: req.Codigo, Nome: req.Nome, Descricao: req.Descricao,
		ReferenciaInstrumento: req.ReferenciaInstrumento, Versao: req.Versao,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, indicadorPlataformaParaResposta(port.ItemIndicadorPlataforma{Indicador: atualizado}))
}

// AlterarSituacao godoc
// @Summary      Inativar ou reativar indicador do INEP
// @Tags         indicadores-plataforma
// @Accept       json
// @Produce      json
// @Success      200 {object} IndicadorPlataformaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores/{id}/situacao [patch]
func (h *IndicadorPlataformaHandler) AlterarSituacao(c *gin.Context) {
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
	if err := h.alterarSituacao.Executar(c.Request.Context(), indicadorplataformacmd.AlterarSituacaoIndicadorPlataformaInput{
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
	c.JSON(http.StatusOK, indicadorPlataformaParaResposta(item))
}

// Excluir godoc
// @Summary      Excluir indicador do INEP
// @Description  Bloqueado enquanto qualquer instituição o referenciar em meta.
// @Tags         indicadores-plataforma
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/plataforma/indicadores/{id} [delete]
func (h *IndicadorPlataformaHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), indicadorplataformacmd.ExcluirIndicadorPlataformaInput{Ator: ator, IndicadorID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
