package port

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/google/uuid"
)

type FiltroListarPlanos struct {
	PeriodoID *uuid.UUID
	CursoID   *uuid.UUID
	Situacao  string // "todas" | "rascunho" | "vigente" | "encerrado"
	Aprovacao string // "todos" | "aprovados" | "sem_aprovacao"
	Page      int
	PageSize  int
	Sort      string
	Order     string
}

// LinhaPlano é a projeção usada pelo grid — sem os itens (design.md §6,
// contrato de API: "metas" na listagem é só a contagem).
type LinhaPlano struct {
	Plano           plano.Plano
	CursoNome       string
	CursoGrau       string
	CursoModalidade string
	CursoCodigoEMec string
	CursoVago       bool
	PeriodoNome     string
	Situacao        string // efetiva — já traduzida por Plano.SituacaoEfetiva
	SemAprovacao    bool
	TotalItens      int
	TotalExigido    int
	TemEntrega      bool
	UltimoDocumento *ResumoDocumento // só preenchido no detalhe
}

type ResumoDocumento struct {
	ID       uuid.UUID
	GeradoEm time.Time
}

// ItemDoPlanoResponse embute os indicadores da meta — a origem de cada
// um, para a tela renderizar sem segunda chamada por linha (IT-03).
type ItemDoPlanoResponse struct {
	ID          uuid.UUID
	MetaID      uuid.UUID
	MetaNome    string
	Indicadores []IndicadorDaMeta
	Quantidade  int
	TemEntrega  bool
	Versao      int
}

// DetalhePlano é o que a página do plano (ux.md) carrega de uma vez:
// cabeçalho, coordenador na data de referência, itens e último documento.
type DetalhePlano struct {
	LinhaPlano
	CoordenadorNome     *string // nil = curso vago, "Sem responsável"
	CoordenadorPortaria *string
	Itens               []ItemDoPlanoResponse
}

type ResultadoListaPlanos struct {
	Itens []LinhaPlano
	Total int
}

// DestinoDeCopia — uma linha da tela de cópia em lote (ux.md, "Cópia em
// lote"). JaTemPlano vem do MESMO período de destino verificado na
// consulta; Vago é o mesmo conceito de cursos (sem designação vigente).
type DestinoDeCopia struct {
	CursoID         uuid.UUID
	CursoNome       string
	CoordenadorNome *string
	Vago            bool
	JaTemPlano      bool
}

// PlanoRepository serve /api/v1/planos e /api/v1/meus-planos
// (specs/plano-acao/design.md §5, §6).
type PlanoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (DetalhePlano, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarPlanos) (ResultadoListaPlanos, error)
	// ListarMeusPlanos usa o mesmo Escopo restrito à carteira
	// (PlanosDaCarteira) — em TODAS as situações, somente leitura (P-10).
	ListarMeusPlanos(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarPlanos) (ResultadoListaPlanos, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano) error
	AtualizarDados(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error
	AtualizarSituacao(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error

	// InstituicaoDaSessao devolve nome e sigla da instituição do Escopo —
	// consulta direta à tabela instituicao (leitura de contexto do
	// próprio ator, não recurso pedido pelo cliente, por isso sem
	// AplicarEscopo/Alvo). Usada só para preencher o cabeçalho do .docx.
	InstituicaoDaSessao(ctx context.Context, escopo autorizacao.Escopo) (nome, sigla string, err error)

	// PeriodoDoPlano lê nome e vigência do período de um plano já
	// validado pelo Escopo — filtrado só por instituição (nunca por
	// carteira): período é entidade da instituição, não do curso, e
	// AlvoPeriodo corretamente recusa restrição de carteira
	// (temColunaCurso: false). Usar port.PeriodoRepository aqui exigiria
	// o alcance PeriodosDaInstituicao, que o Coordenador não tem —
	// mesmo já estando autorizado a ver o próprio plano.
	PeriodoDoPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (periodo.Periodo, error)

	// CursoValidoParaPlano confere existência (404 senão), instituição do
	// escopo e situação ativa (400 CURSO_INATIVO) — consulta direta à
	// tabela curso, que já existe (specs/cursos/design.md), sem depender
	// de CursoRepository (fora do escopo desta feature).
	CursoValidoParaPlano(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID) (ativo bool, err error)

	// ExisteEntregaNoPlano/ExisteEntregaNoItem sustentam PLANO_COM_ENTREGA e
	// ITEM_COM_ENTREGA (PP-3, IT-10). A tabela `entrega` nasce em
	// specs/metas-coordenacao, ainda não construída — por ora devolvem
	// sempre false, no mesmo padrão que port/meta_repository.go usou para
	// Planos=0 antes de item_plano existir. Ver testes-pendentes.md.
	ExisteEntregaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) (bool, error)
	ExisteEntregaNoItem(ctx context.Context, escopo autorizacao.Escopo, itemID uuid.UUID) (bool, error)

	ListarDestinosDeCopia(ctx context.Context, escopo autorizacao.Escopo, planoOrigemID, periodoDestinoID uuid.UUID, busca string) ([]DestinoDeCopia, error)
	// InserirComItens grava o plano copiado e os itens num único
	// INSERT multi-linha (design.md §5.3, V-6) — usada só pela cópia em
	// lote, que já abre sua própria transação por curso.
	InserirComItens(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, itensOrigem []ItemDoPlanoResponse) error
}
