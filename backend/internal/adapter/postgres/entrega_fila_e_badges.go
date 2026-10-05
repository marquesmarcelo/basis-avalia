package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/jmoiron/sqlx"
)

// colunasEntregaComPlano e juncoesEntregaComPlano — separadas de propósito
// (achado de revisão, T-292): antes eram uma única constante com "FROM ...
// JOINs" já embutido, e ListarFilaDeAvaliacao concatenava mais uma coluna
// (EXISTS(...) AS coordenado_pelo_avaliador) DEPOIS dela — ou seja, depois
// do FROM/JOINs. A vírgula que devia separar duas colunas do SELECT virava
// uma vírgula depois do último JOIN, e o Postgres tentava ler
// "EXISTS(...) AS coordenado_pelo_avaliador" como mais um item da lista
// FROM (join implícito à moda antiga) em vez de uma coluna — daí o erro
// "could not determine data type of parameter $1" (42P18): o parser não
// falhava com erro de sintaxe, mas não conseguia mais resolver o tipo do
// parâmetro usado dentro do EXISTS naquela posição. Com as duas
// constantes separadas, quem monta a query decide explicitamente onde
// cada coluna extra entra — antes do FROM, nunca depois.
const colunasEntregaComPlano = `
	entrega.*, item_plano.meta_id, item_plano.quantidade, meta.nome AS meta_nome, curso.nome AS curso_nome,
	enviou.nome AS enviada_por_nome, corrigiu.nome AS corrigida_por_nome, avaliou.nome AS avaliada_por_nome
`

const juncoesEntregaComPlano = `
	FROM entrega
	JOIN item_plano ON item_plano.id = entrega.item_plano_id
	JOIN plano ON plano.id = item_plano.plano_id
	JOIN meta ON meta.id = item_plano.meta_id
	JOIN curso ON curso.id = entrega.curso_id
	JOIN usuario enviou ON enviou.id = entrega.enviada_por
	LEFT JOIN usuario corrigiu ON corrigiu.id = entrega.corrigida_por
	LEFT JOIN usuario avaliou ON avaliou.id = entrega.avaliada_por
`

