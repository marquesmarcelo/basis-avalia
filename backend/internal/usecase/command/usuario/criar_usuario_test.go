package usuario

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type usuarioRepoMock struct {
	inserirChamado            bool
	instituicaoInserida       *uuid.UUID
	perfisInseridos           valueobject.ConjuntoDePerfis
	erroInserir               error
	usuarioParaBuscar         *usuario.Usuario
	erroBuscar                error
	atualizarChamado          bool
	versaoRecebida            int
	erroAtualizar             error
	substituirPerfisChamado   bool
	perfisSubstituidos        valueobject.ConjuntoDePerfis
	totalDetentores           int
	erroContarDetentores      error
	travarPopulacaoChamado    bool
	excluirLogicamenteChamado bool
	definirSenhaChamado       bool
	senhaProvisoriaRecebida   bool
}

func (m *usuarioRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (*usuario.Usuario, error) {
	return m.usuarioParaBuscar, m.erroBuscar
}
func (m *usuarioRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarUsuarios) (port.ResultadoListaUsuarios, error) {
	return port.ResultadoListaUsuarios{}, nil
}
func (m *usuarioRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario) error {
	m.inserirChamado = true
	m.instituicaoInserida = u.InstituicaoID
	m.perfisInseridos = u.Perfis
	return m.erroInserir
}
func (m *usuarioRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario, versaoEsperada int) error {
	m.atualizarChamado = true
	m.versaoRecebida = versaoEsperada
	return m.erroAtualizar
}
func (m *usuarioRepoMock) SubstituirPerfis(ctx context.Context, escopo autorizacao.Escopo, usuarioID uuid.UUID, perfis valueobject.ConjuntoDePerfis) error {
	m.substituirPerfisChamado = true
	m.perfisSubstituidos = perfis
	return nil
}
func (m *usuarioRepoMock) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	m.excluirLogicamenteChamado = true
	return nil
}
func (m *usuarioRepoMock) DefinirSenha(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, hash valueobject.SenhaHash, senhaProvisoria bool, instante time.Time) error {
	m.definirSenhaChamado = true
	m.senhaProvisoriaRecebida = senhaProvisoria
	return nil
}
func (m *usuarioRepoMock) InvalidarSessoes(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	return nil
}
func (m *usuarioRepoMock) ContarDetentoresDoPerfil(ctx context.Context, escopo autorizacao.Escopo, perfil valueobject.Perfil) (int, error) {
	return m.totalDetentores, m.erroContarDetentores
}
func (m *usuarioRepoMock) TravarPopulacao(ctx context.Context, escopo autorizacao.Escopo) error {
	m.travarPopulacaoChamado = true
	return nil
}

type hashMock struct{}

func (hashMock) Gerar(ctx context.Context, senha valueobject.SenhaEmTexto) (valueobject.SenhaHash, error) {
	return valueobject.NovaSenhaHash("$argon2id$v=19$m=1,t=1,p=1$c2FsdA$aGFzaA")
}
func (hashMock) Conferir(ctx context.Context, senha valueobject.SenhaEmTexto, hash valueobject.SenhaHash) (bool, error) {
	return true, nil
}
func (hashMock) ConferirDescartavel(ctx context.Context, senha valueobject.SenhaEmTexto) {}

type auditMock struct{ eventos []auditoria.Evento }

func (m *auditMock) Registrar(ctx context.Context, e auditoria.Evento) error {
	m.eventos = append(m.eventos, e)
	return nil
}

type uowFake struct{}

func (uowFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func atorPI(t *testing.T, instituicaoID uuid.UUID) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator
}

func emailTeste(t *testing.T, s string) valueobject.Email {
	t.Helper()
	e, err := valueobject.NovoEmail(s)
	if err != nil {
		t.Fatalf("email: %v", err)
	}
	return e
}

func senhaTeste(t *testing.T, s string) valueobject.SenhaEmTexto {
	t.Helper()
	senha, err := valueobject.NovaSenhaEmTexto(s)
	if err != nil {
		t.Fatalf("senha: %v", err)
	}
	return senha
}

