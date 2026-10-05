package autorizacao

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// alcancesNovosDeInstituicao e alcancesNovosDeCarteira são as duas listas
// fechadas de fundacao-metas.md §6.2 — usadas pelos testes abaixo E pelo
// guarda "nenhum alcance sem linha em permissaoExigida".
var alcancesNovosDeInstituicao = []Alcance{
	CatalogoDeIndicadores, MetasDaInstituicao, CursosDaInstituicao, DesignacoesDaInstituicao,
	PeriodosDaInstituicao, PlanosDaInstituicao, EntregasDaInstituicao, DesempenhoDaInstituicao,
}

var alcancesNovosDeCarteira = []Alcance{
	CursosDaCarteira, PlanosDaCarteira, EntregasDaCarteira, DesempenhoDaCarteira,
}

// TestAlcancesNovos_TemLinhaEmPermissaoExigida é teste de MECANISMO: todo
// alcance novo de fundacao-metas.md §6.2, mais IndicadoresDaPlataforma,
// precisa ter uma linha em permissaoExigida — um alcance sem linha cairia
// no default (Nenhuma, ErrEscopoInvalido) e pareceria "sempre nega",
// silenciosamente, sem ninguém perceber até a rota correspondente ser
// escrita.
func TestAlcancesNovos_TemLinhaEmPermissaoExigida(t *testing.T) {
	todos := append([]Alcance{IndicadoresDaPlataforma}, alcancesNovosDeInstituicao...)
	todos = append(todos, alcancesNovosDeCarteira...)

	for _, alcance := range todos {
		if _, err := permissaoExigida(alcance, AcaoListar); err != nil {
			t.Errorf("alcance %v não tem linha em permissaoExigida (AcaoListar): %v", alcance, err)
		}
	}
}

func atorPITeste(t *testing.T, instituicaoID uuid.UUID) Ator {
	t.Helper()
	conjunto := novoConjuntoTeste(t, valueobject.PesquisadorInstitucional)
	return novoAtorTeste(t, conjunto, &instituicaoID)
}

func atorCoordenadorTeste(t *testing.T, instituicaoID uuid.UUID) Ator {
	t.Helper()
	conjunto := novoConjuntoTeste(t, valueobject.CoordenadorCurso)
	return novoAtorTeste(t, conjunto, &instituicaoID)
}

// TestAutorizar_AlcancesDeInstituicao_EscopoDevolvidoCampoACampo prova
// fundacao-metas.md §6.2: os oito alcances "Da Instituição" devolvem o
// escopo da instituição do ator — nunca plataforma, nunca restrito a
// carteira.
func TestAutorizar_AlcancesDeInstituicao_EscopoDevolvidoCampoACampo(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator := atorPITeste(t, instituicaoID)

	for _, alcance := range alcancesNovosDeInstituicao {
		escopo, err := Autorizar(ator, alcance, AcaoListar, nil)
		if err != nil {
			t.Errorf("alcance %v: erro inesperado %v", alcance, err)
			continue
		}
		if !escopo.Valido() {
			t.Errorf("alcance %v: escopo deveria ser válido", alcance)
		}
		if escopo.Plataforma() {
			t.Errorf("alcance %v: não deveria ser escopo de plataforma", alcance)
		}
		if escopo.InstituicaoID() == nil || *escopo.InstituicaoID() != instituicaoID {
			t.Errorf("alcance %v: deveria conter a instituição do ator", alcance)
		}
		if escopo.RestritoACarteiraDe() != nil {
			t.Errorf("alcance %v: não deveria restringir por carteira", alcance)
		}
	}
}

// TestAutorizar_IndicadoresDaPlataforma_EscopoDePlataformaSemPerfil prova
// que o único alcance de indicador do Administrador do Sistema devolve
// plataforma, sem exigir perfil do alvo (fundacao-metas.md §6.2).
func TestAutorizar_IndicadoresDaPlataforma_EscopoDePlataformaSemPerfil(t *testing.T) {
	ator := novoAtorTeste(t, valueobject.ConjuntoDeAdministrador(), nil)

	escopo, err := Autorizar(ator, IndicadoresDaPlataforma, AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	if !escopo.Plataforma() {
		t.Error("deveria ser escopo de plataforma")
	}
	if escopo.ExigePerfil() != nil {
		t.Error("não deveria exigir perfil do alvo")
	}
}

// TestAutorizar_AlcancesDeCarteira_RestritoACarteiraDeSempreLigado prova
// fundacao-metas.md §6.2: os quatro alcances de carteira ligam a
// restrição SEMPRE, nunca condicionalmente — restritoACarteiraDe é o
// próprio id do ator (coordenador), e a instituição também é a dele.
func TestAutorizar_AlcancesDeCarteira_RestritoACarteiraDeSempreLigado(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator := atorCoordenadorTeste(t, instituicaoID)

	for _, alcance := range alcancesNovosDeCarteira {
		escopo, err := Autorizar(ator, alcance, AcaoListar, nil)
		if err != nil {
			t.Errorf("alcance %v: erro inesperado %v", alcance, err)
			continue
		}
		if escopo.RestritoACarteiraDe() == nil {
			t.Errorf("alcance %v: deveria restringir por carteira", alcance)
			continue
		}
		if *escopo.RestritoACarteiraDe() != ator.UsuarioID() {
			t.Errorf("alcance %v: restritoACarteiraDe deveria ser o próprio ator", alcance)
		}
		if escopo.InstituicaoID() == nil || *escopo.InstituicaoID() != instituicaoID {
			t.Errorf("alcance %v: deveria conter a instituição do ator", alcance)
		}
	}
}
