package http

import (
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	designacaocmd "github.com/basis-avalia/backend/internal/usecase/command/designacao"
	designacaoquery "github.com/basis-avalia/backend/internal/usecase/query/designacao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoDesignacaoAllowlist = map[string]bool{
	"data_inicio": true, "data_fim": true, "coordenador": true, "portaria": true,
}

// DesignacaoHandler serve /api/v1/cursos/{id}/designacoes,
// /api/v1/designacoes/{id} e /api/v1/designacoes/candidatos.
type DesignacaoHandler struct {
	criar            *designacaocmd.CriarDesignacaoUseCase
	atualizar        *designacaocmd.AtualizarDesignacaoUseCase
	excluir          *designacaocmd.ExcluirDesignacaoUseCase
	listarDoCurso    *designacaoquery.ListarDoCursoUseCase
	buscar           *designacaoquery.BuscarDesignacaoUseCase
	listarCandidatos *designacaoquery.ListarCandidatosUseCase
}

func NovoDesignacaoHandler(
	criar *designacaocmd.CriarDesignacaoUseCase,
	atualizar *designacaocmd.AtualizarDesignacaoUseCase,
	excluir *designacaocmd.ExcluirDesignacaoUseCase,
	listarDoCurso *designacaoquery.ListarDoCursoUseCase,
	buscar *designacaoquery.BuscarDesignacaoUseCase,
	listarCandidatos *designacaoquery.ListarCandidatosUseCase,
) *DesignacaoHandler {
	return &DesignacaoHandler{
		criar: criar, atualizar: atualizar, excluir: excluir,
		listarDoCurso: listarDoCurso, buscar: buscar, listarCandidatos: listarCandidatos,
	}
}

func designacaoParaResposta(item port.ItemDesignacao, hoje valueobject.DataLocal) DesignacaoResponse {
	var dataFim *string
	if item.Designacao.Vigencia.Fim() != nil {
		s := item.Designacao.Vigencia.Fim().String()
		dataFim = &s
	}
	return DesignacaoResponse{
		ID: item.Designacao.ID.String(),
		Coordenador: CoordenadorDaDesignacaoResponse{
			ID: item.Designacao.CoordenadorID.String(), Nome: item.CoordenadorNome,
		},
		Portaria:                  item.Designacao.Portaria.String(),
		DataInicio:                item.Designacao.Vigencia.Inicio().String(),
		DataFim:                   dataFim,
		Situacao:                  string(item.Designacao.Vigencia.SituacaoEm(hoje)),
		Autodesignacao:            item.Designacao.Autodesignacao,
		OutrasDesignacoesVigentes: item.OutrasDesignacoesVigentes,
		Versao:                    item.Designacao.Versao,
	}
}

// ListarDoCurso godoc
// @Summary      Listar designações de um curso
// @Tags         designacoes
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/cursos/{id}/designacoes [get]
func (h *DesignacaoHandler) ListarDoCurso(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	cursoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoDesignacaoAllowlist, "data_inicio", "desc")
	if err != nil {
		c.Error(err)
		return
	}
	var coordenadorID *uuid.UUID
	if bruto := c.Query("coordenador_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "coordenador_id"})
			return
		}
		coordenadorID = &id
	}
	resultado, err := h.listarDoCurso.Executar(c.Request.Context(), designacaoquery.ListarDoCursoInput{
		Ator: ator, CursoID: cursoID, Situacao: c.DefaultQuery("situacao", "todas"), CoordenadorID: coordenadorID,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]DesignacaoResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, designacaoParaResposta(item, ator.DataDeReferencia()))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Candidatos godoc
// @Summary      Autocomplete de candidato a coordenador
// @Tags         designacoes
// @Produce      json
// @Success      200 {array} CandidatoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/designacoes/candidatos [get]
func (h *DesignacaoHandler) Candidatos(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	candidatos, err := h.listarCandidatos.Executar(c.Request.Context(), ator, c.Query("busca"))
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]CandidatoResponse, 0, len(candidatos))
	for _, cand := range candidatos {
		resposta = append(resposta, CandidatoResponse{ID: cand.ID.String(), Nome: cand.Nome, Email: cand.Email, Perfis: cand.Perfis})
	}
	c.JSON(http.StatusOK, resposta)
}

// Buscar godoc
// @Summary      Buscar designação por ID
// @Tags         designacoes
// @Produce      json
// @Success      200 {object} DesignacaoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/designacoes/{id} [get]
func (h *DesignacaoHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, designacaoParaResposta(item, ator.DataDeReferencia()))
}

// Criar godoc
// @Summary      Designar coordenador
// @Tags         designacoes
// @Accept       json
// @Produce      json
// @Success      201 {object} DesignacaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/cursos/{id}/designacoes [post]
func (h *DesignacaoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	cursoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	var req DesignacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	coordenadorID, err := uuid.Parse(req.CoordenadorID)
	if err != nil {
		c.Error(domain.ErrCoordenadorInvalido)
		return
	}
	nova, err := h.criar.Executar(c.Request.Context(), designacaocmd.CriarDesignacaoInput{
		Ator: ator, CursoID: cursoID, CoordenadorID: coordenadorID,
		Portaria: req.Portaria, DataInicio: req.DataInicio, DataFim: req.DataFim,
	})
	if err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, nova.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, designacaoParaResposta(item, ator.DataDeReferencia()))
}

// Atualizar godoc
// @Summary      Atualizar designação
// @Tags         designacoes
// @Accept       json
// @Produce      json
// @Success      200 {object} DesignacaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/designacoes/{id} [put]
func (h *DesignacaoHandler) Atualizar(c *gin.Context) {
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
	var req DesignacaoAtualizarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	coordenadorID, err := uuid.Parse(req.CoordenadorID)
	if err != nil {
		c.Error(domain.ErrCoordenadorInvalido)
		return
	}
	_, err = h.atualizar.Executar(c.Request.Context(), designacaocmd.AtualizarDesignacaoInput{
		Ator: ator, DesignacaoID: id, CoordenadorID: coordenadorID,
		Portaria: req.Portaria, DataInicio: req.DataInicio, DataFim: req.DataFim, Versao: req.Versao,
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
	c.JSON(http.StatusOK, designacaoParaResposta(item, ator.DataDeReferencia()))
}

// Excluir godoc
// @Summary      Excluir designação
// @Tags         designacoes
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/designacoes/{id} [delete]
func (h *DesignacaoHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), designacaocmd.ExcluirDesignacaoInput{Ator: ator, DesignacaoID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
