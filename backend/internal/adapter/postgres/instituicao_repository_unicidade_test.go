package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
)

func novaInstituicaoTeste(t *testing.T, nome, sigla, codigoEMec string) *instituicao.Instituicao {
	t.Helper()
	siglaVO, err := valueobject.NovaSigla(sigla)
	if err != nil {
		t.Fatalf("sigla: %v", err)
	}
	codigoVO, err := valueobject.NovoCodigoEMec(codigoEMec)
	if err != nil {
		t.Fatalf("codigo e-mec: %v", err)
	}
	nova, err := instituicao.NovaInstituicao(nome, siglaVO, codigoVO)
	if err != nil {
		t.Fatalf("NovaInstituicao: %v", err)
	}
	return nova
}

func TestInstituicaoRepository_I04_SiglaDuplicadaEhRecusada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)
	esc := escopoDePlataforma(t)

	primeira := novaInstituicaoTeste(t, "Faculdade Um", "IRU1", "")
	if err := repo.Inserir(context.Background(), esc, primeira); err != nil {
		t.Fatalf("inserir primeira: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM instituicao WHERE id = $1`, primeira.ID) })

	segunda := novaInstituicaoTeste(t, "Faculdade Dois", "IRU1", "")
	err := repo.Inserir(context.Background(), esc, segunda)
	if err != domain.ErrSiglaDuplicada {
		t.Fatalf("esperava ErrSiglaDuplicada, obtido %v", err)
	}
}

// TestInstituicaoRepository_CodigoEMecUnicoEntreTodasNaoExcluidas prova a
// regra atual (revisão de P15 pelo dono do produto): o código e-MEC é
// único entre TODAS as instituições não excluídas, ativas e inativas —
// não só entre as ativas.
func TestInstituicaoRepository_CodigoEMecUnicoEntreTodasNaoExcluidas(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)
	esc := escopoDePlataforma(t)

	inativa := novaInstituicaoTeste(t, "Faculdade Inativa", "IRU2", "99887")
	if err := repo.Inserir(context.Background(), esc, inativa); err != nil {
		t.Fatalf("inserir inativa: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM instituicao WHERE id = $1`, inativa.ID) })
	if err := db.QueryRow(`UPDATE instituicao SET situacao = 'inativa' WHERE id = $1 RETURNING id`, inativa.ID).Scan(new(string)); err != nil {
		t.Fatalf("inativar: %v", err)
	}

	ativa := novaInstituicaoTeste(t, "Faculdade Ativa", "IRU3", "99887")
	err := repo.Inserir(context.Background(), esc, ativa)
	if err != domain.ErrCodigoEMecDuplicado {
		t.Fatalf("esperava ErrCodigoEMecDuplicado mesmo com a outra inativa, obtido %v", err)
	}
}

func TestInstituicaoRepository_I12_AtualizarComVersaoVelhaDevolveConflito(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)
	esc := escopoDePlataforma(t)

	nova := novaInstituicaoTeste(t, "Faculdade Versao", "IRU4", "")
	if err := repo.Inserir(context.Background(), esc, nova); err != nil {
		t.Fatalf("inserir: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM instituicao WHERE id = $1`, nova.ID) })

	nova.Nome = "Faculdade Versao Renomeada"
	err := repo.Atualizar(context.Background(), esc, nova, 999)
	if err != domain.ErrConflitoDeVersao {
		t.Fatalf("esperava ErrConflitoDeVersao, obtido %v", err)
	}
}
