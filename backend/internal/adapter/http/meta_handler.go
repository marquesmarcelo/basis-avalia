package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	metacmd "github.com/basis-avalia/backend/internal/usecase/command/meta"
	metaquery "github.com/basis-avalia/backend/internal/usecase/query/meta"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoMetaAllowlist = map[string]bool{"nome": true, "criado_em": true}

// MetaHandler serve /api/v1/metas — CRUD do catálogo de metas da
// instituição (PI); o Coordenador lê como contexto, embutido noutras
// rotas de specs/metas-coordenacao.
type MetaHandler struct {
	criar           *metacmd.CriarMetaUseCase
	atualizar       *metacmd.AtualizarMetaUseCase
	alterarSituacao *metacmd.AlterarSituacaoMetaUseCase
	excluir         *metacmd.ExcluirMetaUseCase
	listar          *metaquery.ListarMetasUseCase
	buscar          *metaquery.BuscarMetaUseCase
	sugerir         *metaquery.SugerirMetaUseCase
}

func NovoMetaHandler(
	criar *metacmd.CriarMetaUseCase,
	atualizar *metacmd.AtualizarMetaUseCase,
	alterarSituacao *metacmd.AlterarSituacaoMetaUseCase,
	excluir *metacmd.ExcluirMetaUseCase,
	listar *metaquery.ListarMetasUseCase,
	buscar *metaquery.BuscarMetaUseCase,
	sugerir *metaquery.SugerirMetaUseCase,
) *MetaHandler {
	return &MetaHandler{criar: criar, atualizar: atualizar, alterarSituacao: alterarSituacao, excluir: excluir, listar: listar, buscar: buscar, sugerir: sugerir}
}

func indicadoresEmbutidosParaResposta(indicadores []port.IndicadorDaMeta) []IndicadorEmbutidoResponse {
	resposta := make([]IndicadorEmbutidoResponse, 0, len(indicadores))
	for _, i := range indicadores {
		resposta = append(resposta, IndicadorEmbutidoResponse{
			ID: i.ID.String(), Codigo: i.Codigo, Nome: i.Nome, Escopo: i.Escopo,
			ReferenciaInstrumento: i.ReferenciaInstrumento, Situacao: i.Situacao,
		})
	}
	return resposta
}

func metaParaResposta(item port.ItemMeta) MetaResponse {
	var atualizadoEm *string
	if item.Meta.AtualizadoEm != nil {
		s := item.Meta.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	var quantidadeSugerida *int
	if item.Meta.QuantidadeSugerida != nil {
		v := item.Meta.QuantidadeSugerida.Int()
		quantidadeSugerida = &v
	}
	return MetaResponse{
		ID: item.Meta.ID.String(), Nome: item.Meta.Nome.String(), Descricao: item.Meta.Descricao,
		Situacao: string(item.Meta.Situacao), Planos: item.Planos, Indicadores: indicadoresEmbutidosParaResposta(item.Indicadores),
		QuantidadeSugerida: quantidadeSugerida,
		CriadoEm:           item.Meta.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: item.Meta.Versao,
	}
}

func indicadoresDoRequest(brutos []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(brutos))
	for i, bruto := range brutos {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return nil, &domain.ErrValidacao{Campo: "indicadores", Mensagem: "Indicador inválido."}
		}
		ids[i] = id
	}
	return ids, nil
}

// Listar godoc
// @Summary      Listar metas
// @Tags         metas
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/metas [get]
func (h *MetaHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoMetaAllowlist, "nome", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	origem := c.DefaultQuery("origem", "todas")
	if origem != "todas" && origem != "plataforma" && origem != "instituicao" {
		c.Error(&domain.ErrParametro{Nome: "origem"})
		return
	}
	situacao := c.DefaultQuery("situacao", "ativo")

	var indicadorID *uuid.UUID
	if bruto := c.Query("indicador_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "indicador_id"})
			return
		}
		indicadorID = &id
	}

	resultado, err := h.listar.Executar(c.Request.Context(), metaquery.ListarMetasInput{
		Ator: ator, Busca: c.Query("busca"), IndicadorID: indicadorID, Origem: origem, Situacao: situacao,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]MetaResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, metaParaResposta(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Sugerir godoc
// @Summary      Autocomplete de meta
// @Tags         metas
// @Produce      json
// @Success      200 {array} MetaSugestaoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/metas/sugestoes [get]
func (h *MetaHandler) Sugerir(c *gin.Context) {
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
	resposta := make([]MetaSugestaoResponse, 0, len(itens))
	for _, item := range itens {
		resposta = append(resposta, MetaSugestaoResponse{
			ID: item.ID.String(), Nome: item.Nome, Indicadores: indicadoresEmbutidosParaResposta(item.Indicadores),
			QuantidadeSugerida: item.QuantidadeSugerida,
		})
	}
	c.JSON(http.StatusOK, resposta)
}

// Buscar godoc
// @Summary      Buscar meta por ID
// @Tags         metas
// @Produce      json
// @Success      200 {object} MetaResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/metas/{id} [get]
func (h *MetaHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, metaParaResposta(item))
}

// Criar godoc
// @Summary      Cadastrar meta
// @Tags         metas
// @Accept       json
// @Produce      json
// @Success      201 {object} MetaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/metas [post]
func (h *MetaHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req MetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	indicadores, err := indicadoresDoRequest(req.Indicadores)
	if err != nil {
		c.Error(err)
		return
	}
	nova, err := h.criar.Executar(c.Request.Context(), metacmd.CriarMetaInput{
		Ator: ator, Nome: req.Nome, Descricao: req.Descricao, Indicadores: indicadores,
		QuantidadeSugerida: req.QuantidadeSugerida,
	})
	if err != nil {
		c.Error(err)
		return
	}
	// Reconsultada para trazer os indicadores embutidos (código, nome,
	// origem, referência) — o use case devolve só a entidade de domínio,
	// com os indicadores como IDs.
	item, err := h.buscar.Executar(c.Request.Context(), ator, nova.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, metaParaResposta(item))
}

// Atualizar godoc
// @Summary      Atualizar meta
// @Tags         metas
// @Accept       json
// @Produce      json
// @Success      200 {object} MetaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/metas/{id} [put]
func (h *MetaHandler) Atualizar(c *gin.Context) {
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
	var req MetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	indicadores, err := indicadoresDoRequest(req.Indicadores)
	if err != nil {
		c.Error(err)
		return
	}
	_, err = h.atualizar.Executar(c.Request.Context(), metacmd.AtualizarMetaInput{
		Ator: ator, MetaID: id, Nome: req.Nome, Descricao: req.Descricao, Indicadores: indicadores,
		QuantidadeSugerida: req.QuantidadeSugerida, Versao: req.Versao,
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
	c.JSON(http.StatusOK, metaParaResposta(item))
}

// AlterarSituacao godoc
// @Summary      Inativar ou reativar meta
// @Tags         metas
// @Accept       json
// @Produce      json
// @Success      200 {object} MetaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/metas/{id}/situacao [patch]
func (h *MetaHandler) AlterarSituacao(c *gin.Context) {
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
	if err := h.alterarSituacao.Executar(c.Request.Context(), metacmd.AlterarSituacaoMetaInput{
		Ator: ator, MetaID: id, NovaSituacao: nova, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, metaParaResposta(item))
}

// Excluir godoc
// @Summary      Excluir meta
// @Tags         metas
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/metas/{id} [delete]
func (h *MetaHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), metacmd.ExcluirMetaInput{Ator: ator, MetaID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
