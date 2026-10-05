package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// lateralIndicadoresDoItem espelha lateralIndicadoresDaMeta
// (meta_repository.go), correlacionada por item_plano.meta_id — sempre
// via AplicarEscopo(AlvoIndicador), nunca junção nua (fundacao-metas.md
// §3.3): "o item já está recortado" não é autorização para o indicador.
func lateralIndicadoresDoItem(escopo autorizacao.Escopo, proximoPlaceholder int) (string, []any, int, error) {
	clausula, args, err := AplicarEscopo(escopo.SemCarteira(), AlvoIndicador, proximoPlaceholder)
	if err != nil {
		return "", nil, 0, err
	}
	sqlText := fmt.Sprintf(`
	LEFT JOIN LATERAL (
	  SELECT json_agg(json_build_object(
	           'id', indicador.id, 'codigo', indicador.codigo, 'nome', indicador.nome,
	           'escopo', indicador.escopo, 'referencia_instrumento', indicador.referencia_instrumento,
	           'situacao', indicador.situacao
	         ) ORDER BY indicador.codigo) AS indicadores
	    FROM meta_indicador mi
	    JOIN indicador ON indicador.id = mi.indicador_id
	   WHERE mi.meta_id = item_plano.meta_id AND %s
	) ind ON TRUE`, clausula)
	return sqlText, args, proximoPlaceholder + len(args), nil
}

func indicadoresDeJSON(bruto *string) ([]port.IndicadorDaMeta, error) {
	if bruto == nil {
		return nil, nil
	}
	var embutidos []indicadorEmbutidoJSON
	if err := json.Unmarshal([]byte(*bruto), &embutidos); err != nil {
		return nil, err
	}
	ricos := make([]port.IndicadorDaMeta, 0, len(embutidos))
	for _, e := range embutidos {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			return nil, err
		}
		ref := ""
		if e.ReferenciaInstrumento != nil {
			ref = *e.ReferenciaInstrumento
		}
		ricos = append(ricos, port.IndicadorDaMeta{
			ID: id, Codigo: e.Codigo, Nome: e.Nome, Escopo: e.Escopo,
			ReferenciaInstrumento: ref, Situacao: e.Situacao,
		})
	}
	return ricos, nil
}

// ListarPeriodosParaMinhasMetas — ver comentário em port.PeriodoOpcao.
// "Os períodos em que eu tenho metas" é projeção derivada da carteira do
// coordenador, não o catálogo da instituição — a restrição de carteira é
// APLICADA, nunca removida (decisão do arquiteto, T-113). Por isso a
// consulta passa por AlvoItemPlano (que suporta carteira), na MESMA base
// que produz MinhasMetas — plano vigente, item do plano dentro do
// recorte — e não por AlvoPeriodo com escopo.SemCarteira(): um período
// sem nenhum item na carteira do ator não aparece, o que é o
// comportamento certo de um seletor (nunca oferecer uma opção que leva a
// tela vazia). DISTINCT aqui deduplica LINHAS de período, não agregados —
// não é o padrão proibido do relatório (M-10): não há count/soma nesta
// consulta para o DISTINCT mascarar.
func (r *EntregaRepository) ListarPeriodosParaMinhasMetas(ctx context.Context, escopo autorizacao.Escopo) ([]port.PeriodoOpcao, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoItemPlano, 1)
	if err != nil {
		return nil, err
	}
	args = append(args, escopo.DataDeReferencia().String())
	n := len(args)
	ex := Executor(ctx, r.db)
	var linhas []struct {
		ID      uuid.UUID `db:"id"`
		Nome    string    `db:"nome"`
		DataFim string    `db:"data_fim"`
		Aberto  bool      `db:"aberto"`
	}
	consulta := fmt.Sprintf(`
		SELECT DISTINCT periodo.id, periodo.nome, periodo.data_fim::text AS data_fim,
		       (periodo.data_inicio <= $%d::date AND periodo.data_fim >= $%d::date) AS aberto
		  FROM periodo
		  JOIN plano ON plano.periodo_id = periodo.id AND plano.excluido_em IS NULL
		  JOIN item_plano ON item_plano.plano_id = plano.id
		 WHERE %s AND plano.situacao_publicacao = 'vigente'
		 ORDER BY data_fim DESC`, n, n, clausula)
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	opcoes := make([]port.PeriodoOpcao, 0, len(linhas))
	for _, l := range linhas {
		opcoes = append(opcoes, port.PeriodoOpcao{ID: l.ID, Nome: l.Nome, DataFim: l.DataFim, Aberto: l.Aberto})
	}
	return opcoes, nil
}

