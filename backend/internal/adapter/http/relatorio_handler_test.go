package http

import (
	"bufio"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/port"
)

// T-125 (fundacao-metas.md §4.7): aspas RFC 4180 não neutralizam fórmula —
// o Excel e o LibreOffice removem as aspas antes de avaliar. Só o
// apóstrofo de prefixo neutraliza de verdade.
func TestNeutralizarFormulaCSV_PrefixaOsQuatroCaracteresDeFormula(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado string
	}{
		{"=cmd|'/c calc'!A1", "'=cmd|'/c calc'!A1"},
		{"+1+1", "'+1+1"},
		{"-1+1", "'-1+1"},
		{"@SUM(A1:A2)", "'@SUM(A1:A2)"},
		{"Engenharia de Software", "Engenharia de Software"},
		{"", ""},
	}
	for _, c := range casos {
		if got := neutralizarFormulaCSV(c.entrada); got != c.esperado {
			t.Errorf("neutralizarFormulaCSV(%q) = %q, esperado %q", c.entrada, got, c.esperado)
		}
	}
}

// Coluna numérica nunca passa por neutralizarFormulaCSV — um "-10%" de
// variação legítima não pode virar "'-10%" (texto corrompido na planilha).
func TestEscreverLinhaCSV_ColunaNumericaComSinalNegativoNaoEhPrefixada(t *testing.T) {
	var buf strings.Builder
	w := bufio.NewWriter(&buf)
	nome := "Curso Regular"
	l := port.LinhaCSVRelatorio{
		CursoNome: nome, ResponsavelNome: &nome, MetaNome: "Meta X",
		Exigido: 5, Aceitas: -1, Pendentes: 2, EmCorrecao: 0,
		Cumprimento: -10, Situacao: "pendente",
	}
	escreverLinhaCSV(w, l)
	_ = w.Flush()
	linha := buf.String()
	if strings.Contains(linha, "'-1") || strings.Contains(linha, "'-10%") {
		t.Fatalf("coluna numérica não deveria ser prefixada com apóstrofo: %s", linha)
	}
	if !strings.Contains(linha, ";-1;") {
		t.Fatalf("esperava -1 intacto na coluna Aceitas: %s", linha)
	}
	if !strings.Contains(linha, "-10%") {
		t.Fatalf("esperava -10%% intacto na coluna Cumprimento: %s", linha)
	}
}

// Coluna textual controlada pelo usuário (nome de curso) é o vetor real:
// quem cria o curso escolhe o nome, e esse nome vira a primeira célula da
// linha exportada.
func TestEscreverLinhaCSV_NomeDeCursoComFormulaEhNeutralizado(t *testing.T) {
	var buf strings.Builder
	w := bufio.NewWriter(&buf)
	l := port.LinhaCSVRelatorio{
		CursoNome: "=HYPERLINK(\"http://atacante.example\";\"clique\")",
		MetaNome:  "Meta X", Situacao: "pendente",
	}
	escreverLinhaCSV(w, l)
	_ = w.Flush()
	linha := buf.String()
	if !strings.HasPrefix(linha, `"'=HYPERLINK`) {
		t.Fatalf("esperava apóstrofo prefixando a fórmula (dentro das aspas RFC 4180 por causa do ;), obteve: %s", linha)
	}
}
