package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	instituicaocmd "github.com/basis-avalia/backend/internal/usecase/command/instituicao"
	instituicaoquery "github.com/basis-avalia/backend/internal/usecase/query/instituicao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoInstituicaoAllowlist = map[string]bool{"sigla": true, "nome": true, "codigo_emec": true, "criado_em": true}

type InstituicaoHandler struct {
	criar           *instituicaocmd.CriarInstituicaoUseCase
	atualizar       *instituicaocmd.AtualizarInstituicaoUseCase
	alterarSituacao *instituicaocmd.AlterarSituacaoInstituicaoUseCase
	listar          *instituicaoquery.ListarInstituicoesUseCase
	buscar          *instituicaoquery.BuscarInstituicaoUseCase
}

func NovoInstituicaoHandler(
	criar *instituicaocmd.CriarInstituicaoUseCase,
	atualizar *instituicaocmd.AtualizarInstituicaoUseCase,
	alterarSituacao *instituicaocmd.AlterarSituacaoInstituicaoUseCase,
	listar *instituicaoquery.ListarInstituicoesUseCase,
	buscar *instituicaoquery.BuscarInstituicaoUseCase,
) *InstituicaoHandler {
	return &InstituicaoHandler{criar: criar, atualizar: atualizar, alterarSituacao: alterarSituacao, listar: listar, buscar: buscar}
}

func instituicaoParaResposta(item port.ItemInstituicao) InstituicaoResponse {
	var codigo *string
	if !item.Instituicao.CodigoEMec.Nulo() {
		s := item.Instituicao.CodigoEMec.String()
		codigo = &s
	}
	var atualizadoEm *string
	if item.Instituicao.AtualizadoEm != nil {
		s := item.Instituicao.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	return InstituicaoResponse{
		ID: item.Instituicao.ID.String(), Nome: item.Instituicao.Nome, Sigla: item.Instituicao.Sigla.String(),
		CodigoEMec: codigo, Situacao: string(item.Instituicao.Situacao), PesquisadoresAtivos: item.PesquisadoresAtivos,
		CriadoEm: item.Instituicao.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: item.Instituicao.Versao,
	}
}

// Listar godoc
// @Summary      Listar instituições
// @Tags         instituicoes
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/instituicoes [get]
func (h *InstituicaoHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoInstituicaoAllowlist, "nome", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	situacao := c.DefaultQuery("situacao", "todas")
	if situacao != "todas" && situacao != "ativa" && situacao != "inativa" {
		c.Error(&domain.ErrParametro{Nome: "situacao"})
		return
	}

	resultado, err := h.listar.Executar(c.Request.Context(), instituicaoquery.ListarInstituicoesInput{
		Ator: ator, Busca: c.Query("busca"), Situacao: situacao,
		Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
	})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]InstituicaoResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, instituicaoParaResposta(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Buscar godoc
// @Summary      Buscar instituição por ID
// @Tags         instituicoes
// @Produce      json
// @Success      200 {object} InstituicaoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/instituicoes/{id} [get]
func (h *InstituicaoHandler) Buscar(c *gin.Context) {
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
	c.JSON(http.StatusOK, instituicaoParaResposta(item))
}

// Criar godoc
// @Summary      Criar instituição
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Success      201 {object} InstituicaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/instituicoes [post]
func (h *InstituicaoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req InstituicaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Nome == "" {
		c.Error(&domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."})
		return
	}
	if req.Sigla == "" {
		c.Error(&domain.ErrValidacao{Campo: "sigla", Mensagem: "A sigla é obrigatória."})
		return
	}
	sigla, err := valueobject.NovaSigla(req.Sigla)
	if err != nil {
		c.Error(err)
		return
	}
	codigo, err := valueobject.NovoCodigoEMec(req.CodigoEMec)
	if err != nil {
		c.Error(err)
		return
	}

	nova, err := h.criar.Executar(c.Request.Context(), instituicaocmd.CriarInstituicaoInput{
		Ator: ator, Nome: req.Nome, Sigla: sigla, CodigoEMec: codigo,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, instituicaoParaResposta(port.ItemInstituicao{Instituicao: nova, PesquisadoresAtivos: 0}))
}

// Atualizar godoc
// @Summary      Atualizar instituição
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Success      200 {object} InstituicaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/instituicoes/{id} [put]
func (h *InstituicaoHandler) Atualizar(c *gin.Context) {
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
	var req InstituicaoAtualizarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if req.Nome == "" {
		c.Error(&domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."})
		return
	}
	if req.Sigla == "" {
		c.Error(&domain.ErrValidacao{Campo: "sigla", Mensagem: "A sigla é obrigatória."})
		return
	}
	sigla, err := valueobject.NovaSigla(req.Sigla)
	if err != nil {
		c.Error(err)
		return
	}
	codigo, err := valueobject.NovoCodigoEMec(req.CodigoEMec)
	if err != nil {
		c.Error(err)
		return
	}

	atualizada, err := h.atualizar.Executar(c.Request.Context(), instituicaocmd.AtualizarInstituicaoInput{
		Ator: ator, InstituicaoID: id, Nome: req.Nome, Sigla: sigla, CodigoEMec: codigo, Versao: req.Versao,
	})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, instituicaoParaResposta(port.ItemInstituicao{Instituicao: atualizada}))
}

// AlterarSituacao godoc
// @Summary      Ativar ou inativar instituição
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Success      200 {object} InstituicaoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/instituicoes/{id}/situacao [patch]
func (h *InstituicaoHandler) AlterarSituacao(c *gin.Context) {
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
	var novaSituacao valueobject.SituacaoInstituicao
	switch req.Situacao {
	case "ativa":
		novaSituacao = valueobject.Ativa
	case "inativa":
		novaSituacao = valueobject.Inativa
	default:
		c.Error(&domain.ErrValidacao{Campo: "situacao", Mensagem: "Requisição inválida."})
		return
	}

	if err := h.alterarSituacao.Executar(c.Request.Context(), instituicaocmd.AlterarSituacaoInstituicaoInput{
		Ator: ator, InstituicaoID: id, NovaSituacao: novaSituacao, Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}

	item, err := h.buscar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, instituicaoParaResposta(item))
}