// ListarFilaDeAvaliacao — GET /api/v1/avaliacoes (T-277): cada linha traz
// coordenado_pelo_avaliador, para a tela avisar ANTES de abrir, não
// depois (design.md §5.3).
func (r *EntregaRepository) ListarFilaDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo, avaliador autorizacao.Proprio, filtro port.FiltroFilaAvaliacao) (port.ResultadoListaEntregas, error) {
	// avaliador.UsuarioID() NÃO entra em `args` aqui — achado de revisão
	// (T-292): ele só é lido pelo EXISTS de coordenado_pelo_avaliador, na
	// consulta principal, nunca na de contagem. Reservar um placeholder
	// fixo pra ele em `args` (como este método fazia antes) faz a
	// consulta de CONTAGEM receber um parâmetro que nunca aparece no
	// texto dela — e é exatamente isso que o Postgres recusa com "could
	// not determine data type of parameter $1" (42P18): o parâmetro
	// declarado no Parse nunca é referenciado, então não há contexto
	// nenhum para inferir o tipo. Cada consulta agora recebe só os
	// argumentos que de fato usa.
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 1)
	if err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	condicoes := []string{clausula}
	args := append([]any{}, escopoArgs...)
	n := len(args) + 1

	if filtro.PeriodoID != nil {
		condicoes = append(condicoes, fmt.Sprintf("plano.periodo_id = $%d", n))
		args = append(args, *filtro.PeriodoID)
		n++
	}
	if filtro.CursoID != nil {
		condicoes = append(condicoes, fmt.Sprintf("entrega.curso_id = $%d", n))
		args = append(args, *filtro.CursoID)
		n++
	}
	if filtro.MetaID != nil {
		condicoes = append(condicoes, fmt.Sprintf("item_plano.meta_id = $%d", n))
		args = append(args, *filtro.MetaID)
		n++
	}
	if filtro.Situacao != "" {
		condicoes = append(condicoes, fmt.Sprintf("entrega.situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	contagem := "SELECT count(*) FROM entrega JOIN item_plano ON item_plano.id = entrega.item_plano_id JOIN plano ON plano.id = item_plano.plano_id JOIN meta ON meta.id = item_plano.meta_id JOIN curso ON curso.id = entrega.curso_id JOIN usuario enviou ON enviou.id = entrega.enviada_por WHERE " + where
	if err := sqlx.GetContext(ctx, ex, &total, contagem, args...); err != nil {
		return port.ResultadoListaEntregas{}, err
	}

	sortColuna, ok := ordenacaoEntrega[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoEntrega["criado_em"]
	}
	ordem := "ASC"
	if filtro.Order == "desc" {
		ordem = "DESC"
	}
	page, pageSize := paginaEPageSize(filtro.Page, filtro.PageSize)
	offset := (page - 1) * pageSize

	// nAvaliador é o placeholder do avaliador (só usado aqui, na consulta
	// principal); nData é o da data de referência (usado duas vezes por
	// FragmentoDesignacaoVigente); LIMIT/OFFSET vêm logo depois dela, na
	// mesma ordem em que argsComData é montado.
	nAvaliador := n
	nData := n + 1
	consulta := fmt.Sprintf(`SELECT %s,
		EXISTS (SELECT 1 FROM designacao d WHERE d.curso_id = entrega.curso_id AND d.coordenador_id = $%d
		         AND d.excluido_em IS NULL AND %s) AS coordenado_pelo_avaliador
		%s
		WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		colunasEntregaComPlano, nAvaliador, FragmentoDesignacaoVigente("d", nData), juncoesEntregaComPlano, where, sortColuna, ordem, nData+1, nData+2)
	argsComData := append(append([]any{}, args...), avaliador.UsuarioID(), escopo.DataDeReferencia().String(), pageSize, offset)

	var linhas []linhaEntregaDBComMarca
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsComData...); err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	itens := make([]port.LinhaEntrega, 0, len(linhas))
	for _, l := range linhas {
		item := l.linhaEntregaDB.paraLinha()
		item.CoordenadoPeloAvaliador = l.CoordenadoPeloAvaliador
		anexos, err := r.listarAnexos(ctx, l.ID)
		if err != nil {
			return port.ResultadoListaEntregas{}, err
		}
		item.Anexos = anexos
		itens = append(itens, item)
	}
	return port.ResultadoListaEntregas{Itens: itens, Total: total}, nil
}

type linhaEntregaDBComMarca struct {
	linhaEntregaDB
	CoordenadoPeloAvaliador bool `db:"coordenado_pelo_avaliador"`
}

func (r *EntregaRepository) ContarPendentesDeAvaliacao(ctx context.Context, escopo autorizacao.Escopo) (int, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoEntrega, 1)
	if err != nil {
		return 0, err
	}
	args = append(args, "pendente_avaliacao")
	ex := Executor(ctx, r.db)
	var total int
	consulta := fmt.Sprintf("SELECT count(*) FROM entrega WHERE %s AND entrega.situacao = $%d", clausula, len(args))
	if err := sqlx.GetContext(ctx, ex, &total, consulta, args...); err != nil {
		return 0, err
	}
	return total, nil
}

// ContarPendenciasNaoVistas — o badge do coordenador (NT-04): literalmente
// "não vistas", zera por item quando ele abre a entrega — nunca um
// contador de recusas totais.
func (r *EntregaRepository) ContarPendenciasNaoVistas(ctx context.Context, escopo autorizacao.Escopo) (int, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoEntrega, 1)
	if err != nil {
		return 0, err
	}
	args = append(args, "recusada")
	ex := Executor(ctx, r.db)
	var total int
	consulta := fmt.Sprintf("SELECT count(*) FROM entrega WHERE %s AND entrega.situacao = $%d AND entrega.pendencia_vista_em IS NULL", clausula, len(args))
	if err := sqlx.GetContext(ctx, ex, &total, consulta, args...); err != nil {
		return 0, err
	}
	return total, nil
}
