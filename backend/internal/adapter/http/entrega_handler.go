package http

import (
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/anexo"
	"github.com/basis-avalia/backend/internal/port"
	anexocmd "github.com/basis-avalia/backend/internal/usecase/command/anexo"
	entregacmd "github.com/basis-avalia/backend/internal/usecase/command/entrega"
	anexoquery "github.com/basis-avalia/backend/internal/usecase/query/anexo"
	entregaquery "github.com/basis-avalia/backend/internal/usecase/query/entrega"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoEntregaAllowlist = map[string]bool{"criado_em": true, "curso": true, "meta": true}

// limiteCorpoMultipart — 50 MB (PM-5) mais folga do multipart (boundary,
// cabeçalhos, campo de texto) — design.md §6.3, M-08.
const limiteCorpoMultipart = 51 * 1024 * 1024

// EntregaHandler serve entrega, anexo e os dois badges
// (specs/metas-coordenacao/design.md §11).
type EntregaHandler struct {
	registrar           *entregacmd.RegistrarEntregaUseCase
	corrigir            *entregacmd.CorrigirEntregaUseCase
	excluir             *entregacmd.ExcluirEntregaUseCase
	marcarPendenciaVista *entregacmd.MarcarPendenciaVistaUseCase
	adicionarAnexo      *anexocmd.AdicionarAnexoUseCase
	removerAnexo        *anexocmd.RemoverAnexoUseCase
	listarDoItem        *entregaquery.ListarDoItemUseCase
	buscar              *entregaquery.BuscarUseCase
	minhasMetas         *entregaquery.MinhasMetasUseCase
	listarPeriodos      *entregaquery.ListarPeriodosUseCase
	baixarAnexo         *anexoquery.BaixarUseCase
}

func NovoEntregaHandler(
	registrar *entregacmd.RegistrarEntregaUseCase, corrigir *entregacmd.CorrigirEntregaUseCase,
	excluir *entregacmd.ExcluirEntregaUseCase, marcarPendenciaVista *entregacmd.MarcarPendenciaVistaUseCase,
	adicionarAnexo *anexocmd.AdicionarAnexoUseCase, removerAnexo *anexocmd.RemoverAnexoUseCase,
	listarDoItem *entregaquery.ListarDoItemUseCase, buscar *entregaquery.BuscarUseCase,
	minhasMetas *entregaquery.MinhasMetasUseCase, listarPeriodos *entregaquery.ListarPeriodosUseCase,
	baixarAnexo *anexoquery.BaixarUseCase,
) *EntregaHandler {
	return &EntregaHandler{
		registrar: registrar, corrigir: corrigir, excluir: excluir, marcarPendenciaVista: marcarPendenciaVista,
		adicionarAnexo: adicionarAnexo, removerAnexo: removerAnexo, listarDoItem: listarDoItem,
		buscar: buscar, minhasMetas: minhasMetas, listarPeriodos: listarPeriodos, baixarAnexo: baixarAnexo,
	}
}

// ListarPeriodos godoc
// @Summary      Listar períodos para o seletor de "Minhas metas"
// @Tags         entregas
// @Produce      json
// @Success      200 {array} PeriodoOpcaoResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/minhas-metas/periodos [get]
func (h *EntregaHandler) ListarPeriodos(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	opcoes, err := h.listarPeriodos.Executar(c.Request.Context(), ator)
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]PeriodoOpcaoResponse, 0, len(opcoes))
	for _, o := range opcoes {
		resposta = append(resposta, PeriodoOpcaoResponse{ID: o.ID.String(), Nome: o.Nome, DataFim: o.DataFim, Aberto: o.Aberto})
	}
	c.JSON(http.StatusOK, resposta)
}

func anexoParaResposta(a port.AnexoResponse) AnexoResponse {
	return AnexoResponse{
		ID: a.ID.String(), NomeOriginal: a.NomeOriginal, Tipo: a.Tipo, TamanhoBytes: a.TamanhoBytes,
		HashSHA256: a.HashSHA256, CriadoEm: a.CriadoEm.Format(time.RFC3339),
	}
}

