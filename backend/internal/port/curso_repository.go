package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// CoordenadorDoCurso é o campo computado "coordenador" da linha de curso
// (design.md §5.5) — projeção de uma designação vigente, nunca coluna
// persistida. nil quando o curso está vago (CU-05).
type CoordenadorDoCurso struct {
	ID                             uuid.UUID
	Nome                           string
	DataFim                        *valueobject.DataLocal
	TambemPesquisadorInstitucional bool
}

type ItemCurso struct {
	Curso       curso.Curso
	Coordenador *CoordenadorDoCurso
	// TemVinculo — mesmo EXISTS que CursoRepository.Excluir usa para
	// recusar a exclusão (design.md §5.4); exposto aqui para o botão
	// "Excluir" já nascer escondido no grid, sem convidar a um 409
	// garantido (ux.md, mesma decisão de specs/indicadores/ux.md).
	TemVinculo bool
}

type FiltroListarCursos struct {
	Busca         string
	Grau          string
	Modalidade    string
	Situacao      string
	CoordenadorID *uuid.UUID
	// Vago — filtra só cursos sem designação vigente (spec.md §7, opção
	// "Vago" do filtro Coordenador). Ignorado quando CoordenadorID != nil.
	Vago     bool
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ResultadoListaCursos struct {
	Itens []ItemCurso
	Total int
}

// CursoRepository serve /api/v1/cursos e /api/v1/meus-cursos.
type CursoRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (ItemCurso, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarCursos) (ResultadoListaCursos, error)
	// ListarMeusCursos — cursos ativos com designação vigente do ator
	// (CV-01); o Escopo já vem com RestritoACarteiraDe ligado por
	// autorizacao.CursosDaCarteira.
	ListarMeusCursos(ctx context.Context, escopo autorizacao.Escopo) ([]curso.Curso, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso, versaoEsperada int) error
	AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCurso, versaoEsperada int) error
	// Excluir — 409 CURSO_COM_VINCULO se houver designação ou plano (design.md
	// §5.4); verificado sob SELECT ... FOR UPDATE na linha do curso.
	Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
	// TravarSeAtivo — SELECT ... FOR SHARE na linha do curso, ErrNaoEncontrado
	// se excluído ou fora do escopo. É o que fecha a corrida com Excluir
	// (achado de revisão T-171): o FK de designacao→curso só garante que a
	// LINHA existe, nunca que excluido_em continua nulo — a exclusão lógica
	// nunca dispara a FK. Sem uma trava própria aqui, CriarDesignacaoUseCase
	// podia criar designação para um curso que acabou de ser excluído no
	// instante entre a leitura e a escrita. FOR SHARE (não FOR UPDATE):
	// múltiplas criações de designação no mesmo curso não precisam se
	// bloquear entre si — só contra Excluir, que é FOR UPDATE. Chamado
	// sempre dentro da mesma unidade de trabalho da escrita.
	TravarSeAtivo(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error
}
