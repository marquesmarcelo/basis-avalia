package http

import (
	"fmt"
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	documentocmd "github.com/basis-avalia/backend/internal/usecase/command/documento"
	documentoquery "github.com/basis-avalia/backend/internal/usecase/query/documento"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DocumentoHandler serve a geração e o download do .docx do plano
// (specs/plano-acao/design.md §7). Nenhuma resposta expõe a URL do
// armazenamento — o conteúdo é sempre transmitido pela própria rota.
type DocumentoHandler struct {
	gerar  *documentocmd.GerarDocumentoUseCase
	baixar *documentoquery.BaixarDocumentoUseCase
}

func NovoDocumentoHandler(gerar *documentocmd.GerarDocumentoUseCase, baixar *documentoquery.BaixarDocumentoUseCase) *DocumentoHandler {
	return &DocumentoHandler{gerar: gerar, baixar: baixar}
}

func (h *DocumentoHandler) gerarComum(c *gin.Context, alcance autorizacao.Alcance) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	planoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(domain.ErrNaoEncontrado)
		return
	}
	resultado, err := h.gerar.Executar(c.Request.Context(), documentocmd.GerarDocumentoInput{Ator: ator, Alcance: alcance, PlanoID: planoID})
	if err != nil {
		c.Error(err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, resultado.NomeArquivo))
	c.Data(http.StatusCreated, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", resultado.Conteudo)
}

// Gerar godoc
// @Summary      Gerar o documento .docx do plano (Pesquisador Institucional)
// @Tags         documentos
// @Produce      application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Success      201 {file} file
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/planos/{id}/documentos [post]
func (h *DocumentoHandler) Gerar(c *gin.Context) { h.gerarComum(c, autorizacao.PlanosDaInstituicao) }

// GerarMeu godoc
// @Summary      Gerar o documento .docx de um plano da carteira (Coordenador)
// @Tags         documentos
// @Produce      application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Success      201 {file} file
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/meus-planos/{id}/documentos [post]
func (h *DocumentoHandler) GerarMeu(c *gin.Context) { h.gerarComum(c, autorizacao.PlanosDaCarteira) }

// Baixar godoc
// @Summary      Baixar um documento já gerado
// @Tags         documentos
// @Produce      application/vnd.openxmlformats-officedocument.wordprocessingml.document
// @Success      200 {file} file
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Failure      404 {object} erroResposta
// @Router       /api/v1/documentos/{id}/conteudo [get]
func (h *DocumentoHandler) Baixar(c *gin.Context) {
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
	resultado, err := h.baixar.Executar(c.Request.Context(), ator, id)
	if err != nil {
		c.Error(err)
		return
	}
	defer resultado.Conteudo.Close()
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, resultado.NomeArquivo))
	c.DataFromReader(http.StatusOK, -1, resultado.TipoConteudo, resultado.Conteudo, nil)
}
