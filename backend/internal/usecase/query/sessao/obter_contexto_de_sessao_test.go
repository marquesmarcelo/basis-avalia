package sessao

import (
	"context"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// permissoesEsperadasPI — união exata de matrizPermissoes[PesquisadorInstitucional]
// (perfil.go), reaproveitada pelos dois testes abaixo para não divergir em
// silêncio se a matriz mudar sem os testes acompanharem.
var permissoesEsperadasPI = map[string]bool{
	"usuario.listar": true, "usuario.criar": true, "usuario.editar": true,
	"usuario.excluir": true, "usuario.redefinir_senha": true,
	"indicador.listar": true, "indicador.gerenciar": true,
	"meta.listar": true, "meta.gerenciar": true,
	"curso.listar": true, "curso.gerenciar": true,
	"designacao.gerenciar": true, "periodo.gerenciar": true,
	"plano.listar": true, "plano.gerenciar": true,
	"entrega.listar": true, "entrega.avaliar": true,
	"relatorio.ler": true, "relatorio.exportar": true,
}

type autenticacaoRepoContextoMock struct {
	ctxSessao *port.ContextoDeSessao
	erro      error
}

func (m *autenticacaoRepoContextoMock) BuscarCredencial(ctx context.Context, instituicaoID *uuid.UUID, email valueobject.Email) (*port.CredencialUsuario, error) {
	return nil, nil
}
func (m *autenticacaoRepoContextoMock) CarregarContextoDeSessao(ctx context.Context, usuarioID uuid.UUID, hoje valueobject.DataLocal) (*port.ContextoDeSessao, error) {
	return m.ctxSessao, m.erro
}
func (m *autenticacaoRepoContextoMock) BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*port.CredencialUsuario, error) {
	return nil, nil
}
func (m *autenticacaoRepoContextoMock) DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error {
	return nil
}
func (m *autenticacaoRepoContextoMock) InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error {
	return nil
}

var hojeTeste = valueobject.DataLocalDe(time.Now(), time.UTC)

func emailTeste(t *testing.T, s string) valueobject.Email {
	t.Helper()
	e, err := valueobject.NovoEmail(s)
	if err != nil {
		t.Fatalf("email inválido: %v", err)
	}
	return e
}

func conjuntoTeste(t *testing.T, p ...valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(p...)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	return c
}

func TestObterContextoDeSessao_SH02_PerfilInstitucionalDevolveInstituicaoDoBanco(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoContextoMock{ctxSessao: &port.ContextoDeSessao{
		UsuarioID: usuarioID, Nome: "Maria Souza", Email: emailTeste(t, "maria.souza@fsa.edu.br"),
		Perfis: conjuntoTeste(t, valueobject.PesquisadorInstitucional), InstituicaoID: &instituicaoID,
		InstituicaoNome: "Faculdade Serra Azul", InstituicaoSigla: "FSA",
	}}
	uc := NovoObterContextoDeSessaoUseCase(repo)

	out, err := uc.Executar(context.Background(), usuarioID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Instituicao == nil {
		t.Fatal("esperava instituição preenchida")
	}
	if out.Instituicao.Nome != "Faculdade Serra Azul" || out.Instituicao.Sigla != "FSA" {
		t.Fatalf("instituição incorreta: %+v", out.Instituicao)
	}
}

func TestObterContextoDeSessao_AS05_AdministradorDevolveInstituicaoNula(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoContextoMock{ctxSessao: &port.ContextoDeSessao{
		UsuarioID: usuarioID, Nome: "Rafael Toledo", Email: emailTeste(t, "rafael.toledo@basis-avalia.local"),
		Perfis: conjuntoTeste(t, valueobject.AdministradorSistema), InstituicaoID: nil,
	}}
	uc := NovoObterContextoDeSessaoUseCase(repo)

	out, err := uc.Executar(context.Background(), usuarioID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Instituicao != nil {
		t.Fatal("administrador não deveria ter instituição")
	}
}

func TestObterContextoDeSessao_PermissoesRefletemAMatriz39(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoContextoMock{ctxSessao: &port.ContextoDeSessao{
		UsuarioID: usuarioID, Email: emailTeste(t, "maria.souza@fsa.edu.br"),
		Perfis: conjuntoTeste(t, valueobject.PesquisadorInstitucional), InstituicaoID: &instituicaoID,
	}}
	uc := NovoObterContextoDeSessaoUseCase(repo)

	out, err := uc.Executar(context.Background(), usuarioID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	esperadas := permissoesEsperadasPI
	if len(out.Permissoes) != len(esperadas) {
		t.Fatalf("esperava %d permissões, obtido %d: %v", len(esperadas), len(out.Permissoes), out.Permissoes)
	}
	for _, p := range out.Permissoes {
		if !esperadas[string(p)] {
			t.Fatalf("permissão inesperada para PI: %s", p)
		}
	}
}

// TestObterContextoDeSessao_A07_UniaoDePermissoesDeDoisPerfis é o marco de
// T-088: Beatriz tem {professor, pesquisador_institucional} — a resposta
// devolve os dois perfis e a união das permissões dos dois (design.md
// §6.4), não a de um só.
func TestObterContextoDeSessao_A07_UniaoDePermissoesDeDoisPerfis(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	instituicaoID := uuid.Must(uuid.NewV7())
	repo := &autenticacaoRepoContextoMock{ctxSessao: &port.ContextoDeSessao{
		UsuarioID: usuarioID, Nome: "Beatriz Andrade", Email: emailTeste(t, "beatriz.andrade@fsa.edu.br"),
		Perfis:          conjuntoTeste(t, valueobject.Professor, valueobject.PesquisadorInstitucional),
		InstituicaoID:   &instituicaoID,
		InstituicaoNome: "Faculdade Serra Azul", InstituicaoSigla: "FSA",
	}}
	uc := NovoObterContextoDeSessaoUseCase(repo)

	out, err := uc.Executar(context.Background(), usuarioID, hojeTeste)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(out.Perfis) != 2 {
		t.Fatalf("esperava os dois perfis na resposta, obtido %v", out.Perfis)
	}
	esperadas := permissoesEsperadasPI
	if len(out.Permissoes) != len(esperadas) {
		t.Fatalf("esperava a união das permissões de PI, obtido %d: %v", len(out.Permissoes), out.Permissoes)
	}
}
