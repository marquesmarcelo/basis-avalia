package meta

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// TestAtualizarMeta_MC13_AuditaListaAnteriorECompleta prova MC-13: a
// auditoria registra a lista anterior e a nova, completas — nunca a
// diferença.
func TestAtualizarMeta_MC13_AuditaListaAnteriorECompleta(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ind15, err := indicador.NovoDaPlataforma("1.5", "Coordenação de curso", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	ind21, err := indicador.NovoDaPlataforma("2.1", "Núcleo de apoio ao discente", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	existente, err := meta.NovaMeta(instituicaoID, "Plano de ensino revisado", "", []uuid.UUID{ind15.ID})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	indicadorRepo := &indicadorRepoPorIDMock{itens: map[uuid.UUID]port.ItemIndicador{
		ind15.ID: {Indicador: *ind15}, ind21.ID: {Indicador: *ind21},
	}}
	repo := &metaRepoMock{itemParaBuscar: port.ItemMeta{Meta: *existente}}
	audit := &auditMock{}
	uc := NovoAtualizarMetaUseCase(repo, indicadorRepo, audit, uowFake{})

	_, err = uc.Executar(context.Background(), AtualizarMetaInput{
		Ator: atorPI(t, instituicaoID), MetaID: existente.ID, Nome: "Plano de ensino revisado",
		Indicadores: []uuid.UUID{ind21.ID}, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(audit.eventos) != 1 {
		t.Fatalf("esperava 1 evento, obtido %d", len(audit.eventos))
	}
	antes, ok := audit.eventos[0].Detalhes["indicadores_antes"].([]string)
	if !ok || len(antes) != 1 || antes[0] != ind15.ID.String() {
		t.Fatalf("esperava indicadores_antes = [%s], obtido %v", ind15.ID, antes)
	}
	depois, ok := audit.eventos[0].Detalhes["indicadores_depois"].([]string)
	if !ok || len(depois) != 1 || depois[0] != ind21.ID.String() {
		t.Fatalf("esperava indicadores_depois = [%s], obtido %v", ind21.ID, depois)
	}
}
