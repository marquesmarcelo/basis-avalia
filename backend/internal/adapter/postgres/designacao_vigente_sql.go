package postgres

import "fmt"

// FragmentoDesignacaoVigente é a ÚNICA função que escreve em SQL a regra de
// vigência de uma designação (specs/cursos/design.md §4.3, C-03). Espelha
// valueobject.Vigencia.SituacaoEm bit a bit — inclusive na forma de
// escrever a fronteira do fim: NOT (data_fim < hoje), nunca
// "data_fim >= hoje" reescrito por engano com o sinal trocado (DG-06, o dia
// da data de fim entra INTEIRO).
//
// Devolve só o predicado de vigência — nunca "excluido_em IS NULL", que
// cada chamador decide separadamente conforme o alias em jogo (o EXISTS da
// sessão usa o alias da própria designacao; AplicarEscopo já acrescenta a
// exclusão lógica do seu alvo antes de chamar esta função).
//
// n é o número do placeholder $N que recebe a data de referência (formato
// "AAAA-MM-DD", DataLocal.String()) — usado duas vezes no fragmento, uma
// para cada lado do intervalo.
//
// Proibido em qualquer adapter: CURRENT_DATE ou now() aqui dentro — é a
// data do servidor, em UTC, e diverge do dia de exibição perto da meia-noite
// de Brasília (fundacao-metas.md §5.1, DG-06).
func FragmentoDesignacaoVigente(alias string, n int) string {
	return fmt.Sprintf(
		"%s.data_inicio <= $%d AND (%s.data_fim IS NULL OR %s.data_fim >= $%d)",
		alias, n, alias, alias, n,
	)
}
