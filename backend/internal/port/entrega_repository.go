package port

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// AnexoResponse é a projeção de um anexo para a API — nunca inclui a
// chave do objeto nem qualquer URL do armazenamento (fundacao-metas.md §9,
// restrição inegociável).
type AnexoResponse struct {
	ID           uuid.UUID
	NomeOriginal string
	Tipo         string
	TamanhoBytes int64
	HashSHA256   string
	CriadoEm     time.Time
}

// LinhaEntrega é a projeção usada nas listagens (item, fila de avaliação,
// minhas metas).
type LinhaEntrega struct {
	ID                      uuid.UUID
	ItemPlanoID             uuid.UUID
	CursoID                 uuid.UUID
	CursoNome               string
	MetaID                  uuid.UUID
	MetaNome                string
	Situacao                string
	Rodadas                 int
	PrazoCorrecao           *time.Time
	EnviadaPorID            uuid.UUID
	EnviadaPorNome          string
	CorrigidaPorNome        *string
	Observacao              string
	Motivo                  string
	AvaliadaPorNome         *string
	AvaliadaEm              *time.Time
	AvaliadorEraCoordenador bool
	PendenciaVistaEm        *time.Time
	CoordenadoPeloAvaliador bool
	Anexos                  []AnexoResponse
	CriadoEm                time.Time
	AtualizadoEm            *time.Time
	Versao                  int
}

// DetalheEntrega é o que a página de registrar/corrigir/avaliar carrega.
type DetalheEntrega struct {
	LinhaEntrega
	Quantidade  int
	Indicadores []IndicadorDaMeta
}

