package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/port"
	planocmd "github.com/basis-avalia/backend/internal/usecase/command/plano"
	planoquery "github.com/basis-avalia/backend/internal/usecase/query/plano"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoPlanoAllowlist = map[string]bool{"curso": true, "periodo": true, "situacao": true, "criado_em": true}

// PlanoHandler serve /api/v1/planos e /api/v1/meus-planos
// (specs/plano-acao/design.md §6).
type PlanoHandler struct {
	criar               *planocmd.CriarPlanoUseCase
	atualizar           *planocmd.AtualizarPlanoUseCase
	publicar            *planocmd.PublicarPlanoUseCase
	despublicar         *planocmd.DespublicarPlanoUseCase
	encerrar            *planocmd.EncerrarPlanoUseCase
	reabrir             *planocmd.ReabrirPlanoUseCase
	excluir             *planocmd.ExcluirPlanoUseCase
	copiarEmLote        *planocmd.CopiarEmLoteUseCase
	listar              *planoquery.ListarPlanosUseCase
	listarMeus          *planoquery.ListarMeusPlanosUseCase
	buscar              *planoquery.BuscarPlanoUseCase
	listarDestinosCopia *planoquery.ListarDestinosDeCopiaUseCase
}

func NovoPlanoHandler(
	criar *planocmd.CriarPlanoUseCase, atualizar *planocmd.AtualizarPlanoUseCase,
	publicar *planocmd.PublicarPlanoUseCase, despublicar *planocmd.DespublicarPlanoUseCase,
	encerrar *planocmd.EncerrarPlanoUseCase, reabrir *planocmd.ReabrirPlanoUseCase, excluir *planocmd.ExcluirPlanoUseCase,
	copiarEmLote *planocmd.CopiarEmLoteUseCase,
	listar *planoquery.ListarPlanosUseCase, listarMeus *planoquery.ListarMeusPlanosUseCase,
	buscar *planoquery.BuscarPlanoUseCase, listarDestinosCopia *planoquery.ListarDestinosDeCopiaUseCase,
) *PlanoHandler {
	return &PlanoHandler{
		criar: criar, atualizar: atualizar, publicar: publicar, despublicar: despublicar, encerrar: encerrar,
		reabrir: reabrir, excluir: excluir, copiarEmLote: copiarEmLote, listar: listar, listarMeus: listarMeus,
		buscar: buscar, listarDestinosCopia: listarDestinosCopia,
	}
}

func indicadoresDaMetaParaResposta(indicadores []port.IndicadorDaMeta) []IndicadorEmbutidoResponse {
	resposta := make([]IndicadorEmbutidoResponse, 0, len(indicadores))
	for _, i := range indicadores {
		resposta = append(resposta, IndicadorEmbutidoResponse{
			ID: i.ID.String(), Codigo: i.Codigo, Nome: i.Nome, Escopo: i.Escopo,
			ReferenciaInstrumento: i.ReferenciaInstrumento, Situacao: i.Situacao,
		})
	}
	return resposta
}

func itemPlanoParaResposta(item port.ItemDoPlanoResponse) ItemPlanoResponse {
	return ItemPlanoResponse{
		ID: item.ID.String(), MetaID: item.MetaID.String(), MetaNome: item.MetaNome,
		Indicadores: indicadoresDaMetaParaResposta(item.Indicadores), Quantidade: item.Quantidade,
		TemEntrega: item.TemEntrega, Versao: item.Versao,
	}
}

