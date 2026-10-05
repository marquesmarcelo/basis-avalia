package postgres

import "strings"

// escaparCuringasLike escapa "\", "%" e "_" — os três caracteres com
// significado especial em LIKE/ILIKE do Postgres — antes do termo entrar
// entre os "%" que montam a busca por substring (design.md §16 T-108,
// achado O-4). Sem isto, um termo de busca com "%" ou "_" deixa de ser
// tratado como texto literal e vira curinga: buscar "100%" devolveria
// qualquer linha que começasse com "100", e um "%" isolado devolveria a
// tabela inteira. A ordem importa — "\" precisa ser escapado primeiro,
// senão os escapes de "%" e "_" seriam escapados de novo.
func escaparCuringasLike(termo string) string {
	escapado := strings.ReplaceAll(termo, `\`, `\\`)
	escapado = strings.ReplaceAll(escapado, "%", `\%`)
	escapado = strings.ReplaceAll(escapado, "_", `\_`)
	return escapado
}