type FiltroListarEntregas struct {
	Situacao string // "" | "pendente_avaliacao" | "aceita" | "recusada"
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ResultadoListaEntregas struct {
	Itens []LinhaEntrega
	Total int
}

type FiltroFilaAvaliacao struct {
	PeriodoID *uuid.UUID
	CursoID   *uuid.UUID
	MetaID    *uuid.UUID
	Situacao  string
	Page      int
	PageSize  int
	Sort      string
	Order     string
}

// GrupoMinhasMetas — um curso e os itens de plano vigente/encerrado dele,
// para a única tela do produto que abre sem "Pesquisar" (3.11).
type GrupoMinhasMetas struct {
	CursoID   uuid.UUID
	CursoNome string
	Itens     []ItemMinhasMetas
}

type ItemMinhasMetas struct {
	ItemPlanoID       uuid.UUID
	MetaNome          string
	Indicadores       []IndicadorDaMeta
	Quantidade        int
	Aceitas           int
	Pendentes         int
	EmCorrecao        int
	TemDesfazimento   bool
	UltimoDesfeitoTxt string // texto pronto do aviso (AV-12), vazio se não houver
}

// PendenciasResposta — os dois badges do menu (fundacao-metas.md, design.md
// §11): pendentes_de_avaliacao é o tamanho da fila do PI; pendencias_nao_vistas
// é literalmente "não vistas" do coordenador.
type PendenciasResposta struct {
	PendentesDeAvaliacao int
	PendenciasNaoVistas  int
}

// AnexoParaGravar — o que o handler já validou (tipo pelo conteúdo,
// tamanho) antes de qualquer chamada ao repositório.
type AnexoParaGravar struct {
	NomeOriginal string
	Tipo         string
	TamanhoBytes int64
	ChaveObjeto  string
	HashSHA256   string
}

// FiltroRelatorio — specs/metas-coordenacao/design.md §9, §11.
type FiltroRelatorio struct {
	PeriodoID       uuid.UUID
	CursoID         *uuid.UUID
	MetaID          *uuid.UUID
	IndicadorID     *uuid.UUID
	Origem          string // "" | "inep" | "instituicao"
	Situacao        string
	ResponsavelID   *uuid.UUID
	Autoavaliado    *bool
	IncluirInativos bool
	Page            int
	PageSize        int
	Sort            string
	Order           string
}

type LinhaRelatorio struct {
	ItemPlanoID                     uuid.UUID
	CursoNome                       string
	CursoVago                       bool
	ResponsavelNome                 *string
	ResponsavelDesde                *string // "dd/mm/aaaa", vazio quando vago o período inteiro
	VagoDesde                       *string
	MetaNome                        string
	Indicadores                     []IndicadorDaMeta
	Exigido                         int
	Aceitas                         int
	Pendentes                       int
	EmCorrecao                      int
	Cumprimento                     float64 // 0..100
	Situacao                        string  // cumprida | sem_responsavel | em_andamento | em_correcao | nao_cumprida
	InclusEntregasAnteriores        bool
	AvaliacaoPeloProprioCoordenador bool
}

type ResumoRelatorio struct {
	CursosSemCoordenador     int
	MetasNaoCumpridasDeVagos int
}

type ResultadoRelatorio struct {
	Itens  []LinhaRelatorio
	Total  int
	Resumo ResumoRelatorio
}

// LinhaCSVRelatorio — uma linha já formatada para exportação em fluxo
// (design.md §9.3): nunca a lista completa montada em memória.
type LinhaCSVRelatorio = LinhaRelatorio

// ItemDesempenhoPorCurso — uma barra do gráfico "Cumprimento de metas por
// curso" (ux.md, design.md §9.4). Só cursos COM coordenador aparecem aqui
// — curso vago nunca vira item deste slice (ver ResultadoDesempenhoPorCurso).
type ItemDesempenhoPorCurso struct {
	CursoID         uuid.UUID
	CursoNome       string
	ResponsavelNome string
	ExigidoTotal    int
	AceitasTotal    int     // já somado com o cap por item (design.md §9.4)
	Percentual      float64 // 0..100, ExigidoTotal>0 ? min(100, AceitasTotal/ExigidoTotal*100) : 0
}

// ResultadoDesempenhoPorCurso — design.md §9.4. Itens é o ranking (top 10,
// só cursos com coordenador). CursosSemCoordenador é a lista de NOMES dos
// cursos vagos que bateram no filtro — nunca uma barra (ux.md, tabela
// "Distinguir sem coordenador/0%"). TotalCursosComCoordenador é o universo
// inteiro (não só os 10 exibidos), para o texto "Mostrando os 10 melhores
// de N".
type ResultadoDesempenhoPorCurso struct {
	Itens                     []ItemDesempenhoPorCurso
	CursosSemCoordenador      []string
	TotalCursosComCoordenador int
	TotalCursosSemCoordenador int
}

// EntregaParaNotificar — o que o relê lê para montar a mensagem (design.md
// §8.4): tudo vem da PRÓPRIA linha da entrega, nunca de um payload
// congelado à parte.
type EntregaParaNotificar struct {
	EntregaID             uuid.UUID
	Evento                string // "recusa" | "desfazimento" | "prazo_restaurado"
	CursoNome             string
	MetaNome              string
	Motivo                string
	PrazoCorrecao         *time.Time
	Rodadas               int
	DestinatarioEmail     string
	DestinatarioNome      string
	NotificacaoTentativas int
}

// CandidatoRestauracaoDePrazo — PM-4/VG-06: entrega recusada cujo prazo
// expirou durante a vacância do curso.
type CandidatoRestauracaoDePrazo struct {
	EntregaID uuid.UUID
	CursoID   uuid.UUID
}

// PeriodoOpcao — a projeção mínima para o seletor de período de "Minhas
// metas" (ux.md, "Seletor de período"). O coordenador não tem
// periodo.gerenciar (é permissão do PI); esta leitura é liberada pelo
// MESMO alcance de carteira que já autoriza "Minhas metas". Desde T-113,
// ListarPeriodosParaMinhasMetas devolve só os períodos em que o
// coordenador TEM metas (derivado da carteira, não todo o catálogo da
// instituição) — a implementação não usa mais escopo.SemCarteira(): o
// recorte de carteira já é exatamente o que esta consulta precisa.
type PeriodoOpcao struct {
	ID      uuid.UUID
	Nome    string
	DataFim string
	Aberto  bool
}

// EntregaRepository serve entrega, anexo e o relatório de desempenho
// (specs/metas-coordenacao/design.md §5). Todo método recebe
// autorizacao.Escopo — o filtro de isolamento é montado só por
// postgres.AplicarEscopo (fundacao-metas.md §3).
type EntregaRepository interface {
	// ListarPeriodosParaMinhasMetas alimenta o seletor de período da única
	// tela sem "Pesquisar" — não expõe nada que o PeriodoRepository já não
	// exponha, só evita exigir periodo.gerenciar do coordenador para um
	// combo de leitura.
	ListarPeriodosParaMinhasMetas(ctx context.Context, escopo autorizacao.Escopo) ([]PeriodoOpcao, error)

	// avaliador é sempre o próprio ator autenticado (autorizacao.Proprio,
	// nunca uuid.UUID cru) — usado só para computar CoordenadoPeloAvaliador
	// (T-277: a tela avisa ANTES de abrir se "EU, avaliando esta entrega,
	// também sou coordenador deste curso hoje?"). Para o coordenador vendo
	// a própria entrega o campo é irrelevante e ignorado pela tela. Quem
	// pergunta E o recurso alcançado importam aqui, por isso Escopo e
	// Proprio aparecem juntos (design.md de autenticacao-usuarios §4.4:
	// "se o método alcança um recurso, Proprio não substitui Escopo —
	// valem os dois").
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, avaliador autorizacao.Proprio) (DetalheEntrega, error)
	ListarDoItem(ctx context.Context, escopo autorizacao.Escopo, itemPlanoID uuid.UUID, filtro FiltroListarEntregas) (ResultadoListaEntregas, error)

	// CoordenaCursoHoje responde o fato "EU (o próprio ator) coordeno
	// cursoID hoje" — a pergunta decide se EU posso avaliar (M-13, marca
	// gravada no instante do ato) ou o que aparece como
	// CoordenadoPeloAvaliador nas listagens. usuarioID nunca foi de
	// terceiro: os três pontos de chamada sempre passam in.Ator.UsuarioID().
	// autorizacao.Proprio (nunca uuid.UUID cru) — mesma garantia estrutural
	// de Escopo, só no eixo "quem pergunta" em vez de "o que é alcançado"
	// (design.md de autenticacao-usuarios §4.4). cursoID continua uuid.UUID
	// puro, sem Escopo: a consulta CONJUGA curso e ator
	// (curso_id = $1 AND coordenador_id = $2) — um curso de outra
	// instituição só torna a resposta mais restritiva (nunca encontra
	// designação), nunca concede privilégio. É essa conjunção, não o
	// comentário, que torna o método seguro sem Escopo.
	CoordenaCursoHoje(ctx context.Context, proprio autorizacao.Proprio, cursoID uuid.UUID, dataDeReferencia valueobject.DataLocal) (bool, error)

	// InserirIdempotencia e InserirEntregaComAnexos são chamados em
	// sequência, DENTRO da mesma unidade de trabalho (design.md §7, M-04):
	// a garantia é a chave primária de `idempotencia`, não uma verificação
	// prévia. Em caso de violação de PK, InserirIdempotencia devolve
	// domain.ErrChaveDuplicada, a transação sofre rollback e o use case
	// relê a entrega original com BuscarEntregaPorChaveIdempotencia — FORA
	// da transação que já morreu.
	//
	// O identificador do usuário entra como autorizacao.Proprio, não
	// uuid.UUID cru: a chave de idempotência é sempre a do PRÓPRIO ator
	// autenticado (nunca de terceiro), e Proprio só nasce de Ator.Proprio()
	// — a mesma garantia estrutural que port.AutenticacaoRepository já usa
	// para a troca da própria senha (design.md de autenticacao-usuarios
	// §4.4). recursoID continua uuid.UUID puro: é o id gerado pelo próprio
	// processo na mesma transação (proveniência (c) de §4.4), não um
	// identificador de terceiro.
	InserirIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string, recursoID uuid.UUID) error
	InserirEntregaComAnexos(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, anexos []AnexoParaGravar) error
	BuscarEntregaPorChaveIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string) (uuid.UUID, error)

	// AtualizarCorrecao grava a correção (observação, situação de volta a
	// pendente_avaliacao) — SEM tocar anexos, que têm seu próprio comando.
	AtualizarCorrecao(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) error
	ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
	MarcarPendenciaVista(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error

	// Avaliar — UPDATE condicional em id+versão+situação='pendente_avaliacao'
	// (design.md §5.2). rowsAffected=false distingue CONFLITO_DE_VERSAO de
	// ENTREGA_JA_AVALIADA: o chamador relê a linha para decidir qual.
	Avaliar(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) (rowsAffected bool, err error)
	DesfazerAceitacao(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) (rowsAffected bool, err error)

	AdicionarAnexos(ctx context.Context, escopo autorizacao.Escopo, entregaID uuid.UUID, anexos []AnexoParaGravar) ([]AnexoResponse, error)
	// autor é sempre o próprio ator autenticado (autorizacao.Proprio) — só
	// o autor da entrega remove um anexo dela, e o repositório confere
	// isso contra entrega.enviada_por. Escopo recorta o anexo pelo
	// isolamento normal; Proprio verifica quem está pedindo — os dois
	// juntos, mesmo caso de EntregaRepository.BuscarPorID (T-114).
	RemoverAnexo(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID, autor autorizacao.Proprio) (chaveObjeto string, err error)
	BuscarAnexoParaDownload(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID) (nomeOriginal, tipo, chaveObjeto string, err error)

	MinhasMetas(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) ([]GrupoMinhasMetas, error)
	// avaliador é sempre o próprio ator (autorizacao.Proprio) — mesmo
	// padrão de RemoverAnexo/BuscarPorID (T-114, T-118): a fila é SEMPRE
	// a fila de quem pediu, nunca a de um terceiro escolhido pelo cliente.
	ListarFilaDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo, avaliador autorizacao.Proprio, filtro FiltroFilaAvaliacao) (ResultadoListaEntregas, error)
	ContarPendentesDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo) (int, error)
	ContarPendenciasNaoVistas(ctx context.Context, escopo autorizacao.Escopo) (int, error)

	RelatorioDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroRelatorio) (ResultadoRelatorio, error)
	// ExportarDesempenho é streaming: cada linha é entregue via callback,
	// nunca acumulada em slice (design.md §9.3).
	ExportarDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroRelatorio, linha func(LinhaCSVRelatorio) error) (totalLinhas int, err error)

	// DesempenhoPorCurso alimenta o gráfico "Cumprimento de metas por
	// curso" (design.md §9.4). Reaproveita montarRelatorioBase — o MESMO
	// FROM/WHERE de RelatorioDesempenho/ExportarDesempenho, nunca uma
	// consulta nova — para que o universo do gráfico nunca divirja do
	// universo da tabela sob os mesmos filtros. Page/PageSize/Sort/Order
	// de FiltroRelatorio são ignorados: o agregado sempre roda sobre o
	// conjunto filtrado inteiro, nunca sobre uma página.
	DesempenhoPorCurso(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroRelatorio) (ResultadoDesempenhoPorCurso, error)
}

