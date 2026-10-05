package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// TestIndicadorRepository_IV04_CatalogoComumVisivelATodasSemVazarInstituicao
// prova IV-04: a exceção do catálogo comum traz os indicadores de
// plataforma para QUALQUER instituição, mas nunca o indicador próprio de
// outra instituição — e a única linha sem instituição é de escopo
// plataforma.
func TestIndicadorRepository_IV04_CatalogoComumVisivelATodasSemVazarInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoIndicadorRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})

	inep := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "1.4"})
	proprioFSA := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: fsaID, Codigo: "GEST-01"})
	proprioIVV := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: ivvID, Codigo: "GEST-01"})

	escopoFSA := escopoCatalogoDeIndicadores(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarIndicadores{Page: 1, PageSize: 100, Situacao: "todos"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	vistos := map[string]bool{}
	for _, item := range resultado.Itens {
		vistos[item.Indicador.ID.String()] = true
		if item.Indicador.InstituicaoID == nil {
			if item.Indicador.Escopo.PertenceAInstituicao() {
				t.Fatalf("linha sem instituição deveria ser de escopo plataforma: %v", item.Indicador)
			}
		}
	}
	if !vistos[inep.String()] {
		t.Fatal("esperava ver o indicador do catálogo comum")
	}
	if !vistos[proprioFSA.String()] {
		t.Fatal("esperava ver o indicador próprio da FSA")
	}
	if vistos[proprioIVV.String()] {
		t.Fatal("NÃO deveria ver o indicador próprio do IVV — vazamento entre instituições")
	}

	// Busca direta também respeita o recorte (404 fora da carteira).
	if _, err := repo.BuscarPorID(context.Background(), escopoFSA, proprioIVV); err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado ao buscar indicador do IVV pela FSA, obtido %v", err)
	}
	if _, err := repo.BuscarPorID(context.Background(), escopoFSA, inep); err != nil {
		t.Fatalf("esperava enxergar o indicador do catálogo comum, obtido erro %v", err)
	}
}

