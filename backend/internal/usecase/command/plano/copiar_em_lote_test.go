package plano

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func detalheDeOrigem(instituicaoID uuid.UUID, comAprovacao bool) port.DetalhePlano {
	dados := plano.DadosDoPlano{Descricao: "Descrição", ObjetivoGeral: "Objetivo", ResultadosEsperados: "Resultados"}
	if comAprovacao {
		dados.AprovacaoData, dados.AprovacaoOrgao = "2026-02-10", "nde"
	}
	origem, _ := plano.NovoPlano(instituicaoID, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), dados)
	origem.SituacaoPublicacao = "vigente"
	metaA, metaB := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	return port.DetalhePlano{
		LinhaPlano: port.LinhaPlano{Plano: *origem},
		Itens: []port.ItemDoPlanoResponse{
			{MetaID: metaA, Quantidade: 4},
			{MetaID: metaB, Quantidade: 2},
		},
	}
}

// TestCopiarEmLote_CP02_NaoLevaAprovacaoNemSituacao prova CP-02: o pior
// defeito possível da feature seria atribuir aprovação institucional que
// não aconteceu — cada plano copiado nasce em rascunho, sem aprovação.
func TestCopiarEmLote_CP02_NaoLevaAprovacaoNemSituacao(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &planoRepoMock{detalheParaBuscar: detalheDeOrigem(instituicaoID, true)}
	uc := NovoCopiarEmLoteUseCase(repo, &auditMock{}, uowFake{})

	destino := uuid.Must(uuid.NewV7())
	resultado, err := uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()),
		PeriodoDestinoID: uuid.Must(uuid.NewV7()), Cursos: []uuid.UUID{destino},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Criados != 1 {
		t.Fatalf("esperava 1 criado, obtido %d", resultado.Criados)
	}
	if len(repo.inseridos) != 1 {
		t.Fatalf("esperava 1 plano inserido, obtido %d", len(repo.inseridos))
	}
	copiado := repo.inseridos[0]
	if copiado.SituacaoPublicacao != "rascunho" {
		t.Fatalf("plano copiado deveria nascer em rascunho, obtido %q", copiado.SituacaoPublicacao)
	}
	if copiado.Aprovacao != nil {
		t.Fatal("plano copiado NUNCA deveria levar a aprovação da origem")
	}
}

// TestCopiarEmLote_CP05_CursoJaComPlanoEhPulado prova CP-05: a violação
// do índice único é tratada como resultado esperado, nunca como exceção —
// o curso entra em pulados, nunca substitui.
func TestCopiarEmLote_CP05_CursoJaComPlanoEhPulado(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoComPlano := uuid.Must(uuid.NewV7())
	cursoLivre := uuid.Must(uuid.NewV7())
	repo := &planoRepoMock{
		detalheParaBuscar: detalheDeOrigem(instituicaoID, false),
		errosPorCurso:     map[uuid.UUID]error{cursoComPlano: domain.ErrPlanoDuplicado},
	}
	uc := NovoCopiarEmLoteUseCase(repo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()),
		PeriodoDestinoID: uuid.Must(uuid.NewV7()), Cursos: []uuid.UUID{cursoComPlano, cursoLivre},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Criados != 1 {
		t.Fatalf("esperava 1 criado, obtido %d", resultado.Criados)
	}
	if len(resultado.Pulados) != 1 || resultado.Pulados[0].CursoID != cursoComPlano {
		t.Fatalf("esperava o curso com plano pulado, obtido %+v", resultado.Pulados)
	}
	if resultado.Pulados[0].Motivo != "já tem plano neste período" {
		t.Fatalf("motivo inesperado: %q", resultado.Pulados[0].Motivo)
	}
}

// TestCopiarEmLote_CP07_FalhaEmUmNaoDesfazOsDemais prova que uma falha
// isolada (não relacionada a duplicidade) não impede os cursos seguintes.
func TestCopiarEmLote_CP07_FalhaEmUmNaoDesfazOsDemais(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	cursoComFalha := uuid.Must(uuid.NewV7())
	cursoOK1, cursoOK2 := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	repo := &planoRepoMock{
		detalheParaBuscar: detalheDeOrigem(instituicaoID, false),
		errosPorCurso:     map[uuid.UUID]error{cursoComFalha: domain.ErrNaoEncontrado},
	}
	uc := NovoCopiarEmLoteUseCase(repo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()),
		PeriodoDestinoID: uuid.Must(uuid.NewV7()), Cursos: []uuid.UUID{cursoOK1, cursoComFalha, cursoOK2},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Criados != 2 {
		t.Fatalf("esperava 2 criados (os dois sem falha), obtido %d", resultado.Criados)
	}
	if len(resultado.Pulados) != 1 {
		t.Fatalf("esperava 1 pulado, obtido %d", len(resultado.Pulados))
	}
}

// TestCopiarEmLote_CP12_UmRegistroDeAuditoriaPorPlanoCriado prova que não
// existe um único registro representando o lote inteiro.
func TestCopiarEmLote_CP12_UmRegistroDeAuditoriaPorPlanoCriado(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &planoRepoMock{detalheParaBuscar: detalheDeOrigem(instituicaoID, false)}
	audit := &auditMock{}
	uc := NovoCopiarEmLoteUseCase(repo, audit, uowFake{})

	cursos := []uuid.UUID{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())}
	if _, err := uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()), PeriodoDestinoID: uuid.Must(uuid.NewV7()), Cursos: cursos,
	}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(audit.eventos) != 3 {
		t.Fatalf("esperava 3 eventos de auditoria (um por plano criado), obtido %d", len(audit.eventos))
	}
}

// TestCopiarEmLote_CP11_LimitesDoLote prova os dois limites.
func TestCopiarEmLote_CP11_LimitesDoLote(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &planoRepoMock{detalheParaBuscar: detalheDeOrigem(instituicaoID, false)}
	uc := NovoCopiarEmLoteUseCase(repo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()), PeriodoDestinoID: uuid.Must(uuid.NewV7()),
	})
	if err != domain.ErrCursosObrigatorios {
		t.Fatalf("esperava ErrCursosObrigatorios, obtido %v", err)
	}

	cursos := make([]uuid.UUID, 101)
	for i := range cursos {
		cursos[i] = uuid.Must(uuid.NewV7())
	}
	_, err = uc.Executar(context.Background(), CopiarEmLoteInput{
		Ator: atorPI(instituicaoID), PlanoOrigemID: uuid.Must(uuid.NewV7()), PeriodoDestinoID: uuid.Must(uuid.NewV7()), Cursos: cursos,
	})
	if err != domain.ErrLoteAcimaDoLimite {
		t.Fatalf("esperava ErrLoteAcimaDoLimite, obtido %v", err)
	}
}