func linhaPlanoParaResposta(l port.LinhaPlano) PlanoResponse {
	var atualizadoEm *string
	if l.Plano.AtualizadoEm != nil {
		s := l.Plano.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	var aprovacao *AprovacaoResponse
	if l.Plano.Aprovacao != nil {
		aprovacao = &AprovacaoResponse{Data: l.Plano.Aprovacao.Data().String(), Orgao: string(l.Plano.Aprovacao.Orgao())}
	}
	var ultimoDocumento *UltimoDocumentoResponse
	if l.UltimoDocumento != nil {
		ultimoDocumento = &UltimoDocumentoResponse{ID: l.UltimoDocumento.ID.String(), GeradoEm: l.UltimoDocumento.GeradoEm.Format(time.RFC3339)}
	}
	return PlanoResponse{
		ID: l.Plano.ID.String(),
		Curso: CursoEmbutidoResponse{
			ID: l.Plano.CursoID.String(), Nome: l.CursoNome, Grau: l.CursoGrau, Modalidade: l.CursoModalidade,
			CodigoEMec: l.CursoCodigoEMec, Vago: l.CursoVago,
		},
		Periodo:             PeriodoEmbutidoResponse{ID: l.Plano.PeriodoID.String(), Nome: l.PeriodoNome},
		Descricao:           l.Plano.Descricao,
		ObjetivoGeral:       l.Plano.ObjetivoGeral,
		ResultadosEsperados: l.Plano.ResultadosEsperados,
		AlinhamentoPDI:      l.Plano.AlinhamentoPDI,
		AlinhamentoPPC:      l.Plano.AlinhamentoPPC,
		Situacao:            l.Situacao,
		Metas:               l.TotalItens,
		TotalExigido:        l.TotalExigido,
		Aprovacao:           aprovacao,
		SemAprovacao:        l.SemAprovacao,
		TemEntrega:          l.TemEntrega,
		EncerramentoMotivo:  l.Plano.EncerramentoMotivo,
		UltimoDocumento:     ultimoDocumento,
		CriadoEm:            l.Plano.CriadoEm.Format(time.RFC3339),
		AtualizadoEm:        atualizadoEm,
		Versao:              l.Plano.Versao,
	}
}

func detalhePlanoParaResposta(d port.DetalhePlano) PlanoResponse {
	resp := linhaPlanoParaResposta(d.LinhaPlano)
	resp.CoordenadorNome = d.CoordenadorNome
	itens := make([]ItemPlanoResponse, 0, len(d.Itens))
	for _, item := range d.Itens {
		itens = append(itens, itemPlanoParaResposta(item))
	}
	resp.Itens = itens
	return resp
}

func filtroPlanoDoRequest(c *gin.Context) (port.FiltroListarPlanos, error) {
	filtro := port.FiltroListarPlanos{
		Situacao:  c.DefaultQuery("situacao", "todas"),
		Aprovacao: c.DefaultQuery("aprovacao", "todos"),
	}
	if bruto := c.Query("periodo_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroListarPlanos{}, &domain.ErrParametro{Nome: "periodo_id"}
		}
		filtro.PeriodoID = &id
	}
	if bruto := c.Query("curso_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroListarPlanos{}, &domain.ErrParametro{Nome: "curso_id"}
		}
		filtro.CursoID = &id
	}
	return filtro, nil
}

func (h *PlanoHandler) listarComum(c *gin.Context, executar func(port.FiltroListarPlanos) (port.ResultadoListaPlanos, error)) {
	paginacao, err := ParsePaginacao(c, ordenacaoPlanoAllowlist, "curso", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	filtro, err := filtroPlanoDoRequest(c)
	if err != nil {
		c.Error(err)
		return
	}
	filtro.Page, filtro.PageSize, filtro.Sort, filtro.Order = paginacao.Page, paginacao.PageSize, paginacao.Sort, paginacao.Order
	resultado, err := executar(filtro)
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]PlanoResponse, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		itens = append(itens, linhaPlanoParaResposta(item))
	}
	c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
		Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
		TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
	}})
}

// Listar godoc
// @Summary      Listar planos da instituição
// @Tags         planos
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/planos [get]
func (h *PlanoHandler) Listar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	h.listarComum(c, func(filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
		return h.listar.Executar(c.Request.Context(), planoquery.ListarPlanosInput{Ator: ator, Filtro: filtro})
	})
}

// ListarMeus godoc
// @Summary      Listar os planos dos cursos que o coordenador coordena
// @Tags         planos
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/meus-planos [get]
func (h *PlanoHandler) ListarMeus(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	h.listarComum(c, func(filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
		return h.listarMeus.Executar(c.Request.Context(), planoquery.ListarMeusPlanosInput{Ator: ator, Filtro: filtro})
	})
}

