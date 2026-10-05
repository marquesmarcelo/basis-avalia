package indicador

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type indicadorRepoMock struct {
	itemParaBuscar         port.ItemIndicador
	erroBuscar             error
	inserirChamado         bool
	atualizarChamado       bool
	alterarSituacaoChamado bool
	excluirChamado         bool
	erroInserir            error
	erroAtualizar          error
	erroAlterarSituacao    error
	erroExcluir            error
}

func (m *indicadorRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemIndicador, error) {
	return m.itemParaBuscar, m.erroBuscar
}
func (m *indicadorRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarIndicadores) (port.ResultadoListaIndicadores, error) {
	return port.ResultadoListaIndicadores{}, nil
}
func (m *indicadorRepoMock) Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemSugestaoIndicador, error) {
	return nil, nil
}
func (m *indicadorRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error {
	m.inserirChamado = true
	return m.erroInserir
}
func (m *indicadorRepoMock) Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error {
	m.atualizarChamado = true
	return m.erroAtualizar
}
func (m *indicadorRepoMock) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
	m.alterarSituacaoChamado = true
	return m.erroAlterarSituacao
}
func (m *indicadorRepoMock) ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	m.excluirChamado = true
	return m.erroExcluir
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
