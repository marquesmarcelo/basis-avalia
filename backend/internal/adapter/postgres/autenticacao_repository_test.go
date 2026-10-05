package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// hojeTeste — os testes deste arquivo não exercitam vigência de
// designação, só existência/exclusão da linha de usuário; qualquer data
// serve.
var hojeTeste = valueobject.DataLocalDe(time.Now(), time.UTC)

func TestAutenticacaoRepository_CarregarContextoDeSessao_DistingueAusenteDeExcluido(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "AUTA"})
	excluidoID := testhelpers.CriarUsuarioExcluido(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	ctxExcluido, err := repo.CarregarContextoDeSessao(context.Background(), excluidoID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ctxExcluido == nil {
		t.Fatal("esperava encontrar a linha do usuário excluído (para distinguir de ausente)")
	}
	if ctxExcluido.ExcluidoEm == nil {
		t.Fatal("excluido_em deveria estar preenchido")
	}

	idInexistente := excluidoID
	idInexistente[0] ^= 0xFF
	ctxAusente, err := repo.CarregarContextoDeSessao(context.Background(), idInexistente, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ctxAusente != nil {
		t.Fatal("esperava nil para usuário inexistente")
	}
}

func TestAutenticacaoRepository_InvalidarSessoes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "AUTB"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	conjunto, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("conjunto de teste: %v", err)
	}
	ator, err := autorizacao.NovoAtor(usuarioID, conjunto, &fsaID)
	if err != nil {
		t.Fatalf("ator de teste: %v", err)
	}

	instante := time.Now().Add(time.Hour)
	if err := repo.InvalidarSessoesProprias(context.Background(), ator.Proprio(), instante); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	ctxSessao, err := repo.CarregarContextoDeSessao(context.Background(), usuarioID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Postgres TIMESTAMPTZ tem precisão de microssegundo — comparar truncado
	// evita falso negativo pelo nanossegundo residual do time.Now() em Go.
	if !ctxSessao.SessoesValidasAPartirDe.Truncate(time.Microsecond).Equal(instante.Truncate(time.Microsecond)) {
		t.Fatalf("sessoes_validas_a_partir_de não foi atualizado: %v != %v", ctxSessao.SessoesValidasAPartirDe, instante)
	}
}

// TestAutenticacaoRepository_ConjuntoVazioDevolveErro prova o guarda de
// design.md §5.6: uma linha de usuario sem nenhum vínculo em
// usuario_perfil (estado impossível em uso normal, só alcançável aqui por
// inserir sem passar pelo caminho do domínio) faz CarregarContextoDeSessao
// devolver erro — nunca um ContextoDeSessao com conjunto vazio, que
// deixaria o middleware decidir 401 por engano.
func TestAutenticacaoRepository_ConjuntoVazioDevolveErro(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoAutenticacaoRepository(db)

	id := uuid.Must(uuid.NewV7())
	if _, err := db.Exec(
		`INSERT INTO usuario (id, nome, email, senha_hash) VALUES ($1,$2,$3,$4)`,
		id, "Sem Perfil", "sem.perfil.v9@teste.local", testhelpers.HashDeTeste,
	); err != nil {
		t.Fatalf("inserir usuário sem vínculo: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM usuario WHERE id = $1`, id) })

	if _, err := repo.CarregarContextoDeSessao(context.Background(), id, hojeTeste); err == nil {
		t.Fatal("esperava erro para conjunto de perfis vazio")
	}
}

// TestAutenticacaoRepository_CP06_DesignacaoVigenteDeCursoInativoMantemOPerfil
// é a armadilha nomeada em specs/cursos/design.md §4.2: o EXISTS de
// designação vigente NÃO junta curso, então inativar o curso não derruba
// coordena_hoje/cursos_coordenados de quem tem designação vigente nele.
func TestAutenticacaoRepository_CP06_DesignacaoVigenteDeCursoInativoMantemOPerfil(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoAutenticacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CP06"})
	diegoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Diego Nunes", Perfil: "professor"})
	nutricaoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Nutrição", Situacao: "inativo"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: nutricaoID, InstituicaoID: fsaID, CoordenadorID: diegoID,
		Portaria: "51/2026", DataInicio: "2026-02-01",
	})

	hoje := valueobject.DataLocalDe(time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC), time.UTC)
	ctxSessao, err := repo.CarregarContextoDeSessao(context.Background(), diegoID, hoje)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !ctxSessao.CoordenaHoje {
		t.Fatal("designação vigente em curso inativo deveria manter coordena_hoje verdadeiro (CP-06)")
	}
	if ctxSessao.CursosCoordenados != 1 {
		t.Fatalf("esperava cursos_coordenados = 1, obtido %d", ctxSessao.CursosCoordenados)
	}
}

// TestAutenticacaoRepository_V5_IndiceDeCoordenadorExiste prova o que dá
// para provar honestamente neste volume (mesmo achado de
// TestInstituicaoPublicaQuery_V3_IndiceParcialExiste): com a massa de um
// teste de integração (poucas linhas em designacao), o planner do
// Postgres escolhe corretamente Seq Scan por ser mais barato — verificar
// o plano de execução real fazia este teste falhar de forma dependente da
// ordem/volume de outros testes na mesma suíte (intermitente, não um
// defeito de produção). O que fica provado aqui é que a migration criou o
// índice que design.md §4.4 V-5 exige; o planner de fato usá-lo sob
// volume representativo fica em testes-pendentes.md.
func TestAutenticacaoRepository_V5_IndiceDeCoordenadorExiste(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)

	var definicao string
	err := db.Get(&definicao, `
		SELECT indexdef FROM pg_indexes
		 WHERE tablename = 'designacao' AND indexname = 'idx_designacao_coordenador'`)
	if err != nil {
		t.Fatalf("índice idx_designacao_coordenador não encontrado no catálogo: %v", err)
	}
	if !strings.Contains(definicao, "coordenador_id") {
		t.Fatalf("esperava índice por coordenador_id, obtido: %s", definicao)
	}
}
