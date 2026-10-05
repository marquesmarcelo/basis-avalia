package usuario

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func usuarioAlvoTeste(t *testing.T, perfis valueobject.ConjuntoDePerfis, instituicaoID *uuid.UUID) *usuario.Usuario {
	t.Helper()
	hash, err := valueobject.NovaSenhaHash("$argon2id$v=19$m=1,t=1,p=1$c2FsdA$aGFzaA")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	u, err := usuario.NovoUsuario("Alvo Teste", emailTeste(t, "alvo.teste@fsa.edu.br"), hash, perfis, instituicaoID)
	if err != nil {
		t.Fatalf("NovoUsuario: %v", err)
	}
	return u
}

func atorComID(t *testing.T, id uuid.UUID, instituicaoID uuid.UUID) autorizacao.Ator {
	t.Helper()
	conjunto := mustConjuntoTeste(t, valueobject.PesquisadorInstitucional)
	ator, err := autorizacao.NovoAtor(id, conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator
}

func TestAtualizarUsuario_E14_PromoverMantemOQueJaExistia(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Aluno), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo}
	uc := NovoAtualizarUsuarioUseCase(repo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, Nome: alvo.Nome, Email: alvo.Email,
		Perfis: []valueobject.Perfil{valueobject.Aluno, valueobject.Professor}, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resultado.Perfis.Igual(mustConjuntoTeste(t, valueobject.Aluno, valueobject.Professor)) {
		t.Fatalf("esperava {aluno, professor} — aluno preservado, obtido %v", resultado.Perfis.Ordenado())
	}
	if !repo.substituirPerfisChamado {
		t.Fatal("SubstituirPerfis deveria ter sido chamado")
	}
}

func TestAtualizarUsuario_E15_RetirarUmPerfilNaoRetiraOsOutros(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Aluno, valueobject.Professor), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo, totalDetentores: 5}
	uc := NovoAtualizarUsuarioUseCase(repo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, Nome: alvo.Nome, Email: alvo.Email,
		Perfis: []valueobject.Perfil{valueobject.Aluno}, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resultado.Perfis.Igual(mustConjuntoTeste(t, valueobject.Aluno)) {
		t.Fatalf("esperava {aluno}, obtido %v", resultado.Perfis.Ordenado())
	}
}

func TestAtualizarUsuario_E16_ConjuntoVazioRecaiParaAluno(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Professor), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo, totalDetentores: 5}
	uc := NovoAtualizarUsuarioUseCase(repo, &auditMock{}, uowFake{})

	resultado, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, Nome: alvo.Nome, Email: alvo.Email,
		Perfis: nil, Versao: 1,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resultado.Perfis.Igual(mustConjuntoTeste(t, valueobject.Aluno)) {
		t.Fatalf("esperava {aluno} — conjunto vazio recai para aluno, obtido %v", resultado.Perfis.Ordenado())
	}
}

func TestAtualizarUsuario_E04_AlterarOsProprioPerfisNegada(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Aluno), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo}
	uc := NovoAtualizarUsuarioUseCase(repo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorComID(t, alvo.ID, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, Nome: alvo.Nome, Email: alvo.Email,
		Perfis: []valueobject.Perfil{valueobject.Professor}, Versao: 1,
	})
	if !errors.Is(err, domain.ErrAlteracaoDosPropriosPerfisNegada) {
		t.Fatalf("esperava ErrAlteracaoDosPropriosPerfisNegada, obtido %v", err)
	}
	if repo.atualizarChamado {
		t.Fatal("nada deveria ser gravado")
	}
}

func TestAtualizarUsuario_E03_VersaoDivergenteDevolveConflito(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	alvo := usuarioAlvoTeste(t, mustConjuntoTeste(t, valueobject.Aluno), &instituicaoID)
	repo := &usuarioRepoMock{usuarioParaBuscar: alvo, erroAtualizar: domain.ErrConflitoDeVersao}
	uc := NovoAtualizarUsuarioUseCase(repo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), AtualizarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		UsuarioID: alvo.ID, Nome: "Novo Nome", Email: alvo.Email,
		Perfis: []valueobject.Perfil{valueobject.Aluno}, Versao: 1,
	})
	if !errors.Is(err, domain.ErrConflitoDeVersao) {
		t.Fatalf("esperava ErrConflitoDeVersao, obtido %v", err)
	}
	if repo.substituirPerfisChamado {
		t.Fatal("SubstituirPerfis não deveria ter sido chamado após conflito de versão")
	}
}
