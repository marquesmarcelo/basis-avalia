package designacao

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// TestExcluirDesignacao_Futura_Funciona prova DG-08: designação futura
// pode ser excluída inteira.
func TestExcluirDesignacao_Futura_Funciona(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorID, "2026-08-01", nil) // futura
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoExcluirDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	if err := uc.Executar(context.Background(), ExcluirDesignacaoInput{Ator: ator, DesignacaoID: d.ID}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !repo.excluirChamado {
		t.Fatal("esperava Excluir chamado")
	}
}

// TestExcluirDesignacao_Vigente_RecusaComEfeito prova que uma designação
// vigente não pode ser excluída — 409 DESIGNACAO_COM_EFEITO.
func TestExcluirDesignacao_Vigente_RecusaComEfeito(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorID, "2026-01-01", nil) // vigente
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoExcluirDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	err := uc.Executar(context.Background(), ExcluirDesignacaoInput{Ator: ator, DesignacaoID: d.ID})
	if err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito, obtido %v", err)
	}
	if repo.excluirChamado {
		t.Fatal("não deveria ter chamado Excluir")
	}
}

// TestExcluirDesignacao_Encerrada_RecusaComEfeito — mesma regra para
// encerrada: já é histórico.
func TestExcluirDesignacao_Encerrada_RecusaComEfeito(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	coordenadorID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	fim := "2026-01-31"
	d := designacaoDeTeste(t, cursoID, instituicaoID, coordenadorID, "2025-06-01", &fim) // encerrada
	repo := &designacaoRepoMock{itemParaBuscar: port.ItemDesignacao{Designacao: *d}}
	uc := NovoExcluirDesignacaoUseCase(repo, &auditMock{}, uowFake{})

	err := uc.Executar(context.Background(), ExcluirDesignacaoInput{Ator: ator, DesignacaoID: d.ID})
	if err != domain.ErrDesignacaoComEfeito {
		t.Fatalf("esperava ErrDesignacaoComEfeito, obtido %v", err)
	}
}
