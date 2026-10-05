package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func TestUnidadeDeTrabalho_T013_ErroDesfazAsDuasInsercoes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	uow := NovaUnidadeDeTrabalho(db)

	idA := uuid.Must(uuid.NewV7())
	idB := uuid.Must(uuid.NewV7())
	t.Cleanup(func() {
		db.Exec(`DELETE FROM instituicao WHERE id IN ($1,$2)`, idA, idB)
	})

	err := uow.Executar(context.Background(), func(ctx context.Context) error {
		ex := Executor(ctx, db)
		if _, err := ex.ExecContext(ctx, `INSERT INTO instituicao (id, nome, sigla) VALUES ($1,$2,$3)`, idA, "Fac Uow A", "UOWA"); err != nil {
			return err
		}
		// Segunda inserção falha propositalmente (violação de NOT NULL em nome).
		if _, err := ex.ExecContext(ctx, `INSERT INTO instituicao (id, nome, sigla) VALUES ($1,$2,$3)`, idB, nil, "UOWB"); err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		t.Fatal("esperava erro na segunda inserção")
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM instituicao WHERE id IN ($1,$2)`, idA, idB); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if total != 0 {
		t.Fatalf("esperava 0 linhas após rollback, obtido %d", total)
	}
}

func TestUnidadeDeTrabalho_SucessoPersisteAsDuasInsercoes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	uow := NovaUnidadeDeTrabalho(db)

	idA := uuid.Must(uuid.NewV7())
	idB := uuid.Must(uuid.NewV7())
	t.Cleanup(func() {
		db.Exec(`DELETE FROM instituicao WHERE id IN ($1,$2)`, idA, idB)
	})

	err := uow.Executar(context.Background(), func(ctx context.Context) error {
		ex := Executor(ctx, db)
		if _, err := ex.ExecContext(ctx, `INSERT INTO instituicao (id, nome, sigla) VALUES ($1,$2,$3)`, idA, "Fac Uow C", "UOWC"); err != nil {
			return err
		}
		if _, err := ex.ExecContext(ctx, `INSERT INTO instituicao (id, nome, sigla) VALUES ($1,$2,$3)`, idB, "Fac Uow D", "UOWD"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM instituicao WHERE id IN ($1,$2)`, idA, idB); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if total != 2 {
		t.Fatalf("esperava 2 linhas após commit, obtido %d", total)
	}
}
