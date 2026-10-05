package autorizacao_test

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func data(t *testing.T, iso string) valueobject.DataLocal {
	t.Helper()
	d, err := valueobject.DataLocalTexto(iso)
	if err != nil {
		t.Fatalf("data inválida: %v", err)
	}
	return d
}

// contarVigentes espelha, em Go puro e com o relógio injetado (nunca
// time.Now), o que postgres.FragmentoDesignacaoVigente calcula em SQL —
// suficiente para os cenários CP-02 a CP-04 em unidade, sem precisar do
// banco (o mecanismo do predicado em si já está preso por
// TestVigencia_RotuloEPredicadoConcordam).
func contarVigentes(designacoes []*designacao.Designacao, hoje valueobject.DataLocal) int {
	n := 0
	for _, d := range designacoes {
		if d.SituacaoEm(hoje) == valueobject.Vigente {
			n++
		}
	}
	return n
}

func novaDesignacaoTeste(t *testing.T, coordenadorID uuid.UUID, inicio valueobject.DataLocal, fim *valueobject.DataLocal) *designacao.Designacao {
	t.Helper()
	d, err := designacao.NovaDesignacao(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), coordenadorID, "1/2026", inicio, fim, false)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return d
}

// TestConjuntoEfetivo_CP02_CoordenadorAcrescentaNaoSubstitui: Ana com
// Professor e Aluno atribuídos, designação entra em vigência — efetivo
// passa a ter os três, sem perder o que já tinha.
func TestConjuntoEfetivo_CP02_CoordenadorAcrescentaNaoSubstitui(t *testing.T) {
	anaID := uuid.Must(uuid.NewV7())
	atribuidos, err := valueobject.NovoConjunto(valueobject.Professor, valueobject.Aluno)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	hoje := data(t, "2026-03-15")
	designacoes := []*designacao.Designacao{novaDesignacaoTeste(t, anaID, data(t, "2026-01-01"), nil)}

	efetivo, err := autorizacao.MontarConjuntoEfetivo(atribuidos, contarVigentes(designacoes, hoje) > 0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	for _, p := range []valueobject.Perfil{valueobject.Professor, valueobject.Aluno, valueobject.CoordenadorCurso} {
		if !efetivo.Possui(p) {
			t.Fatalf("esperava %s no conjunto efetivo, obtido %v", p, efetivo.Ordenado())
		}
	}
}

// TestConjuntoEfetivo_CP03_EncerrarAUnicaDesignacaoRemoveOPerfilDerivado:
// a única designação de Ana encerra — o efetivo volta a ser só o
// atribuído, e o atribuído em si (o valor de entrada) nunca é tocado por
// esta função — é sempre o repositório que decide o que grava.
func TestConjuntoEfetivo_CP03_EncerrarAUnicaDesignacaoRemoveOPerfilDerivado(t *testing.T) {
	anaID := uuid.Must(uuid.NewV7())
	atribuidos, err := valueobject.NovoConjunto(valueobject.Professor, valueobject.Aluno)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	fimEmFevereiro := data(t, "2026-02-28")
	designacoes := []*designacao.Designacao{novaDesignacaoTeste(t, anaID, data(t, "2026-01-01"), &fimEmFevereiro)}
	depoisDoFim := data(t, "2026-03-01")

	efetivo, err := autorizacao.MontarConjuntoEfetivo(atribuidos, contarVigentes(designacoes, depoisDoFim) > 0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !efetivo.Igual(atribuidos) {
		t.Fatalf("esperava o efetivo igual ao atribuído sem a designação vigente, obtido %v", efetivo.Ordenado())
	}
	if efetivo.Possui(valueobject.CoordenadorCurso) {
		t.Fatal("designação encerrada não deveria manter Coordenador de Curso")
	}
}

// TestConjuntoEfetivo_CP04_SairDeUmaEntreVariasNaoTiraOPerfil: Ana
// coordena três cursos; uma das três designações encerra — as outras duas
// ainda vigentes mantêm o perfil.
func TestConjuntoEfetivo_CP04_SairDeUmaEntreVariasNaoTiraOPerfil(t *testing.T) {
	anaID := uuid.Must(uuid.NewV7())
	atribuidos, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	fimDaPrimeira := data(t, "2026-02-28")
	designacoes := []*designacao.Designacao{
		novaDesignacaoTeste(t, anaID, data(t, "2026-01-01"), &fimDaPrimeira), // encerrada em março
		novaDesignacaoTeste(t, anaID, data(t, "2026-01-01"), nil),            // ainda vigente
		novaDesignacaoTeste(t, anaID, data(t, "2026-01-01"), nil),            // ainda vigente
	}
	hoje := data(t, "2026-03-15")

	efetivo, err := autorizacao.MontarConjuntoEfetivo(atribuidos, contarVigentes(designacoes, hoje) > 0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !efetivo.Possui(valueobject.CoordenadorCurso) {
		t.Fatal("com duas das três designações ainda vigentes, deveria continuar Coordenador de Curso")
	}
}

// TestConjuntoEfetivo_QuandoNaoCoordenaDevolveOMesmoAtribuido prova que a
// função é identidade quando coordenaHoje é falso — nenhuma alocação nem
// mutação escondida.
func TestConjuntoEfetivo_QuandoNaoCoordenaDevolveOMesmoAtribuido(t *testing.T) {
	atribuidos, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	efetivo, err := autorizacao.MontarConjuntoEfetivo(atribuidos, false)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !efetivo.Igual(atribuidos) {
		t.Fatalf("esperava o mesmo conjunto atribuído, obtido %v", efetivo.Ordenado())
	}
}

// TestConjuntoEfetivo_CP14_LerSoOAtribuidoFazCoordenadorReceber403 é a
// comprovação negativa do cenário CP-14: sabota a leitura (usa só o
// conjunto atribuído, como se MontarConjuntoEfetivo nunca tivesse sido
// chamado) e prova que um coordenador legítimo — com designação vigente,
// sem o perfil atribuído — recebe 403 numa rota exclusiva de coordenador
// (CursoLerProprio só existe na matriz de CoordenadorCurso).
func TestConjuntoEfetivo_CP14_LerSoOAtribuidoFazCoordenadorReceber403(t *testing.T) {
	anaID := uuid.Must(uuid.NewV7())
	instituicaoID := uuid.Must(uuid.NewV7())
	atribuidos, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Sabotagem: constrói o Ator direto do atribuído, sem passar por
	// MontarConjuntoEfetivo — exatamente o bug que este teste existe para
	// pegar se alguém reintroduzir.
	atorSabotado, err := autorizacao.NovoAtor(anaID, atribuidos, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	if _, err := autorizacao.Autorizar(atorSabotado, autorizacao.CursosDaCarteira, autorizacao.AcaoListar, nil); err == nil {
		t.Fatal("CP-14: a leitura sabotada deveria negar — se passou, o teste não está pegando o defeito")
	}

	// Caminho correto: passa por MontarConjuntoEfetivo antes de construir
	// o Ator — o mesmo coordenador, agora autorizado.
	efetivo, err := autorizacao.MontarConjuntoEfetivo(atribuidos, true)
	if err != nil {
		t.Fatalf("MontarConjuntoEfetivo: %v", err)
	}
	atorCorreto, err := autorizacao.NovoAtor(anaID, efetivo, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	if _, err := autorizacao.Autorizar(atorCorreto, autorizacao.CursosDaCarteira, autorizacao.AcaoListar, nil); err != nil {
		t.Fatalf("coordenador legítimo com o conjunto efetivo deveria ser autorizado, obtido %v", err)
	}
}
