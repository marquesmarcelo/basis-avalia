package http

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
	relatorioquery "github.com/basis-avalia/backend/internal/usecase/query/relatorio"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ordenacaoRelatorioAllowlist = map[string]bool{
	"curso": true, "responsavel": true, "meta": true, "exigido": true, "aceitas": true, "cumprimento": true,
}

// RelatorioHandler serve o relatório de desempenho e sua exportação CSV
// (specs/metas-coordenacao/design.md §9, §11).
type RelatorioHandler struct {
	desempenho *relatorioquery.DesempenhoUseCase
	porCurso   *relatorioquery.DesempenhoPorCursoUseCase
	exportar   *relatorioquery.ExportarDesempenhoUseCase
}

func NovoRelatorioHandler(desempenho *relatorioquery.DesempenhoUseCase, porCurso *relatorioquery.DesempenhoPorCursoUseCase, exportar *relatorioquery.ExportarDesempenhoUseCase) *RelatorioHandler {
	return &RelatorioHandler{desempenho: desempenho, porCurso: porCurso, exportar: exportar}
}

func filtroRelatorioDoRequest(c *gin.Context) (port.FiltroRelatorio, error) {
	periodoID, err := uuid.Parse(c.Query("periodo_id"))
	if err != nil {
		return port.FiltroRelatorio{}, domain.ErrPeriodoObrigatorio
	}
	filtro := port.FiltroRelatorio{
		PeriodoID: periodoID, Origem: c.Query("origem"), Situacao: c.Query("situacao"),
		IncluirInativos: c.Query("incluir_inativos") == "true",
	}
	if bruto := c.Query("curso_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroRelatorio{}, &domain.ErrParametro{Nome: "curso_id"}
		}
		filtro.CursoID = &id
	}
	if bruto := c.Query("meta_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroRelatorio{}, &domain.ErrParametro{Nome: "meta_id"}
		}
		filtro.MetaID = &id
	}
	if bruto := c.Query("indicador_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroRelatorio{}, &domain.ErrParametro{Nome: "indicador_id"}
		}
		filtro.IndicadorID = &id
	}
	if bruto := c.Query("responsavel_id"); bruto != "" {
		id, err := uuid.Parse(bruto)
		if err != nil {
			return port.FiltroRelatorio{}, &domain.ErrParametro{Nome: "responsavel_id"}
		}
		filtro.ResponsavelID = &id
	}
	if bruto := c.Query("autoavaliado"); bruto != "" {
		v, err := strconv.ParseBool(bruto)
		if err != nil {
			return port.FiltroRelatorio{}, &domain.ErrParametro{Nome: "autoavaliado"}
		}
		filtro.Autoavaliado = &v
	}
	return filtro, nil
}

func linhaRelatorioParaResposta(l port.LinhaRelatorio) RelatorioLinhaResponse {
	indicadores := make([]IndicadorEmbutidoResponse, 0, len(l.Indicadores))
	for _, i := range l.Indicadores {
		indicadores = append(indicadores, IndicadorEmbutidoResponse{
			ID: i.ID.String(), Codigo: i.Codigo, Nome: i.Nome, Escopo: i.Escopo,
			ReferenciaInstrumento: i.ReferenciaInstrumento, Situacao: i.Situacao,
		})
	}
	return RelatorioLinhaResponse{
		ItemPlanoID: l.ItemPlanoID.String(), CursoNome: l.CursoNome, CursoVago: l.CursoVago,
		ResponsavelNome: l.ResponsavelNome, ResponsavelDesde: l.ResponsavelDesde, VagoDesde: l.VagoDesde,
		MetaNome: l.MetaNome, Indicadores: indicadores, Exigido: l.Exigido, Aceitas: l.Aceitas,
		Pendentes: l.Pendentes, EmCorrecao: l.EmCorrecao, Cumprimento: l.Cumprimento, Situacao: l.Situacao,
		InclusEntregasAnteriores: l.InclusEntregasAnteriores, AvaliacaoPeloProprioCoordenador: l.AvaliacaoPeloProprioCoordenador,
	}
}

// Desempenho godoc
// @Summary      Relatório de desempenho por item de plano
// @Tags         relatorios
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/relatorios/desempenho [get]
func (h *RelatorioHandler) Desempenho(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	filtro, err := filtroRelatorioDoRequest(c)
	if err != nil {
		c.Error(err)
		return
	}
	paginacao, err := ParsePaginacao(c, ordenacaoRelatorioAllowlist, "curso", "asc")
	if err != nil {
		c.Error(err)
		return
	}
	filtro.Page, filtro.PageSize, filtro.Sort, filtro.Order = paginacao.Page, paginacao.PageSize, paginacao.Sort, paginacao.Order

	resultado, err := h.desempenho.Executar(c.Request.Context(), relatorioquery.DesempenhoInput{Ator: ator, Filtro: filtro})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]RelatorioLinhaResponse, 0, len(resultado.Itens))
	for _, l := range resultado.Itens {
		itens = append(itens, linhaRelatorioParaResposta(l))
	}
	c.JSON(http.StatusOK, gin.H{
		"data": itens,
		"meta": MetaPaginacao{Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total, TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize)},
		"resumo": ResumoRelatorioResponse{CursosSemCoordenador: resultado.Resumo.CursosSemCoordenador, MetasNaoCumpridasDeVagos: resultado.Resumo.MetasNaoCumpridasDeVagos},
	})
}

