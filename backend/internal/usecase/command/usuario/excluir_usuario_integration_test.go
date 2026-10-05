package usuario

import (
	"context"
	"sync"
	"testing"

	"github.com/basis-avalia/backend/internal/adapter/auditoria"
	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func atorDoUsuario(t *testing.T, id uuid.UUID, perfis []valueobject.Perfil, instituicaoID *uuid.UUID) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(perfis...)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(id, conjunto, instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator
}

func contarDetentoresDeTeste(t *testing.T, db *sqlx.DB, instituicaoID uuid.UUID, perfil string) int {
	t.Helper()
	var total int
	if err := db.Get(&total, `
		SELECT count(*) FROM usuario_perfil up JOIN usuario u ON u.id = up.usuario_id
		 WHERE up.instituicao_id = $1 AND up.perfil = $2 AND u.excluido_em IS NULL`,
		instituicaoID, perfil); err != nil {
		t.Fatalf("contar detentores: %v", err)
	}
	return total
}

// TestAtualizarUsuario_E17_ConcorrenciaDoUltimoPI prova a trava de
// design.md §5.4/§5.8: Maria e Beatriz são as duas únicas detentoras
// ativas do perfil de pesquisador_institucional na mesma instituição;
// duas requisições concorrentes tentam, cada uma, RETIRAR o perfil de uma
// delas (via atualizar_usuario, não exclusão — é o gatilho mais amplo que
// T-090 introduz). Exatamente uma tem sucesso; a instituição nunca fica
// sem PI ativo (E-17, absorve a antiga T-053).
func TestAtualizarUsuario_E17_ConcorrenciaDoUltimoPI(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CCPI"})
	maria := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "pesquisador_institucional"})
	beatriz := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfis: []string{"professor", "pesquisador_institucional"}})

	repo := postgres.NovoUsuarioRepository(db)
	uow := postgres.NovaUnidadeDeTrabalho(db)
	sys, _ := auditoria.NovoSyslog("", "tcp", "teste")
	auditRepo := postgres.NovoAuditoriaRepository(db)
	audit := auditoria.NovoComposto(auditRepo, sys)
	t.Cleanup(func() { db.Exec(`DELETE FROM auditoria WHERE instituicao_id = $1`, instituicaoID) })

	uc := NovoAtualizarUsuarioUseCase(repo, audit, uow)

	var wg sync.WaitGroup
	erros := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		email, _ := valueobject.NovoEmail("maria.retirada@fsa.edu.br")
		_, erros[0] = uc.Executar(context.Background(), AtualizarUsuarioInput{
			Ator: atorDoUsuario(t, beatriz, []valueobject.Perfil{valueobject.Professor, valueobject.PesquisadorInstitucional}, &instituicaoID),
			Alcance: autorizacao.UsuariosDaPropriaInstituicao, UsuarioID: maria,
			Nome: "Maria Souza", Email: email, Perfis: []valueobject.Perfil{valueobject.Aluno}, Versao: 1,
		})
	}()
	go func() {
		defer wg.Done()
		email, _ := valueobject.NovoEmail("beatriz.retirada@fsa.edu.br")
		_, erros[1] = uc.Executar(context.Background(), AtualizarUsuarioInput{
			Ator: atorDoUsuario(t, maria, []valueobject.Perfil{valueobject.PesquisadorInstitucional}, &instituicaoID),
			Alcance: autorizacao.UsuariosDaPropriaInstituicao, UsuarioID: beatriz,
			Nome: "Beatriz Andrade", Email: email, Perfis: []valueobject.Perfil{valueobject.Professor}, Versao: 1,
		})
	}()
	wg.Wait()

	sucessos, conflitos := 0, 0
	for _, err := range erros {
		switch {
		case err == nil:
			sucessos++
		case err == domain.ErrUltimoPesquisadorInstitucional:
			conflitos++
		default:
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if sucessos != 1 || conflitos != 1 {
		t.Fatalf("esperava exatamente 1 sucesso e 1 conflito, obtido %d sucesso(s) e %d conflito(s) — erros: %v", sucessos, conflitos, erros)
	}
	if total := contarDetentoresDeTeste(t, db, instituicaoID, "pesquisador_institucional"); total != 1 {
		t.Fatalf("a instituição deveria ficar com exatamente 1 detentor de PI, obtido %d", total)
	}
}

// TestAtualizarUsuario_E10_ComDuasDetentorasRetirarDeUmaEhAceito prova o
// caminho feliz da mesma invariante (E-10): com duas detentoras, retirar
// o perfil de uma é aceito, e os demais perfis dela permanecem intactos.
func TestAtualizarUsuario_E10_ComDuasDetentorasRetirarDeUmaEhAceito(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "E10I"})
	maria := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "pesquisador_institucional"})
	beatriz := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfis: []string{"professor", "pesquisador_institucional"}})

	repo := postgres.NovoUsuarioRepository(db)
	uow := postgres.NovaUnidadeDeTrabalho(db)
	sys, _ := auditoria.NovoSyslog("", "tcp", "teste")
	auditRepo := postgres.NovoAuditoriaRepository(db)
	audit := auditoria.NovoComposto(auditRepo, sys)
	t.Cleanup(func() { db.Exec(`DELETE FROM auditoria WHERE instituicao_id = $1`, instituicaoID) })

	uc := NovoAtualizarUsuarioUseCase(repo, audit, uow)
	email, _ := valueobject.NovoEmail("beatriz.rebaixada@fsa.edu.br")
	resultado, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorDoUsuario(t, maria, []valueobject.Perfil{valueobject.PesquisadorInstitucional}, &instituicaoID),
		Alcance: autorizacao.UsuariosDaPropriaInstituicao, UsuarioID: beatriz,
		Nome: "Beatriz Andrade", Email: email, Perfis: []valueobject.Perfil{valueobject.Professor}, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Perfis.Possui(valueobject.PesquisadorInstitucional) {
		t.Fatal("pesquisador_institucional deveria ter sido retirado")
	}
	if !resultado.Perfis.Possui(valueobject.Professor) {
		t.Fatal("professor deveria permanecer intacto (E-10, E-15)")
	}
}

