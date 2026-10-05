package docx

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
)

//go:embed modelo_plano_acao.docx
var modeloFS embed.FS

const nomeParteDocumento = "word/document.xml"

const (
	marcadorInicioItens = "{{#itens}}"
	marcadorFimItens    = "{{/itens}}"
)

// Gerador — adapter de port.GeradorDeDocumento sobre archive/zip +
// encoding/xml, sem biblioteca e sem container de conversão
// (specs/plano-acao/design.md §7.1, P-07). O modelo é um .docx real,
// versionado no repositório: substituir {{marcador}} em word/document.xml
// preserva estilo, cabeçalho e rodapé que a instituição definiu lá.
type Gerador struct{}

func Novo() *Gerador { return &Gerador{} }

func (g *Gerador) Gerar(dados port.DadosDoDocumento) ([]byte, error) {
	modeloBytes, err := modeloFS.ReadFile("modelo_plano_acao.docx")
	if err != nil {
		return nil, err
	}
	leitor, err := zip.NewReader(bytes.NewReader(modeloBytes), int64(len(modeloBytes)))
	if err != nil {
		return nil, err
	}

	var saida bytes.Buffer
	escritor := zip.NewWriter(&saida)
	for _, arquivo := range leitor.File {
		conteudo, err := lerArquivoDoZip(arquivo)
		if err != nil {
			return nil, err
		}
		if arquivo.Name == nomeParteDocumento {
			renderizado, err := renderizarDocumento(string(conteudo), dados)
			if err != nil {
				return nil, err
			}
			conteudo = []byte(renderizado)
		}
		w, err := escritor.Create(arquivo.Name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(conteudo)); err != nil {
			return nil, err
		}
	}
	if err := escritor.Close(); err != nil {
		return nil, err
	}
	return saida.Bytes(), nil
}

func lerArquivoDoZip(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// renderizarDocumento aplica as três mitigações de
// specs/plano-acao/design.md §7.2, nesta ordem: 1) clona a linha da
// tabela entre {{#itens}} e {{/itens}}; 2) substitui os marcadores
// escalares com xml.EscapeText e \n -> <w:br/>; 3) falha alto (P-08) se
// sobrar qualquer "{{" — marcador declarado no modelo sem correspondente
// nos dados, nunca um documento com marcador vazado.
func renderizarDocumento(xmlBruto string, dados port.DadosDoDocumento) (string, error) {
	comItens, err := clonarLinhaDeItens(xmlBruto, dados)
	if err != nil {
		return "", err
	}
	final := substituirEscalares(comItens, escalaresDe(dados))
	if strings.Contains(final, "{{") {
		return "", domain.ErrMarcadorNaoEncontrado
	}
	return final, nil
}

func clonarLinhaDeItens(xmlBruto string, dados port.DadosDoDocumento) (string, error) {
	inicioMarcador := strings.Index(xmlBruto, marcadorInicioItens)
	fimMarcador := strings.Index(xmlBruto, marcadorFimItens)
	if inicioMarcador == -1 || fimMarcador == -1 {
		return "", domain.ErrMarcadorNaoEncontrado
	}

	inicioLinha := strings.LastIndex(xmlBruto[:inicioMarcador], "<w:tr>")
	fimLinha := strings.Index(xmlBruto[fimMarcador:], "</w:tr>")
	if inicioLinha == -1 || fimLinha == -1 {
		return "", domain.ErrMarcadorNaoEncontrado
	}
	fimLinha = fimMarcador + fimLinha + len("</w:tr>")

	linhaModelo := xmlBruto[inicioLinha:fimLinha]
	linhaModelo = strings.ReplaceAll(linhaModelo, marcadorInicioItens, "")
	linhaModelo = strings.ReplaceAll(linhaModelo, marcadorFimItens, "")

	var linhasRenderizadas strings.Builder
	for _, item := range dados.Itens {
		linha := linhaModelo
		linha = strings.ReplaceAll(linha, "{{item_meta}}", escaparTexto(item.MetaNome))
		linha = strings.ReplaceAll(linha, "{{item_indicadores}}", escaparTexto(strings.Join(item.Indicadores, "; ")))
		linha = strings.ReplaceAll(linha, "{{item_quantidade}}", strconv.Itoa(item.Quantidade))
		linhasRenderizadas.WriteString(linha)
	}

	return xmlBruto[:inicioLinha] + linhasRenderizadas.String() + xmlBruto[fimLinha:], nil
}

func escalaresDe(d port.DadosDoDocumento) map[string]string {
	aprovacaoData, aprovacaoOrgao, aprovacaoObservacao := "Não aprovado", "", ""
	if d.AprovacaoData != "" {
		aprovacaoData, aprovacaoOrgao = d.AprovacaoData, d.AprovacaoOrgao
	}
	if d.AprovacaoPendente {
		aprovacaoObservacao = "Aprovação ainda não registrada"
	}
	coordenadorNome, coordenadorPortaria := "Sem responsável", ""
	if d.CoordenadorNome != "" {
		coordenadorNome, coordenadorPortaria = d.CoordenadorNome, "(Portaria "+d.CoordenadorPortaria+")"
	}
	return map[string]string{
		"{{instituicao_nome}}":     d.InstituicaoNome,
		"{{instituicao_sigla}}":    d.InstituicaoSigla,
		"{{curso_nome}}":           d.CursoNome,
		"{{curso_grau}}":           d.CursoGrau,
		"{{curso_modalidade}}":     d.CursoModalidade,
		"{{curso_codigo_emec}}":    d.CursoCodigoEMec,
		"{{coordenador_nome}}":     coordenadorNome,
		"{{coordenador_portaria}}": coordenadorPortaria,
		"{{periodo_nome}}":         d.PeriodoNome,
		"{{periodo_inicio}}":       d.PeriodoInicio,
		"{{periodo_fim}}":          d.PeriodoFim,
		"{{descricao}}":            d.Descricao,
		"{{objetivo_geral}}":       d.ObjetivoGeral,
		"{{resultados_esperados}}": d.ResultadosEsperados,
		"{{alinhamento_pdi}}":      d.AlinhamentoPDI,
		"{{alinhamento_ppc}}":      d.AlinhamentoPPC,
		"{{aprovacao_data}}":       aprovacaoData,
		"{{aprovacao_orgao}}":      aprovacaoOrgao,
		"{{aprovacao_observacao}}": aprovacaoObservacao,
		"{{marca_situacao}}":       d.MarcaSituacao,
		"{{gerado_em}}":            d.GeradoEmTexto,
	}
}

func substituirEscalares(xmlBruto string, valores map[string]string) string {
	for marcador, valor := range valores {
		xmlBruto = strings.ReplaceAll(xmlBruto, marcador, escaparTexto(valor))
	}
	return xmlBruto
}

// escaparTexto implementa a segunda armadilha de design.md §7.2: escape
// de XML obrigatório (um "&" ou "<" solto produz .docx que o Word recusa
// abrir) e quebra de linha vira <w:br/> — um "\n" dentro de <w:t> é
// descartado pelo Word, e o documento sairia com os parágrafos colados.
func escaparTexto(bruto string) string {
	linhas := strings.Split(bruto, "\n")
	partes := make([]string, len(linhas))
	for i, linha := range linhas {
		var b strings.Builder
		_ = xml.EscapeText(&b, []byte(linha))
		partes[i] = b.String()
	}
	return strings.Join(partes, `</w:t></w:r><w:r><w:br/><w:t xml:space="preserve">`)
}
