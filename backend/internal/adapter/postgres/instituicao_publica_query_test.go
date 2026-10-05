package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
)

func TestInstituicaoPublicaQuery_L11_SoAtivaComPIAtivo(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	query := NovoInstituicaoPublicaQuery(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Nome: "Zeta Faculdade Combo", Sigla: "ZFC1"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Nome: "Alfa Instituto Combo", Sigla: "AIC1"})
	ceaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Nome: "Centro Combo Inativo", Sigla: "CCI1", Situacao: "inativa"})
	fnaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Nome: "Faculdade Combo Sem PI", Sigla: "FCS1"})

	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &ivvID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &ceaID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fnaID, Perfil: "professor"})

	itens, err := query.ListarParaCombo(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	var siglasRelevantes []string
	posicoes := map[string]int{}
	for i, item := range itens {
		if item.Sigla == "ZFC1" || item.Sigla == "AIC1" || item.Sigla == "CCI1" || item.Sigla == "FCS1" {
			siglasRelevantes = append(siglasRelevantes, item.Sigla)
			posicoes[item.Sigla] = i
		}
	}
	if len(siglasRelevantes) != 2 {
		t.Fatalf("esperava exatamente 2 instituições relevantes (ativas com PI), obtido %v", siglasRelevantes)
	}
	if posicoes["AIC1"] >= posicoes["ZFC1"] {
		t.Fatalf("ordem incorreta: Alfa Instituto Combo deveria vir antes de Zeta Faculdade Combo, posicoes=%v", posicoes)
	}
	for _, item := range itens {
		if item.Sigla == "CCI1" {
			t.Fatal("instituição inativa não deveria aparecer no combo")
		}
		if item.Sigla == "FCS1" {
			t.Fatal("instituição sem PI ativo não deveria aparecer no combo")
		}
	}
}

// TestInstituicaoPublicaQuery_V3_IndiceParcialExiste prova o que dá para
// provar honestamente neste volume: que a migration criou o índice que
// design.md §5.9 especifica para a sonda do combo. Antes verificava o
// plano de execução real via EXPLAIN, mas com a massa de um teste de
// integração (poucas linhas) o planner do Postgres escolhe corretamente
// Seq Scan por ser mais barato — o teste falhava sempre no ambiente de
// teste e passaria em produção, ensinando a ignorar a suíte. A verificação
// de que o planner de fato usa o índice sob volume representativo fica
// registrada em testes-pendentes.md.
func TestInstituicaoPublicaQuery_V3_IndiceParcialExiste(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)

	var definicao string
	err := db.Get(&definicao, `
		SELECT indexdef FROM pg_indexes
		 WHERE tablename = 'usuario_perfil' AND indexname = 'idx_usuario_perfil_pi'`)
	if err != nil {
		t.Fatalf("índice idx_usuario_perfil_pi não encontrado no catálogo: %v", err)
	}
	if !strings.Contains(definicao, "pesquisador_institucional") {
		t.Fatalf("esperava índice parcial filtrado por perfil = pesquisador_institucional, obtido: %s", definicao)
	}
}
