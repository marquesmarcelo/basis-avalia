package sessao

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type autenticacaoRepoMock struct {
	credencial *port.CredencialUsuario
	erro       error
}

func (m *autenticacaoRepoMock) BuscarCredencial(ctx context.Context, instituicaoID *uuid.UUID, email valueobject.Email) (*port.CredencialUsuario, error) {
	return m.credencial, m.erro
}
func (m *autenticacaoRepoMock) CarregarContextoDeSessao(ctx context.Context, usuarioID uuid.UUID, hoje valueobject.DataLocal) (*port.ContextoDeSessao, error) {
	return nil, nil
}
func (m *autenticacaoRepoMock) BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*port.CredencialUsuario, error) {
	return nil, nil
}
func (m *autenticacaoRepoMock) DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error {
	return nil
}
func (m *autenticacaoRepoMock) InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error {
	return nil
}

type hashMock struct {
	conferirResultado         bool
	conferirErro              error
	chamouConferir            bool
	chamouConferirDescartavel bool
}

func (m *hashMock) Gerar(ctx context.Context, senha valueobject.SenhaEmTexto) (valueobject.SenhaHash, error) {
	return valueobject.SenhaHash{}, nil
}
func (m *hashMock) Conferir(ctx context.Context, senha valueobject.SenhaEmTexto, hash valueobject.SenhaHash) (bool, error) {
	m.chamouConferir = true
	return m.conferirResultado, m.conferirErro
}
func (m *hashMock) ConferirDescartavel(ctx context.Context, senha valueobject.SenhaEmTexto) {
	m.chamouConferirDescartavel = true
}

type relogioMock struct{ agora time.Time }

func (r relogioMock) Agora() time.Time { return r.agora }

func emailTeste(t *testing.T, s string) valueobject.Email {
	t.Helper()
	e, err := valueobject.NovoEmail(s)
	if err != nil {
		t.Fatalf("email de teste inválido: %v", err)
	}
	return e
}

func senhaTeste(t *testing.T, s string) valueobject.SenhaEmTexto {
	t.Helper()
	senha, err := valueobject.NovaSenhaEmTexto(s)
	if err != nil {
		t.Fatalf("senha de teste inválida: %v", err)
	}
	return senha
}

func hashTeste(t *testing.T) valueobject.SenhaHash {
	t.Helper()
	h, err := valueobject.NovaSenhaHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA")
	if err != nil {
		t.Fatalf("hash de teste inválido: %v", err)
	}
	return h
}

func conjuntoTesteAuth(t *testing.T, p valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(p)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	return c
}

func TestAutenticar_L01_LoginBemSucedido(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: usuarioID, InstituicaoID: &instituicaoID, Nome: "Maria Souza",
		Email: emailTeste(t, "maria.souza@fsa.edu.br"), SenhaHash: hashTeste(t),
		Perfis: conjuntoTesteAuth(t, valueobject.PesquisadorInstitucional),
	}}
	hash := &hashMock{conferirResultado: true}
	agora := time.Now()
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: agora}, time.UTC)

	out, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &instituicaoID, Email: emailTeste(t, "maria.souza@fsa.edu.br"), Senha: senhaTeste(t, "reuniao-nde-2026"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Ator.UsuarioID() != usuarioID {
		t.Fatalf("ator com usuario_id incorreto")
	}
	if !out.InstanteDeAutenticacao.Equal(agora) {
		t.Fatalf("instante de autenticação incorreto")
	}
	if hash.chamouConferirDescartavel {
		t.Fatal("login bem-sucedido não deveria chamar ConferirDescartavel")
	}
}

