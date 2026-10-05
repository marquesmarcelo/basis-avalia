package avaliacao

import (
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/google/uuid"
)

// AV-11: qualquer PI desfaz a aceitação — consome rodada, abre prazo, e o
// cumprimento diminui (verificado no repositório real, não aqui).
func TestDesfazerAceitacao_Sucesso(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("aceita", 3), desfazerOK: true}
	local := fusoSaoPaulo()
	agora := time.Date(2026, 8, 12, 10, 0, 0, 0, local)
	uc := NovoDesfazerAceitacaoUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: agora}, local)
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), DesfazerAceitacaoInput{
		Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Motivo: "a ata anexada é de outra reunião", Versao: 3,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.desfeitas) != 1 {
		t.Fatalf("esperava um desfazimento gravado, obteve %d", len(repo.desfeitas))
	}
	if repo.desfeitas[0].Rodadas.Int() != 1 {
		t.Fatalf("esperava rodada 1, obteve %d", repo.desfeitas[0].Rodadas.Int())
	}
}

// AV-13: desfazer exige motivo e só vale sobre entrega aceita.
func TestDesfazerAceitacao_Guardas(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("recusada", 1), desfazerOK: true}
	uc := NovoDesfazerAceitacaoUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), DesfazerAceitacaoInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Motivo: "motivo", Versao: 1})
	if err != domain.ErrEntregaNaoEstaAceita {
		t.Fatalf("esperava ErrEntregaNaoEstaAceita, obteve %v", err)
	}
}

func TestDesfazerAceitacao_MotivoObrigatorio(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("aceita", 1), desfazerOK: true}
	uc := NovoDesfazerAceitacaoUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), DesfazerAceitacaoInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Motivo: "", Versao: 1})
	if err != domain.ErrMotivoObrigatorio {
		t.Fatalf("esperava ErrMotivoObrigatorio, obteve %v", err)
	}
}
