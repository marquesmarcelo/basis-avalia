package postgres

import (
	"os"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// TestIndicador_IV06_BancoRecusaEscopoInstituicaoIncoerente prova IV-06: o
// par (escopo, instituição) incoerente é impedido pelo BANCO, não só pela
// aplicação — ck_indicador_coerencia.
func TestIndicador_IV06_BancoRecusaEscopoInstituicaoIncoerente(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})

	// plataforma COM instituição — deve ser recusado.
	id1 := uuid.Must(uuid.NewV7())
	_, err := db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, referencia_instrumento)
		 VALUES ($1,'plataforma',$2,'X','Nome','referência')`, id1, fsaID)
	if err == nil {
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id1)
		t.Fatal("esperava o banco recusar escopo plataforma com instituição preenchida")
	}

	// instituicao SEM instituição — deve ser recusado.
	id2 := uuid.Must(uuid.NewV7())
	_, err = db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome)
		 VALUES ($1,'instituicao',NULL,'X','Nome')`, id2)
	if err == nil {
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id2)
		t.Fatal("esperava o banco recusar escopo instituicao sem instituição")
	}
}

// TestIndicador_IE02_IE08_BancoRecusaReferenciaIncoerente prova IE-02
// (referência obrigatória em plataforma) e a metade de IN-02 (proibida em
// instituição) — ck_indicador_referencia, pelo BANCO.
func TestIndicador_IE02_IN02_BancoRecusaReferenciaIncoerente(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})

	// plataforma SEM referência — deve ser recusado (IE-02).
	id1 := uuid.Must(uuid.NewV7())
	_, err := db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, referencia_instrumento)
		 VALUES ($1,'plataforma',NULL,'X','Nome',NULL)`, id1)
	if err == nil {
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id1)
		t.Fatal("esperava o banco recusar indicador de plataforma sem referência")
	}

	// instituicao COM referência — deve ser recusado (IN-02).
	id2 := uuid.Must(uuid.NewV7())
	_, err = db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, referencia_instrumento)
		 VALUES ($1,'instituicao',$2,'X','Nome','não deveria existir')`, id2, fsaID)
	if err == nil {
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id2)
		t.Fatal("esperava o banco recusar indicador institucional com referência do instrumento")
	}
}

// TestIndicador_IE08_EscopoNuncaEstaNoSET é o guarda de mecanismo de IE-08:
// nenhum UPDATE de indicador, nas duas famílias de repositório, lista a
// coluna "escopo" — é isso que torna o escopo imutável, não uma checagem
// em tempo de execução.
func TestIndicador_IE08_EscopoNuncaEstaNoSET(t *testing.T) {
	for _, arquivo := range []string{"indicador_repository.go", "indicador_plataforma_repository.go"} {
		fonte, err := os.ReadFile(arquivo)
		if err != nil {
			t.Fatalf("lendo %s: %v", arquivo, err)
		}
		for _, linha := range strings.Split(string(fonte), "\n") {
			alvo := strings.SplitN(linha, "//", 2)[0]
			if strings.Contains(alvo, "UPDATE indicador SET") && strings.Contains(alvo, "escopo=") {
				t.Fatalf("%s: UPDATE lista a coluna escopo — o escopo deixaria de ser imutável: %q", arquivo, linha)
			}
		}
	}
}
