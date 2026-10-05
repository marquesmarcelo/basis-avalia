package usuario

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/google/uuid"
)

func TestExcluirUsuario_E08_NinguemExcluiASiMesmo(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	proprioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoExcluirUsuarioUseCase(repo, &auditMock{}, uowFake{})

	err := uc.Executar(context.Background(), ExcluirUsuarioInput{
		Ator: atorComID(t, proprioID, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: proprioID,
	})

	if !errors.Is(err, domain.ErrAutoExclusaoNegada) {
		t.Fatalf("esperava ErrAutoExclusaoNegada, obtido %v", err)
	}
	if repo.excluirLogicamenteChamado {
		t.Fatal("ExcluirLogicamente não deveria ter sido chamado — a recusa acontece antes de qualquer acesso ao repositório")
	}
}
