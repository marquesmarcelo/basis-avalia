package plano

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func periodoAbertoMock(t *testing.T) port.ItemPeriodo {
	t.Helper()
	inicio, _ := valueobject.DataLocalTexto("2026-01-01")
	fim, _ := valueobject.DataLocalTexto("2026-07-30")
	p, err := periodo.NovoPeriodo(uuid.Must(uuid.NewV7()), "2026.1", inicio, fim)
	if err != nil {
		t.Fatalf("setup periodo: %v", err)
	}
	return port.ItemPeriodo{Periodo: *p}
}

func periodoEncerradoMock(t *testing.T) port.ItemPeriodo {
	t.Helper()
	inicio, _ := valueobject.DataLocalTexto("2025-01-01")
	fim, _ := valueobject.DataLocalTexto("2025-07-30")
	p, err := periodo.NovoPeriodo(uuid.Must(uuid.NewV7()), "2025.1", inicio, fim)
	if err != nil {
		t.Fatalf("setup periodo: %v", err)
	}
	return port.ItemPeriodo{Periodo: *p}
}

func planoDeTeste(t *testing.T, instituicaoID uuid.UUID) *plano.Plano {
	t.Helper()
	p, err := plano.NovoPlano(instituicaoID, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()),
		plano.DadosDoPlano{Descricao: "D", ObjetivoGeral: "O", ResultadosEsperados: "R"})
	if err != nil {
		t.Fatalf("setup plano: %v", err)
	}
	return p
}

// TestPublicarPlano_SI02_SemItemRecusa prova SI-02 no use case completo
// (com a contagem de itens vindo do repositório).
func TestPublicarPlano_SI02_SemItemRecusa(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	p := planoDeTeste(t, instituicaoID)
	repo := &planoRepoMock{detalheParaBuscar: port.DetalhePlano{LinhaPlano: port.LinhaPlano{Plano: *p, TotalItens: 0}}}
	periodoRepo := &periodoRepoMock{item: periodoAbertoMock(t)}
	uc := NovoPublicarPlanoUseCase(repo, periodoRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), PublicarPlanoInput{Ator: atorPI(instituicaoID), PlanoID: p.ID, Versao: 1})
	if err != domain.ErrPlanoSemItem {
		t.Fatalf("esperava ErrPlanoSemItem, obtido %v", err)
	}
}

// TestPublicarPlano_SI09_PeriodoEncerradoRecusa prova SI-09.
func TestPublicarPlano_SI09_PeriodoEncerradoRecusa(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	p := planoDeTeste(t, instituicaoID)
	repo := &planoRepoMock{detalheParaBuscar: port.DetalhePlano{LinhaPlano: port.LinhaPlano{Plano: *p, TotalItens: 1}}}
	periodoRepo := &periodoRepoMock{item: periodoEncerradoMock(t)}
	uc := NovoPublicarPlanoUseCase(repo, periodoRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), PublicarPlanoInput{Ator: atorPI(instituicaoID), PlanoID: p.ID, Versao: 1})
	if err != domain.ErrPeriodoEncerrado {
		t.Fatalf("esperava ErrPeriodoEncerrado, obtido %v", err)
	}
}

// TestPublicarPlano_SI04_SemAprovacaoEhPermitidoComAviso prova SI-04: a
// publicação é permitida, e o resultado avisa da pendência — nunca um
// erro de "plano não aprovado".
func TestPublicarPlano_SI04_SemAprovacaoEhPermitidoComAviso(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	p := planoDeTeste(t, instituicaoID)
	repo := &planoRepoMock{detalheParaBuscar: port.DetalhePlano{LinhaPlano: port.LinhaPlano{Plano: *p, TotalItens: 1}}}
	periodoRepo := &periodoRepoMock{item: periodoAbertoMock(t)}
	uc := NovoPublicarPlanoUseCase(repo, periodoRepo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), PublicarPlanoInput{Ator: atorPI(instituicaoID), PlanoID: p.ID, Versao: 1})
	if err != nil {
		t.Fatalf("publicar sem aprovação deveria ser PERMITIDO, obtido erro: %v", err)
	}
	if !resultado.SemAprovacao {
		t.Fatal("resultado deveria avisar que falta aprovação")
	}
}
