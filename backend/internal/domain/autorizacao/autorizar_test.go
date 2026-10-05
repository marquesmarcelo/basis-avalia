package autorizacao

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Fora do pacote, Escopo não pode ser construído diretamente — campos não
// exportados e nenhum construtor exportado além do que Autorizar devolve.
// O bloco abaixo NÃO compila caso descomentado (é o teste do item d de
// T-003), por isso fica em comentário:
//
//   package outro_pacote
//   import "github.com/basis-avalia/backend/internal/domain/autorizacao"
//   func exemplo() {
//       _ = autorizacao.Escopo{} // erro: campos não exportados
//   }

func novoConjuntoTeste(t *testing.T, perfis ...valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(perfis...)
	if err != nil {
		t.Fatalf("NovoConjunto(%v): %v", perfis, err)
	}
	return c
}

func novoAtorTeste(t *testing.T, conjunto valueobject.ConjuntoDePerfis, instituicaoID *uuid.UUID) Ator {
	t.Helper()
	ator, err := NovoAtor(uuid.Must(uuid.NewV7()), conjunto, instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator
}

// TestAutorizar_MatrizDeAlcancesAcoesEConjuntos cobre A-01 a A-07: os quatro
// alcances × as sete ações × cinco conjuntos representativos, incluindo o
// caso que o modelo de perfil único não tinha — um conjunto com dois
// perfis (design.md T-085).
func TestAutorizar_MatrizDeAlcancesAcoesEConjuntos(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	outraInstituicaoID := uuid.Must(uuid.NewV7())

	conjuntos := []valueobject.ConjuntoDePerfis{
		novoConjuntoTeste(t, valueobject.Aluno),
		novoConjuntoTeste(t, valueobject.Professor),
		novoConjuntoTeste(t, valueobject.PesquisadorInstitucional),
		novoConjuntoTeste(t, valueobject.Professor, valueobject.PesquisadorInstitucional),
		novoConjuntoTeste(t, valueobject.AdministradorSistema),
	}
	acoes := []Acao{AcaoListar, AcaoBuscar, AcaoCriar, AcaoEditar, AcaoExcluir, AcaoRedefinirSenha, AcaoInativar}
	alcances := []Alcance{UsuariosDaPropriaInstituicao, PesquisadoresDeUmaInstituicao, AdministradoresDaPlataforma, InstituicoesDaPlataforma}

	for _, conjunto := range conjuntos {
		var atorInstituicaoID *uuid.UUID
		if conjunto.PertenceAInstituicao() {
			atorInstituicaoID = &instituicaoID
		}
		ator := novoAtorTeste(t, conjunto, atorInstituicaoID)

		for _, alcance := range alcances {
			for _, acao := range acoes {
				permissaoExigidaEsperada, erroPermissao := permissaoExigida(alcance, acao)

				escopo, err := Autorizar(ator, alcance, acao, &outraInstituicaoID)

				if erroPermissao != nil {
					if err == nil {
						t.Errorf("alcance=%v acao=%v conjunto=%v: esperava erro de combinação inválida", alcance, acao, conjunto.Ordenado())
					}
					continue
				}

				podeAcao := conjunto.Pode(permissaoExigidaEsperada)
				if !podeAcao {
					if err != domain.ErrPermissaoNegada {
						t.Errorf("alcance=%v acao=%v conjunto=%v: esperava ErrPermissaoNegada, obtido %v", alcance, acao, conjunto.Ordenado(), err)
					}
					continue
				}

				if err != nil {
					t.Errorf("alcance=%v acao=%v conjunto=%v: erro inesperado %v", alcance, acao, conjunto.Ordenado(), err)
					continue
				}
				if !escopo.Valido() {
					t.Errorf("alcance=%v acao=%v conjunto=%v: escopo deveria ser válido", alcance, acao, conjunto.Ordenado())
				}
				switch alcance {
				case UsuariosDaPropriaInstituicao:
					if escopo.Plataforma() {
						t.Errorf("escopo de usuários da própria instituição não deveria ser de plataforma")
					}
					if escopo.InstituicaoID() == nil || *escopo.InstituicaoID() != instituicaoID {
						t.Errorf("escopo deveria conter a instituição do ator")
					}
					if escopo.ExigePerfil() != nil {
						t.Errorf("escopo de usuários da própria instituição não deveria exigir perfil do alvo")
					}
				case PesquisadoresDeUmaInstituicao:
					if escopo.InstituicaoID() == nil || *escopo.InstituicaoID() != outraInstituicaoID {
						t.Errorf("escopo deveria conter a instituição do caminho, não a do ator")
					}
					if escopo.ExigePerfil() == nil || *escopo.ExigePerfil() != valueobject.PesquisadorInstitucional {
						t.Errorf("escopo deveria exigir pesquisador_institucional do alvo")
					}
				case AdministradoresDaPlataforma:
					if !escopo.Plataforma() {
						t.Errorf("escopo de administradores deveria ser de plataforma")
					}
					if escopo.ExigePerfil() == nil || *escopo.ExigePerfil() != valueobject.AdministradorSistema {
						t.Errorf("escopo deveria exigir administrador_sistema do alvo")
					}
				case InstituicoesDaPlataforma:
					if !escopo.Plataforma() {
						t.Errorf("escopo de instituições deveria ser de plataforma")
					}
					if escopo.ExigePerfil() != nil {
						t.Errorf("escopo de instituições não deveria exigir perfil do alvo")
					}
				}
			}
		}
	}
}

func TestNovoAtor_RecusaConjuntoVazio(t *testing.T) {
	var vazio valueobject.ConjuntoDePerfis
	if _, err := NovoAtor(uuid.Must(uuid.NewV7()), vazio, nil); err == nil {
		t.Fatal("esperava erro: conjunto vazio")
	}
}

func TestNovoAtor_RecusaVinculoNuloComConjuntoInstitucional(t *testing.T) {
	conjunto := novoConjuntoTeste(t, valueobject.PesquisadorInstitucional)
	if _, err := NovoAtor(uuid.Must(uuid.NewV7()), conjunto, nil); err == nil {
		t.Fatal("esperava erro: conjunto institucional exige vínculo")
	}
}

func TestNovoAtor_RecusaAdministradorComVinculoPreenchido(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	conjunto := novoConjuntoTeste(t, valueobject.AdministradorSistema)
	if _, err := NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID); err == nil {
		t.Fatal("esperava erro: administrador não tem vínculo institucional")
	}
}

func TestEscopo_ValorZeroInvalido(t *testing.T) {
	var vazio Escopo
	if vazio.Valido() {
		t.Fatal("Escopo{} deveria ser inválido")
	}
}
