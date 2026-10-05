package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func escopoDeInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	ator = ator.ComDataDeReferencia(hoje)
	esc, err := autorizacao.Autorizar(ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return esc
}

// TestPlanoRepository_PL02_SegundoPlanoDoMesmoCursoEPeriodoERecusado prova
// PL-02/design.md V-1: a garantia é o índice único, não a verificação
// prévia — testado com DUAS transações concorrentes reais.
func TestPlanoRepository_PL02_SegundoPlanoDoMesmoCursoEPeriodoERecusado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})

	repo := NovoPlanoRepository(db)
	esc := escopoDeInstituicao(t, instituicaoID)

	criarPlanoDeTeste := func() *plano.Plano {
		p, err := plano.NovoPlano(instituicaoID, cursoID, periodoID,
			plano.DadosDoPlano{Descricao: "D", ObjetivoGeral: "O", ResultadosEsperados: "R"})
		if err != nil {
			t.Fatalf("NovoPlano: %v", err)
		}
		return p
	}

	var wg sync.WaitGroup
	erros := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			erros[i] = repo.Inserir(context.Background(), esc, criarPlanoDeTeste())
		}(i)
	}
	wg.Wait()

	sucessos, duplicados := 0, 0
	for _, err := range erros {
		switch err {
		case nil:
			sucessos++
		case domain.ErrPlanoDuplicado:
			duplicados++
		default:
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if sucessos != 1 || duplicados != 1 {
		t.Fatalf("esperava exatamente 1 sucesso e 1 duplicado, obtido %d sucesso(s) e %d duplicado(s)", sucessos, duplicados)
	}

	t.Cleanup(func() { db.Exec(`DELETE FROM plano WHERE curso_id = $1 AND periodo_id = $2`, cursoID, periodoID) })

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM plano WHERE curso_id = $1 AND periodo_id = $2 AND excluido_em IS NULL`, cursoID, periodoID); err != nil {
		t.Fatalf("contar planos: %v", err)
	}
	if total != 1 {
		t.Fatalf("esperava exatamente 1 plano gravado, obtido %d", total)
	}
}
