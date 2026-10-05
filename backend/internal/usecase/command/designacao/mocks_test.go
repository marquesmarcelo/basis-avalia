package designacao

import (
	"context"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/curso"
	designacaodomain "github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type designacaoRepoMock struct {
	itemParaBuscar       port.ItemDesignacao
	erroBuscar           error
	inserirChamado       bool
	erroInserir          error
	atualizarCompChamado bool
	erroAtualizarComp    error
	atualizarFimChamado  bool
	erroAtualizarFim     error
	excluirChamado       bool
	erroExcluir          error
}

func (m *designacaoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemDesignacao, error) {
	return m.itemParaBuscar, m.erroBuscar
}
func (m *designacaoRepoMock) ListarDoCurso(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID, filtro port.FiltroListarDesignacoes) (port.ResultadoListaDesignacoes, error) {
	return port.ResultadoListaDesignacoes{}, nil
}
func (m *designacaoRepoMock) ListarCandidatos(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemCandidato, error) {
	return nil, nil
}
func (m *designacaoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, d *designacaodomain.Designacao) error {
	m.inserirChamado = true
	return m.erroInserir
}
func (m *designacaoRepoMock) AtualizarCompleta(ctx context.Context, escopo autorizacao.Escopo, d *designacaodomain.Designacao, versaoEsperada int) error {
	m.atualizarCompChamado = true
	return m.erroAtualizarComp
}
func (m *designacaoRepoMock) AtualizarFimEPortaria(ctx context.Context, escopo autorizacao.Escopo, d *designacaodomain.Designacao, versaoEsperada int) error {
	m.atualizarFimChamado = true
	return m.erroAtualizarFim
}
func (m *designacaoRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	m.excluirChamado = true
	return m.erroExcluir
}

type cursoRepoMock struct {
	itemParaBuscar port.ItemCurso
	erroBuscar     error
}

func (m *cursoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemCurso, error) {
	return m.itemParaBuscar, m.erroBuscar
}
func (m *cursoRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarCursos) (port.ResultadoListaCursos, error) {
	return port.ResultadoListaCursos{}, nil
}
func (m *cursoRepoMock) ListarMeusCursos(ctx context.Context, escopo autorizacao.Escopo) ([]curso.Curso, error) {
	return nil, nil
}
func (m *cursoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso) error {
	return nil
}
func (m *cursoRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso, versaoEsperada int) error {
	return nil
}
func (m *cursoRepoMock) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCurso, versaoEsperada int) error {
	return nil
}
func (m *cursoRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *cursoRepoMock) TravarSeAtivo(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return m.erroBuscar
}

type usuarioRepoMock struct {
	itemParaBuscar *usuario.Usuario
	erroBuscar     error
}

func (m *usuarioRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (*usuario.Usuario, error) {
	return m.itemParaBuscar, m.erroBuscar
}
func (m *usuarioRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarUsuarios) (port.ResultadoListaUsuarios, error) {
	return port.ResultadoListaUsuarios{}, nil
}
func (m *usuarioRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario) error {
	return nil
}
func (m *usuarioRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario, versaoEsperada int) error {
	return nil
}
func (m *usuarioRepoMock) SubstituirPerfis(ctx context.Context, escopo autorizacao.Escopo, usuarioID uuid.UUID, perfis valueobject.ConjuntoDePerfis) error {
	return nil
}
func (m *usuarioRepoMock) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	return nil
}
func (m *usuarioRepoMock) DefinirSenha(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, hash valueobject.SenhaHash, senhaProvisoria bool, instante time.Time) error {
	return nil
}
func (m *usuarioRepoMock) InvalidarSessoes(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	return nil
}
func (m *usuarioRepoMock) ContarDetentoresDoPerfil(ctx context.Context, escopo autorizacao.Escopo, perfil valueobject.Perfil) (int, error) {
	return 0, nil
}
func (m *usuarioRepoMock) TravarPopulacao(ctx context.Context, escopo autorizacao.Escopo) error {
	return nil
}

type auditMock struct{ eventos []auditoria.Evento }

func (m *auditMock) Registrar(ctx context.Context, e auditoria.Evento) error {
	m.eventos = append(m.eventos, e)
	return nil
}

type uowFake struct{}

func (uowFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func dataTeste(t *testing.T, iso string) valueobject.DataLocal {
	t.Helper()
	d, err := valueobject.DataLocalTexto(iso)
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	return d
}

func atorPIComData(t *testing.T, instituicaoID uuid.UUID, hoje valueobject.DataLocal) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator.ComDataDeReferencia(hoje)
}
