package meta

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// TestCriarMeta_MC07_IndicadorDeOutraInstituicaoRecebe404 prova a segunda
// metade de MC-07: indicador de outra instituição — 404, nunca 403 nem
// 400 (o mock devolve ErrNaoEncontrado quando o id não está na carteira,
// simulando exatamente o que AplicarEscopo faria no adapter real).
func TestCriarMeta_MC07_IndicadorDeOutraInstituicaoRecebe404(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	idDeOutraInstituicao := uuid.Must(uuid.NewV7())
	indicadorRepo := &indicadorRepoPorIDMock{itens: map[uuid.UUID]port.ItemIndicador{}}
	uc := NovoCriarMetaUseCase(&metaRepoMock{}, indicadorRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarMetaInput{
		Ator: atorPI(t, instituicaoID), Nome: "Meta de teste", Indicadores: []uuid.UUID{idDeOutraInstituicao},
	})
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado (404), obtido %v", err)
	}
}

// TestCriarMeta_MC07_IndicadorInativoRecebe400 prova a primeira metade de
// MC-07.
func TestCriarMeta_MC07_IndicadorInativoRecebe400(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	inativo, err := indicador.NovoDaInstituicao(instituicaoID, "GEST-02", "Painel interno de acompanhamento", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	inativo.AlterarSituacao(valueobject.CatalogoInativo)
	indicadorRepo := &indicadorRepoPorIDMock{itens: map[uuid.UUID]port.ItemIndicador{
		inativo.ID: {Indicador: *inativo},
	}}
	uc := NovoCriarMetaUseCase(&metaRepoMock{}, indicadorRepo, &auditMock{}, uowFake{})

	_, err = uc.Executar(context.Background(), CriarMetaInput{
		Ator: atorPI(t, instituicaoID), Nome: "Meta de teste", Indicadores: []uuid.UUID{inativo.ID},
	})
	if err != domain.ErrIndicadorInativo {
		t.Fatalf("esperava ErrIndicadorInativo (400), obtido %v", err)
	}
}

// TestCriarMeta_MC03_SemIndicadorNaoChamaRepositorio prova que a validação
// estrutural do domínio (MC-03) acontece antes de qualquer I/O.
func TestCriarMeta_MC03_SemIndicadorNaoChamaRepositorio(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &metaRepoMock{}
	uc := NovoCriarMetaUseCase(repo, &indicadorRepoPorIDMock{itens: map[uuid.UUID]port.ItemIndicador{}}, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarMetaInput{Ator: atorPI(t, instituicaoID), Nome: "Meta sem indicador"})
	if err != domain.ErrIndicadorObrigatorio {
		t.Fatalf("esperava ErrIndicadorObrigatorio, obtido %v", err)
	}
	if repo.inserirChamado {
		t.Fatal("Inserir não deveria ter sido chamado")
	}
}

// TestCriarMeta_MC02_ComDoisIndicadoresAtivos prova o caminho feliz do caso
// do dono: dois indicadores válidos, meta criada com sucesso e auditada.
func TestCriarMeta_MC02_ComDoisIndicadoresAtivos(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ind14, err := indicador.NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	ind15, err := indicador.NovoDaPlataforma("1.5", "Coordenação de curso", "", "referência")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	indicadorRepo := &indicadorRepoPorIDMock{itens: map[uuid.UUID]port.ItemIndicador{
		ind14.ID: {Indicador: *ind14}, ind15.ID: {Indicador: *ind15},
	}}
	repo := &metaRepoMock{}
	audit := &auditMock{}
	uc := NovoCriarMetaUseCase(repo, indicadorRepo, audit, uowFake{})

	resultado, err := uc.Executar(context.Background(), CriarMetaInput{
		Ator: atorPI(t, instituicaoID), Nome: "Registrar reuniões de NDE em ata",
		Indicadores: []uuid.UUID{ind14.ID, ind15.ID},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(resultado.Indicadores) != 2 {
		t.Fatalf("esperava 2 indicadores, obtido %d", len(resultado.Indicadores))
	}
	if !repo.inserirChamado {
		t.Fatal("Inserir deveria ter sido chamado")
	}
	if len(audit.eventos) != 1 {
		t.Fatalf("esperava 1 evento de auditoria, obtido %d", len(audit.eventos))
	}
}