func (h *PlanoHandler) buscarComum(c *gin.Context, alcance autorizacao.Alcance) {
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
	item, err := h.buscar.Executar(c.Request.Context(), ator, alcance, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, detalhePlanoParaResposta(item))
}

// Buscar godoc
// @Summary      Buscar plano por ID
// @Tags         planos
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/planos/{id} [get]
func (h *PlanoHandler) Buscar(c *gin.Context) { h.buscarComum(c, autorizacao.PlanosDaInstituicao) }

// BuscarMeu godoc
// @Summary      Buscar plano de um curso da carteira do coordenador
// @Tags         planos
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/meus-planos/{id} [get]
func (h *PlanoHandler) BuscarMeu(c *gin.Context) { h.buscarComum(c, autorizacao.PlanosDaCarteira) }

// Criar godoc
// @Summary      Cadastrar plano de ação curso/coordenador
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      201 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos [post]
func (h *PlanoHandler) Criar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	var req PlanoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	cursoID, err := uuid.Parse(req.CursoID)
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	periodoID, err := uuid.Parse(req.PeriodoID)
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	novo, err := h.criar.Executar(c.Request.Context(), planocmd.CriarPlanoInput{
		Ator: ator, CursoID: cursoID, PeriodoID: periodoID, Dados: dadosDoPlanoRequest(req),
	})
	if err != nil {
		c.Error(err)
		return
	}
	item, err := h.buscar.Executar(c.Request.Context(), ator, autorizacao.PlanosDaInstituicao, novo.ID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, detalhePlanoParaResposta(item))
}

func dadosDoPlanoRequest(req PlanoRequest) plano.DadosDoPlano {
	return plano.DadosDoPlano{
		Descricao: req.Descricao, ObjetivoGeral: req.ObjetivoGeral, ResultadosEsperados: req.ResultadosEsperados,
		AlinhamentoPDI: req.AlinhamentoPDI, AlinhamentoPPC: req.AlinhamentoPPC,
		AprovacaoData: req.AprovacaoData, AprovacaoOrgao: req.AprovacaoOrgao,
	}
}

func (h *PlanoHandler) atualizarResposta(c *gin.Context, ator autorizacao.Ator, id uuid.UUID) {
	item, err := h.buscar.Executar(c.Request.Context(), ator, autorizacao.PlanosDaInstituicao, id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, detalhePlanoParaResposta(item))
}

// Atualizar godoc
// @Summary      Atualizar dados e aprovação do plano
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id} [put]
func (h *PlanoHandler) Atualizar(c *gin.Context) {
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
	var req PlanoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	if req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if _, err := h.atualizar.Executar(c.Request.Context(), planocmd.AtualizarPlanoInput{
		Ator: ator, PlanoID: id, Dados: dadosDoPlanoRequest(req), Versao: req.Versao,
	}); err != nil {
		c.Error(err)
		return
	}
	h.atualizarResposta(c, ator, id)
}

// Publicar godoc
// @Summary      Publicar plano (rascunho -> vigente)
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/publicar [post]
func (h *PlanoHandler) Publicar(c *gin.Context) {
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
	var req VersaoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if _, err := h.publicar.Executar(c.Request.Context(), planocmd.PublicarPlanoInput{Ator: ator, PlanoID: id, Versao: req.Versao}); err != nil {
		c.Error(err)
		return
	}
	h.atualizarResposta(c, ator, id)
}

// Despublicar godoc
// @Summary      Despublicar plano (vigente -> rascunho)
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/despublicar [post]
func (h *PlanoHandler) Despublicar(c *gin.Context) {
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
	var req VersaoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.despublicar.Executar(c.Request.Context(), planocmd.DespublicarPlanoInput{Ator: ator, PlanoID: id, Versao: req.Versao}); err != nil {
		c.Error(err)
		return
	}
	h.atualizarResposta(c, ator, id)
}