func TestCriarUsuario_U11_InstituicaoVemDaSessaoNuncaDoPayload(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		Nome: "João Ribeiro", Email: emailTeste(t, "joao.ribeiro@ies.edu.br"), Perfis: []valueobject.Perfil{valueobject.Professor},
		Senha: senhaTeste(t, "primeiro-acesso-2026"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if repo.instituicaoInserida == nil || *repo.instituicaoInserida != instituicaoID {
		t.Fatalf("instituição inserida deveria ser a da sessão, obtido %v", repo.instituicaoInserida)
	}
}

func TestCriarUsuario_U10_PINaoAtribuiAdministradorSistema(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		Nome: "X", Email: emailTeste(t, "x@ies.edu.br"), Perfis: []valueobject.Perfil{valueobject.AdministradorSistema}, Senha: senhaTeste(t, "qualquer"),
	})
	if !errors.Is(err, domain.ErrPerfilNaoAtribuivel) {
		t.Fatalf("esperava ErrPerfilNaoAtribuivel, obtido %v", err)
	}
	if repo.inserirChamado {
		t.Fatal("nada deveria ser inserido")
	}
}

func TestCriarUsuario_AS01_PerfilImpostoPeloAlcanceNaoLidoDoPayload(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}

	_, err = uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: ator, Alcance: autorizacao.PesquisadoresDeUmaInstituicao, InstituicaoDoCaminho: &instituicaoID,
		Nome: "Renata Coimbra", Email: emailTeste(t, "renata.coimbra@ivv.edu.br"), Senha: senhaTeste(t, "nde-vale-verde-26"),
		// Perfis não informado — o alcance é quem decide (AS-01).
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !repo.perfisInseridos.Igual(mustConjuntoTeste(t, valueobject.PesquisadorInstitucional)) {
		t.Fatalf("perfis deveriam ser {pesquisador_institucional}, obtido %v", repo.perfisInseridos.Ordenado())
	}
	if repo.instituicaoInserida == nil || *repo.instituicaoInserida != instituicaoID {
		t.Fatal("instituição deveria vir do caminho")
	}
}

func TestCriarUsuario_U12_SemPerfisNoPayloadGravaAluno(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	novo, err := uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		Nome: "Letícia Moraes", Email: emailTeste(t, "leticia.moraes@fsa.edu.br"), Senha: senhaTeste(t, "primeiro-acesso-26"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !novo.Perfis.Igual(mustConjuntoTeste(t, valueobject.Aluno)) {
		t.Fatalf("esperava {aluno}, obtido %v", novo.Perfis.Ordenado())
	}
	if !repo.perfisInseridos.Igual(mustConjuntoTeste(t, valueobject.Aluno)) {
		t.Fatalf("perfis gravados deveriam ser {aluno}, obtido %v", repo.perfisInseridos.Ordenado())
	}
}

func TestCriarUsuario_U13_ComProfessorNaoAcrescentaAluno(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	novo, err := uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		Nome: "Rafael Toledo", Email: emailTeste(t, "rafael.toledo@fsa.edu.br"), Perfis: []valueobject.Perfil{valueobject.Professor},
		Senha: senhaTeste(t, "primeiro-acesso-26"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !novo.Perfis.Igual(mustConjuntoTeste(t, valueobject.Professor)) {
		t.Fatalf("esperava {professor} sem aluno acrescentado, obtido %v", novo.Perfis.Ordenado())
	}
}

func TestCriarUsuario_U07_PerfilInexistenteDevolveErro(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoMock{}
	uc := NovoCriarUsuarioUseCase(repo, hashMock{}, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarUsuarioInput{
		Ator: atorPI(t, instituicaoID), Alcance: autorizacao.UsuariosDaPropriaInstituicao,
		Nome: "X", Email: emailTeste(t, "x@fsa.edu.br"), Perfis: []valueobject.Perfil{"bibliotecario"},
		Senha: senhaTeste(t, "primeiro-acesso-26"),
	})
	if !errors.Is(err, domain.ErrPerfilInvalido) {
		t.Fatalf("esperava ErrPerfilInvalido, obtido %v", err)
	}
	if repo.inserirChamado {
		t.Fatal("nada deveria ser inserido")
	}
}

func mustConjuntoTeste(t *testing.T, p ...valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(p...)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	return c
}
