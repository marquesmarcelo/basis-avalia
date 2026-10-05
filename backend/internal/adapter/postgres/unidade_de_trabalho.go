package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type chaveContexto int

const (
	chaveTx chaveContexto = iota
	chaveGanchosPosCommit
)

// ExecutorSQL é o subconjunto de métodos que *sqlx.DB e *sqlx.Tx têm em
// comum — os repositórios programam contra esta interface, nunca contra
// o tipo concreto, para poderem rodar dentro ou fora de transação.
type ExecutorSQL = sqlx.ExtContext

// UnidadeDeTrabalho — adapter Postgres de port.UnidadeDeTrabalho.
type UnidadeDeTrabalho struct {
	db *sqlx.DB
}

func NovaUnidadeDeTrabalho(db *sqlx.DB) *UnidadeDeTrabalho {
	return &UnidadeDeTrabalho{db: db}
}

func (u *UnidadeDeTrabalho) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	ganchos := &[]func(){}
	ctxComTx := context.WithValue(ctx, chaveTx, tx)
	ctxComTx = context.WithValue(ctxComTx, chaveGanchosPosCommit, ganchos)

	if err := fn(ctxComTx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, gancho := range *ganchos {
		go gancho()
	}
	return nil
}

// Executor devolve a *sqlx.Tx presente no contexto, ou o pool de conexões
// quando a operação roda fora de uma unidade de trabalho (ex: consultas).
func Executor(ctx context.Context, db *sqlx.DB) ExecutorSQL {
	if tx, ok := ctx.Value(chaveTx).(*sqlx.Tx); ok {
		return tx
	}
	return db
}

// AgendarAposCommit registra fn para rodar (em goroutine própria) só depois
// que a transação da unidade de trabalho corrente for commitada com
// sucesso — é o que resolve "syslog fire-and-forget após o commit" sem
// arriscar mandar o evento de uma transação que acabou revertendo
// (design.md §8.1, D-09). Fora de uma unidade de trabalho, dispara direto:
// não há commit a esperar.
func AgendarAposCommit(ctx context.Context, fn func()) {
	if ganchos, ok := ctx.Value(chaveGanchosPosCommit).(*[]func()); ok {
		*ganchos = append(*ganchos, fn)
		return
	}
	go fn()
}
