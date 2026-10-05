package meta

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type metaRepoMock struct {
	itemParaBuscar   port.ItemMeta
	erroBuscar       error
	inserirChamado   bool
	atualizarChamado bool
	erroInserir      error
	erroAtualizar    error
}

func (m *metaRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemMeta, error) {
	return m.itemParaBuscar, m.erroBuscar
}
func (m *metaRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarMetas) (port.ResultadoListaMetas, error) {
	return port.ResultadoListaMetas{}, nil
}
func (m *metaRepoMock) Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemSugestaoMeta, error) {
	return nil, nil
}
func (m *metaRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, mt *meta.Meta) error {
	m.inserirChamado = true
	return m.erroInserir
}
func (m *metaRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, mt *meta.Meta, versaoEsperada int) error {
	m.atualizarChamado = true
	return m.erroAtualizar
}
func (m *metaRepoMock) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
	return nil
}
func (m *metaRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}

// indicadorRepoPorIDMock devolve, por id, um item pré-cadastrado — ou
// domain.ErrNaoEncontrado quando o id não está no mapa (simula 404 por
// escopo, ex: indicador de outra instituição).
type indicadorRepoPorIDMock struct {
	itens map[uuid.UUID]port.ItemIndicador
	erros map[uuid.UUID]error
}

func (m *indicadorRepoPorIDMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemIndicador, error) {
	if err, ok := m.erros[id]; ok {
		return port.ItemIndicador{}, err
	}
	item, ok := m.itens[id]
	if !ok {
		return port.ItemIndicador{}, domain.ErrNaoEncontrado
	}
	return item, nil
}
func (m *indicadorRepoPorIDMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarIndicadores) (port.ResultadoListaIndicadores, error) {
	return port.ResultadoListaIndicadores{}, nil
}
func (m *indicadorRepoPorIDMock) Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemSugestaoIndicador, error) {
	return nil, nil
}
func (m *indicadorRepoPorIDMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error {
	return nil
}
func (m *indicadorRepoPorIDMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error {
	return nil
}
func (m *indicadorRepoPorIDMock) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
	return nil
}
func (m *indicadorRepoPorIDMock) ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
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
