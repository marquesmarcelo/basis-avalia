package postgres

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	metadomain "github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// TestMetaRepository_MC14_FiltrarPorOrigemNaoDuplicaAMeta prova MC-14: a
// meta com 1.4 e 1.5 (ambos "Do INEP"), filtrada por origem "Do INEP",
// aparece UMA vez — a junção multiplicaria; o filtro é EXISTS.
func TestMetaRepository_MC14_FiltrarPorOrigemNaoDuplicaAMeta(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoMetaRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ind14 := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "1.4"})
	ind15 := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "1.5"})
	testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{
		InstituicaoID: fsaID, Nome: "Registrar reuniões de NDE em ata",
		Indicadores: []uuid.UUID{ind14, ind15},
	})

	esc := escopoMetasDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), esc, port.FiltroListarMetas{
		Page: 1, PageSize: 100, Situacao: "todas", Origem: "plataforma",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 1 {
		t.Fatalf("esperava a meta UMA vez (total=1), obtido total=%d", resultado.Total)
	}
	if len(resultado.Itens) != 1 {
		t.Fatalf("esperava 1 item, obtido %d", len(resultado.Itens))
	}
	if len(resultado.Itens[0].Indicadores) != 2 {
		t.Fatalf("esperava os 2 indicadores embutidos, obtido %d", len(resultado.Itens[0].Indicadores))
	}

	// Mesmo vale filtrando por um indicador específico (regra 2 de 3.3).
	resultadoPorIndicador, err := repo.Listar(context.Background(), esc, port.FiltroListarMetas{
		Page: 1, PageSize: 100, Situacao: "todas", IndicadorID: &ind14,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultadoPorIndicador.Total != 1 {
		t.Fatalf("esperava a meta UMA vez ao filtrar por indicador, obtido total=%d", resultadoPorIndicador.Total)
	}
}

// TestMetaRepository_MC14_ConsultaNuncaContemDISTINCT é a comprovação
// estática de design.md §7.1: DISTINCT pareceria corrigir a duplicação,
// mas deixaria o total/agregados errados (relatório plausível e falso).
// Só examina código SQL (linhas com aspas), nunca comentário — o próprio
// texto explicativo deste arquivo cita a palavra "DISTINCT" para dizer que
// ela é proibida.
func TestMetaRepository_MC14_ConsultaNuncaContemDISTINCT(t *testing.T) {
	fonte, err := os.ReadFile("meta_repository.go")
	if err != nil {
		t.Fatalf("lendo o fonte do adapter: %v", err)
	}
	for _, linha := range strings.Split(string(fonte), "\n") {
		semComentario := strings.SplitN(linha, "//", 2)[0]
		if strings.Contains(strings.ToUpper(semComentario), "DISTINCT") {
			t.Fatalf("linha de código com DISTINCT — proibido, ver design.md §7.1: %q", linha)
		}
	}
}

// TestMetaRepository_IsolamentoInstitucional prova que uma meta da FSA não
// é vista nem alcançada pelo IVV.
func TestMetaRepository_IsolamentoInstitucional(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoMetaRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ind := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{})
	metaFSA := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: fsaID, Indicadores: []uuid.UUID{ind}})

	escIVV := escopoMetasDaInstituicao(t, ivvID)
	if _, err := repo.BuscarPorID(context.Background(), escIVV, metaFSA); err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado para meta de outra instituição, obtido %v", err)
	}

	resultado, err := repo.Listar(context.Background(), escIVV, port.FiltroListarMetas{Page: 1, PageSize: 100, Situacao: "todas"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	for _, item := range resultado.Itens {
		if item.Meta.ID == metaFSA {
			t.Fatal("meta da FSA vazou para a listagem do IVV")
		}
	}
}

// TestMeta_MC05_NaoExisteColunaDeQuantidadeExigida é a metade estrutural
// de MC-05 (a parte inteira é coberta em specs/metas-coordenacao): o
// modelo não permite QUANTIDADE EXIGIDA na meta — nunca multiplicar a
// entrega pelo número de indicadores é impossível de representar porque
// essa coluna não existe. Isso continua vale integralmente: a apuração do
// relatório só lê item_plano.quantidade, nunca nada em meta.
//
// Exceção única, registrada (specs/indicadores/design.md, revisão
// pós-produto): `quantidade_sugerida` é um conceito DIFERENTE de
// "quantidade exigida" — é sugestão herdada pelo item do plano só na
// criação, nunca lida pela apuração, e o campo permanece opcional. Por
// isso a checagem usa uma allowlist de uma linha em vez de recusar
// qualquer substring "quantidade" — a regra continua ampla para pegar
// qualquer OUTRA coluna de quantidade (curso, período, INEP), só abre
// exceção nomeada para a que foi deliberadamente aprovada.
func TestMeta_MC05_NaoExisteColunaDeQuantidadeExigida(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	var colunas []string
	err := db.Select(&colunas, `SELECT column_name FROM information_schema.columns WHERE table_name = 'meta'`)
	if err != nil {
		t.Fatalf("consultando colunas de meta: %v", err)
	}
	const excecaoAprovada = "quantidade_sugerida"
	for _, c := range colunas {
		lc := strings.ToLower(c)
		if lc == excecaoAprovada {
			continue
		}
		if strings.Contains(lc, "quantidade") || strings.Contains(lc, "curso") || strings.Contains(lc, "periodo") || strings.Contains(lc, "inep") {
			t.Fatalf("meta não pode ter coluna %q — meta é catálogo, sem quantidade exigida, período, curso ou campo de INEP (design.md §3.3)", c)
		}
	}
}

// TestMetaRepository_MC13_DefinirIndicadoresReescreveOVinculo prova que a
// troca de indicadores (MC-13) é persistida como substituição total.
func TestMetaRepository_MC13_DefinirIndicadoresReescreveOVinculo(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoMetaRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ind15 := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "1.5"})
	ind21 := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "2.1"})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: fsaID, Nome: "Plano de ensino revisado", Indicadores: []uuid.UUID{ind15}})

	esc := escopoMetasDaInstituicao(t, fsaID)
	item, err := repo.BuscarPorID(context.Background(), esc, metaID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	atual := item.Meta
	if _, err := atual.DefinirIndicadores([]uuid.UUID{ind21}); err != nil {
		t.Fatalf("definir indicadores: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, &atual, item.Meta.Versao); err != nil {
		t.Fatalf("atualizar: %v", err)
	}

	relido, err := repo.BuscarPorID(context.Background(), esc, metaID)
	if err != nil {
		t.Fatalf("reler: %v", err)
	}
	if len(relido.Meta.Indicadores) != 1 || relido.Meta.Indicadores[0] != ind21 {
		t.Fatalf("esperava só o indicador 2.1, obtido %v", relido.Meta.Indicadores)
	}
}

// TestMetaRepository_QuantidadeSugerida_PersisteAtualizaELimpa prova o
// ciclo completo da sugestão (indicadores/design.md, revisão pós-produto):
// nasce ausente, Inserir grava um valor, Atualizar troca o valor e depois
// limpa (nil) — cada leitura reflete exatamente o último gravado.
func TestMetaRepository_QuantidadeSugerida_PersisteAtualizaELimpa(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoMetaRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	ind := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "3.1"})
	esc := escopoMetasDaInstituicao(t, fsaID)

	nova, err := metadomain.NovaMeta(fsaID, "Meta com sugestão", "", []uuid.UUID{ind})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Inserir(context.Background(), esc, nova); err != nil {
		t.Fatalf("inserir: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM meta WHERE id = $1`, nova.ID) })

	semSugestao, err := repo.BuscarPorID(context.Background(), esc, nova.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if semSugestao.Meta.QuantidadeSugerida != nil {
		t.Fatalf("esperava sugestão nil ao nascer, obtido %v", semSugestao.Meta.QuantidadeSugerida)
	}

	atual := semSugestao.Meta
	duas := 2
	if err := atual.DefinirQuantidadeSugerida(&duas); err != nil {
		t.Fatalf("definir sugestão: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, &atual, semSugestao.Meta.Versao); err != nil {
		t.Fatalf("atualizar com sugestão: %v", err)
	}
	comSugestao, err := repo.BuscarPorID(context.Background(), esc, nova.ID)
	if err != nil {
		t.Fatalf("reler: %v", err)
	}
	if comSugestao.Meta.QuantidadeSugerida == nil || comSugestao.Meta.QuantidadeSugerida.Int() != 2 {
		t.Fatalf("esperava sugestão 2, obtido %v", comSugestao.Meta.QuantidadeSugerida)
	}

	depois := comSugestao.Meta
	if err := depois.DefinirQuantidadeSugerida(nil); err != nil {
		t.Fatalf("limpar sugestão: %v", err)
	}
	if err := repo.Atualizar(context.Background(), esc, &depois, comSugestao.Meta.Versao); err != nil {
		t.Fatalf("atualizar limpando sugestão: %v", err)
	}
	limpa, err := repo.BuscarPorID(context.Background(), esc, nova.ID)
	if err != nil {
		t.Fatalf("reler após limpar: %v", err)
	}
	if limpa.Meta.QuantidadeSugerida != nil {
		t.Fatalf("esperava sugestão nil após limpar, obtido %v", limpa.Meta.QuantidadeSugerida)
	}
}
