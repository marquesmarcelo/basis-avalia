package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/testhelpers"
)

func TestAuditoriaRepository_RegistroLocalEhDesfeitoComATransacao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	uow := NovaUnidadeDeTrabalho(db)
	repo := NovoAuditoriaRepository(db)

	evento := auditoria.NovoEvento(auditoria.CriarUsuario, auditoria.ResultadoSucesso)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE id = $1`, evento.ID)
	})

	err := uow.Executar(context.Background(), func(ctx context.Context) error {
		if err := repo.Registrar(ctx, evento); err != nil {
			return err
		}
		return errors.New("falha proposital depois de registrar a auditoria")
	})
	if err == nil {
		t.Fatal("esperava erro propagado")
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM auditoria WHERE id = $1`, evento.ID); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if total != 0 {
		t.Fatal("o registro de auditoria deveria ter sido desfeito junto com a transação")
	}
}

func TestAuditoriaRepository_SucessoPersisteORegistro(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	uow := NovaUnidadeDeTrabalho(db)
	repo := NovoAuditoriaRepository(db)

	evento := auditoria.NovoEvento(auditoria.CriarInstituicao, auditoria.ResultadoSucesso)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE id = $1`, evento.ID)
	})

	err := uow.Executar(context.Background(), func(ctx context.Context) error {
		return repo.Registrar(ctx, evento)
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM auditoria WHERE id = $1`, evento.ID); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if total != 1 {
		t.Fatal("o registro de auditoria deveria ter sido persistido")
	}
}
