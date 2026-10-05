package port

import "context"

// UnidadeDeTrabalho executa fn dentro de uma transação: sucesso comita,
// erro desfaz tudo. A conexão de transação viaja pelo ctx — os
// repositórios a recuperam de lá, caindo no pool quando ausente
// (design.md §4.2).
type UnidadeDeTrabalho interface {
	Executar(ctx context.Context, fn func(ctx context.Context) error) error
}
