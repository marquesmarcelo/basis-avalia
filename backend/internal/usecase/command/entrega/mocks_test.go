package entrega

import (
	"bytes"
	"context"
	"io"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	entregadomain "github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/itemplano"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type entregaRepoMock struct {
	detalhe    port.DetalheEntrega
	erroBuscar error

	inseridos         []*entregadomain.Entrega
	anexosInseridos   [][]port.AnexoParaGravar
	idempotenciaVista map[string]uuid.UUID
	erroIdempotencia  error

	corrigidas []*entregadomain.Entrega
	excluidas  []uuid.UUID
}

func (m *entregaRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, avaliador autorizacao.Proprio) (port.DetalheEntrega, error) {
	return m.detalhe, m.erroBuscar
}
func (m *entregaRepoMock) ListarDoItem(ctx context.Context, escopo autorizacao.Escopo, itemPlanoID uuid.UUID, filtro port.FiltroListarEntregas) (port.ResultadoListaEntregas, error) {
	return port.ResultadoListaEntregas{}, nil
}
func (m *entregaRepoMock) CoordenaCursoHoje(ctx context.Context, proprio autorizacao.Proprio, cursoID uuid.UUID, dataDeReferencia valueobject.DataLocal) (bool, error) {
	return false, nil
}
func (m *entregaRepoMock) ListarPeriodosParaMinhasMetas(ctx context.Context, escopo autorizacao.Escopo) ([]port.PeriodoOpcao, error) {
	return nil, nil
}
func (m *entregaRepoMock) InserirIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string, recursoID uuid.UUID) error {
	if m.erroIdempotencia != nil {
		return m.erroIdempotencia
	}
	if m.idempotenciaVista == nil {
		m.idempotenciaVista = map[string]uuid.UUID{}
	}
	if _, ok := m.idempotenciaVista[chave]; ok {
		return domain.ErrChaveDuplicada
	}
	m.idempotenciaVista[chave] = recursoID
	return nil
}
func (m *entregaRepoMock) InserirEntregaComAnexos(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, anexos []port.AnexoParaGravar) error {
	m.inseridos = append(m.inseridos, e)
	m.anexosInseridos = append(m.anexosInseridos, anexos)
	return nil
}
func (m *entregaRepoMock) BuscarEntregaPorChaveIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string) (uuid.UUID, error) {
	if id, ok := m.idempotenciaVista[chave]; ok {
		return id, nil
	}
	return uuid.UUID{}, domain.ErrNaoEncontrado
}
func (m *entregaRepoMock) AtualizarCorrecao(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) error {
	m.corrigidas = append(m.corrigidas, e)
	return nil
}
func (m *entregaRepoMock) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	m.excluidas = append(m.excluidas, id)
	return nil
}
func (m *entregaRepoMock) MarcarPendenciaVista(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *entregaRepoMock) Avaliar(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) (bool, error) {
	return true, nil
}
func (m *entregaRepoMock) DesfazerAceitacao(ctx context.Context, escopo autorizacao.Escopo, e *entregadomain.Entrega, versaoEsperada int) (bool, error) {
	return true, nil
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

type itemPlanoRepoMock struct {
	item itemplano.ItemDoPlano
	erro error
}

func (m *itemPlanoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (itemplano.ItemDoPlano, error) {
	return m.item, m.erro
}
func (m *itemPlanoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano) error {
	return nil
}
func (m *itemPlanoRepoMock) AtualizarQuantidade(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano, versaoEsperada int) error {
	return nil
}
func (m *itemPlanoRepoMock) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	return nil
}
func (m *itemPlanoRepoMock) ExisteMetaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID, metaID uuid.UUID) (bool, error) {
	return false, nil
}

type planoRepoMock struct {
	detalhe     port.DetalhePlano
	erroBuscar  error
	cursoAtivo  bool
	periodo     periodo.Periodo
	erroPeriodo error
}

func (m *planoRepoMock) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.DetalhePlano, error) {
	return m.detalhe, m.erroBuscar
}
func (m *planoRepoMock) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	return port.ResultadoListaPlanos{}, nil
}
func (m *planoRepoMock) ListarMeusPlanos(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	return port.ResultadoListaPlanos{}, nil
}
func (m *planoRepoMock) Inserir(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano) error {
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
	return m.periodo, m.erroPeriodo
}
func (m *planoRepoMock) CursoValidoParaPlano(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID) (bool, error) {
	return m.cursoAtivo, nil
}
func (m *planoRepoMock) ExisteEntregaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *planoRepoMock) ExisteEntregaNoItem(ctx context.Context, escopo autorizacao.Escopo, itemID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *planoRepoMock) ListarDestinosDeCopia(ctx context.Context, escopo autorizacao.Escopo, planoOrigemID, periodoDestinoID uuid.UUID, busca string) ([]port.DestinoDeCopia, error) {
	return nil, nil
}
func (m *planoRepoMock) InserirComItens(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, itensOrigem []port.ItemDoPlanoResponse) error {
	return nil
}

type armazenamentoMock struct {
	gravados   []string
	removidos  []string
	erroGravar error
}

func (m *armazenamentoMock) Gravar(ctx context.Context, chave string, conteudo io.Reader, tipo string, tamanho int64) error {
	if m.erroGravar != nil {
		return m.erroGravar
	}
	m.gravados = append(m.gravados, chave)
	return nil
}
func (m *armazenamentoMock) Ler(ctx context.Context, chave string) (io.ReadCloser, string, error) {
	return io.NopCloser(bytes.NewReader(nil)), "", nil
}
func (m *armazenamentoMock) Remover(ctx context.Context, chave string) error {
	m.removidos = append(m.removidos, chave)
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

func atorCoordenador(instituicaoID uuid.UUID) autorizacao.Ator {
	conjunto, _ := valueobject.NovoConjunto(valueobject.CoordenadorCurso)
	ator, _ := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	return ator.ComDataDeReferencia(hoje)
}

func periodoAberto() periodo.Periodo {
	inicio, _ := valueobject.DataLocalTexto("2026-01-01")
	fim, _ := valueobject.DataLocalTexto("2026-07-30")
	vigencia, _ := valueobject.NovaVigencia(inicio, &fim)
	return periodo.Periodo{Vigencia: vigencia}
}