func ponteiroTempo(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func entregaParaResposta(d port.DetalheEntrega) EntregaResponse {
	anexos := make([]AnexoResponse, 0, len(d.Anexos))
	for _, a := range d.Anexos {
		anexos = append(anexos, anexoParaResposta(a))
	}
	indicadores := make([]IndicadorEmbutidoResponse, 0, len(d.Indicadores))
	for _, i := range d.Indicadores {
		indicadores = append(indicadores, IndicadorEmbutidoResponse{
			ID: i.ID.String(), Codigo: i.Codigo, Nome: i.Nome, Escopo: i.Escopo,
			ReferenciaInstrumento: i.ReferenciaInstrumento, Situacao: i.Situacao,
		})
	}
	return EntregaResponse{
		ID: d.ID.String(), ItemPlanoID: d.ItemPlanoID.String(), CursoID: d.CursoID.String(),
		CursoNome: d.CursoNome, MetaNome: d.MetaNome, Quantidade: d.Quantidade, Indicadores: indicadores,
		Situacao: d.Situacao, Rodadas: d.Rodadas, PrazoCorrecao: ponteiroTempo(d.PrazoCorrecao),
		EnviadaPorID: d.EnviadaPorID.String(), EnviadaPorNome: d.EnviadaPorNome, CorrigidaPorNome: d.CorrigidaPorNome,
		Observacao: d.Observacao, Motivo: d.Motivo, AvaliadaPorNome: d.AvaliadaPorNome, AvaliadaEm: ponteiroTempo(d.AvaliadaEm),
		AvaliadorEraCoordenador: d.AvaliadorEraCoordenador, PendenciaVistaEm: ponteiroTempo(d.PendenciaVistaEm),
		CoordenadoPeloAvaliador: d.CoordenadoPeloAvaliador, Anexos: anexos,
		CriadoEm: d.CriadoEm.Format(time.RFC3339), Versao: d.Versao,
	}
}

// lerArquivosMultipart aplica os limites ANTES de qualquer byte chegar ao
// use case (M-08, AN-03): MaxBytesReader na requisição inteira,
// io.LimitReader de LimiteBytesPorArquivo+1 por parte — nunca
// io.ReadAll sobre um stream ilimitado. O "+1" é o que distingue "o
// arquivo tem exatamente o limite" de "o arquivo é maior e foi cortado".
func lerArquivosMultipart(c *gin.Context) (string, []entregacmd.ArquivoRecebido, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limiteCorpoMultipart)
	mr, err := c.Request.MultipartReader()
	if err != nil {
		return "", nil, &domain.ErrValidacao{Mensagem: "Requisição inválida."}
	}

	observacao := ""
	arquivos := make([]entregacmd.ArquivoRecebido, 0, anexo.LimiteArquivosPorEntrega)
	var somaBytes int64

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, &domain.ErrValidacao{Mensagem: "Requisição inválida."}
		}
		if part.FormName() == "observacao" {
			limitado := io.LimitReader(part, 5000)
			bruto, err := io.ReadAll(limitado)
			if err != nil {
				return "", nil, &domain.ErrValidacao{Mensagem: "Requisição inválida."}
			}
			observacao = string(bruto)
			continue
		}
		if part.FormName() != "arquivos" {
			continue
		}
		if len(arquivos) >= anexo.LimiteArquivosPorEntrega {
			return "", nil, domain.ErrAnexosAcimaDoLimite
		}
		conteudo, err := lerParteLimitada(part)
		if err != nil {
			return "", nil, err
		}
		somaBytes += int64(len(conteudo))
		if somaBytes > anexo.LimiteBytesPorEntrega {
			return "", nil, domain.ErrAnexosAcimaDoLimite
		}
		arquivos = append(arquivos, entregacmd.ArquivoRecebido{NomeOriginal: part.FileName(), Conteudo: conteudo})
	}
	return observacao, arquivos, nil
}

func lerParteLimitada(part *multipart.Part) ([]byte, error) {
	return lerConteudoLimitado(part, anexo.LimiteBytesPorArquivo)
}

