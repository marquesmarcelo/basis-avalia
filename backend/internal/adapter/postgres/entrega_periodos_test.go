package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// T-113: "os períodos em que eu tenho metas" é projeção da CARTEIRA do
// coordenador, não o catálogo da instituição — um período em que ele não
// tem nenhum item de plano vigente não pode aparecer no seletor (levaria
// a uma tela vazia). Prova as três exclusões: curso fora da carteira,
// plano em rascunho, e confirma a inclusão do período correto.
func TestEntregaRepository_T113_ListarPeriodosParaMinhasMetas(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})

	// Período A: curso da carteira do coordenador, plano vigente — DEVE aparecer.
	periodoA := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})
	cursoDaCarteira := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})
	planoA := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoDaCarteira, PeriodoID: periodoA, SituacaoPublicacao: "vigente"})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoA, CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1})

	// Período B: curso FORA da carteira (sem designação para este
	// coordenador), plano vigente — NÃO deve aparecer.
	periodoB := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-08-01", DataFim: "2026-12-20"})
	cursoAlheio := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	planoB := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoAlheio, PeriodoID: periodoB, SituacaoPublicacao: "vigente"})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoB, CursoID: cursoAlheio, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1})

	// Período C: curso da carteira, mas plano em RASCUNHO — NÃO deve aparecer.
	periodoC := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2025-01-01", DataFim: "2025-07-30"})
	planoC := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoDaCarteira, PeriodoID: periodoC, SituacaoPublicacao: "rascunho"})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoC, CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1})

	repo := NovoEntregaRepository(db)
	esc := escopoEntregaDaCarteira(t, instituicaoID, coordenadorID)

	opcoes, err := repo.ListarPeriodosParaMinhasMetas(context.Background(), esc)
	if err != nil {
		t.Fatalf("ListarPeriodosParaMinhasMetas: %v", err)
	}

	ids := map[uuid.UUID]bool{}
	for _, o := range opcoes {
		ids[o.ID] = true
	}
	if !ids[periodoA] {
		t.Fatal("esperava o período A (curso da carteira, plano vigente) na lista")
	}
	if ids[periodoB] {
		t.Fatal("período B (curso fora da carteira) NÃO deveria aparecer — levaria a uma tela vazia")
	}
	if ids[periodoC] {
		t.Fatal("período C (plano em rascunho) NÃO deveria aparecer — rascunho não é obrigação")
	}
	if len(opcoes) != 1 {
		t.Fatalf("esperava exatamente 1 período (sem duplicata do DISTINCT), obteve %d", len(opcoes))
	}
}
