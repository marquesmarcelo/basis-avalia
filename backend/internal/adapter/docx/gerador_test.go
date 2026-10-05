package docx

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
)

func dadosDeTeste() port.DadosDoDocumento {
	return port.DadosDoDocumento{
		InstituicaoNome: "Faculdade São Aleixo", InstituicaoSigla: "FSA",
		CursoNome: "Engenharia de Software", CursoGrau: "bacharelado", CursoModalidade: "presencial", CursoCodigoEMec: "123456",
		CoordenadorNome: "Ana Costa", CoordenadorPortaria: "12/2026",
		PeriodoNome: "2026.1", PeriodoInicio: "01/01/2026", PeriodoFim: "30/07/2026",
		Descricao: "Descrição do plano", ObjetivoGeral: "Objetivo geral", ResultadosEsperados: "Resultados esperados",
		AlinhamentoPDI: "Alinhamento PDI", AlinhamentoPPC: "Alinhamento PPC",
		Itens: []port.ItemDoDocumento{
			{MetaNome: "Registrar reuniões de NDE em ata", Indicadores: []string{"1.4", "1.5"}, Quantidade: 4},
			{MetaNome: "Relatório de acompanhamento", Indicadores: []string{"1.5"}, Quantidade: 1},
		},
		AprovacaoData: "10/02/2026", AprovacaoOrgao: "NDE",
		MarcaSituacao: "", GeradoEmTexto: "15/03/2026 14:30",
	}
}

func documentoXMLDe(t *testing.T, saida []byte) string {
	t.Helper()
	leitor, err := zip.NewReader(bytes.NewReader(saida), int64(len(saida)))
	if err != nil {
		t.Fatalf("zip inválido: %v", err)
	}
	for _, f := range leitor.File {
		if f.Name == nomeParteDocumento {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("abrir document.xml: %v", err)
			}
			defer rc.Close()
			var b bytes.Buffer
			b.ReadFrom(rc)
			return b.String()
		}
	}
	t.Fatal("word/document.xml não encontrado no .docx gerado")
	return ""
}

// TestGerar_ZeroMarcadoresVazados prova P-08: o modelo versionado, uma vez
// renderizado com todos os campos, não deixa nenhum "{{" no resultado.
func TestGerar_ZeroMarcadoresVazados(t *testing.T) {
	g := Novo()
	saida, err := g.Gerar(dadosDeTeste())
	if err != nil {
		t.Fatalf("Gerar: %v", err)
	}
	doc := documentoXMLDe(t, saida)
	if strings.Contains(doc, "{{") {
		t.Fatalf("marcador vazado no documento gerado:\n%s", doc)
	}
}

// TestGerar_ClonaLinhaPorItem prova a repetição de linha (design.md §7.2,
// item 3): duas metas produzem duas ocorrências do nome de cada uma.
func TestGerar_ClonaLinhaPorItem(t *testing.T) {
	g := Novo()
	saida, err := g.Gerar(dadosDeTeste())
	if err != nil {
		t.Fatalf("Gerar: %v", err)
	}
	doc := documentoXMLDe(t, saida)
	if !strings.Contains(doc, "Registrar reuniões de NDE em ata") {
		t.Fatal("primeira meta não apareceu no documento")
	}
	if !strings.Contains(doc, "Relatório de acompanhamento") {
		t.Fatal("segunda meta não apareceu no documento")
	}
	if strings.Count(doc, "<w:tr>") < 3 { // cabeçalho + 2 itens
		t.Fatalf("esperava ao menos 3 linhas de tabela, obtido XML: %s", doc)
	}
}

// TestGerar_EscapaXMLEQuebraDeLinha prova a segunda armadilha
// (design.md §7.2, item 2): "&", "<" e quebra de linha não produzem XML
// inválido nem descartam o parágrafo.
func TestGerar_EscapaXMLEQuebraDeLinha(t *testing.T) {
	dados := dadosDeTeste()
	dados.Descricao = "Comissão & Avaliação <interna>\nSegunda linha"
	g := Novo()
	saida, err := g.Gerar(dados)
	if err != nil {
		t.Fatalf("Gerar: %v", err)
	}
	doc := documentoXMLDe(t, saida)
	if strings.Contains(doc, "Comissão & Avaliação") {
		t.Fatal("'&' não deveria aparecer cru no XML")
	}
	if !strings.Contains(doc, "&amp;") {
		t.Fatal("'&' deveria ter sido escapado para &amp;")
	}
	if !strings.Contains(doc, "<w:br/>") {
		t.Fatal("quebra de linha deveria ter virado <w:br/>")
	}
}

// TestGerar_MarcadorAusenteFalhaAlto prova P-08: se o modelo referenciar
// um marcador que os dados não cobrem, o gerador falha — nunca produz
// documento com "{{" vazado.
func TestGerar_MarcadorAusenteFalhaAlto(t *testing.T) {
	_, err := renderizarDocumento(`<w:document><w:body><w:p><w:r><w:t>{{campo_inexistente}}</w:t></w:r></w:p></w:body></w:document>`, dadosDeTeste())
	if err != domain.ErrMarcadorNaoEncontrado {
		t.Fatalf("esperava ErrMarcadorNaoEncontrado, obtido %v", err)
	}
}
