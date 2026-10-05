package plano

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func atorCoordenadorDeTeste(t *testing.T, usuarioID, instituicaoID uuid.UUID) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.CoordenadorCurso)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(usuarioID, conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	return ator.ComDataDeReferencia(hoje)
}

// TestBuscarPlano_SI01_CoordenadorVeRascunhoDoProprioCursoSomenteLeitura
// prova SI-01/VI-01 (P-10, design.md §5.4): o coordenador vê o plano do
// curso dele mesmo em rascunho — 200, nunca 404 — com os itens embutidos
// (regressão do bug de Escopo.SemCarteira: a junção com o catálogo de
// indicador não pode recusar por causa da carteira).
func TestBuscarPlano_SI01_CoordenadorVeRascunhoDoProprioCursoSomenteLeitura(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenador := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenador, DataInicio: "2026-01-01"})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "rascunho"})
	indicador := testhelpers.CriarIndicadorPlataforma(t, db, testhelpers.OpcoesIndicadorPlataforma{})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID, Indicadores: []uuid.UUID{indicador}})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 4})

	repo := postgres.NovoPlanoRepository(db)
	uc := NovoBuscarPlanoUseCase(repo)

	ator := atorCoordenadorDeTeste(t, coordenador, instituicaoID)
	detalhe, err := uc.Executar(context.Background(), ator, autorizacao.PlanosDaCarteira, planoID)
	if err != nil {
		t.Fatalf("coordenador deveria ver o plano em rascunho do próprio curso: %v", err)
	}
	if detalhe.Situacao != "rascunho" {
		t.Fatalf("esperava situacao rascunho, obtido %q", detalhe.Situacao)
	}
	if len(detalhe.Itens) != 1 || detalhe.Itens[0].MetaNome == "" {
		t.Fatalf("esperava 1 item com o nome da meta preenchido, obtido %+v", detalhe.Itens)
	}
	if len(detalhe.Itens[0].Indicadores) != 1 {
		t.Fatalf("esperava o indicador da meta embutido, obtido %+v", detalhe.Itens[0].Indicadores)
	}
}

// TestBuscarPlano_VI01_CoordenadorForaDaCarteiraRecebe404 prova que o
// coordenador NÃO vê plano de curso que não coordena.
func TestBuscarPlano_VI01_CoordenadorForaDaCarteiraRecebe404(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenador := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "professor"})
	cursoOutro := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoOutro, PeriodoID: periodoID})

	repo := postgres.NovoPlanoRepository(db)
	uc := NovoBuscarPlanoUseCase(repo)

	ator := atorCoordenadorDeTeste(t, coordenador, instituicaoID)
	_, err := uc.Executar(context.Background(), ator, autorizacao.PlanosDaCarteira, planoID)
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado (404), obtido %v", err)
	}
}