// lerConteudoLimitado — o núcleo de M-08 (design.md §6.3): nunca
// io.ReadAll sobre um reader ilimitado. io.LimitReader corta a leitura em
// limite+1 bytes ANTES do ReadAll rodar, não importa quantos bytes o lado
// de origem tente oferecer — é isto que impede um envio de 2 GB de
// derrubar o container antes de qualquer validação de tamanho. Extraído
// para io.Reader genérico (em vez de *multipart.Part) para ser testável
// com um reader sintético que simula um cliente malicioso oferecendo um
// stream indefinido, sem precisar montar 2 GB de verdade em teste.
func lerConteudoLimitado(r io.Reader, limite int64) ([]byte, error) {
	limitado := io.LimitReader(r, limite+1)
	conteudo, err := io.ReadAll(limitado)
	if err != nil {
		return nil, &domain.ErrValidacao{Mensagem: "Requisição inválida."}
	}
	if int64(len(conteudo)) > limite {
		return nil, domain.ErrAnexoAcimaDoLimite
	}
	return conteudo, nil
}

// Registrar godoc
// @Summary      Registrar entrega de comprovante de meta
// @Tags         entregas
// @Accept       multipart/form-data
// @Produce      json
// @Success      201 {object} EntregaResponse
// @Success      200 {object} EntregaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/itens/{itemId}/entregas [post]
func (h *EntregaHandler) Registrar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	chave := c.GetHeader("Idempotency-Key")
	if chave == "" {
		c.Error(domain.ErrIdempotencyKeyObrigatoria)
		return
	}
	observacao, arquivos, err := lerArquivosMultipart(c)
	if err != nil {
		c.Error(err)
		return
	}

	resultado, err := h.registrar.Executar(c.Request.Context(), entregacmd.RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: itemID, Observacao: observacao, IdempotencyKey: chave, Arquivos: arquivos,
	})
	if err != nil {
		c.Error(err)
		return
	}
	detalhe, err := h.buscar.Executar(c.Request.Context(), entregaquery.BuscarInput{Ator: ator, EntregaID: resultado.EntregaID})
	if err != nil {
		c.Error(err)
		return
	}
	status := http.StatusCreated
	if resultado.Reaproveitada {
		status = http.StatusOK
	}
	c.JSON(status, entregaParaResposta(detalhe))
}

// Buscar godoc
// @Summary      Buscar entrega por ID
// @Tags         entregas
// @Produce      json
// @Success      200 {object} EntregaResponse
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/entregas/{id} [get]
func (h *EntregaHandler) Buscar(c *gin.Context) {
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
	detalhe, err := h.buscar.Executar(c.Request.Context(), entregaquery.BuscarInput{Ator: ator, EntregaID: id})
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, entregaParaResposta(detalhe))
}

