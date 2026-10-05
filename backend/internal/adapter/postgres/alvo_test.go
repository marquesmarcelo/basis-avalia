package postgres

import "testing"

// TestAlvos_ExcecaoDoCatalogoEmExatamenteUm é teste de MECANISMO
// (fundacao-metas.md §3.7, restrição 4): admiteCatalogoComum precisa ser
// verdadeiro em exatamente um Alvo, e esse um precisa ser AlvoIndicador.
// Zero apaga a exceção que indicadores precisa; dois ou mais reproduzem a
// exceção do catálogo comum em uma tabela que não deveria tê-la.
func TestAlvos_ExcecaoDoCatalogoEmExatamenteUm(t *testing.T) {
	var comExcecao []string
	for _, alvo := range todosOsAlvos {
		if alvo.admiteCatalogoComum {
			comExcecao = append(comExcecao, alvo.alias)
		}
	}
	if len(comExcecao) != 1 {
		t.Fatalf("esperava exatamente 1 alvo com admiteCatalogoComum, obtido %d: %v", len(comExcecao), comExcecao)
	}
	if comExcecao[0] != AlvoIndicador.alias {
		t.Fatalf("o único alvo com a exceção deveria ser AlvoIndicador, obtido %q", comExcecao[0])
	}
}
