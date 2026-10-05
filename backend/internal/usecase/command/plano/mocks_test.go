package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type planoRepoMock struct {
	detalheParaBuscar port.DetalhePlano
	erroBuscar        error

	inseridos     []*plano.Plano
	errosPorCurso map[uuid.UUID]error // simula a corrida/duplicidade de CP-05
	cursosAtivos  map[uuid.UUID]bool
	destinos      []port.DestinoDeCopia
}

func (m *planoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.DetalhePlano, error) {
	return m.detalheParaBuscar, m.erroBuscar
}
func (m *planoRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	return port.ResultadoListaPlanos{}, nil
}
func (m *planoRepoMock) ListarMeusPlanos(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	return port.ResultadoListaPlanos{}, nil
}
func (m *planoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano) error {
	m.inseridos = append(m.inseridos, p)
	return nil
}
func (m *planoRepoMock) AtualizarDados(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error {
	return nil
}
func (m *planoRepoMock) AtualizarSituacao(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error {
	return nil
}
func (m *planoRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *planoRepoMock) InstituicaoDaSessao(ctx context.Context, escopo autorizacao.Escopo) (string, string, error) {
	return "Faculdade São Aleixo", "FSA", nil
}
func (m *planoRepoMock) PeriodoDoPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (periodo.Periodo, error) {
	return periodo.Periodo{}, nil
}
func (m *planoRepoMock) CursoValidoParaPlano(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID) (bool, error) {
	if m.cursosAtivos == nil {
		return true, nil
	}
	return m.cursosAtivos[cursoID], nil
}
func (m *planoRepoMock) ExisteEntregaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *planoRepoMock) ExisteEntregaNoItem(ctx context.Context, escopo autorizacao.Escopo, itemID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *planoRepoMock) ListarDestinosDeCopia(ctx context.Context, escopo autorizacao.Escopo, planoOrigemID, periodoDestinoID uuid.UUID, busca string) ([]port.DestinoDeCopia, error) {
	return m.destinos, nil
}
func (m *planoRepoMock) InserirComItens(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, itensOrigem []port.ItemDoPlanoResponse) error {
	if err, ok := m.errosPorCurso[p.CursoID]; ok {
		return err
	}
	m.inseridos = append(m.inseridos, p)
	return nil
}

type periodoRepoMock struct {
	item port.ItemPeriodo
	erro error
}

func (m *periodoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemPeriodo, error) {
	return m.item, m.erro
}
func (m *periodoRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPeriodos) (port.ResultadoListaPeriodos, error) {
	return port.ResultadoListaPeriodos{}, nil
}
func (m *periodoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo) error {
	return nil
}
func (m *periodoRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo, versaoEsperada int) error {
	return nil
}
func (m *periodoRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *periodoRepoMock) TemPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (bool, error) {
	return false, nil
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

// atorPI fixa a data de referência em 15/03/2026 (mesmo "hoje" dos
// exemplos concretos de spec.md) — sem isso, a data de referência fica no
// valor zero e qualquer período parece "não iniciado", nunca "encerrado".
func atorPI(instituicaoID uuid.UUID) autorizacao.Ator {
	conjunto, _ := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	return ator.ComDataDeReferencia(hoje)
}
