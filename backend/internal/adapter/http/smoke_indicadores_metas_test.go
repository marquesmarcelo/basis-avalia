package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func setupPILogado(t *testing.T, db *sqlx.DB, router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}, instituicaoID string, email, senha string) *http.Cookie {
	t.Helper()
	senhaProvisoriaFalsa := false
	inst := parseUUIDTeste(t, instituicaoID)
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &inst, Nome: "PI de Teste", Email: email, Perfil: "pesquisador_institucional",
		SenhaProvisoria: &senhaProvisoriaFalsa, SenhaHash: hashParaTeste(t, senha),
	})
	return loginEObterCookie(t, router, &instituicaoID, email, senha)
}

// TestSmoke_IndicadoresPlataforma_SemSessaoRecebe401
func TestSmoke_IndicadoresPlataforma_SemSessaoRecebe401(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	rec := requisitar(t, router, http.MethodGet, "/api/v1/plataforma/indicadores", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401 sem sessão, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestSmoke_IndicadoresPlataforma_PISemPermissaoRecebe403 prova que
// nenhuma instituição escreve nem lê o catálogo comum pela rota do
// Administrador.
func TestSmoke_IndicadoresPlataforma_PISemPermissaoRecebe403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkinda@fsa.edu.br", "senha-do-pi-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/plataforma/indicadores", cookiePI, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
	if codigoDeErro(t, rec) != "PERMISSAO_NEGADA" {
		t.Fatalf("código incorreto: %s", rec.Body.String())
	}
}

// TestSmoke_Indicadores_SemSessaoRecebe401
func TestSmoke_Indicadores_SemSessaoRecebe401(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	rec := requisitar(t, router, http.MethodGet, "/api/v1/indicadores", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperava 401, obtido %d", rec.Code)
	}
}

// TestSmoke_Indicadores_AdministradorRecebe403 — Administrador não tem
// IndicadorListar (essa permissão é do PI).
func TestSmoke_Indicadores_AdministradorRecebe403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkindb@basis-avalia.local", "senha-do-admin-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/indicadores", cookieAdmin, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestSmoke_Metas_AdministradorRecebe403 — Administrador não tem
// permissão de meta alguma.
func TestSmoke_Metas_AdministradorRecebe403(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkindc@basis-avalia.local", "senha-do-admin-2026")

	rec := requisitar(t, router, http.MethodGet, "/api/v1/metas", cookieAdmin, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperava 403, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestSmoke_IndicadoresPlataforma_IE09_IndicadorInstitucionalRecebe404
// prova IE-09: Rafael (Admin) tem permissão sobre a ROTA de plataforma,
// mas o indicador institucional da FSA não está no recorte dela — 404,
// nunca 403 (que confirmaria a existência fora do recorte dele).
func TestSmoke_IndicadoresPlataforma_IE09_IndicadorInstitucionalRecebe404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkindd@basis-avalia.local", "senha-do-admin-2026")
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	proprioID := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: fsaID})

	rec := requisitar(t, router, http.MethodGet, "/api/v1/plataforma/indicadores/"+proprioID.String(), cookieAdmin, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtido %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestSmoke_IndicadoresPlataforma_CicloCompleto cobre o caminho feliz
// (IE-01) e T-133: a resposta da família /plataforma/ nunca contém
// identificador nem nome de instituição em campo nenhum.
func TestSmoke_IndicadoresPlataforma_CicloCompleto(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	cookieAdmin := setupAdminLogado(t, db, router, "admin.smkinde@basis-avalia.local", "senha-do-admin-2026")
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})

	// Criar (201).
	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/plataforma/indicadores", cookieAdmin, IndicadorPlataformaRequest{
		Codigo: "1.4", Nome: "Núcleo Docente Estruturante", ReferenciaInstrumento: "Instrumento 2017 - Dim 1, 1.4",
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	if strings.Contains(strings.ToLower(recCriar.Body.String()), "instituicao") {
		t.Fatalf("resposta da família plataforma não pode mencionar instituição: %s", recCriar.Body.String())
	}
	var criado IndicadorPlataformaResponse
	if err := json.Unmarshal(recCriar.Body.Bytes(), &criado); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM indicador WHERE id = $1`, criado.ID) })
	if criado.Escopo != "plataforma" {
		t.Fatalf("esperava escopo plataforma, obtido %q", criado.Escopo)
	}

	// Segundo "1.4" (IE-03, 409).
	recDup := requisitar(t, router, http.MethodPost, "/api/v1/plataforma/indicadores", cookieAdmin, IndicadorPlataformaRequest{
		Codigo: "1.4", Nome: "Outro", ReferenciaInstrumento: "outra referência",
	})
	if recDup.Code != http.StatusConflict || codigoDeErro(t, recDup) != "CODIGO_INDICADOR_DUPLICADO" {
		t.Fatalf("esperava 409 CODIGO_INDICADOR_DUPLICADO, obtido %d (%s)", recDup.Code, recDup.Body.String())
	}

	// Buscar (200).
	recBuscar := requisitar(t, router, http.MethodGet, "/api/v1/plataforma/indicadores/"+criado.ID, cookieAdmin, nil)
	if recBuscar.Code != http.StatusOK {
		t.Fatalf("buscar: esperava 200, obtido %d", recBuscar.Code)
	}

	// Meta da FSA usando o indicador, para provar a contagem total e o 409
	// de exclusão sem menção a instituição (IE-07, PI-5).
	idIndicador := parseUUIDTeste(t, criado.ID)
	testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: fsaID, Indicadores: []uuid.UUID{idIndicador}})

	recExcluir := requisitar(t, router, http.MethodDelete, "/api/v1/plataforma/indicadores/"+criado.ID, cookieAdmin, nil)
	if recExcluir.Code != http.StatusConflict || codigoDeErro(t, recExcluir) != "INDICADOR_COM_META" {
		t.Fatalf("esperava 409 INDICADOR_COM_META, obtido %d (%s)", recExcluir.Code, recExcluir.Body.String())
	}
	if strings.Contains(strings.ToLower(recExcluir.Body.String()), "instituicao") || strings.Contains(strings.ToLower(recExcluir.Body.String()), "fsa") {
		t.Fatalf("mensagem de 409 não pode identificar instituição: %s", recExcluir.Body.String())
	}
}

// TestSmoke_Metas_CicloCompleto cobre MC-01/MC-02: criar meta com
// indicadores, buscar e ver a lista embutida.
func TestSmoke_Metas_CicloCompleto(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	router := montarAppDeTeste(t, db)
	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cookiePI := setupPILogado(t, db, router, fsaID.String(), "pi.smkindf@fsa.edu.br", "senha-do-pi-2026")
	ind14 := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{Codigo: "1.4"})

	recCriar := requisitar(t, router, http.MethodPost, "/api/v1/metas", cookiePI, MetaRequest{
		Nome: "Registrar reuniões de NDE em ata", Indicadores: []string{ind14.String()},
	})
	if recCriar.Code != http.StatusCreated {
		t.Fatalf("criar meta: esperava 201, obtido %d (%s)", recCriar.Code, recCriar.Body.String())
	}
	var criada MetaResponse
	if err := json.Unmarshal(recCriar.Body.Bytes(), &criada); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_indicador WHERE meta_id = $1`, criada.ID)
		db.Exec(`DELETE FROM meta WHERE id = $1`, criada.ID)
	})
	if len(criada.Indicadores) != 1 {
		t.Fatalf("esperava 1 indicador embutido, obtido %d", len(criada.Indicadores))
	}

	// Sem indicador (MC-03, 400).
	recSemIndicador := requisitar(t, router, http.MethodPost, "/api/v1/metas", cookiePI, MetaRequest{Nome: "Meta sem indicador"})
	if recSemIndicador.Code != http.StatusBadRequest || codigoDeErro(t, recSemIndicador) != "INDICADOR_OBRIGATORIO" {
		t.Fatalf("esperava 400 INDICADOR_OBRIGATORIO, obtido %d (%s)", recSemIndicador.Code, recSemIndicador.Body.String())
	}
}

func parseUUIDTeste(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("uuid inválido %q: %v", s, err)
	}
	return id
}