// TestIndicadorPlataformaRepository_IE03_CodigoUnicoNaInstalacao prova
// IE-03: dois "1.4" no catálogo do INEP colidem, mas um "1.4" próprio de
// uma instituição não colide com o do INEP.
func TestIndicadorPlataformaRepository_IE03_CodigoUnicoNaInstalacao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repoPlataforma := NovoIndicadorPlataformaRepository(db)
	repoInstituicao := NovoIndicadorRepository(db)
	escPlataforma := escopoIndicadoresDaPlataforma(t)

	primeiro, err := indicador.NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repoPlataforma.Inserir(context.Background(), escPlataforma, primeiro); err != nil {
		t.Fatalf("inserir primeiro: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, primeiro.ID) })

	segundo, err := indicador.NovoDaPlataforma("1.4", "Outro nome", "", "outra referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repoPlataforma.Inserir(context.Background(), escPlataforma, segundo); err != domain.ErrCodigoIndicadorDuplicado {
		t.Fatalf("esperava ErrCodigoIndicadorDuplicado, obtido %v", err)
	}

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	proprio, err := indicador.NovoDaInstituicao(fsaID, "1.4", "Comissão própria de NDE", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	escFSA := escopoCatalogoDeIndicadores(t, fsaID)
	if err := repoInstituicao.Inserir(context.Background(), escFSA, proprio); err != nil {
		t.Fatalf("indicador próprio com o mesmo código do INEP deveria ser permitido: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, proprio.ID) })
}

// TestIndicadorRepository_IN03_CodigoUnicoPorInstituicao prova IN-03: o
// mesmo código em instituições diferentes não colide.
func TestIndicadorRepository_IN03_CodigoUnicoPorInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoIndicadorRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})

	primeiro, err := indicador.NovoDaInstituicao(fsaID, "GEST-01", "Reuniões com representação discente", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Inserir(context.Background(), escopoCatalogoDeIndicadores(t, fsaID), primeiro); err != nil {
		t.Fatalf("inserir na FSA: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, primeiro.ID) })

	repetidoNaFSA, err := indicador.NovoDaInstituicao(fsaID, "GEST-01", "Outro", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Inserir(context.Background(), escopoCatalogoDeIndicadores(t, fsaID), repetidoNaFSA); err != domain.ErrCodigoIndicadorDuplicado {
		t.Fatalf("esperava ErrCodigoIndicadorDuplicado na mesma instituição, obtido %v", err)
	}

	noIVV, err := indicador.NovoDaInstituicao(ivvID, "GEST-01", "Reuniões com representação discente", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Inserir(context.Background(), escopoCatalogoDeIndicadores(t, ivvID), noIVV); err != nil {
		t.Fatalf("esperava permitir o mesmo código em outra instituição, obtido %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, noIVV.ID) })
}

// TestIndicadorRepository_AtualizarEAlterarSituacao_ComEscopoInstitucional
// exercita Atualizar e AlterarSituacao com escopo INSTITUCIONAL de verdade
// (ao contrário do escopo de plataforma, que não emite placeholder algum
// e por isso não pega erro de contagem de $N) — é o teste que teria pego
// o desalinhamento de placeholder entre a cláusula estática e a que
// AplicarEscopo devolve.
func TestIndicadorRepository_AtualizarEAlterarSituacao_ComEscopoInstitucional(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoIndicadorRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	proprioID := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: fsaID, Codigo: "GEST-05"})

	esc := escopoCatalogoDeIndicadores(t, fsaID)
	item, err := repo.BuscarPorID(context.Background(), esc, proprioID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	atual := item.Indicador
	if err := atual.AtualizarDaInstituicao("GEST-05", "Nome corrigido", "descrição nova"); err != nil {
		t.Fatalf("atualizar domínio: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, &atual, item.Indicador.Versao); err != nil {
		t.Fatalf("Atualizar com escopo institucional: %v", err)
	}

	if err := repo.AlterarSituacao(context.Background(), esc, proprioID, valueobject.CatalogoInativo, atual.Versao); err != nil {
		t.Fatalf("AlterarSituacao com escopo institucional: %v", err)
	}

	relido, err := repo.BuscarPorID(context.Background(), esc, proprioID)
	if err != nil {
		t.Fatalf("reler: %v", err)
	}
	if relido.Indicador.Nome.String() != "Nome corrigido" {
		t.Fatalf("esperava nome corrigido, obtido %q", relido.Indicador.Nome.String())
	}
	if relido.Indicador.Situacao != valueobject.CatalogoInativo {
		t.Fatal("esperava situação inativa")
	}
}

// TestIndicadorPlataformaRepository_IE07_ExcluirBloqueadoComUso prova
// IE-07: exclusão bloqueada com a contagem TOTAL (todas as instituições).
func TestIndicadorPlataformaRepository_IE07_ExcluirBloqueadoComUso(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repoPlataforma := NovoIndicadorPlataformaRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	inep := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{})
	testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: fsaID, Indicadores: []uuid.UUID{inep}})
	testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: ivvID, Indicadores: []uuid.UUID{inep}})

	escPlataforma := escopoIndicadoresDaPlataforma(t)
	err := repoPlataforma.ExcluirSeSemUso(context.Background(), escPlataforma, inep)
	var comMeta *domain.ErrIndicadorComMeta
	if !errors.As(err, &comMeta) {
		t.Fatalf("esperava *domain.ErrIndicadorComMeta, obtido %v", err)
	}
	if comMeta.Total != 2 {
		t.Fatalf("esperava contagem total 2 (das duas instituições), obtido %d", comMeta.Total)
	}
}

// TestIndicadorRepository_IN05_ExcluirBloqueadoComUsoNaInstituicao prova
// IN-05: mesma mecânica, restrita à instituição.
func TestIndicadorRepository_IN05_ExcluirBloqueadoComUsoNaInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoIndicadorRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	proprio := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: fsaID})
	testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: fsaID, Indicadores: []uuid.UUID{proprio}})

	esc := escopoCatalogoDeIndicadores(t, fsaID)
	err := repo.ExcluirSeSemUso(context.Background(), esc, proprio)
	var comMeta *domain.ErrIndicadorComMeta
	if !errors.As(err, &comMeta) {
		t.Fatalf("esperava *domain.ErrIndicadorComMeta, obtido %v", err)
	}
	if comMeta.Total != 1 {
		t.Fatalf("esperava contagem 1, obtido %d", comMeta.Total)
	}

	// Um recém-criado, sem meta, é excluído normalmente (IN-05, segunda parte).
	semUso := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: fsaID})
	if err := repo.ExcluirSeSemUso(context.Background(), esc, semUso); err != nil {
		t.Fatalf("esperava excluir sem erro, obtido %v", err)
	}
}

// TestIndicadorPlataformaRepository_IE10_ConflitoDeVersao prova IE-10.
func TestIndicadorPlataformaRepository_IE10_ConflitoDeVersao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoIndicadorPlataformaRepository(db)
	esc := escopoIndicadoresDaPlataforma(t)

	ind, err := indicador.NovoDaPlataforma("2.1", "Núcleo de apoio ao discente", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Inserir(context.Background(), esc, ind); err != nil {
		t.Fatalf("inserir: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, ind.ID) })

	copiaDesatualizada := *ind
	if err := ind.AtualizarDaPlataforma("2.1", "Nome corrigido", "", "referência corrigida"); err != nil {
		t.Fatalf("atualizar: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, ind, 1); err != nil {
		t.Fatalf("primeira atualização deveria funcionar: %v", err)
	}

	if err := copiaDesatualizada.AtualizarDaPlataforma("2.1", "Outra tentativa", "", "outra referência"); err != nil {
		t.Fatalf("atualizar cópia: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, &copiaDesatualizada, 1); err != domain.ErrConflitoDeVersao {
		t.Fatalf("esperava ErrConflitoDeVersao na segunda escrita com versão desatualizada, obtido %v", err)
	}
}