func TestAutenticar_L02_EmailNaoCadastrado(t *testing.T) {
	repo := &autenticacaoRepoMock{credencial: nil}
	hash := &hashMock{}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	_, err := uc.Executar(context.Background(), AutenticarInput{
		Email: emailTeste(t, "ninguem@fsa.edu.br"), Senha: senhaTeste(t, "qualquercoisa"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
	if !hash.chamouConferirDescartavel {
		t.Fatal("esperava que ConferirDescartavel fosse chamado (L-08)")
	}
}

func TestAutenticar_L03_SenhaIncorreta(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: &instituicaoID,
		Email: emailTeste(t, "maria.souza@fsa.edu.br"), SenhaHash: hashTeste(t),
		Perfis: conjuntoTesteAuth(t, valueobject.PesquisadorInstitucional),
	}}
	hash := &hashMock{conferirResultado: false}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	_, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &instituicaoID, Email: emailTeste(t, "maria.souza@fsa.edu.br"), Senha: senhaTeste(t, "senha-errada"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
	if !hash.chamouConferir {
		t.Fatal("esperava que Conferir fosse chamado com a senha errada")
	}
}

func TestAutenticar_L04_UsuarioExcluidoNaoEntra(t *testing.T) {
	// BuscarCredencial já filtra excluido_em — para o use case, é idêntico
	// a "sem conta" (credencial nil).
	repo := &autenticacaoRepoMock{credencial: nil}
	hash := &hashMock{}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	_, err := uc.Executar(context.Background(), AutenticarInput{
		Email: emailTeste(t, "carlos.pereira@fsa.edu.br"), Senha: senhaTeste(t, "qualquer-coisa-antiga"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
	if !hash.chamouConferirDescartavel {
		t.Fatal("esperava ConferirDescartavel para usuário excluído (mesma mensagem genérica)")
	}
}

func TestAutenticar_L12_MesmoEmailEmDuasInstituicoesAutenticaContaEscolhida(t *testing.T) {
	fsaID := uuid.Must(uuid.NewV7())
	ivvID := uuid.Must(uuid.NewV7())
	joaoFSA := uuid.Must(uuid.NewV7())

	repoFSA := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: joaoFSA, InstituicaoID: &fsaID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"),
		SenhaHash: hashTeste(t), Perfis: conjuntoTesteAuth(t, valueobject.Professor),
	}}
	hashFSA := &hashMock{conferirResultado: true}
	ucFSA := NovoAutenticarUseCase(repoFSA, hashFSA, relogioMock{agora: time.Now()}, time.UTC)
	outFSA, err := ucFSA.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &fsaID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"), Senha: senhaTeste(t, "senha-fsa-2026"),
	})
	if err != nil {
		t.Fatalf("erro inesperado (FSA): %v", err)
	}
	if !outFSA.Ator.Perfis().Possui(valueobject.Professor) {
		t.Fatalf("esperava perfil professor na FSA")
	}

	joaoIVV := uuid.Must(uuid.NewV7())
	repoIVV := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: joaoIVV, InstituicaoID: &ivvID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"),
		SenhaHash: hashTeste(t), Perfis: conjuntoTesteAuth(t, valueobject.CoordenadorCurso),
	}}
	hashIVV := &hashMock{conferirResultado: true}
	ucIVV := NovoAutenticarUseCase(repoIVV, hashIVV, relogioMock{agora: time.Now()}, time.UTC)
	outIVV, err := ucIVV.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &ivvID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"), Senha: senhaTeste(t, "senha-ivv-2026"),
	})
	if err != nil {
		t.Fatalf("erro inesperado (IVV): %v", err)
	}
	if !outIVV.Ator.Perfis().Possui(valueobject.CoordenadorCurso) {
		t.Fatalf("esperava perfil coordenador_curso no IVV")
	}
	if outFSA.Ator.UsuarioID() == outIVV.Ator.UsuarioID() {
		t.Fatal("as duas contas deveriam ser independentes")
	}
}

func TestAutenticar_L13_SenhaDaOutraInstituicaoNaoServe(t *testing.T) {
	fsaID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: &fsaID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"),
		SenhaHash: hashTeste(t), Perfis: conjuntoTesteAuth(t, valueobject.Professor),
	}}
	hash := &hashMock{conferirResultado: false}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	_, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &fsaID, Email: emailTeste(t, "joao.ribeiro@ies.edu.br"), Senha: senhaTeste(t, "senha-ivv-2026"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
}

func TestAutenticar_L14_EmailQueExisteEmOutraInstituicaoNaoNaEscolhida(t *testing.T) {
	// BuscarCredencial filtra por instituicao_id — pedir a FSA quando só
	// existe conta no IVV devolve credencial nil.
	repo := &autenticacaoRepoMock{credencial: nil}
	hash := &hashMock{}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	fsaID := uuid.Must(uuid.NewV7())
	_, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &fsaID, Email: emailTeste(t, "renata.coimbra@ivv.edu.br"), Senha: senhaTeste(t, "nde-vale-verde-26"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
	if !hash.chamouConferirDescartavel {
		t.Fatal("esperava ConferirDescartavel")
	}
}

func TestAutenticar_L15_LoginEmInstituicaoInativada(t *testing.T) {
	// BuscarCredencial já exige situacao='ativa' — instituição inativa
	// produz credencial nil, idêntico a "sem conta".
	repo := &autenticacaoRepoMock{credencial: nil}
	hash := &hashMock{}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	fsaID := uuid.Must(uuid.NewV7())
	_, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &fsaID, Email: emailTeste(t, "maria.souza@fsa.edu.br"), Senha: senhaTeste(t, "reuniao-nde-2026"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
}

func TestAutenticar_L16_IdentificadorDeInstituicaoManipulado(t *testing.T) {
	repo := &autenticacaoRepoMock{credencial: nil}
	hash := &hashMock{}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	inexistente := uuid.Must(uuid.NewV7())
	_, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: &inexistente, Email: emailTeste(t, "qualquer@x.local"), Senha: senhaTeste(t, "qualquer"),
	})
	if !errors.Is(err, domain.ErrCredenciaisInvalidas) {
		t.Fatalf("esperava ErrCredenciaisInvalidas, obtido %v", err)
	}
}

func TestAutenticar_L17_LoginDoAdministradorDoSistema(t *testing.T) {
	repo := &autenticacaoRepoMock{credencial: &port.CredencialUsuario{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: nil, Email: emailTeste(t, "rafael.toledo@basis-avalia.local"),
		SenhaHash: hashTeste(t), Perfis: conjuntoTesteAuth(t, valueobject.AdministradorSistema),
	}}
	hash := &hashMock{conferirResultado: true}
	uc := NovoAutenticarUseCase(repo, hash, relogioMock{agora: time.Now()}, time.UTC)

	out, err := uc.Executar(context.Background(), AutenticarInput{
		InstituicaoID: nil, Email: emailTeste(t, "rafael.toledo@basis-avalia.local"), Senha: senhaTeste(t, "plataforma-2026"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Ator.InstituicaoID() != nil {
		t.Fatal("o ator do administrador não deveria ter instituicao_id")
	}
}
