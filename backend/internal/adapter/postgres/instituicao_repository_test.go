package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func escopoDePlataforma(t *testing.T) autorizacao.Escopo {
	t.Helper()
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

func TestInstituicaoRepository_I08_ContagemDePesquisadoresAtivosPorLinha(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IFSA"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IIVV"})
	ceaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "ICEA", Situacao: "inativa"})
	testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IFNA"})

	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &ivvID, Perfil: "pesquisador_institucional"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &ceaID, Perfil: "pesquisador_institucional"})

	escopo := escopoDePlataforma(t)
	contagem := map[string]int{}
	itens, err := repo.Listar(context.Background(), escopo, port.FiltroListarInstituicoes{Page: 1, PageSize: 100, Situacao: "todas", Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	for _, item := range itens.Itens {
		contagem[item.Instituicao.Sigla.String()] = item.PesquisadoresAtivos
	}
	if contagem["IFSA"] != 2 {
		t.Errorf("FSA: esperado 2, obtido %d", contagem["IFSA"])
	}
	if contagem["IIVV"] != 1 {
		t.Errorf("IVV: esperado 1, obtido %d", contagem["IIVV"])
	}
	if contagem["ICEA"] != 1 {
		t.Errorf("CEA: esperado 1, obtido %d", contagem["ICEA"])
	}
	if contagem["IFNA"] != 0 {
		t.Errorf("FNA: esperado 0, obtido %d", contagem["IFNA"])
	}
}

func TestInstituicaoRepository_I11_InativaApareceNaListagem(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)

	ceaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "ICE2", Situacao: "inativa"})

	escopo := escopoDePlataforma(t)
	resultado, err := repo.Listar(context.Background(), escopo, port.FiltroListarInstituicoes{Page: 1, PageSize: 100, Situacao: "todas", Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	achou := false
	for _, item := range resultado.Itens {
		if item.Instituicao.ID == ceaID {
			achou = true
		}
	}
	if !achou {
		t.Fatal("instituição inativa deveria aparecer na listagem (3.19)")
	}
}

func TestInstituicaoRepository_FiltroSituacaoInativaTrazSoAsInativas(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)

	testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IAT1", Situacao: "ativa"})
	ceaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IIN1", Situacao: "inativa"})

	escopo := escopoDePlataforma(t)
	resultado, err := repo.Listar(context.Background(), escopo, port.FiltroListarInstituicoes{Page: 1, PageSize: 100, Situacao: "inativa", Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	for _, item := range resultado.Itens {
		if item.Instituicao.Situacao != valueobject.Inativa {
			t.Fatalf("filtro inativa trouxe registro ativo: %s", item.Instituicao.Sigla.String())
		}
	}
	achouCEA := false
	for _, item := range resultado.Itens {
		if item.Instituicao.ID == ceaID {
			achouCEA = true
		}
	}
	if !achouCEA {
		t.Fatal("esperava encontrar a instituição inativa criada no teste")
	}
}

func TestInstituicaoRepository_EscopoInstitucionalDevolveErro(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoInstituicaoRepository(db)

	instituicaoID := uuid.Must(uuid.NewV7())
	escopoInstitucional := escopoDaInstituicao(t, instituicaoID)

	_, err := repo.Listar(context.Background(), escopoInstitucional, port.FiltroListarInstituicoes{Page: 1, PageSize: 20})
	if err != domain.ErrEscopoInvalido {
		t.Fatalf("esperava ErrEscopoInvalido, obtido %v", err)
	}
}
