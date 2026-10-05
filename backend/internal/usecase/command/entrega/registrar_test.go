package entrega

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/itemplano"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func montarUseCase(itemRepo *itemPlanoRepoMock, planoRepo *planoRepoMock, entregaRepo *entregaRepoMock, arm *armazenamentoMock) *RegistrarEntregaUseCase {
	if entregaRepo == nil {
		entregaRepo = &entregaRepoMock{}
	}
	if arm == nil {
		arm = &armazenamentoMock{}
	}
	return NovoRegistrarEntregaUseCase(entregaRepo, itemRepo, planoRepo, arm, &auditMock{}, uowFake{})
}

func itemPadrao(instituicaoID, cursoID, planoID uuid.UUID) itemplano.ItemDoPlano {
	item, _ := itemplano.NovoItemDoPlano(planoID, cursoID, instituicaoID, uuid.Must(uuid.NewV7()), 4)
	return *item
}

func pdfDeTeste(nome string) ArquivoRecebido {
	return ArquivoRecebido{NomeOriginal: nome, Conteudo: []byte("%PDF-1.4\nconteudo de teste")}
}

// EN-03 reescrito (M-14): plano em rascunho responde 409
// PLANO_NAO_VIGENTE, nunca 404 — o coordenador já leu o plano.
func TestRegistrarEntrega_PlanoRascunho(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	planoID := uuid.Must(uuid.NewV7())
	item := itemPadrao(instituicaoID, cursoID, planoID)

	itemRepo := &itemPlanoRepoMock{item: item}
	planoRepo := &planoRepoMock{detalhe: port.DetalhePlano{LinhaPlano: port.LinhaPlano{Situacao: "rascunho"}}}

	uc := montarUseCase(itemRepo, planoRepo, nil, nil)
	ator := atorCoordenador(instituicaoID)

	_, err := uc.Executar(t.Context(), RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: item.ID, IdempotencyKey: "chave-1", Arquivos: []ArquivoRecebido{pdfDeTeste("ata.pdf")},
	})
	if err != domain.ErrPlanoNaoVigente {
		t.Fatalf("esperava ErrPlanoNaoVigente, obteve %v", err)
	}
}

// EN-02: entrega sem nenhum comprovante responde 400 ENTREGA_SEM_ANEXO —
// e nenhum objeto é gravado no armazenamento.
func TestRegistrarEntrega_SemAnexo(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	planoID := uuid.Must(uuid.NewV7())
	item := itemPadrao(instituicaoID, cursoID, planoID)

	itemRepo := &itemPlanoRepoMock{item: item}
	planoRepo := &planoRepoMock{
		detalhe:    port.DetalhePlano{LinhaPlano: port.LinhaPlano{Situacao: "vigente"}},
		cursoAtivo: true,
		periodo:    periodoAberto(),
	}
	arm := &armazenamentoMock{}
	uc := montarUseCase(itemRepo, planoRepo, nil, arm)
	ator := atorCoordenador(instituicaoID)

	_, err := uc.Executar(t.Context(), RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: item.ID, IdempotencyKey: "chave-1", Arquivos: nil,
	})
	if err != domain.ErrEntregaSemAnexo {
		t.Fatalf("esperava ErrEntregaSemAnexo, obteve %v", err)
	}
	if len(arm.gravados) != 0 {
		t.Fatalf("esperava nenhum objeto gravado, gravou %d", len(arm.gravados))
	}
}