// MinhasMetas — GET /api/v1/minhas-metas (T-270). O SomenteVigentes
// (situacao_publicacao = 'vigente', que cobre tanto "vigente" quanto
// "encerrado" efetivos — rascunho é o único valor de coluna que os
// distingue) mora AQUI, veio da consulta de planos (plano-acao §5.4):
// item de plano em rascunho não é obrigação e não aparece.
func (r *EntregaRepository) MinhasMetas(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) ([]port.GrupoMinhasMetas, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoItemPlano, 3)
	if err != nil {
		return nil, err
	}
	lateralSQL, lateralArgs, _, err := lateralIndicadoresDoItem(escopo, 3+len(escopoArgs))
	if err != nil {
		return nil, err
	}
	args := append([]any{periodoID}, escopoArgs...)
	args = append(args, lateralArgs...)

	consulta := `
		SELECT curso.id AS curso_id, curso.nome AS curso_nome,
		       item_plano.id AS item_plano_id, meta.nome AS meta_nome, item_plano.quantidade,
		       coalesce(ent.aceitas, 0) AS aceitas, coalesce(ent.pendentes, 0) AS pendentes,
		       coalesce(ent.em_correcao, 0) AS em_correcao, ind.indicadores
		  FROM item_plano
		  JOIN plano ON plano.id = item_plano.plano_id AND plano.excluido_em IS NULL
		  JOIN curso ON curso.id = item_plano.curso_id
		  JOIN meta ON meta.id = item_plano.meta_id
		  LEFT JOIN LATERAL (
		    SELECT count(*) FILTER (WHERE e.situacao = 'aceita')             AS aceitas,
		           count(*) FILTER (WHERE e.situacao = 'pendente_avaliacao') AS pendentes,
		           count(*) FILTER (WHERE e.situacao = 'recusada')           AS em_correcao
		      FROM entrega e WHERE e.item_plano_id = item_plano.id AND e.excluido_em IS NULL
		  ) ent ON TRUE
		  ` + lateralSQL + `
		 WHERE ` + clausula + `
		   AND plano.situacao_publicacao = 'vigente'
		   AND plano.periodo_id = $1
		 ORDER BY curso.nome COLLATE "pt-BR-x-icu", meta.nome COLLATE "pt-BR-x-icu"`

	ex := Executor(ctx, r.db)
	var linhas []struct {
		CursoID     uuid.UUID `db:"curso_id"`
		CursoNome   string    `db:"curso_nome"`
		ItemPlanoID uuid.UUID `db:"item_plano_id"`
		MetaNome    string    `db:"meta_nome"`
		Quantidade  int       `db:"quantidade"`
		Aceitas     int       `db:"aceitas"`
		Pendentes   int       `db:"pendentes"`
		EmCorrecao  int       `db:"em_correcao"`
		Indicadores *string   `db:"indicadores"`
	}
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}

	grupos := make([]port.GrupoMinhasMetas, 0)
	indicePorCurso := map[uuid.UUID]int{}
	for _, l := range linhas {
		indicadores, err := indicadoresDeJSON(l.Indicadores)
		if err != nil {
			return nil, err
		}
		item := port.ItemMinhasMetas{
			ItemPlanoID: l.ItemPlanoID, MetaNome: l.MetaNome, Indicadores: indicadores,
			Quantidade: l.Quantidade, Aceitas: l.Aceitas, Pendentes: l.Pendentes, EmCorrecao: l.EmCorrecao,
		}
		idx, ok := indicePorCurso[l.CursoID]
		if !ok {
			grupos = append(grupos, port.GrupoMinhasMetas{CursoID: l.CursoID, CursoNome: l.CursoNome})
			idx = len(grupos) - 1
			indicePorCurso[l.CursoID] = idx
		}
		grupos[idx].Itens = append(grupos[idx].Itens, item)
	}
	return grupos, nil
}
