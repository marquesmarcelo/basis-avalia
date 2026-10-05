package usuario

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func TestRedefinirSenhaUsuario_E11_RedefinicaoComSucesso(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Professor), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo}
	audit := &auditMock{}
	uc := NovoRedefinirSenhaUsuarioUseCase(repo, hashMock{}, audit, uowFake{})

	err := uc.Executar(context.Background(), RedefinirSenhaUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, SenhaNova: senhaTeste(t, "senha-provisoria-nova"),
	})

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !repo.definirSenhaChamado {
		t.Fatal("DefinirSenha deveria ter sido chamado")
	}
	if !repo.senhaProvisoriaRecebida {
		t.Fatal("a senha redefinida por outro usuário precisa nascer provisória (E-11)")
	}
	if len(audit.eventos) != 1 {
		t.Fatalf("esperava 1 evento de auditoria, obtido %d", len(audit.eventos))
	}
}

func TestRedefinirSenhaUsuario_E12_NaoValeParaSiMesmo(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	proprioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoRedefinirSenhaUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	err := uc.Executar(context.Background(), RedefinirSenhaUsuarioInput{
		Ator: atorComID(t, proprioID, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: proprioID, SenhaNova: senhaTeste(t, "senha-provisoria-nova"),
	})

	if !errors.Is(err, domain.ErrRedefinirPropriaSenhaNegada) {
		t.Fatalf("esperava ErrRedefinirPropriaSenhaNegada, obtido %v", err)
	}
	if repo.definirSenhaChamado {
		t.Fatal("DefinirSenha não deveria ter sido chamado — a recusa acontece antes de qualquer acesso ao repositório")
	}
}
