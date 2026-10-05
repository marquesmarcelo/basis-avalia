package meta

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/google/uuid"
)

// TestNovaMeta_RecusaSemIndicador prova MC-03: não existe meta sem
// indicador, em nenhuma rota.
func TestNovaMeta_RecusaSemIndicador(t *testing.T) {
	_, err := NovaMeta(uuid.Must(uuid.NewV7()), "Registrar reuniões de NDE em ata", "", nil)
	if err != domain.ErrIndicadorObrigatorio {
		t.Fatalf("esperava ErrIndicadorObrigatorio, obtido %v", err)
	}
}

// TestNovaMeta_RecusaAcimaDoLimite prova MC-04, primeira metade.
func TestNovaMeta_RecusaAcimaDoLimite(t *testing.T) {
	seis := make([]uuid.UUID, 6)
	for i := range seis {
		seis[i] = uuid.Must(uuid.NewV7())
	}
	_, err := NovaMeta(uuid.Must(uuid.NewV7()), "Meta com seis indicadores", "", seis)
	if err != domain.ErrIndicadoresAcimaDoLimite {
		t.Fatalf("esperava ErrIndicadoresAcimaDoLimite, obtido %v", err)
	}
}

// TestNovaMeta_RecusaIndicadorDuplicado prova MC-04, segunda metade.
func TestNovaMeta_RecusaIndicadorDuplicado(t *testing.T) {
	repetido := uuid.Must(uuid.NewV7())
	_, err := NovaMeta(uuid.Must(uuid.NewV7()), "Meta com indicador repetido", "", []uuid.UUID{repetido, repetido})
	if err != domain.ErrIndicadorDuplicadoNaMeta {
		t.Fatalf("esperava ErrIndicadorDuplicadoNaMeta, obtido %v", err)
	}
}

// TestNovaMeta_ComDoisIndicadores prova MC-02 — o caso do dono (ata de NDE
// apontando para 1.4 e 1.5).
func TestNovaMeta_ComDoisIndicadores(t *testing.T) {
	ind1, ind2 := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	m, err := NovaMeta(uuid.Must(uuid.NewV7()), "Registrar reuniões de NDE em ata", "", []uuid.UUID{ind1, ind2})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(m.Indicadores) != 2 {
		t.Fatalf("esperava 2 indicadores, obtido %d", len(m.Indicadores))
	}
}

// TestMeta_DefinirIndicadores_DevolveListaAnteriorCompleta prova MC-13: a
// auditoria registra a lista anterior e a nova, nunca a diferença.
func TestMeta_DefinirIndicadores_DevolveListaAnteriorCompleta(t *testing.T) {
	ind15 := uuid.Must(uuid.NewV7())
	m, err := NovaMeta(uuid.Must(uuid.NewV7()), "Plano de ensino revisado", "", []uuid.UUID{ind15})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	ind21 := uuid.Must(uuid.NewV7())
	anteriores, err := m.DefinirIndicadores([]uuid.UUID{ind21})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(anteriores) != 1 || anteriores[0] != ind15 {
		t.Fatalf("esperava a lista anterior completa [%v], obtido %v", ind15, anteriores)
	}
	if len(m.Indicadores) != 1 || m.Indicadores[0] != ind21 {
		t.Fatalf("esperava a nova lista [%v], obtido %v", ind21, m.Indicadores)
	}
}

// TestMeta_DefinirIndicadores_RemoverOUltimoERecusado prova a segunda
// metade de MC-03: também ao editar, tentando remover o último.
func TestMeta_DefinirIndicadores_RemoverOUltimoERecusado(t *testing.T) {
	m, err := NovaMeta(uuid.Must(uuid.NewV7()), "Meta", "", []uuid.UUID{uuid.Must(uuid.NewV7())})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err = m.DefinirIndicadores(nil)
	if err != domain.ErrIndicadorObrigatorio {
		t.Fatalf("esperava ErrIndicadorObrigatorio, obtido %v", err)
	}
}

func TestNovaMeta_RecusaNomeVazio(t *testing.T) {
	_, err := NovaMeta(uuid.Must(uuid.NewV7()), "", "", []uuid.UUID{uuid.Must(uuid.NewV7())})
	if err == nil {
		t.Fatal("esperava erro para nome vazio")
	}
}

// TestMeta_DefinirQuantidadeSugerida_OpcionalENilLimpa prova que a
// sugestão é opcional (nil aceito, nunca obriga preenchimento) e que
// definir nil de novo limpa uma sugestão já gravada.
func TestMeta_DefinirQuantidadeSugerida_OpcionalENilLimpa(t *testing.T) {
	m, err := NovaMeta(uuid.Must(uuid.NewV7()), "Meta", "", []uuid.UUID{uuid.Must(uuid.NewV7())})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if m.QuantidadeSugerida != nil {
		t.Fatal("esperava QuantidadeSugerida nil por padrão")
	}
	quatro := 4
	if err := m.DefinirQuantidadeSugerida(&quatro); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if m.QuantidadeSugerida == nil || m.QuantidadeSugerida.Int() != 4 {
		t.Fatalf("esperava sugestão 4, obtido %v", m.QuantidadeSugerida)
	}
	if err := m.DefinirQuantidadeSugerida(nil); err != nil {
		t.Fatalf("erro inesperado ao limpar: %v", err)
	}
	if m.QuantidadeSugerida != nil {
		t.Fatal("esperava sugestão limpa após DefinirQuantidadeSugerida(nil)")
	}
}

// TestMeta_DefinirQuantidadeSugerida_RecusaMenorQueUm prova que a mesma
// regra de Quantidade (>= 1) vale para a sugestão.
func TestMeta_DefinirQuantidadeSugerida_RecusaMenorQueUm(t *testing.T) {
	m, err := NovaMeta(uuid.Must(uuid.NewV7()), "Meta", "", []uuid.UUID{uuid.Must(uuid.NewV7())})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	zero := 0
	if err := m.DefinirQuantidadeSugerida(&zero); err != domain.ErrQuantidadeInvalida {
		t.Fatalf("esperava ErrQuantidadeInvalida, obtido %v", err)
	}
}