// Encerrar godoc
// @Summary      Encerrar plano antecipadamente
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/encerrar [post]
func (h *PlanoHandler) Encerrar(c *gin.Context) {
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
	var req EncerrarPlanoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.encerrar.Executar(c.Request.Context(), planocmd.EncerrarPlanoInput{Ator: ator, PlanoID: id, Motivo: req.Motivo, Versao: req.Versao}); err != nil {
		c.Error(err)
		return
	}
	h.atualizarResposta(c, ator, id)
}

// Reabrir godoc
// @Summary      Reabrir plano encerrado antecipadamente
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} PlanoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id}/reabrir [post]
func (h *PlanoHandler) Reabrir(c *gin.Context) {
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
	var req VersaoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.reabrir.Executar(c.Request.Context(), planocmd.ReabrirPlanoInput{Ator: ator, PlanoID: id, Versao: req.Versao}); err != nil {
		c.Error(err)
		return
	}
	h.atualizarResposta(c, ator, id)
}

// Excluir godoc
// @Summary      Excluir plano em rascunho
// @Tags         planos
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/planos/{id} [delete]
func (h *PlanoHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), planocmd.ExcluirPlanoInput{Ator: ator, PlanoID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListarDestinosDeCopia godoc
// @Summary      Listar cursos elegíveis para cópia em lote do plano
// @Tags         planos
// @Produce      json
// @Success      200 {array} DestinoDeCopiaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/planos/{id}/destinos-copia [get]
func (h *PlanoHandler) ListarDestinosDeCopia(c *gin.Context) {
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
	periodoDestinoID, err := uuid.Parse(c.Query("periodo_destino_id"))
	if err != nil {
		c.Error(&domain.ErrParametro{Nome: "periodo_destino_id"})
		return
	}
	destinos, err := h.listarDestinosCopia.Executar(c.Request.Context(), planoquery.ListarDestinosDeCopiaInput{
		Ator: ator, PlanoOrigemID: id, PeriodoDestinoID: periodoDestinoID, Busca: c.Query("busca"),
	})
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]DestinoDeCopiaResponse, 0, len(destinos))
	for _, d := range destinos {
		resposta = append(resposta, DestinoDeCopiaResponse{
			CursoID: d.CursoID.String(), CursoNome: d.CursoNome, CoordenadorNome: d.CoordenadorNome,
			Vago: d.Vago, JaTemPlano: d.JaTemPlano,
		})
	}
	c.JSON(http.StatusOK, resposta)
}

// CopiarEmLote godoc
// @Summary      Copiar plano para vários cursos
// @Tags         planos
// @Accept       json
// @Produce      json
// @Success      200 {object} CopiaEmLoteResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/planos/{id}/copias [post]
func (h *PlanoHandler) CopiarEmLote(c *gin.Context) {
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
	var req CopiaEmLoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
		return
	}
	periodoDestinoID, err := uuid.Parse(req.PeriodoDestinoID)
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	cursos := make([]uuid.UUID, 0, len(req.Cursos))
	for _, bruto := range req.Cursos {
		cursoID, err := uuid.Parse(bruto)
		if err != nil {
			c.Error(&domain.ErrValidacao{Campo: "cursos", Mensagem: "Curso inválido."})
			return
		}
		cursos = append(cursos, cursoID)
	}
	resultado, err := h.copiarEmLote.Executar(c.Request.Context(), planocmd.CopiarEmLoteInput{
		Ator: ator, PlanoOrigemID: id, PeriodoDestinoID: periodoDestinoID, Cursos: cursos,
	})
	if err != nil {
		c.Error(err)
		return
	}
	pulados := make([]CursoPuladoResponse, 0, len(resultado.Pulados))
	for _, p := range resultado.Pulados {
		pulados = append(pulados, CursoPuladoResponse{Curso: CursoEmbutidoResponse{ID: p.CursoID.String()}, Motivo: p.Motivo})
	}
	c.JSON(http.StatusOK, CopiaEmLoteResponse{Criados: resultado.Criados, Pulados: pulados})
}
