package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
)

// TestSmoke_CP01_AtribuidoNuncaContemCoordenadorEfetivoContem prova
// specs/cursos/spec.md CP-01 fim a fim: o conjunto ATRIBUÍDO (devolvido
// pelo CRUD de usuários) nunca contém coordenador_curso; o conjunto
// EFETIVO (devolvido por GET /auth/eu) contém, junto com
// perfis_derivados e cursos_coordenados — para quem tem designação
// vigente de verdade, criada direto no banco (não pelo perfil, que não
// aceita mais o valor — DC-4).
func TestSmoke_CP01_AtribuidoNuncaContemCoordenadorEfetivoContem(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CP01"})
	senhaProvisoriaFalsa := false
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Ana Lima", Email: "ana.cp01@fsa.edu.br", Perfil: "professor",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, "senha-ana-2026"),
	})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Engenharia de Software CP01"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: anaID,
		Portaria: "47/2026", DataInicio: "2026-01-01",
	})

	fsaIDStr := fsaID.String()
	cookieAna := loginEObterCookie(t, router, &fsaIDStr, "ana.cp01@fsa.edu.br", "senha-ana-2026")

	// Lado atribuído: quem administra usuários (PI) nunca vê
	// coordenador_curso no conjunto de Ana.
	senhaProvisoriaFalsaMaria := false
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Maria Souza", Email: "maria.cp01@fsa.edu.br", Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsaMaria, SenhaHash: hashParaTeste(t, "senha-maria-2026"),
	})
	cookieMaria := loginEObterCookie(t, router, &fsaIDStr, "maria.cp01@fsa.edu.br", "senha-maria-2026")
	recUsuario := requisitar(t, router, http.MethodGet, "/api/v1/usuarios/"+anaID.String(), cookieMaria, nil)
	if recUsuario.Code != http.StatusOK {
		t.Fatalf("buscar usuário: esperava 200, obtido %d (%s)", recUsuario.Code, recUsuario.Body.String())
	}
	var atribuido UsuarioResponse
	if err := json.Unmarshal(recUsuario.Body.Bytes(), &atribuido); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	for _, p := range atribuido.Perfis {
		if p == "coordenador_curso" {
			t.Fatal("CP-01: o conjunto ATRIBUÍDO não deveria conter coordenador_curso")
		}
	}

	// Lado efetivo: GET /auth/eu, autenticada como a própria Ana.
	recEu := requisitar(t, router, http.MethodGet, "/api/v1/auth/eu", cookieAna, nil)
	if recEu.Code != http.StatusOK {
		t.Fatalf("GET /auth/eu: esperava 200, obtido %d (%s)", recEu.Code, recEu.Body.String())
	}
	var eu ContextoDeSessaoResponse
	if err := json.Unmarshal(recEu.Body.Bytes(), &eu); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	temCoordenadorNoEfetivo := false
	for _, p := range eu.Perfis {
		if p == "coordenador_curso" {
			temCoordenadorNoEfetivo = true
		}
	}
	if !temCoordenadorNoEfetivo {
		t.Fatalf("CP-01: o conjunto EFETIVO deveria conter coordenador_curso, obtido %v", eu.Perfis)
	}
	if len(eu.PerfisDerivados) != 1 || eu.PerfisDerivados[0] != "coordenador_curso" {
		t.Fatalf("esperava perfis_derivados = [\"coordenador_curso\"], obtido %v", eu.PerfisDerivados)
	}
	if eu.CursosCoordenados != 1 {
		t.Fatalf("esperava cursos_coordenados = 1, obtido %d", eu.CursosCoordenados)
	}

	// T-164: tentar atribuir coordenador_curso pelo CRUD responde 403
	// PERFIL_NAO_ATRIBUIVEL, nunca 400 — mesmo tratamento de
	// administrador_sistema (U-10), estendido pelo DC-4.
	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/usuarios", cookieMaria, UsuarioRequest{
		Nome: "Tentativa CP01", Email: "tentativa.cp01@fsa.edu.br", Perfis: []string{"coordenador_curso"}, Senha: "primeiro-acesso-2026",
	})
	if recCriar.Code != http.StatusForbidden {
		t.Fatalf("criar com coordenador_curso: esperava 403, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	if codigoDeErro(t, recCriar) != "PERFIL_NAO_ATRIBUIVEL" {
		t.Fatalf("código incorreto: %s", recCriar.Body.String())
	}
}
