package postgres

import (
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
)

// TestMigracaoNormalizacao_CP15_ComLinhasPresentes é o teste de
// integração de T-181 (specs/cursos/tasks.md): reproduz, com linhas de
// verdade, os três passos da migration 000005_cursos_normalizacao —
// auditar, apagar, estreitar o CHECK — provando o enunciado de CP-15.
//
// O banco de teste já rodou a migration (o CHECK já está estreito), então
// não dá para inserir coordenador_curso pelo caminho normal para montar o
// cenário "antes". A saída: tudo dentro de uma ÚNICA transação — larga o
// CHECK, insere o cenário, roda os três passos exatamente como a
// migration, confere, e sempre desfaz com ROLLBACK no fim (DDL é
// transacional no Postgres). Nenhuma outra suíte rodando em paralelo
// enxerga nada disto: é isolado por MVCC até o commit, que nunca
// acontece.
func TestMigracaoNormalizacao_CP15_ComLinhasPresentes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CP15"})
	comDesignacao := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Ana CP15", Perfil: "professor"})
	semDesignacao := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "João CP15", Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso CP15"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: comDesignacao, DataInicio: "2026-01-01",
	})

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	t.Cleanup(func() { tx.Rollback() }) // sempre desfaz — nunca commita esta transação

	// Cenário "antes": larga o CHECK estreito, atribui coordenador_curso
	// aos dois (um com designação vigente, outro sem), religa o CHECK
	// largo só para o INSERT ser aceito.
	if _, err := tx.Exec(`ALTER TABLE usuario_perfil DROP CONSTRAINT ck_usuario_perfil_valor,
		ADD CONSTRAINT ck_usuario_perfil_valor CHECK (perfil IN (
			'administrador_sistema','pesquisador_institucional','coordenador_curso','professor','aluno'))`); err != nil {
		t.Fatalf("largar o CHECK: %v", err)
	}
	if _, err := tx.Exec(`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,'coordenador_curso',$2)`, comDesignacao, fsaID); err != nil {
		t.Fatalf("inserir coordenador_curso (com designação): %v", err)
	}
	if _, err := tx.Exec(`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,'coordenador_curso',$2)`, semDesignacao, fsaID); err != nil {
		t.Fatalf("inserir coordenador_curso (sem designação): %v", err)
	}

	// Os três passos da migration, na mesma ordem: auditar, apagar,
	// estreitar.
	if _, err := tx.Exec(`
		INSERT INTO auditoria (id, acao, resultado, ator_id, instituicao_id, recurso_tipo, recurso_id, detalhes, executado_em)
		SELECT gen_random_uuid(), 'normalizar_perfil_coordenador', 'sucesso', NULL, u.instituicao_id, 'Usuario', u.id,
		       jsonb_build_object('motivo', 'perfil de coordenador passou a ser derivado de designação'), now()
		  FROM usuario u JOIN usuario_perfil up ON up.usuario_id = u.id
		 WHERE up.perfil = 'coordenador_curso' AND u.instituicao_id = $1`, fsaID); err != nil {
		t.Fatalf("auditar: %v", err)
	}
	if _, err := tx.Exec(`DELETE FROM usuario_perfil WHERE perfil = 'coordenador_curso' AND instituicao_id = $1`, fsaID); err != nil {
		t.Fatalf("apagar: %v", err)
	}
	if _, err := tx.Exec(`ALTER TABLE usuario_perfil DROP CONSTRAINT ck_usuario_perfil_valor,
		ADD CONSTRAINT ck_usuario_perfil_valor CHECK (perfil IN (
			'administrador_sistema','pesquisador_institucional','professor','aluno'))`); err != nil {
		t.Fatalf("estreitar o CHECK: %v", err)
	}

	// CP-15, cláusula 1: o perfil sai do conjunto atribuído de TODOS.
	var restam int
	if err := tx.QueryRow(`SELECT count(*) FROM usuario_perfil WHERE perfil = 'coordenador_curso' AND instituicao_id = $1`, fsaID).Scan(&restam); err != nil {
		t.Fatalf("contar restantes: %v", err)
	}
	if restam != 0 {
		t.Fatalf("esperava 0 linhas com coordenador_curso atribuído, restaram %d", restam)
	}

	// CP-15, cláusula 2: quem tem designação vigente continua coordenador
	// — pela derivação, não pelo atribuído. Roda, DENTRO DESTA MESMA
	// TRANSAÇÃO (para enxergar o estado pós-DELETE), o mesmo predicado
	// que CarregarContextoDeSessao usa — FragmentoDesignacaoVigente, a
	// única fonte da regra de vigência (design.md §4.3, C-03).
	var coordenaHoje bool
	consultaCoordenaHoje := `SELECT EXISTS (
		SELECT 1 FROM designacao d WHERE d.coordenador_id = $1 AND d.excluido_em IS NULL AND ` +
		FragmentoDesignacaoVigente("d", 2) + `)`
	if err := tx.QueryRow(consultaCoordenaHoje, comDesignacao, hojeTeste.String()).Scan(&coordenaHoje); err != nil {
		t.Fatalf("consultar derivação: %v", err)
	}
	if !coordenaHoje {
		t.Fatal("CP-15: quem tem designação vigente deveria continuar coordenador pela derivação")
	}

	// CP-15, cláusula 3: quem NÃO tem designação mantém os demais perfis
	// (aqui, só 'professor' — que nunca foi tocado pelo DELETE, que é por
	// (usuario_id, perfil), nunca a linha inteira do usuário).
	var perfilRestante string
	if err := tx.QueryRow(`SELECT perfil FROM usuario_perfil WHERE usuario_id = $1`, semDesignacao).Scan(&perfilRestante); err != nil {
		t.Fatalf("ler perfil restante: %v", err)
	}
	if perfilRestante != "professor" {
		t.Fatalf("esperava que 'professor' permanecesse intacto, obtido %q", perfilRestante)
	}

	// CP-15, cláusula 4: a auditoria tem UMA linha por usuário afetado —
	// nunca um "rebaixamento" por pessoa (não é isso que este evento
	// descreve).
	var totalAuditoria int
	if err := tx.QueryRow(`SELECT count(*) FROM auditoria WHERE acao = 'normalizar_perfil_coordenador' AND instituicao_id = $1`, fsaID).Scan(&totalAuditoria); err != nil {
		t.Fatalf("contar auditoria: %v", err)
	}
	if totalAuditoria != 2 {
		t.Fatalf("esperava 2 linhas de auditoria (uma por usuário afetado), obtido %d", totalAuditoria)
	}
}