// EN-01: registro com sucesso grava a entrega vinculada ao item.
func TestRegistrarEntrega_Sucesso(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	planoID := uuid.Must(uuid.NewV7())
	item := itemPadrao(instituicaoID, cursoID, planoID)

	itemRepo := &itemPlanoRepoMock{item: item}
	planoRepo := &planoRepoMock{
		detalhe:    port.DetalhePlano{LinhaPlano: port.LinhaPlano{Situacao: "vigente"}},
		cursoAtivo: true,
		periodo:    periodoAberto(),
	}
	entregaRepo := &entregaRepoMock{}
	arm := &armazenamentoMock{}
	uc := montarUseCase(itemRepo, planoRepo, entregaRepo, arm)
	ator := atorCoordenador(instituicaoID)

	resultado, err := uc.Executar(t.Context(), RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: item.ID, Observacao: "Reunião ordinária do NDE de 12/03/2026.",
		IdempotencyKey: "chave-1", Arquivos: []ArquivoRecebido{pdfDeTeste("ata.pdf")},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Reaproveitada {
		t.Fatal("primeira chamada não deveria ser reaproveitada")
	}
	if len(entregaRepo.inseridos) != 1 {
		t.Fatalf("esperava uma entrega inserida, obteve %d", len(entregaRepo.inseridos))
	}
	if entregaRepo.inseridos[0].ItemPlanoID != item.ID {
		t.Fatal("entrega não vinculada ao item correto")
	}
	if len(arm.gravados) != 1 {
		t.Fatalf("esperava um objeto gravado, obteve %d", len(arm.gravados))
	}
}

// EN-06 (mocado — a corrida real está em teste de integração): reenvio
// com a mesma Idempotency-Key devolve a MESMA entrega, marcada como
// reaproveitada — e nenhuma segunda linha é inserida.
func TestRegistrarEntrega_ReenvioComMesmaChave(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	planoID := uuid.Must(uuid.NewV7())
	item := itemPadrao(instituicaoID, cursoID, planoID)

	itemRepo := &itemPlanoRepoMock{item: item}
	planoRepo := &planoRepoMock{
		detalhe:    port.DetalhePlano{LinhaPlano: port.LinhaPlano{Situacao: "vigente"}},
		cursoAtivo: true,
		periodo:    periodoAberto(),
	}
	entregaRepo := &entregaRepoMock{}
	arm := &armazenamentoMock{}
	uc := montarUseCase(itemRepo, planoRepo, entregaRepo, arm)
	ator := atorCoordenador(instituicaoID)

	in := RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: item.ID, IdempotencyKey: "chave-repetida",
		Arquivos: []ArquivoRecebido{pdfDeTeste("ata.pdf")},
	}
	primeira, err := uc.Executar(t.Context(), in)
	if err != nil {
		t.Fatalf("erro na primeira chamada: %v", err)
	}
	segunda, err := uc.Executar(t.Context(), in)
	if err != nil {
		t.Fatalf("erro na segunda chamada: %v", err)
	}
	if !segunda.Reaproveitada {
		t.Fatal("esperava a segunda chamada reaproveitada")
	}
	if segunda.EntregaID != primeira.EntregaID {
		t.Fatalf("esperava a MESMA entrega (%s), obteve %s", primeira.EntregaID, segunda.EntregaID)
	}
	if len(entregaRepo.inseridos) != 1 {
		t.Fatalf("esperava exatamente UMA entrega inserida, obteve %d", len(entregaRepo.inseridos))
	}
	// Os objetos gravados para a segunda tentativa são removidos
	// (best-effort) — o armazenamento não deve acumular órfãos referentes
	// a uma inserção que não aconteceu.
	if len(arm.removidos) != 1 {
		t.Fatalf("esperava um objeto removido (órfão da segunda tentativa), obteve %d", len(arm.removidos))
	}
}

// AN-02, nível 1, exercitado no fluxo completo: arquivo renomeado é
// recusado antes de qualquer gravação.
func TestRegistrarEntrega_AnexoTipoNaoPermitido(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoID := uuid.Must(uuid.NewV7())
	planoID := uuid.Must(uuid.NewV7())
	item := itemPadrao(instituicaoID, cursoID, planoID)

	itemRepo := &itemPlanoRepoMock{item: item}
	planoRepo := &planoRepoMock{
		detalhe:    port.DetalhePlano{LinhaPlano: port.LinhaPlano{Situacao: "vigente"}},
		cursoAtivo: true,
		periodo:    periodoAberto(),
	}
	arm := &armazenamentoMock{}
	uc := montarUseCase(itemRepo, planoRepo, nil, arm)
	ator := atorCoordenador(instituicaoID)

	falso := ArquivoRecebido{NomeOriginal: "ata-nde.pdf", Conteudo: []byte("MZ\x90\x00conteudo de executavel")}
	_, err := uc.Executar(t.Context(), RegistrarEntregaInput{
		Ator: ator, ItemPlanoID: item.ID, IdempotencyKey: "chave-1", Arquivos: []ArquivoRecebido{falso},
	})
	if err != domain.ErrAnexoTipoNaoPermitido {
		t.Fatalf("esperava ErrAnexoTipoNaoPermitido, obteve %v", err)
	}
	if len(arm.gravados) != 0 {
		t.Fatal("nenhum byte deveria ser gravado")
	}
}
