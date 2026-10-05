package indicador

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// TestAtualizarIndicador_IE04_PIRecusadoAoEscreverNoINEP prova a
// assimetria de IE-04: o PI ENXERGA o indicador do INEP (BuscarPorID o
// devolve, exceção do catálogo) mas é recusado ao ESCREVER — 403, não 404.
func TestAtualizarIndicador_IE04_PIRecusadoAoEscreverNoINEP(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	doInep, err := indicador.NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	repo := &indicadorRepoMock{itemParaBuscar: port.ItemIndicador{Indicador: *doInep}}
	audit := &auditMock{}
	uc := NovoAtualizarIndicadorUseCase(repo, audit, uowFake{})

	_, err = uc.Executar(context.Background(), AtualizarIndicadorInput{
		Ator: atorPI(t, instituicaoID), IndicadorID: doInep.ID, Codigo: "1.4", Nome: "Outro nome", Versao: 1,
	})
	if err != domain.ErrPermissaoNegada {
		t.Fatalf("esperava ErrPermissaoNegada (403), obtido %v", err)
	}
	if repo.atualizarChamado {
		t.Fatal("Atualizar não deveria ter sido chamado")
	}
	if len(audit.eventos) != 0 {
		t.Fatal("nenhum evento de auditoria deveria ter sido registrado")
	}
}

func TestAtualizarIndicador_PermiteEscritaNoIndicadorProprio(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	proprio, err := indicador.NovoDaInstituicao(instituicaoID, "GEST-01", "Reuniões com representação discente", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	repo := &indicadorRepoMock{itemParaBuscar: port.ItemIndicador{Indicador: *proprio}}
	audit := &auditMock{}
	uc := NovoAtualizarIndicadorUseCase(repo, audit, uowFake{})

	_, err = uc.Executar(context.Background(), AtualizarIndicadorInput{
		Ator: atorPI(t, instituicaoID), IndicadorID: proprio.ID, Codigo: "GEST-01", Nome: "Nome corrigido", Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !repo.atualizarChamado {
		t.Fatal("Atualizar deveria ter sido chamado")
	}
	if len(audit.eventos) != 1 {
		t.Fatalf("esperava 1 evento de auditoria, obtido %d", len(audit.eventos))
	}
}

// TestAlterarSituacaoIndicador_IE04 e TestExcluirIndicador_IE04 provam a
// mesma assimetria nas outras duas operações.
func TestAlterarSituacaoIndicador_IE04_PIRecusadoNoINEP(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	doInep, err := indicador.NovoDaPlataforma("1.4", "Nome", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	repo := &indicadorRepoMock{itemParaBuscar: port.ItemIndicador{Indicador: *doInep}}
	uc := NovoAlterarSituacaoIndicadorUseCase(repo, &auditMock{}, uowFake{})

	err = uc.Executar(context.Background(), AlterarSituacaoIndicadorInput{
		Ator: atorPI(t, instituicaoID), IndicadorID: doInep.ID, NovaSituacao: valueobject.CatalogoInativo, Versao: 1,
	})
	if err != domain.ErrPermissaoNegada {
		t.Fatalf("esperava ErrPermissaoNegada (403), obtido %v", err)
	}
	if repo.alterarSituacaoChamado {
		t.Fatal("AlterarSituacao não deveria ter sido chamado")
	}
}

func TestExcluirIndicador_IE04_PIRecusadoNoINEP(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	doInep, err := indicador.NovoDaPlataforma("1.4", "Nome", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	repo := &indicadorRepoMock{itemParaBuscar: port.ItemIndicador{Indicador: *doInep}}
	uc := NovoExcluirIndicadorUseCase(repo, &auditMock{}, uowFake{})

	err = uc.Executar(context.Background(), ExcluirIndicadorInput{
		Ator: atorPI(t, instituicaoID), IndicadorID: doInep.ID,
	})
	if err != domain.ErrPermissaoNegada {
		t.Fatalf("esperava ErrPermissaoNegada (403), obtido %v", err)
	}
	if repo.excluirChamado {
		t.Fatal("ExcluirSeSemUso não deveria ter sido chamado")
	}
}
