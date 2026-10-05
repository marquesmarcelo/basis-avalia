package designacao

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	designacaodomain "github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func designacaoDeTeste(t *testing.T, cursoID, instituicaoID, coordenadorID uuid.UUID, inicioISO string, fimISO *string) *designacaodomain.Designacao {
	t.Helper()
	inicio := dataTeste(t, inicioISO)
	if fimISO != nil {
		f := dataTeste(t, *fimISO)
		d, err := designacaodomain.NovaDesignacao(cursoID, instituicaoID, coordenadorID, "1/2026", inicio, &f, false)
		if err != nil {
			t.Fatalf("setup designação: %v", err)
		}
		return d
	}
	d, err := designacaodomain.NovaDesignacao(cursoID, instituicaoID, coordenadorID, "1/2026", inicio, nil, false)
	if err != nil {
		t.Fatalf("setup designação: %v", err)
	}
	return d
}

// TestAtualizarDesignacao_Futura_TrocaCoordenadorEInicioLivremente prova
// design.md §5.3: situação futura aceita mudar tudo.
func TestAtualizarDesignacao_Futura_TrocaCoordenadorEInicioLivremente(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorAntigo := uuid.Must(uuid.NewV7())
	coordenadorNovo := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorAntigo, "2026-08-01", nil) // futura em relação a hoje
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoAtualizarDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	out, err := uc.Executar(context.Background(), AtualizarDesignacaoInput{
		Ator: ator, DesignacaoID: d.ID, CoordenadorID: coordenadorNovo, Portaria: "88/2026",
		DataInicio: "2026-09-01", Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.CoordenadorID != coordenadorNovo {
		t.Fatal("esperava coordenador trocado")
	}
	if !repo.atualizarCompChamado {
		t.Fatal("esperava AtualizarCompleta chamado (situação futura)")
	}
	if repo.atualizarFimChamado {
		t.Fatal("não deveria ter chamado AtualizarFimEPortaria")
	}
}

// TestAtualizarDesignacao_Vigente_TrocarCoordenadorDevolve409 prova a
// reconciliação 2 do dono: o backend RECUSA, não só desabilita na tela.
func TestAtualizarDesignacao_Vigente_TrocarCoordenadorDevolve409(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorAtual := uuid.Must(uuid.NewV7())
	outroCoordenador := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorAtual, "2026-01-01", nil) // vigente
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoAtualizarDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), AtualizarDesignacaoInput{
		Ator: ator, DesignacaoID: d.ID, CoordenadorID: outroCoordenador, Portaria: "1/2026",
		DataInicio: "2026-01-01", Versao: 1,
	})
	if err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito, obtido %v", err)
	}
	if repo.atualizarCompChamado || repo.atualizarFimChamado {
		t.Fatal("não deveria ter persistido nada")
	}
}

// TestAtualizarDesignacao_Vigente_ProrrogarPortariaEFimFunciona: mesma
// designação vigente, só portaria e data_fim mudando — permitido.
func TestAtualizarDesignacao_Vigente_ProrrogarPortariaEFimFunciona(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorID, "2026-01-01", nil)
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoAtualizarDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	novoFim := "2026-12-31"
	out, err := uc.Executar(context.Background(), AtualizarDesignacaoInput{
		Ator: ator, DesignacaoID: d.ID, CoordenadorID: coordenadorID, Portaria: "1/2026 (prorrogada)",
		DataInicio: "2026-01-01", DataFim: &novoFim, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Vigencia.Fim() == nil || out.Vigencia.Fim().String() != novoFim {
		t.Fatalf("esperava data_fim prorrogada para %s, obtido %v", novoFim, out.Vigencia.Fim())
	}
	if !repo.atualizarFimChamado {
		t.Fatal("esperava AtualizarFimEPortaria chamado (situação vigente)")
	}
}

// TestAtualizarDesignacao_Encerrada_TrocarInicioDevolve409 — mesma regra
// para encerrada.
func TestAtualizarDesignacao_Encerrada_TrocarInicioDevolve409(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	fim := "2026-01-31"
	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorID, "2025-06-01", &fim) // encerrada
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoAtualizarDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), AtualizarDesignacaoInput{
		Ator: ator, DesignacaoID: d.ID, CoordenadorID: coordenadorID, Portaria: "1/2026",
		DataInicio: "2025-07-01", Versao: 1,
	})
	if err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito, obtido %v", err)
	}
}