// ListarDoItem godoc
// @Summary      Listar entregas de um item do plano
// @Tags         entregas
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/itens/{itemId}/entregas [get]
func (h *EntregaHandler) ListarDoItem(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoEntregaAllowlist, "criado_em", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	resultado, err := h.listarDoItem.Executar(c.Request.Context(), entregaquery.ListarDoItemInput{
		Ator: ator, ItemPlanoID: itemID,
		Filtro: port.FiltroListarEntregas{Situacao: c.Query("situacao"), Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order},
	})
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

// Corrigir godoc
// @Summary      Corrigir e reenviar entrega recusada
// @Tags         entregas
// @Accept       json
// @Produce      json
// @Success      200 {object} EntregaResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/entregas/{id} [put]
func (h *EntregaHandler) Corrigir(c *gin.Context) {
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
	var req EntregaCorrigirRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Versao <= 0 {
		c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
		return
	}
	if err := h.corrigir.Executar(c.Request.Context(), entregacmd.CorrigirEntregaInput{
		Ator: ator, EntregaID: id, Observacao: req.Observacao, Versao: req.Versao,
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

// Excluir godoc
// @Summary      Excluir entrega não aceita
// @Tags         entregas
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Failure      409 {object} erroResposta
// @Router       /api/v1/entregas/{id} [delete]
func (h *EntregaHandler) Excluir(c *gin.Context) {
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
	if err := h.excluir.Executar(c.Request.Context(), entregacmd.ExcluirEntregaInput{Ator: ator, EntregaID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MarcarPendenciaVista godoc
// @Summary      Marcar a pendência de recusa como vista
// @Tags         entregas
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/entregas/{id}/pendencia-vista [post]
func (h *EntregaHandler) MarcarPendenciaVista(c *gin.Context) {
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
	if err := h.marcarPendenciaVista.Executar(c.Request.Context(), entregacmd.MarcarPendenciaVistaInput{Ator: ator, EntregaID: id}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// MinhasMetas godoc
// @Summary      Listar as metas dos cursos que o coordenador coordena (única tela sem "Pesquisar")
// @Tags         entregas
// @Produce      json
// @Success      200 {array} GrupoMinhasMetasResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/minhas-metas [get]
func (h *EntregaHandler) MinhasMetas(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	periodoID, err := uuid.Parse(c.Query("periodo_id"))
	if err != nil {
		c.Error(domain.ErrPeriodoObrigatorio)
		return
	}
	grupos, err := h.minhasMetas.Executar(c.Request.Context(), entregaquery.MinhasMetasInput{Ator: ator, PeriodoID: periodoID})
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]GrupoMinhasMetasResponse, 0, len(grupos))
	for _, g := range grupos {
		itens := make([]ItemMinhasMetasResponse, 0, len(g.Itens))
		for _, item := range g.Itens {
			indicadores := make([]IndicadorEmbutidoResponse, 0, len(item.Indicadores))
			for _, i := range item.Indicadores {
				indicadores = append(indicadores, IndicadorEmbutidoResponse{
					ID: i.ID.String(), Codigo: i.Codigo, Nome: i.Nome, Escopo: i.Escopo,
					ReferenciaInstrumento: i.ReferenciaInstrumento, Situacao: i.Situacao,
				})
			}
			itens = append(itens, ItemMinhasMetasResponse{
				ItemPlanoID: item.ItemPlanoID.String(), MetaNome: item.MetaNome, Indicadores: indicadores,
				Quantidade: item.Quantidade, Aceitas: item.Aceitas, Pendentes: item.Pendentes, EmCorrecao: item.EmCorrecao,
			})
		}
		resposta = append(resposta, GrupoMinhasMetasResponse{CursoID: g.CursoID.String(), CursoNome: g.CursoNome, Itens: itens})
	}
	c.JSON(http.StatusOK, resposta)
}

// AdicionarAnexo godoc
// @Summary      Adicionar anexo a uma entrega em correção
// @Tags         anexos
// @Accept       multipart/form-data
// @Produce      json
// @Success      201 {array} AnexoResponse
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/entregas/{id}/anexos [post]
func (h *EntregaHandler) AdicionarAnexo(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	entregaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	_, arquivos, err := lerArquivosMultipart(c)
	if err != nil {
		c.Error(err)
		return
	}
	brutos := make([]anexocmd.ArquivoRecebido, 0, len(arquivos))
	for _, a := range arquivos {
		brutos = append(brutos, anexocmd.ArquivoRecebido{NomeOriginal: a.NomeOriginal, Conteudo: a.Conteudo})
	}
	anexos, err := h.adicionarAnexo.Executar(c.Request.Context(), anexocmd.AdicionarAnexoInput{Ator: ator, EntregaID: entregaID, Arquivos: brutos})
	if err != nil {
		c.Error(err)
		return
	}
	resposta := make([]AnexoResponse, 0, len(anexos))
	for _, a := range anexos {
		resposta = append(resposta, anexoParaResposta(a))
	}
	c.JSON(http.StatusCreated, resposta)
}

// RemoverAnexo godoc
// @Summary      Remover anexo de uma entrega
// @Tags         anexos
// @Success      204
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/anexos/{id} [delete]
func (h *EntregaHandler) RemoverAnexo(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	anexoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	if err := h.removerAnexo.Executar(c.Request.Context(), anexocmd.RemoverAnexoInput{Ator: ator, AnexoID: anexoID}); err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// BaixarAnexo godoc
// @Summary      Baixar o conteúdo de um anexo
// @Tags         anexos
// @Success      200 {file} file
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/anexos/{id}/conteudo [get]
func (h *EntregaHandler) BaixarAnexo(c *gin.Context) {
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
	resultado, err := h.baixarAnexo.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	defer resultado.Conteudo.Close()
	c.Header("Content-Disposition", "attachment; filename=\""+resultado.NomeArquivo+"\"")
	c.DataFromReader(http.StatusOK, -1, resultado.TipoConteudo, resultado.Conteudo, nil)
}