// TestExcluirUsuario_E09_UltimaDetentoraNaoPodeSerExcluida prova o
// caminho não concorrente da invariante: excluir a última PI ativa é
// recusado, e uma detentora em OUTRA instituição não satisfaz a
// invariante local.
func TestExcluirUsuario_E09_UltimaDetentoraNaoPodeSerExcluida(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "E09A"})
	outraInstituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "E09B"})
	unicaPI := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &outraInstituicaoID, Perfil: "pesquisador_institucional"})

	repo := postgres.NovoUsuarioRepository(db)
	uow := postgres.NovaUnidadeDeTrabalho(db)
	sys, _ := auditoria.NovoSyslog("", "tcp", "teste")
	auditRepo := postgres.NovoAuditoriaRepository(db)
	audit := auditoria.NovoComposto(auditRepo, sys)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE instituicao_id IN ($1,$2)`, instituicaoID, outraInstituicaoID)
	})

	uc := NovoExcluirUsuarioUseCase(repo, audit, uow)

	// O ator só precisa ter, na sessão, a permissão de excluir (qualquer
	// PI serve para autorização — Autorizar nunca consulta o banco pelo
	// conjunto do próprio ator, só o que já veio da sessão). O que decide
	// a invariante é a contagem real no banco, não quem chama.
	ator := atorDoUsuario(t, uuid.Must(uuid.NewV7()), []valueobject.Perfil{valueobject.PesquisadorInstitucional}, &instituicaoID)
	err := uc.Executar(context.Background(), ExcluirUsuarioInput{
		Ator: ator, Alcance: autorizacao.UsuariosDaPropriaInstituicao, UsuarioID: unicaPI,
	})
	if err != domain.ErrUltimoPesquisadorInstitucional {
		t.Fatalf("esperava ErrUltimoPesquisadorInstitucional (a PI da outra instituição não deveria contar), obtido %v", err)
	}
}

// TestExcluirUsuario_AS06_ConcorrenciaDoUltimoAdministrador prova o
// pg_advisory_xact_lock de design.md §5.4 no alcance de plataforma —
// administradores não têm caminho de "retirar o perfil" via atualização
// (o conjunto deles é sempre fixo), então a invariante só é exercitada
// pela exclusão.
func TestExcluirUsuario_AS06_ConcorrenciaDoUltimoAdministrador(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	admin1 := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{Perfil: "administrador_sistema"})
	admin2 := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{Perfil: "administrador_sistema"})
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auditoria WHERE ator_id IN ($1,$2)`, admin1, admin2)
	})

	repo := postgres.NovoUsuarioRepository(db)
	uow := postgres.NovaUnidadeDeTrabalho(db)
	sys, _ := auditoria.NovoSyslog("", "tcp", "teste")
	auditRepo := postgres.NovoAuditoriaRepository(db)
	audit := auditoria.NovoComposto(auditRepo, sys)

	uc := NovoExcluirUsuarioUseCase(repo, audit, uow)

	var wg sync.WaitGroup
	erros := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		erros[0] = uc.Executar(context.Background(), ExcluirUsuarioInput{
			Ator: atorDoUsuario(t, admin1, []valueobject.Perfil{valueobject.AdministradorSistema}, nil),
			Alcance: autorizacao.AdministradoresDaPlataforma, UsuarioID: admin2,
		})
	}()
	go func() {
		defer wg.Done()
		erros[1] = uc.Executar(context.Background(), ExcluirUsuarioInput{
			Ator: atorDoUsuario(t, admin2, []valueobject.Perfil{valueobject.AdministradorSistema}, nil),
			Alcance: autorizacao.AdministradoresDaPlataforma, UsuarioID: admin1,
		})
	}()
	wg.Wait()

	sucessos, conflitos := 0, 0
	for _, err := range erros {
		switch {
		case err == nil:
			sucessos++
		case err == domain.ErrUltimoAdministradorSistema:
			conflitos++
		default:
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if sucessos != 1 || conflitos != 1 {
		t.Fatalf("esperava exatamente 1 sucesso e 1 conflito, obtido %d sucesso(s) e %d conflito(s) — erros: %v", sucessos, conflitos, erros)
	}
}
