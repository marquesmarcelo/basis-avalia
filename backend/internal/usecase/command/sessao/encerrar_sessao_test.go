package sessao

import (
	"context"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type invalidarSessoesChamada struct {
	usuarioID uuid.UUID
	instante  time.Time
}

type autenticacaoRepoLogoutMock struct {
	chamadas []invalidarSessoesChamada
}

func (m *autenticacaoRepoLogoutMock) InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error {
	m.chamadas = append(m.chamadas, invalidarSessoesChamada{usuarioID: p.UsuarioID(), instante: instante})
	return nil
}

func TestEncerrarSessao_SE01_GravaSessoesValidasAPartirDeAgora(t *testing.T) {
	repo := &autenticacaoRepoLogoutMock{}
	agora := time.Now()
	uc := NovoEncerrarSessaoUseCase(repo, relogioMock{agora: agora})

	usuarioID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(usuarioID, valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("ator de teste: %v", err)
	}
	if err := uc.Executar(context.Background(), ator); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.chamadas) != 1 {
		t.Fatalf("esperava 1 chamada, obtido %d", len(repo.chamadas))
	}
	if !repo.chamadas[0].instante.Equal(agora) {
		t.Fatal("instante incorreto")
	}
}

func TestEncerrarSessao_SE01_ChamarDuasVezesEhIdempotente(t *testing.T) {
	repo := &autenticacaoRepoLogoutMock{}
	uc := NovoEncerrarSessaoUseCase(repo, relogioMock{agora: time.Now()})
	usuarioID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(usuarioID, valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("ator de teste: %v", err)
	}

	if err := uc.Executar(context.Background(), ator); err != nil {
		t.Fatalf("primeira chamada: %v", err)
	}
	if err := uc.Executar(context.Background(), ator); err != nil {
		t.Fatalf("segunda chamada: %v", err)
	}
	if len(repo.chamadas) != 2 {
		t.Fatalf("esperava 2 chamadas bem-sucedidas, obtido %d", len(repo.chamadas))
	}
}
