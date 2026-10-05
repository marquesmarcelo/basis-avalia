package avaliacao

import (
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func detalheDeTeste(situacao string, versao int) port.DetalheEntrega {
	return port.DetalheEntrega{LinhaEntrega: port.LinhaEntrega{
		ID: uuid.Must(uuid.NewV7()), CursoID: uuid.Must(uuid.NewV7()), Situacao: situacao, Versao: versao,
	}}
}

// AV-01: aceitar a entrega grava quem avaliou e o cumprimento sobe.
func TestAvaliar_Aceitar(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("pendente_avaliacao", 1), avaliarOK: true}
	uc := NovoAvaliarUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), AvaliarInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Resultado: "aceita", Versao: 1})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.avaliadas) != 1 {
		t.Fatalf("esperava uma avaliação gravada, obteve %d", len(repo.avaliadas))
	}
}

// AV-02: recusar sem motivo responde 400 MOTIVO_OBRIGATORIO, a entrega
// continua pendente.
func TestAvaliar_RecusarSemMotivo(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("pendente_avaliacao", 1), avaliarOK: true}
	uc := NovoAvaliarUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), AvaliarInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Resultado: "recusada", Motivo: "", Versao: 1})
	if err != domain.ErrMotivoObrigatorio {
		t.Fatalf("esperava ErrMotivoObrigatorio, obteve %v", err)
	}
	if len(repo.avaliadas) != 0 {
		t.Fatal("nenhuma avaliação deveria ter sido gravada")
	}
}

// AV-06: entrega já avaliada não é avaliada de novo.
func TestAvaliar_JaAvaliada(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("aceita", 2), avaliarOK: false}
	uc := NovoAvaliarUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), AvaliarInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Resultado: "aceita", Versao: 2})
	if err != domain.ErrEntregaJaAvaliada {
		t.Fatalf("esperava ErrEntregaJaAvaliada, obteve %v", err)
	}
}

// AV-07 (mocado — a corrida real com banco está em integração): versão
// divergente responde CONFLITO_DE_VERSAO, não ENTREGA_JA_AVALIADA.
func TestAvaliar_ConflitoDeVersao(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("pendente_avaliacao", 5), avaliarOK: false}
	uc := NovoAvaliarUseCase(repo, &auditMock{}, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	err := uc.Executar(t.Context(), AvaliarInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Resultado: "aceita", Versao: 3})
	if err != domain.ErrConflitoDeVersao {
		t.Fatalf("esperava ErrConflitoDeVersao, obteve %v", err)
	}
}

// AV-15/AV-18: a marca avaliador_e_coordenador_do_curso é gravada
// conforme CoordenaCursoHoje, e aparece na auditoria.
func TestAvaliar_MarcaDeCoincidencia(t *testing.T) {
	repo := &entregaRepoMock{detalhe: detalheDeTeste("pendente_avaliacao", 1), avaliarOK: true, coordena: true}
	auditor := &auditMock{}
	uc := NovoAvaliarUseCase(repo, auditor, uowFake{}, relogioFixo{instante: time.Now()}, fusoSaoPaulo())
	instituicaoID := uuid.Must(uuid.NewV7())

	if err := uc.Executar(t.Context(), AvaliarInput{Ator: atorPI(instituicaoID), EntregaID: repo.detalhe.ID, Resultado: "aceita", Versao: 1}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !repo.avaliadas[0].AvaliadorEraCoordenador {
		t.Fatal("esperava AvaliadorEraCoordenador = true na entrega avaliada")
	}
	if auditor.eventos[0].Detalhes["avaliador_e_coordenador_do_curso"] != true {
		t.Fatal("esperava a marca na auditoria")
	}
}
