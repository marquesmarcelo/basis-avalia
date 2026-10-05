package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func escopoDesignacoesDaInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, err := valueobject.DataLocalTexto("2026-03-15")
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	ator = ator.ComDataDeReferencia(hoje)
	escopo, err := autorizacao.Autorizar(ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

// TestDesignacaoRepository_OutrasDesignacoesVigentes (C-08): campo único
// que substitui os dois que o ux.md propôs — conta as OUTRAS vigentes,
// excluindo a própria linha.
func TestDesignacaoRepository_OutrasDesignacoesVigentes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "DR01"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	curso1 := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso 1 DR01"})
	curso2 := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso 2 DR01"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: curso1, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2026-01-01"})
	desig2ID := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: curso2, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2026-01-01"})

	escopo := escopoDesignacoesDaInstituicao(t, fsaID)
	item, err := repo.BuscarPorID(context.Background(), escopo, desig2ID)
	if err != nil {
		t.Fatalf("BuscarPorID: %v", err)
	}
	if item.OutrasDesignacoesVigentes != 1 {
		t.Fatalf("esperava 1 outra designação vigente, obtido %d", item.OutrasDesignacoesVigentes)
	}
}

// TestDesignacaoRepository_FiltroSituacao (as três fronteiras, no
// repositório): futura, vigente, encerrada devolvem só a linha certa.
func TestDesignacaoRepository_FiltroSituacao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "DR02"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	paulinhoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	diegoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Situacao DR02"})

	fimPassado := "2026-01-31"
	encerradaID := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: diegoID, DataInicio: "2025-06-01", DataFim: &fimPassado,
	})
	fimDaVigente := "2026-07-31"
	vigenteID := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2026-02-01", DataFim: &fimDaVigente,
	})
	futuraID := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: paulinhoID, DataInicio: "2026-08-01",
	})

	escopo := escopoDesignacoesDaInstituicao(t, fsaID)
	casos := []struct {
		situacao   string
		esperadoID uuid.UUID
	}{
		{"encerrada", encerradaID},
		{"vigente", vigenteID},
		{"futura", futuraID},
	}
	for _, c := range casos {
		t.Run(c.situacao, func(t *testing.T) {
			resultado, err := repo.ListarDoCurso(context.Background(), escopo, cursoID, port.FiltroListarDesignacoes{
				Situacao: c.situacao, Page: 1, PageSize: 20, Sort: "data_inicio", Order: "desc",
			})
			if err != nil {
				t.Fatalf("ListarDoCurso: %v", err)
			}
			if len(resultado.Itens) != 1 || resultado.Itens[0].Designacao.ID != c.esperadoID {
				t.Fatalf("esperava só %s, obtido %d itens", c.situacao, len(resultado.Itens))
			}
		})
	}
}

// TestDesignacaoRepository_FiltroCoordenadorID: o filtro "Coordenador" da
// tela de designações de um curso (specs/cursos/ux.md, tela 3) devolve só
// as designações da pessoa filtrada, sem tocar nas de outra.
func TestDesignacaoRepository_FiltroCoordenadorID(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "DR03"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	paulinhoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Filtro Coord DR03"})

	fimDaAna := "2026-01-31"
	anaDesigID := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2025-06-01", DataFim: &fimDaAna,
	})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: paulinhoID, DataInicio: "2026-02-01",
	})

	escopo := escopoDesignacoesDaInstituicao(t, fsaID)
	resultado, err := repo.ListarDoCurso(context.Background(), escopo, cursoID, port.FiltroListarDesignacoes{
		Situacao: "todas", CoordenadorID: &anaID, Page: 1, PageSize: 20, Sort: "data_inicio", Order: "desc",
	})
	if err != nil {
		t.Fatalf("ListarDoCurso: %v", err)
	}
	if len(resultado.Itens) != 1 || resultado.Itens[0].Designacao.ID != anaDesigID {
		t.Fatalf("esperava só a designação da Ana, obtido %d itens", len(resultado.Itens))
	}
}

// TestDesignacaoRepository_CP07_Candidatos: traz quem tem professor ou PI
// (inclusive os dois), não traz quem só tem aluno.
func TestDesignacaoRepository_CP07_Candidatos(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "DR03"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Professor Candidato DR03", Perfil: "professor"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Professor Aluno Candidato DR03", Perfis: []string{"professor", "aluno"},
	})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "So Aluno DR03", Perfil: "aluno"})

	escopo := escopoDesignacoesDaInstituicao(t, fsaID)
	candidatos, err := repo.ListarCandidatos(context.Background(), escopo, "DR03")
	if err != nil {
		t.Fatalf("ListarCandidatos: %v", err)
	}
	nomes := map[string]bool{}
	for _, c := range candidatos {
		nomes[c.Nome] = true
	}
	if !nomes["Professor Candidato DR03"] {
		t.Fatal("esperava o professor entre os candidatos")
	}
	if !nomes["Professor Aluno Candidato DR03"] {
		t.Fatal("esperava quem tem professor E aluno entre os candidatos")
	}
	if nomes["So Aluno DR03"] {
		t.Fatal("não deveria trazer quem só tem o perfil de aluno")
	}
}

// TestDesignacaoRepository_DG03_ConcorrenciaSoUmaVence é o teste de
// integração que a tarefa T-176 pede: duas inserções concorrentes com
// intervalos que se tocam no mesmo curso — o EXCLUDE do banco garante que
// só uma vence, mesmo sob corrida real (goroutines, não sequencial).
func TestDesignacaoRepository_DG03_ConcorrenciaSoUmaVence(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "DR04"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	paulinhoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Concorrencia DR04"})
	t.Cleanup(func() { db.Exec(`DELETE FROM designacao WHERE curso_id = $1`, cursoID) })

	escopo := escopoDesignacoesDaInstituicao(t, fsaID)
	inicio, err := valueobject.DataLocalTexto("2026-01-01")
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	fim, err := valueobject.DataLocalTexto("2026-12-31")
	if err != nil {
		t.Fatalf("data: %v", err)
	}

	novaDesignacao := func(coordenadorID uuid.UUID) *designacao.Designacao {
		d, err := designacao.NovaDesignacao(cursoID, fsaID, coordenadorID, "1/2026", inicio, &fim, false)
		if err != nil {
			t.Fatalf("NovaDesignacao: %v", err)
		}
		return d
	}

	var wg sync.WaitGroup
	erros := make([]error, 2)
	coordenadores := []uuid.UUID{anaID, paulinhoID}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			erros[i] = repo.Inserir(context.Background(), escopo, novaDesignacao(coordenadores[i]))
		}(i)
	}
	wg.Wait()

	sucessos, conflitos := 0, 0
	for _, err := range erros {
		switch err {
		case nil:
			sucessos++
		case domain.ErrDesignacaoSobreposta:
			conflitos++
		default:
			t.Fatalf("erro inesperado na corrida: %v", err)
		}
	}
	if sucessos != 1 || conflitos != 1 {
		t.Fatalf("esperava exatamente 1 sucesso e 1 conflito, obtido %d sucesso(s) e %d conflito(s)", sucessos, conflitos)
	}
}
