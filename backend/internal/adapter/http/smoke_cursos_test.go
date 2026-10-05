package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// TestSmoke_Cursos_CicloCompletoDeStatusCodes prova T-179/T-180: os
// endpoints de curso respondem com o contrato exato de design.md §6 —
// criar, buscar, atualizar, conflito de versão, inativar, excluir.
func TestSmoke_Cursos_CicloCompletoDeStatusCodes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkcurso@fsa.edu.br", "senha-do-pi-2026")

	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/cursos", cookiePI, CursoRequest{
		Nome: "Curso Smoke", Grau: "bacharelado", Modalidade: "presencial",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar curso: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var criado CursoResponse
	if err := json.Unmarshal(recCriar.Body.Bytes(), &criado); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	if criado.Coordenador != nil {
		t.Fatal("curso recém-criado deveria nascer sem coordenador (vago)")
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM curso WHERE id = $1`, criado.ID) })

	recListar := requisitar(t, router, http.MethodGet, "/api/v1/cursos?situacao=todas", cookiePI, nil)
	if recListar.Code != http.StatusOK {
		t.Fatalf("listar: esperava 200, obtido %d", recListar.Code)
	}

	recBuscar := requisitar(t, router, http.MethodGet, "/api/v1/cursos/"+criado.ID, cookiePI, nil)
	if recBuscar.Code != http.StatusOK {
		t.Fatalf("buscar: esperava 200, obtido %d", recBuscar.Code)
	}

	recAtualizar := requisitar(t, router, http.MethodPut, "/api/v1/cursos/"+criado.ID, cookiePI, CursoAtualizarRequest{
		Nome: "Curso Smoke Renomeado", Grau: "bacharelado", Modalidade: "presencial", Versao: criado.Versao,
	})
	if recAtualizar.Code != http.StatusOK {
		t.Fatalf("atualizar: esperava 200, obtido %d (%s)", recAtualizar.Code, recAtualizar.Body.String())
	}

	recConflito := requisitar(t, router, http.MethodPut, "/api/v1/cursos/"+criado.ID, cookiePI, CursoAtualizarRequest{
		Nome: "Outra vez", Grau: "bacharelado", Modalidade: "presencial", Versao: criado.Versao,
	})
	if recConflito.Code != http.StatusConflict {
		t.Fatalf("conflito de versão: esperava 409, obtido %d", recConflito.Code)
	}

	recInativar := requisitar(t, router, http.MethodPatch, "/api/v1/cursos/"+criado.ID+"/situacao", cookiePI, AlterarSituacaoRequest{
		Situacao: "inativo", Versao: criado.Versao + 1,
	})
	if recInativar.Code != http.StatusOK {
		t.Fatalf("inativar: esperava 200, obtido %d (%s)", recInativar.Code, recInativar.Body.String())
	}

	recExcluir := requisitar(t, router, http.MethodDelete, "/api/v1/cursos/"+criado.ID, cookiePI, nil)
	if recExcluir.Code != http.StatusNoContent {
		t.Fatalf("excluir: esperava 204, obtido %d", recExcluir.Code)
	}
}

// TestSmoke_Cursos_CriarJaInativo prova design.md §6 (a extensão aditiva
// de POST /cursos): ux.md pede o campo Situação no formulário de criação,
// para o caso raro de cadastrar já como inativo.
func TestSmoke_Cursos_CriarJaInativo(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkinativo@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodPost, "/api/v1/cursos", cookiePI, CursoRequest{
		Nome: "Curso Já Inativo", Grau: "bacharelado", Modalidade: "presencial", Situacao: "inativo",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar curso inativo: esperava 201, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	var criado CursoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &criado); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM curso WHERE id = $1`, criado.ID) })
	if criado.Situacao != "inativo" {
		t.Fatalf("esperava nascer inativo, obtido %q", criado.Situacao)
	}
}