// EntregaReleRepository — TERCEIRA porta estreita do sistema, ao lado de
// AutenticacaoRepository e InstituicaoPublicaQuery (design.md §4.4 de
// autenticacao-usuarios, fundacao-metas.md §7). Justificativa, paralela à
// das outras duas: o relê em segundo plano não tem sessão, não tem ator,
// não tem instituição — ele varre TODAS as instituições de uma vez, por
// desenho (é o único processo do sistema que precisa disso). AplicarEscopo
// pressupõe um recorte de UMA instituição ou plataforma; um job de
// manutenção cross-tenant está, por natureza, fora do caminho que
// autorizacao.Escopo modela. Nenhum destes métodos é alcançável por uma
// rota HTTP.
type EntregaReleRepository interface {
	// ReivindicarNotificacoesPendentes — T-124/T-127 (fundacao-metas.md
	// §4.7, §7.2, design.md §8.4, revisão pós code-review). Substitui o
	// antigo par "SELECT ... FOR UPDATE SKIP LOCKED" + envio: aquele SELECT
	// roda em autocommit e libera o lock no instante em que termina — MUITO
	// antes do envio de e-mail (latência imprevisível), então N réplicas
	// enxergavam e enviavam a MESMA notificação. A reivindicação agora é
	// uma ESCRITA (grava notificacao_reivindicada_em) num único statement
	// atômico — sobrevive à chamada SMTP porque é dado persistido, não lock.
	//
	// `hoje` é a data de referência no fuso de exibição (nunca UTC: decide
	// QUEM recebe, via designação vigente — DG-06), calculada pelo
	// chamador com o relógio e o fuso do próprio processo.
	//
	// A mesma coluna de reivindicação serve de "última tentativa" para
	// backoff exponencial: uma linha só volta a ficar elegível depois de
	// 2^tentativas minutos (com teto) desde a última reivindicação — uma
	// coluna fecha as duas pendências (reivindicação atômica e backoff).
	//
	// Semântica honesta: ao menos uma vez (at-least-once). Se o processo
	// cair entre o envio bem-sucedido e MarcarNotificacaoEnviada, a
	// reivindicação expira e a mesma entrega é reivindicada de novo —
	// duplicando o e-mail. Preferimos duplicar a perder o alerta.
	ReivindicarNotificacoesPendentes(ctx context.Context, hoje valueobject.DataLocal, limite int) ([]EntregaParaNotificar, error)
	// ContarNotificacoesPendentes alimenta o gauge notificacoes_pendentes
	// (design.md §8.4) — conta TODAS as pendentes, não só o lote do
	// limite; gauge crescendo é relê parado.
	ContarNotificacoesPendentes(ctx context.Context) (int, error)
	MarcarNotificacaoEnviada(ctx context.Context, entregaID uuid.UUID) error
	RegistrarFalhaDeNotificacao(ctx context.Context, entregaID uuid.UUID, erro string) error
	ListarCandidatosARestauracaoDePrazo(ctx context.Context, limite int) ([]CandidatoRestauracaoDePrazo, error)
	RestaurarPrazoPorVacancia(ctx context.Context, entregaID uuid.UUID, novoPrazo time.Time) error

	// ExpurgarIdempotenciaAntesDe remove chaves de idempotência com mais
	// de `antes` de idade (V-7, janela de 24h de M-05) — também
	// cross-tenant por natureza (rotina de retenção, não caminho de
	// cliente).
	ExpurgarIdempotenciaAntesDe(ctx context.Context, antes time.Time) (int64, error)
}
