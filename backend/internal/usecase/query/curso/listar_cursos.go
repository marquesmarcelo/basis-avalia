package curso

import (
	"context"
	"fmt"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarCursosInput struct {
	Ator          autorizacao.Ator
	Busca         string
	Grau          string
	Modalidade    string
	Situacao      string
	CoordenadorID *uuid.UUID
	Vago          bool
	Page          int
	PageSize      int
	Sort          string
	Order         string
}

// ItemCursoComPlano compõe a linha de curso com a situação do plano no
// período aberto (design.md §5.5/§5.7, T-115) — nil quando não há
// período aberto ou o curso não tem plano nele ("—" no grid).
type ItemCursoComPlano struct {
	port.ItemCurso
	PlanoDoPeriodo *string
}

type ListarCursosOutput struct {
	Itens []ItemCursoComPlano
	Total int
}

// ListarCursosUseCase cobre CU-01..CU-11 e, desde T-115, CU-12. Compõe com
// plano-acao SEM CursoRepository conhecer plano — decisão do arquiteto:
// juntar no SELECT de curso inverteria a dependência (cursos é mais
// fundamental que plano-acao). period.Repository/PlanoRepository são
// injetados aqui, no use case de consulta, e usados só com os métodos que
// já existem nas duas portas — nenhuma mudou de contrato por causa disto.
type ListarCursosUseCase struct {
	repo        port.CursoRepository
	periodoRepo port.PeriodoRepository
	planoRepo   port.PlanoRepository
}

func NovoListarCursosUseCase(repo port.CursoRepository, periodoRepo port.PeriodoRepository, planoRepo port.PlanoRepository) *ListarCursosUseCase {
	return &ListarCursosUseCase{repo: repo, periodoRepo: periodoRepo, planoRepo: planoRepo}
}

func (uc *ListarCursosUseCase) Executar(ctx context.Context, in ListarCursosInput) (ListarCursosOutput, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return ListarCursosOutput{}, err
	}
	resultado, err := uc.repo.Listar(ctx, esc, port.FiltroListarCursos{
		Busca: in.Busca, Grau: in.Grau, Modalidade: in.Modalidade, Situacao: in.Situacao,
		CoordenadorID: in.CoordenadorID, Vago: in.Vago,
		Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
	if err != nil {
		return ListarCursosOutput{}, err
	}

	planoPorCurso, err := uc.planoDoPeriodoAbertoPorCurso(ctx, in.Ator)
	if err != nil {
		return ListarCursosOutput{}, err
	}

	itens := make([]ItemCursoComPlano, 0, len(resultado.Itens))
	for _, item := range resultado.Itens {
		composto := ItemCursoComPlano{ItemCurso: item}
		if texto, ok := planoPorCurso[item.Curso.ID]; ok {
			composto.PlanoDoPeriodo = &texto
		}
		itens = append(itens, composto)
	}
	return ListarCursosOutput{Itens: itens, Total: resultado.Total}, nil
}

// planoDoPeriodoAbertoPorCurso — a segunda consulta do "1 + 1" (mais uma
// terceira, para achar QUAL é o período aberto — sem ela não dá para
// filtrar a segunda; nenhuma das duas escala com o número de cursos da
// página, só com o tamanho, fixo, da instituição, o que é o que a regra
// de design.md §5 "1 + 1 consultas" realmente protege: nunca uma consulta
// por linha). Sem período aberto, mapa vazio — toda linha mostra "—".
func (uc *ListarCursosUseCase) planoDoPeriodoAbertoPorCurso(ctx context.Context, ator autorizacao.Ator) (map[uuid.UUID]string, error) {
	escPeriodo, err := autorizacao.Autorizar(ator, autorizacao.PeriodosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	periodos, err := uc.periodoRepo.Listar(ctx, escPeriodo, port.FiltroListarPeriodos{
		Situacao: "aberto", Page: 1, PageSize: 1, Sort: "nome", Order: "asc",
	})
	if err != nil {
		return nil, err
	}
	if len(periodos.Itens) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	periodoAbertoID := periodos.Itens[0].Periodo.ID

	escPlano, err := autorizacao.Autorizar(ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	planos, err := uc.planoRepo.Listar(ctx, escPlano, port.FiltroListarPlanos{
		PeriodoID: &periodoAbertoID, Situacao: "todas", Page: 1, PageSize: 500, Sort: "curso", Order: "asc",
	})
	if err != nil {
		return nil, err
	}

	resultado := make(map[uuid.UUID]string, len(planos.Itens))
	for _, linha := range planos.Itens {
		resultado[linha.Plano.CursoID] = textoPlanoDoPeriodo(linha)
	}
	return resultado, nil
}

func textoPlanoDoPeriodo(linha port.LinhaPlano) string {
	if valueobject.SituacaoPlano(linha.Situacao) == valueobject.PlanoRascunho {
		return "Rascunho"
	}
	return fmt.Sprintf("Vigente (%d metas)", linha.TotalItens)
}