// TestSmoke_Designacoes_CicloCompletoComErros prova a cadeia de
// designação: candidato inválido (400), criação com sucesso,
// sobreposição (409), 409 DESIGNACAO_COM_EFEITO ao tentar trocar o
// coordenador de uma vigente, e 409 CURSO_COM_VINCULO.
func TestSmoke_Designacoes_CicloCompletoComErros(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkdesig@fsa.edu.br", "senha-do-pi-2026")

	recCurso := requisitar(t, router, http.MethodPost, "/api/v1/cursos", cookiePI, CursoRequest{
		Nome: "Curso Designacao Smoke", Grau: "bacharelado", Modalidade: "presencial",
	})
	if recCurso.Code != http.StatusCreated {
		t.Fatalf("criar curso: esperava 201, obtido %d", recCurso.Code)
	}
	var curso CursoResponse
	json.Unmarshal(recCurso.Body.Bytes(), &curso)
	t.Cleanup(func() { db.Exec(`DELETE FROM curso WHERE id = $1`, curso.ID) })

	// Candidato inválido (formato de UUID errado) — 400.
	recInvalido := requisitar(t, router, http.MethodPost, "/api/v1/cursos/"+curso.ID+"/designacoes", cookiePI, DesignacaoRequest{
		CoordenadorID: "nao-e-um-uuid", Portaria: "1/2026", DataInicio: "2026-01-01",
	})
	if recInvalido.Code != http.StatusBadRequest {
		t.Fatalf("candidato inválido: esperava 400, obtido %d (%s)", recInvalido.Code, recInvalido.Body.String())
	}

	senhaProvisoriaFalsa := false
	professorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Professor Designacao Smoke", Perfil: "professor",
		SenhaProvisoria: &senhaProvisoriaFalsa,
	})

	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/cursos/"+curso.ID+"/designacoes", cookiePI, DesignacaoRequest{
		CoordenadorID: professorID.String(), Portaria: "1/2026", DataInicio: "2026-01-01",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar designação: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var designada DesignacaoResponse
	json.Unmarshal(recCriar.Body.Bytes(), &designada)
	t.Cleanup(func() { db.Exec(`DELETE FROM designacao WHERE id = $1`, designada.ID) })
	// Achado de revisão: designacaoParaResposta esquecia de preencher
	// Situacao — a resposta chegava com "" em vez de "vigente" em TODO
	// endpoint (criar, buscar, listar, atualizar), sem que nenhum teste de
	// status code pegasse. Só a asserção no valor exato do campo capturava.
	if designada.Situacao != "vigente" {
		t.Fatalf("esperava situacao=vigente para início 2026-01-01 sem fim, obtido %q", designada.Situacao)
	}

	// Sobreposição: mesma data de início, mesmo curso — 409.
	recSobreposta := requisitar(t, router, http.MethodPost, "/api/v1/cursos/"+curso.ID+"/designacoes", cookiePI, DesignacaoRequest{
		CoordenadorID: professorID.String(), Portaria: "2/2026", DataInicio: "2026-01-01",
	})
	if recSobreposta.Code != http.StatusConflict {
		t.Fatalf("designação sobreposta: esperava 409, obtido %d (%s)", recSobreposta.Code, recSobreposta.Body.String())
	}

	// A designação criada é vigente hoje (relógio real, início em
	// 2026-01-01) — tentar trocar o coordenador é 409.
	outroProfessorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Perfil: "professor", SenhaProvisoria: &senhaProvisoriaFalsa,
	})
	recComEfeito := requisitar(t, router, http.MethodPut, "/api/v1/designacoes/"+designada.ID, cookiePI, DesignacaoAtualizarRequest{
		CoordenadorID: outroProfessorID.String(), Portaria: designada.Portaria, DataInicio: designada.DataInicio, Versao: designada.Versao,
	})
	if recComEfeito.Code != http.StatusConflict {
		t.Fatalf("trocar coordenador de designação vigente: esperava 409, obtido %d (%s)", recComEfeito.Code, recComEfeito.Body.String())
	}

	recCandidatos := requisitar(t, router, http.MethodGet, "/api/v1/designacoes/candidatos?busca=Designacao+Smoke", cookiePI, nil)
	if recCandidatos.Code != http.StatusOK {
		t.Fatalf("candidatos: esperava 200, obtido %d", recCandidatos.Code)
	}

	// Mesmo achado de revisão, mas na listagem (T-190 do E2E foi quem
	// pegou este caminho especificamente): GET .../designacoes também
	// precisa devolver situacao preenchida, não só POST/PUT/GET-por-id.
	recListar := requisitar(t, router, http.MethodGet, "/api/v1/cursos/"+curso.ID+"/designacoes", cookiePI, nil)
	if recListar.Code != http.StatusOK {
		t.Fatalf("listar designações: esperava 200, obtido %d", recListar.Code)
	}
	var listaDesignacoes struct {
		Data []DesignacaoResponse `json:"data"`
	}
	json.Unmarshal(recListar.Body.Bytes(), &listaDesignacoes)
	if len(listaDesignacoes.Data) != 1 || listaDesignacoes.Data[0].Situacao != "vigente" {
		t.Fatalf("esperava 1 designação com situacao=vigente na listagem, obtido %+v", listaDesignacoes.Data)
	}

	// Curso com designação vigente não pode ser excluído.
	recExcluirCurso := requisitar(t, router, http.MethodDelete, "/api/v1/cursos/"+curso.ID, cookiePI, nil)
	if recExcluirCurso.Code != http.StatusConflict {
		t.Fatalf("excluir curso com designação: esperava 409, obtido %d", recExcluirCurso.Code)
	}
}

