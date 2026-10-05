package valueobject

import (
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func TestConjuntoInstitucional_E16_VazioOuNilDevolveAluno(t *testing.T) {
	c, err := ConjuntoInstitucional(nil)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !c.Igual(mustConjunto(t, Aluno)) {
		t.Fatalf("esperava {aluno}, obtido %v", c.Ordenado())
	}

	c2, err := ConjuntoInstitucional([]Perfil{})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !c2.Igual(mustConjunto(t, Aluno)) {
		t.Fatalf("esperava {aluno}, obtido %v", c2.Ordenado())
	}
}

func TestConjuntoInstitucional_U13_NaoAcrescentaAlunoAUmConjuntoNaoVazio(t *testing.T) {
	c, err := ConjuntoInstitucional([]Perfil{Professor})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if c.Possui(Aluno) {
		t.Fatal("U-13: aluno não pode ser acrescentado automaticamente a um conjunto não vazio")
	}
	if !c.Igual(mustConjunto(t, Professor)) {
		t.Fatalf("esperava {professor}, obtido %v", c.Ordenado())
	}
}

func TestConjuntoInstitucional_U14_QualquerCombinacaoDosTresEhAceita(t *testing.T) {
	casos := [][]Perfil{
		{Aluno, Professor},
		{Professor, PesquisadorInstitucional},
		{Aluno, Professor, PesquisadorInstitucional},
	}
	for _, entrada := range casos {
		c, err := ConjuntoInstitucional(entrada)
		if err != nil {
			t.Fatalf("combinação %v deveria ser aceita: %v", entrada, err)
		}
		for _, p := range entrada {
			if !c.Possui(p) {
				t.Fatalf("conjunto %v deveria possuir %s", c.Ordenado(), p)
			}
		}
	}
}

// TestConjuntoInstitucional_DC4_CoordenadorCursoNuncaEhAtribuivel prova
// specs/cursos/design.md C-09: coordenador_curso é sempre 403
// PERFIL_NAO_ATRIBUIVEL neste caminho, mesmo misturado com perfil válido
// — o mesmo tratamento que já existia para administrador_sistema (U-10).
func TestConjuntoInstitucional_DC4_CoordenadorCursoNuncaEhAtribuivel(t *testing.T) {
	_, err := ConjuntoInstitucional([]Perfil{CoordenadorCurso, Professor})
	if !errors.Is(err, domain.ErrPerfilNaoAtribuivel) {
		t.Fatalf("esperava ErrPerfilNaoAtribuivel, obtido %v", err)
	}
}

// U-10: no caminho de criação institucional (ConjuntoInstitucional),
// administrador_sistema é sempre um perfil não atribuível pelo PI — 403,
// nunca 400, mesmo quando vem misturado com outro perfil.
func TestConjuntoInstitucional_U10_AdministradorNuncaEhAtribuivel(t *testing.T) {
	_, err := ConjuntoInstitucional([]Perfil{AdministradorSistema, Professor})
	if !errors.Is(err, domain.ErrPerfilNaoAtribuivel) {
		t.Fatalf("esperava ErrPerfilNaoAtribuivel, obtido %v", err)
	}
}

// U-15: a regra estrutural "administrador não acumula" é do construtor
// geral NovoConjunto, não da coerção institucional — é o que protege
// qualquer reconstrução de conjunto no sistema, não só o payload do PI.
func TestNovoConjunto_U15_AdministradorNaoAcumula(t *testing.T) {
	_, err := NovoConjunto(AdministradorSistema, Professor)
	if !errors.Is(err, domain.ErrCombinacaoDePerfisInvalida) {
		t.Fatalf("esperava ErrCombinacaoDePerfisInvalida, obtido %v", err)
	}
}

func TestConjuntoInstitucional_U07_ValorForaDosCincoEhRecusado(t *testing.T) {
	_, err := ConjuntoInstitucional([]Perfil{Professor, Perfil("superusuario")})
	if !errors.Is(err, domain.ErrPerfilInvalido) {
		t.Fatalf("esperava ErrPerfilInvalido, obtido %v", err)
	}
}

func TestNovoConjunto_ConjuntoVazioEhRecusado(t *testing.T) {
	_, err := NovoConjunto()
	if !errors.Is(err, domain.ErrConjuntoDePerfisVazio) {
		t.Fatalf("esperava ErrConjuntoDePerfisVazio, obtido %v", err)
	}
}

func TestConjuntoDeAdministrador_DevolveApenasAdministrador(t *testing.T) {
	c := ConjuntoDeAdministrador()
	if !c.Igual(mustConjunto(t, AdministradorSistema)) {
		t.Fatalf("esperava {administrador_sistema}, obtido %v", c.Ordenado())
	}
	if c.PertenceAInstituicao() {
		t.Fatal("administrador não pertence a instituição")
	}
}

func TestConjuntoDePerfis_A07_PodeEhAUniao(t *testing.T) {
	c := mustConjunto(t, Professor, PesquisadorInstitucional)

	if !c.Pode(UsuarioListar) {
		t.Fatal("conjunto com pesquisador_institucional deveria poder usuario.listar")
	}
	if !c.Pode(UsuarioCriar) {
		t.Fatal("conjunto com pesquisador_institucional deveria poder usuario.criar")
	}
	if c.Pode(AdministradorGerenciar) {
		t.Fatal("nenhum dos dois perfis pode administrador.gerenciar")
	}
}

func TestConjuntoDePerfis_Ordenado_OrdemCanonicaIndependenteDaEntrada(t *testing.T) {
	a := mustConjunto(t, Professor, Aluno)
	b := mustConjunto(t, Aluno, Professor)

	if len(a.Ordenado()) != len(b.Ordenado()) {
		t.Fatalf("tamanhos diferentes: %v vs %v", a.Ordenado(), b.Ordenado())
	}
	for i := range a.Ordenado() {
		if a.Ordenado()[i] != b.Ordenado()[i] {
			t.Fatalf("ordem diverge por entrada diferente: %v vs %v", a.Ordenado(), b.Ordenado())
		}
	}
	if !a.Igual(b) {
		t.Fatal("conjuntos com os mesmos perfis, em ordem de entrada diferente, deveriam ser iguais")
	}
}

func mustConjunto(t *testing.T, p ...Perfil) ConjuntoDePerfis {
	t.Helper()
	c, err := NovoConjunto(p...)
	if err != nil {
		t.Fatalf("mustConjunto(%v): %v", p, err)
	}
	return c
}
