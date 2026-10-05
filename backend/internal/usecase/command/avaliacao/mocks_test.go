package avaliacao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	entregadomain "github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type entregaRepoMock struct {
	detalhe    port.DetalheEntrega
	erroBuscar error
	coordena   bool
	avaliarOK  bool
	desfazerOK bool
	avaliadas  []*entregadomain.Entrega
	desfeitas  []*entregadomain.Entrega
}

func (m *entregaRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, avaliador autorizacao.Proprio) (port.DetalheEntrega, error) {
	return m.detalhe, m.erroBuscar
}
func (m *entregaRepoMock) ListarDoItem(ctx context.Context, escopo autorizacao.Escopo, itemPlanoID uuid.UUID, filtro port.FiltroListarEntregas) (port.ResultadoListaEntregas, error) {
	return port.ResultadoListaEntregas{}, nil
}
func (m *entregaRepoMock) CoordenaCursoHoje(ctx context.Context, proprio autorizacao.Proprio, cursoID uuid.UUID, dataDeReferencia valueobject.DataLocal) (bool, error) {
	return m.coordena, nil
}
func (m *entregaRepoMock) ListarPeriodosParaMinhasMetas(ctx context.Context, escopo autorizacao.Escopo) ([]port.PeriodoOpcao, error) {
	return nil, nil
}
func (m *entregaRepoMock) InserirIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string, recursoID uuid.UUID) error {
	return nil
}
func (m *entregaRepoMock) InserirEntregaComAnexos(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, anexos []port.AnexoParaGravar) error {
	return nil
}
func (m *entregaRepoMock) BuscarEntregaPorChaveIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string) (uuid.UUID, error) {
	return uuid.UUID{}, nil
}
func (m *entregaRepoMock) AtualizarCorrecao(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) error {
	return nil
}
func (m *entregaRepoMock) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *entregaRepoMock) MarcarPendenciaVista(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *entregaRepoMock) Avaliar(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) (bool, error) {
	m.avaliadas = append(m.avaliadas, e)
	return m.avaliarOK, nil
}
func (m *entregaRepoMock) DesfazerAceitacao(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) (bool, error) {
	m.desfeitas = append(m.desfeitas, e)
	return m.desfazerOK, nil
}
func (m *entregaRepoMock) AdicionarAnexos(ctx context.Context, escopo autorizacao.Escopo, entregaID uuid.UUID, anexos []port.AnexoParaGravar) ([]port.AnexoResponse, error) {
	return nil, nil
}
func (m *entregaRepoMock) RemoverAnexo(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID, autor autorizacao.Proprio) (string, error) {
	return "", nil
}
func (m *entregaRepoMock) BuscarAnexoParaDownload(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID) (string, string, string, error) {
	return "", "", "", nil
}
func (m *entregaRepoMock) MinhasMetas(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) ([]port.GrupoMinhasMetas, error) {
	return nil, nil
}
func (m *entregaRepoMock) ListarFilaDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo, avaliador autorizacao.Proprio, filtro port.FiltroFilaAvaliacao) (port.ResultadoListaEntregas, error) {
	return port.ResultadoListaEntregas{}, nil
}
func (m *entregaRepoMock) ContarPendentesDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo) (int, error) {
	return 0, nil
}
func (m *entregaRepoMock) ContarPendenciasNaoVistas(ctx context.Context, escopo autorizacao.Escopo) (int, error) {
	return 0, nil
}
func (m *entregaRepoMock) RelatorioDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio) (port.ResultadoRelatorio, error) {
	return port.ResultadoRelatorio{}, nil
}
func (m *entregaRepoMock) ExportarDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio, linha func(port.LinhaCSVRelatorio) error) (int, error) {
	return 0, nil
}
func (m *entregaRepoMock) DesempenhoPorCurso(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio) (port.ResultadoDesempenhoPorCurso, error) {
	return port.ResultadoDesempenhoPorCurso{}, nil
}

var _ port.EntregaRepository = (*entregaRepoMock)(nil)

type auditMock struct{ eventos []auditoria.Evento }

func (m *auditMock) Registrar(ctx context.Context, e auditoria.Evento) error {
	m.eventos = append(m.eventos, e)
	return nil
}

type uowFake struct{}

func (uowFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type relogioFixo struct{ instante time.Time }

func (r relogioFixo) Agora() time.Time { return r.instante }

func fusoSaoPaulo() *time.Location {
	l, _ := time.LoadLocation("America/Sao_Paulo")
	return l
}

func atorPI(instituicaoID uuid.UUID) autorizacao.Ator {
	conjunto, _ := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	return ator.ComDataDeReferencia(hoje)
}