// TestSmoke_Cursos_T115_PlanoDoPeriodoAbertoNaListagem prova T-115: a
// listagem de cursos compõe, no use case de consulta (nunca no SELECT de
// CursoRepository — design.md §5.4), a situação do plano no período
// aberto da instituição, via PeriodoRepository + PlanoRepository.
func TestSmoke_Cursos_T115_PlanoDoPeriodoAbertoNaListagem(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkt115@fsa.edu.br", "senha-do-pi-2026")

	recCurso := requisitar(t, router, http.MethodPost, "/api/v1/cursos", cookiePI, CursoRequest{
		Nome: "Curso T115 Com Plano", Grau: "bacharelado", Modalidade: "presencial",
	})
	if recCurso.Code != http.StatusCreated {
		t.Fatalf("criar curso: esperava 201, obtido %d", recCurso.Code)
	}
	var comPlano CursoResponse
	json.Unmarshal(recCurso.Body.Bytes(), &comPlano)
	t.Cleanup(func() { db.Exec(`DELETE FROM curso WHERE id = $1`, comPlano.ID) })

	recCursoSemPlano := requisitar(t, router, http.MethodPost, "/api/v1/cursos", cookiePI, CursoRequest{
		Nome: "Curso T115 Sem Plano", Grau: "bacharelado", Modalidade: "presencial",
	})
	var semPlano CursoResponse
	json.Unmarshal(recCursoSemPlano.Body.Bytes(), &semPlano)
	t.Cleanup(func() { db.Exec(`DELETE FROM curso WHERE id = $1`, semPlano.ID) })

	cursoComPlanoID, err := uuid.Parse(comPlano.ID)
	if err != nil {
		t.Fatalf("uuid do curso: %v", err)
	}
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{
		InstituicaoID: fsaID, DataInicio: "2026-01-01", DataFim: "2026-12-31",
	})
	testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{
		InstituicaoID: fsaID, CursoID: cursoComPlanoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente",
	})

	recListar := requisitar(t, router, http.MethodGet, "/api/v1/cursos?situacao=todas&page_size=50", cookiePI, nil)
	if recListar.Code != http.StatusOK {
		t.Fatalf("listar cursos: esperava 200, obtido %d (%s)", recListar.Code, recListar.Body.String())
	}
	var lista struct {
		Data []CursoResponse `json:"data"`
	}
	json.Unmarshal(recListar.Body.Bytes(), &lista)

	var linhaComPlano, linhaSemPlano *CursoResponse
	for i := range lista.Data {
		switch lista.Data[i].ID {
		case comPlano.ID:
			linhaComPlano = &lista.Data[i]
		case semPlano.ID:
			linhaSemPlano = &lista.Data[i]
		}
	}
	if linhaComPlano == nil || linhaSemPlano == nil {
		t.Fatalf("esperava os dois cursos na listagem, obtido %d itens", len(lista.Data))
	}
	if linhaComPlano.PlanoDoPeriodo == nil || *linhaComPlano.PlanoDoPeriodo != "Vigente (0 metas)" {
		t.Fatalf("esperava \"Vigente (0 metas)\" para o curso com plano, obtido %v", linhaComPlano.PlanoDoPeriodo)
	}
	if linhaSemPlano.PlanoDoPeriodo != nil {
		t.Fatalf("esperava plano_do_periodo nulo para o curso sem plano, obtido %v", *linhaSemPlano.PlanoDoPeriodo)
	}

	recBuscar := requisitar(t, router, http.MethodGet, "/api/v1/cursos/"+comPlano.ID, cookiePI, nil)
	var buscado CursoResponse
	json.Unmarshal(recBuscar.Body.Bytes(), &buscado)
	if buscado.PlanoDoPeriodo != nil {
		t.Fatalf("GET de curso único não deveria trazer plano_do_periodo, obtido %v", *buscado.PlanoDoPeriodo)
	}
}