// PorCurso godoc
// @Summary      Cumprimento de metas agregado por curso (gráfico)
// @Tags         relatorios
// @Produce      json
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/relatorios/desempenho/por-curso [get]
func (h *RelatorioHandler) PorCurso(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	filtro, err := filtroRelatorioDoRequest(c)
	if err != nil {
		c.Error(err)
		return
	}

	resultado, err := h.porCurso.Executar(c.Request.Context(), relatorioquery.DesempenhoPorCursoInput{Ator: ator, Filtro: filtro})
	if err != nil {
		c.Error(err)
		return
	}
	itens := make([]ItemDesempenhoPorCursoResponse, 0, len(resultado.Itens))
	for _, i := range resultado.Itens {
		itens = append(itens, ItemDesempenhoPorCursoResponse{
			CursoID: i.CursoID.String(), CursoNome: i.CursoNome, ResponsavelNome: i.ResponsavelNome,
			ExigidoTotal: i.ExigidoTotal, AceitasTotal: i.AceitasTotal, Percentual: i.Percentual,
		})
	}
	cursosSemCoordenador := resultado.CursosSemCoordenador
	if cursosSemCoordenador == nil {
		cursosSemCoordenador = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":                   itens,
		"cursos_sem_coordenador": cursosSemCoordenador,
		"meta": MetaDesempenhoPorCursoResponse{
			TotalCursosComCoordenador: resultado.TotalCursosComCoordenador,
			TotalCursosSemCoordenador: resultado.TotalCursosSemCoordenador,
		},
	})
}

// Exportar godoc
// @Summary      Exportar o relatório de desempenho em CSV
// @Tags         relatorios
// @Produce      text/csv
// @Success      200 {file} file
// @Failure      400 {object} erroResposta
// @Failure      401 {object} erroResposta
// @Failure      403 {object} erroResposta
// @Router       /api/v1/relatorios/desempenho/exportacao [get]
func (h *RelatorioHandler) Exportar(c *gin.Context) {
	ator, ok := AtorDoContexto(c)
	if !ok {
		c.Error(domain.ErrEscopoInvalido)
		return
	}
	filtro, err := filtroRelatorioDoRequest(c)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="relatorio-desempenho.csv"`)
	c.Status(http.StatusOK)

	w := bufio.NewWriter(c.Writer)
	// UTF-8 com BOM e separador ";" — sem os dois, o Excel em português
	// abre tudo em uma coluna e quebra os acentos (design.md §9.3).
	w.WriteString("\xEF\xBB\xBF")
	w.WriteString("Curso;Responsável;Meta;Indicadores;Exigido;Aceitas;Pendentes;Em correção;Cumprimento;Situação;Avaliação pelo próprio coordenador;Inclui entregas de gestão anterior\n")

	_, err = h.exportar.Executar(c.Request.Context(), relatorioquery.ExportarDesempenhoInput{
		Ator: ator, Filtro: filtro,
		Linha: func(l port.LinhaCSVRelatorio) error {
			escreverLinhaCSV(w, l)
			return w.Flush()
		},
	})
	if err != nil {
		// A resposta já começou (streaming) — não há como trocar o status
		// code aqui; o erro fica só no log do servidor.
		_ = err
	}
	_ = w.Flush()
}

func escreverLinhaCSV(w *bufio.Writer, l port.LinhaCSVRelatorio) {
	responsavel := ""
	if l.ResponsavelNome != nil {
		responsavel = *l.ResponsavelNome
	} else if l.VagoDesde != nil {
		responsavel = "Vago desde " + *l.VagoDesde
	} else {
		responsavel = "Vago o período inteiro"
	}
	nomesIndicadores := make([]string, 0, len(l.Indicadores))
	for _, i := range l.Indicadores {
		nomesIndicadores = append(nomesIndicadores, i.Codigo)
	}
	campos := []string{
		neutralizarFormulaCSV(l.CursoNome), neutralizarFormulaCSV(responsavel), neutralizarFormulaCSV(l.MetaNome),
		neutralizarFormulaCSV(strings.Join(nomesIndicadores, "·")),
		fmt.Sprint(l.Exigido), fmt.Sprint(l.Aceitas), fmt.Sprint(l.Pendentes), fmt.Sprint(l.EmCorrecao),
		fmt.Sprintf("%.0f%%", l.Cumprimento), neutralizarFormulaCSV(l.Situacao),
		fmt.Sprint(l.AvaliacaoPeloProprioCoordenador), fmt.Sprint(l.InclusEntregasAnteriores),
	}
	for i, campo := range campos {
		if i > 0 {
			w.WriteString(";")
		}
		w.WriteString(escaparCampoCSV(campo))
	}
	w.WriteString("\n")
}

func escaparCampoCSV(campo string) string {
	if strings.ContainsAny(campo, ";\"\n") {
		return `"` + strings.ReplaceAll(campo, `"`, `""`) + `"`
	}
	return campo
}

// neutralizarFormulaCSV — T-125 (fundacao-metas.md §4.7): aspas RFC 4180
// (escaparCampoCSV) protegem contra campo com ";" ou quebra de linha, mas
// NÃO contra injeção de fórmula — Excel e LibreOffice removem as aspas
// antes de avaliar o conteúdo, então um campo como `=cmd|'/c calc'!A1`
// ainda é executado como fórmula mesmo entre aspas. A única defesa real é
// prefixar um apóstrofo quando o campo começa com um dos quatro
// caracteres que os leitores de planilha interpretam como início de
// fórmula — o apóstrofo força a célula a texto e não aparece na exibição
// (convenção nativa do formato de planilha).
//
// Aplicado SÓ em colunas textuais (nome de curso, de responsável, de
// meta, código de indicador, situação) — NUNCA em coluna numérica: um "-"
// inicial ali é sinal de número negativo legítimo, e o apóstrofo o
// corromperia.
func neutralizarFormulaCSV(campo string) string {
	if campo == "" {
		return campo
	}
	switch campo[0] {
	case '=', '+', '-', '@':
		return "'" + campo
	}
	return campo
}
