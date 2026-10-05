package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	cursocmd "github.com/basis-avalia/backend/internal/usecase/command/curso"
	cursoquery "github.com/basis-avalia/backend/internal/usecase/query/curso"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoCursoAllowlist = map[string]bool{
	"nome": true, "codigo_emec": true, "grau": true, "modalidade": true, "coordenador": true, "criado_em": true,
}

// CursoHandler serve /api/v1/cursos e /api/v1/meus-cursos.
type CursoHandler struct {
	criar           *cursocmd.CriarCursoUseCase
	atualizar       *cursocmd.AtualizarCursoUseCase
	alterarSituacao *cursocmd.AlterarSituacaoCursoUseCase
	excluir         *cursocmd.ExcluirCursoUseCase
	listar          *cursoquery.ListarCursosUseCase
	buscar          *cursoquery.BuscarCursoUseCase
	listarMeus      *cursoquery.ListarMeusCursosUseCase
}

func NovoCursoHandler(
	criar *cursocmd.CriarCursoUseCase,
	atualizar *cursocmd.AtualizarCursoUseCase,
	alterarSituacao *cursocmd.AlterarSituacaoCursoUseCase,
	excluir *cursocmd.ExcluirCursoUseCase,
	listar *cursoquery.ListarCursosUseCase,
	buscar *cursoquery.BuscarCursoUseCase,
	listarMeus *cursoquery.ListarMeusCursosUseCase,
) *CursoHandler {
	return &CursoHandler{
		criar: criar, atualizar: atualizar, alterarSituacao: alterarSituacao, excluir: excluir,
		listar: listar, buscar: buscar, listarMeus: listarMeus,
	}
}

func coordenadorParaResposta(c *port.CoordenadorDoCurso) *CoordenadorDoCursoResponse {
	if c == nil {
		return nil
	}
	var dataFim *string
	if c.DataFim != nil {
		s := c.DataFim.String()
		dataFim = &s
	}
	return &CoordenadorDoCursoResponse{
		ID: c.ID.String(), Nome: c.Nome, DataFim: dataFim, TambemPesquisadorInstitucional: c.TambemPesquisadorInstitucional,
	}
}

func cursoParaResposta(item port.ItemCurso) CursoResponse {
	var codigoEMec *string
	if item.Curso.CodigoEMec != nil {
		s := item.Curso.CodigoEMec.String()
		codigoEMec = &s
	}
	var atualizadoEm *string
	if item.Curso.AtualizadoEm != nil {
		s := item.Curso.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	return CursoResponse{
		ID: item.Curso.ID.String(), Nome: item.Curso.Nome.String(), CodigoEMec: codigoEMec,
		Grau: string(item.Curso.Grau), Modalidade: string(item.Curso.Modalidade), Situacao: string(item.Curso.Situacao),
		Coordenador: coordenadorParaResposta(item.Coordenador), TemVinculo: item.TemVinculo,
		CriadoEm: item.Curso.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: item.Curso.Versao,
	}
}

// Listar godoc
// @Summary      Listar cursos
// @Tags         cursos
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/cursos [get]
func (h *CursoHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoCursoAllowlist, "nome", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	var coordenadorID *uuid.UUID
	vago := false
	if bruto := c.Query("coordenador_id"); bruto == "__vago__" {
		vago = true
	} else if bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrParametro{Nome: "coordenador_id"})
			return
		}
		coordenadorID = &id
	}

	resultado, err := h.listar.Executar(c.Request.Context(), cursoquery.ListarCursosInput{
		Ator: ator, Busca: c.Query("busca"), Grau: c.Query("grau"), Modalidade: c.Query("modalidade"),
		Situacao: c.DefaultQuery("situacao", "ativo"), CoordenadorID: coordenadorID, Vago: vago,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]CursoResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		resposta := cursoParaResposta(item.ItemCurso)
		resposta.PlanoDoPeriodo = item.PlanoDoPeriodo
		itens = append(itens, resposta)
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// MeusCursos godoc
// @Summary      Cursos que o coordenador autenticado coordena hoje
// @Tags         cursos
// @Produce      json
// @Success      200 {array} CursoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/meus-cursos [get]
func (h *CursoHandler) MeusCursos(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	cursos, err := h.listarMeus.Executar(c.Request.Context(), ator)
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]CursoResponse, 0, len(cursos))
	for _, curso := range cursos {
		var codigoEMec *string
		if curso.CodigoEMec != nil {
			s := curso.CodigoEMec.String()
			codigoEMec = &s
		}
		var atualizadoEm *string
		if curso.AtualizadoEm != nil {
			s := curso.AtualizadoEm.Format(time.RFC3339)
			atualizadoEm = &s
		}
		resposta = append(resposta, CursoResponse{
			ID: curso.ID.String(), Nome: curso.Nome.String(), CodigoEMec: codigoEMec,
			Grau: string(curso.Grau), Modalidade: string(curso.Modalidade), Situacao: string(curso.Situacao),
			CriadoEm: curso.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: curso.Versao,
		})
	}
	c.JSON(http.StatusOK, resposta)
}

// Buscar godoc
// @Summary      Buscar curso por ID
// @Tags         cursos
// @Produce      json
// @Success      200 {object} CursoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/cursos/{id} [get]
func (h *CursoHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, cursoParaResposta(item))
}

// Criar godoc
// @Summary      Cadastrar curso
// @Tags         cursos
// @Accept       json
// @Produce      json
// @Success      201 {object} CursoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/cursos [post]
func (h *CursoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req CursoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	novo, err := h.criar.Executar(c.Request.Context(), cursocmd.CriarCursoInput{
		Ator: ator, Nome: req.Nome, CodigoEMec: req.CodigoEMec, Grau: req.Grau, Modalidade: req.Modalidade,
		Situacao: req.Situacao,
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
	c.JSON(http.StatusCreated, cursoParaResposta(item))
}

// Atualizar godoc
// @Summary      Atualizar curso
// @Tags         cursos
// @Accept       json
// @Produce      json
// @Success      200 {object} CursoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/cursos/{id} [put]
func (h *CursoHandler) Atualizar(c *gin.Context) {
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
	var req CursoAtualizarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	_, err = h.atualizar.Executar(c.Request.Context(), cursocmd.AtualizarCursoInput{
		Ator: ator, CursoID: id, Nome: req.Nome, CodigoEMec: req.CodigoEMec, Grau: req.Grau, Modalidade: req.Modalidade, Versao: req.Versao,
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
	c.JSON(http.StatusOK, cursoParaResposta(item))
}

// AlterarSituacao godoc
// @Summary      Inativar ou reativar curso
// @Tags         cursos
// @Accept       json
// @Produce      json
// @Success      200 {object} CursoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/cursos/{id}/situacao [patch]
func (h *CursoHandler) AlterarSituacao(c *gin.Context) {
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
	var req AlterarSituacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	var nova valueobject.SituacaoCurso
	switch req.Situacao {
	case "ativo":
		nova = valueobject.CursoAtivo
	case "inativo":
		nova = valueobject.CursoInativo
	default:
		c.Error(domain.ErrValorInvalido)
		return
	}
	if err := h.alterarSituacao.Executar(c.Request.Context(), cursocmd.AlterarSituacaoCursoInput{
		Ator: ator, CursoID: id, NovaSituacao: nova, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, cursoParaResposta(item))
}

// Excluir godoc
// @Summary      Excluir curso
// @Tags         cursos
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/cursos/{id} [delete]
func (h *CursoHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), cursocmd.ExcluirCursoInput{Ator: ator, CursoID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
