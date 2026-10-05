package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ordenacaoRelatorio = map[string]string{
	"curso":       `curso.nome COLLATE "pt-BR-x-icu"`,
	"responsavel": "responsavel_nome_ord",
	"meta":        `meta.nome COLLATE "pt-BR-x-icu"`,
	"exigido":     "item_plano.quantidade",
	"aceitas":     "aceitas_ord",
	"cumprimento": "cumprimento_ord",
}

// situacaoCalcSQL — a MESMA árvore de decisão do diagrama 2.2 do design,
// em SQL: cumprida > sem responsável > em andamento > em correção > não
// cumprida, nessa ordem — nenhum ramo alternativo, nenhuma outra saída.
// hojeArg é o placeholder da data de referência (DATE); jamais now()/
// CURRENT_DATE direto (DG-06, invalidaria o índice e divergiria do fuso).
func situacaoCalcSQL(hojeArg int) string {
	return fmt.Sprintf(`CASE
		WHEN coalesce(ent.aceitas,0) >= item_plano.quantidade THEN 'cumprida'
		WHEN resp.coordenador_id IS NULL THEN 'sem_responsavel'
		WHEN NOT (plano.encerrado_em IS NOT NULL OR periodo.data_fim < $%d::date) THEN 'em_andamento'
		WHEN EXISTS (SELECT 1 FROM entrega ec WHERE ec.item_plano_id = item_plano.id AND ec.excluido_em IS NULL
		              AND ec.situacao = 'recusada' AND ec.prazo_correcao_ate >= now()) THEN 'em_correcao'
		ELSE 'nao_cumprida'
	END`, hojeArg)
}

// montarRelatorioBase monta FROM/JOIN/LATERAL e WHERE comuns a
// RelatorioDesempenho e ExportarDesempenho — a MESMA consulta serve as
// duas saídas, nunca duas consultas divergentes (V-3, V-4: o total da
// paginação tem de bater com o conjunto filtrado).
func (r *EntregaRepository) montarRelatorioBase(escopo autorizacao.Escopo, filtro port.FiltroRelatorio) (selectCampos, from, where string, args []any, hojeArg int, err error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoItemPlano, 3)
	if err != nil {
		return "", "", "", nil, 0, err
	}
	lateralInd, lateralIndArgs, proximo, err := lateralIndicadoresDoItem(escopo, 3+len(escopoArgs))
	if err != nil {
		return "", "", "", nil, 0, err
	}

	args = append([]any{filtro.PeriodoID, escopo.DataDeReferencia().String()}, escopoArgs...)
	args = append(args, lateralIndArgs...)
	hojeArg = 2
	n := proximo

	condicoes := []string{clausula, "plano.situacao_publicacao = 'vigente'", "plano.periodo_id = $1"}
	if !filtro.IncluirInativos {
		condicoes = append(condicoes, "curso.situacao = 'ativo'")
	}
	if filtro.CursoID != nil {
		condicoes = append(condicoes, fmt.Sprintf("item_plano.curso_id = $%d", n))
		args = append(args, *filtro.CursoID)
		n++
	}
	if filtro.MetaID != nil {
		condicoes = append(condicoes, fmt.Sprintf("item_plano.meta_id = $%d", n))
		args = append(args, *filtro.MetaID)
		n++
	}
	if filtro.ResponsavelID != nil {
		condicoes = append(condicoes, fmt.Sprintf("resp.coordenador_id = $%d", n))
		args = append(args, *filtro.ResponsavelID)
		n++
	}
	if filtro.Autoavaliado != nil {
		if *filtro.Autoavaliado {
			condicoes = append(condicoes, "resp.coordenador_id IS NOT NULL AND EXISTS (SELECT 1 FROM entrega ea WHERE ea.item_plano_id = item_plano.id AND ea.excluido_em IS NULL AND ea.avaliada_por = resp.coordenador_id)")
		} else {
			condicoes = append(condicoes, "NOT (resp.coordenador_id IS NOT NULL AND EXISTS (SELECT 1 FROM entrega ea WHERE ea.item_plano_id = item_plano.id AND ea.excluido_em IS NULL AND ea.avaliada_por = resp.coordenador_id))")
		}
	}
	// Filtro por indicador e por origem — SEMPRE por EXISTS, nunca por
	// junção (M-11, RD-10): semi-junção não multiplica, e é isso que torna
	// a deduplicação explícita em vez de efeito colateral de DISTINCT.
	if filtro.IndicadorID != nil {
		condicoes = append(condicoes, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM meta_indicador mi WHERE mi.meta_id = item_plano.meta_id AND mi.indicador_id = $%d)", n))
		args = append(args, *filtro.IndicadorID)
		n++
	}
	if filtro.Origem == "inep" || filtro.Origem == "instituicao" {
		origemBanco := "plataforma"
		if filtro.Origem == "instituicao" {
			origemBanco = "instituicao"
		}
		condicoes = append(condicoes, fmt.Sprintf(
			`EXISTS (SELECT 1 FROM meta_indicador mi2 JOIN indicador i2 ON i2.id = mi2.indicador_id
			          WHERE mi2.meta_id = item_plano.meta_id AND i2.escopo = $%d)`, n))
		args = append(args, origemBanco)
		n++
	}
	if filtro.Situacao != "" {
		condicoes = append(condicoes, fmt.Sprintf("%s = $%d", situacaoCalcSQL(hojeArg), n))
		args = append(args, filtro.Situacao)
		n++
	}

	from = `
		FROM item_plano
		JOIN plano ON plano.id = item_plano.plano_id AND plano.excluido_em IS NULL
		JOIN periodo ON periodo.id = plano.periodo_id
		JOIN curso ON curso.id = item_plano.curso_id AND curso.excluido_em IS NULL
		JOIN meta ON meta.id = item_plano.meta_id AND meta.excluido_em IS NULL

		LEFT JOIN LATERAL (
		  SELECT count(*) FILTER (WHERE e.situacao = 'aceita')             AS aceitas,
		         count(*) FILTER (WHERE e.situacao = 'pendente_avaliacao') AS pendentes,
		         count(*) FILTER (WHERE e.situacao = 'recusada')           AS em_correcao
		    FROM entrega e WHERE e.item_plano_id = item_plano.id AND e.excluido_em IS NULL
		) ent ON TRUE
		` + lateralInd + `
		LEFT JOIN LATERAL (
		  SELECT d.coordenador_id, u.nome AS coordenador_nome, d.data_inicio AS desde
		    FROM designacao d JOIN usuario u ON u.id = d.coordenador_id
		   WHERE d.curso_id = item_plano.curso_id AND d.excluido_em IS NULL AND ` + FragmentoDesignacaoVigente("d", 2) + `
		   LIMIT 1
		) resp ON TRUE
		LEFT JOIN LATERAL (
		  SELECT max(d2.data_fim) AS vago_desde
		    FROM designacao d2
		   WHERE d2.curso_id = item_plano.curso_id AND d2.excluido_em IS NULL AND d2.data_fim IS NOT NULL
		) vago ON TRUE
	`
	where = strings.Join(condicoes, " AND ")
	selectCampos = `
		item_plano.id AS item_plano_id, item_plano.curso_id AS curso_id,
		curso.nome AS curso_nome, curso.situacao AS curso_situacao,
		resp.coordenador_id, resp.coordenador_nome, resp.desde AS responsavel_desde, vago.vago_desde,
		meta.nome AS meta_nome, item_plano.quantidade AS exigido,
		coalesce(ent.aceitas,0) AS aceitas, coalesce(ent.pendentes,0) AS pendentes, coalesce(ent.em_correcao,0) AS em_correcao,
		ind.indicadores,
		CASE WHEN resp.coordenador_id IS NOT NULL THEN EXISTS (
		  SELECT 1 FROM entrega ea2 WHERE ea2.item_plano_id = item_plano.id AND ea2.excluido_em IS NULL AND ea2.avaliada_por = resp.coordenador_id
		) ELSE false END AS avaliacao_pelo_proprio_coordenador,
		EXISTS (
		  SELECT 1 FROM entrega ex WHERE ex.item_plano_id = item_plano.id AND ex.excluido_em IS NULL
		    AND resp.coordenador_id IS NOT NULL AND ex.enviada_por <> resp.coordenador_id
		) AS inclui_entregas_anteriores,
		` + situacaoCalcSQL(hojeArg) + ` AS situacao_calc
	`
	return selectCampos, from, where, args, hojeArg, nil
}

type linhaRelatorioDB struct {
	ItemPlanoID      uuid.UUID  `db:"item_plano_id"`
	CursoID          uuid.UUID  `db:"curso_id"`
	CursoNome        string     `db:"curso_nome"`
	CursoSituacao    string     `db:"curso_situacao"`
	CoordenadorID    *uuid.UUID `db:"coordenador_id"`
	CoordenadorNome  *string    `db:"coordenador_nome"`
	ResponsavelDesde *time.Time `db:"responsavel_desde"`
	VagoDesde        *time.Time `db:"vago_desde"`
	MetaNome         string     `db:"meta_nome"`
	Exigido          int        `db:"exigido"`
	Aceitas          int        `db:"aceitas"`
	Pendentes        int        `db:"pendentes"`
	EmCorrecao       int        `db:"em_correcao"`
	Indicadores      *string    `db:"indicadores"`
	Autoavaliado     bool       `db:"avaliacao_pelo_proprio_coordenador"`
	InclusAnteriores bool       `db:"inclui_entregas_anteriores"`
	SituacaoCalc     string     `db:"situacao_calc"`
}

func (l linhaRelatorioDB) paraLinha() (port.LinhaRelatorio, error) {
	indicadores, err := indicadoresDeJSON(l.Indicadores)
	if err != nil {
		return port.LinhaRelatorio{}, err
	}
	cumprimento := 0.0
	if l.Exigido > 0 {
		cumprimento = float64(l.Aceitas) / float64(l.Exigido) * 100
		if cumprimento > 100 {
			cumprimento = 100
		}
	}
	var responsavelDesde *string
	if l.ResponsavelDesde != nil {
		txt := l.ResponsavelDesde.Format("02/01/2006")
		responsavelDesde = &txt
	}
	var vagoDesde *string
	if l.VagoDesde != nil {
		txt := l.VagoDesde.AddDate(0, 0, 1).Format("02/01/2006")
		vagoDesde = &txt
	}
	return port.LinhaRelatorio{
		ItemPlanoID: l.ItemPlanoID, CursoNome: l.CursoNome, CursoVago: l.CoordenadorID == nil,
		ResponsavelNome: l.CoordenadorNome, ResponsavelDesde: responsavelDesde, VagoDesde: vagoDesde,
		MetaNome: l.MetaNome, Indicadores: indicadores, Exigido: l.Exigido, Aceitas: l.Aceitas,
		Pendentes: l.Pendentes, EmCorrecao: l.EmCorrecao, Cumprimento: cumprimento, Situacao: l.SituacaoCalc,
		InclusEntregasAnteriores: l.InclusAnteriores, AvaliacaoPeloProprioCoordenador: l.Autoavaliado,
	}, nil
}

// resumoSQL — a ÚNICA ocorrência sancionada de DISTINCT nesta feature
// (decisão do arquiteto, T-120). A consulta de `montarRelatorioBase` tem
// grão de item de plano (curso × meta) — o resumo conta CURSOS, uma
// entidade de grão mais grosso, sobre esse mesmo conjunto. Contar entidade
// sobre um conjunto de grão mais fino exige DISTINCT: aqui ele CONVERTE
// GRÃO, que é o uso correto — nunca esconde multiplicação de junção, que é
// o que M-10 proíbe.
//
// A chave é `curso_id`, nunca `curso_nome`: nome não é identidade — dois
// cursos homônimos (dois campi da mesma instituição, por exemplo) contam
// como um se a chave for o nome, e o relatório sai plausível e falso. A
// regra geral, para qualquer `DISTINCT` de agregado futuro nesta base:
// sancionado quando converte grão, proibido quando esconde cardinalidade
// de junção — e a chave é sempre o identificador da entidade contada,
// nunca um atributo textual.
func resumoSQL(selectCampos, from, where string) string {
	return `SELECT count(DISTINCT CASE WHEN coordenador_id IS NULL THEN curso_id END) AS cursos_sem_coordenador,
	               count(*) FILTER (WHERE coordenador_id IS NULL AND situacao_calc <> 'cumprida') AS metas_nao_cumpridas_de_vagos
	          FROM (SELECT ` + selectCampos + ` ` + from + ` WHERE ` + where + `) resumo_base`
}

func (r *EntregaRepository) RelatorioDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio) (port.ResultadoRelatorio, error) {
	if filtro.PeriodoID == uuid.Nil {
		return port.ResultadoRelatorio{}, domain.ErrPeriodoObrigatorio
	}
	selectCampos, from, where, args, _, err := r.montarRelatorioBase(escopo, filtro)
	if err != nil {
		return port.ResultadoRelatorio{}, err
	}
	ex := Executor(ctx, r.db)

	// V-4: o total é contado SOBRE O MESMO conjunto filtrado, sem repetir
	// as junções laterais (que não mudam a cardinalidade, mas custariam
	// caro à toa) — FROM/WHERE são os mesmos, count(*) substitui o SELECT.
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) "+from+" WHERE "+where, args...); err != nil {
		return port.ResultadoRelatorio{}, err
	}

	// O resumo é calculado sobre o CONJUNTO FILTRADO INTEIRO, nunca sobre
	// a página (design.md §9.2).
	var resumo struct {
		CursosSemCoordenador     int `db:"cursos_sem_coordenador"`
		MetasNaoCumpridasDeVagos int `db:"metas_nao_cumpridas_de_vagos"`
	}
	if err := sqlx.GetContext(ctx, ex, &resumo, resumoSQL(selectCampos, from, where), args...); err != nil {
		return port.ResultadoRelatorio{}, err
	}

	sortColuna, ok := ordenacaoRelatorio[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoRelatorio["curso"]
	}
	ordem := "ASC"
	if filtro.Order == "desc" {
		ordem = "DESC"
	}
	page, pageSize := paginaEPageSize(filtro.Page, filtro.PageSize)
	offset := (page - 1) * pageSize
	n := len(args) + 1

	consulta := fmt.Sprintf("SELECT %s %s WHERE %s ORDER BY %s %s, meta.nome COLLATE \"pt-BR-x-icu\" ASC LIMIT $%d OFFSET $%d",
		selectCampos, from, where, sortColuna, ordem, n, n+1)
	argsPaginados := append(append([]any{}, args...), pageSize, offset)

	var linhas []linhaRelatorioDB
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoRelatorio{}, err
	}
	itens := make([]port.LinhaRelatorio, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraLinha()
		if err != nil {
			return port.ResultadoRelatorio{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoRelatorio{
		Itens: itens, Total: total,
		Resumo: port.ResumoRelatorio{CursosSemCoordenador: resumo.CursosSemCoordenador, MetasNaoCumpridasDeVagos: resumo.MetasNaoCumpridasDeVagos},
	}, nil
}

// porCursoAgregadoSQL — CTE única, reaproveitando o MESMO selectCampos/
// from/where de montarRelatorioBase (design.md §9.4): a divisão em
// ranking/vagos/totais acontece DEPOIS da agregação, nunca por uma
// segunda consulta com FROM/WHERE próprio — é essa reutilização que
// impede o gráfico e a tabela de contarem universos diferentes.
//
// O cap por item (LEAST(aceitas, exigido)) acontece ANTES da soma, dentro
// da própria CTE — decisão do dono (ux.md): um item entregue acima do
// exigido não pode compensar o déficit de outro item do mesmo curso.
// Somar cru (sem o cap) faria um curso com duas metas de 2 comprovantes,
// 4 entregues numa e 0 na outra, aparecer com 100% (4 de 4) quando a
// metade das metas está zerada — o cap corrige para 2 de 4 (50%).
func porCursoAgregadoSQL(selectCampos, from, where string) string {
	return `
		WITH agregado AS (
			SELECT curso_id, curso_nome, coordenador_id, coordenador_nome,
			       sum(exigido) AS exigido_total,
			       sum(LEAST(aceitas, exigido)) AS aceitas_total
			  FROM (SELECT ` + selectCampos + ` ` + from + ` WHERE ` + where + `) base
			 GROUP BY curso_id, curso_nome, coordenador_id, coordenador_nome
		)
	`
}

type linhaPorCursoDB struct {
	CursoID         uuid.UUID `db:"curso_id"`
	CursoNome       string    `db:"curso_nome"`
	ResponsavelNome string    `db:"responsavel_nome"`
	ExigidoTotal    int       `db:"exigido_total"`
	AceitasTotal    int       `db:"aceitas_total"`
}

type cursoVagoDB struct {
	CursoNome string `db:"curso_nome"`
}

// DesempenhoPorCurso — design.md §9.4. Reaproveita montarRelatorioBase
// (mesmo FROM/WHERE de RelatorioDesempenho/ExportarDesempenho) e roda TRÊS
// consultas sobre a mesma CTE: ranking (top 10, só com coordenador), a
// lista de nomes dos cursos vagos, e os totais do universo inteiro — nunca
// sobre uma página, nunca por uma consulta com FROM/WHERE próprio.
func (r *EntregaRepository) DesempenhoPorCurso(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio) (port.ResultadoDesempenhoPorCurso, error) {
	if filtro.PeriodoID == uuid.Nil {
		return port.ResultadoDesempenhoPorCurso{}, domain.ErrPeriodoObrigatorio
	}
	selectCampos, from, where, args, _, err := r.montarRelatorioBase(escopo, filtro)
	if err != nil {
		return port.ResultadoDesempenhoPorCurso{}, err
	}
	ex := Executor(ctx, r.db)
	cte := porCursoAgregadoSQL(selectCampos, from, where)

	// Ranking: só cursos com coordenador, top 10, decrescente por
	// percentual (calculado a partir dos totais já capados por item),
	// desempate por nome com collation pt-BR (design.md §9.1).
	var ranking []linhaPorCursoDB
	consultaRanking := cte + `
		SELECT curso_id, curso_nome, coordenador_nome AS responsavel_nome, exigido_total, aceitas_total
		  FROM agregado
		 WHERE coordenador_id IS NOT NULL
		 ORDER BY CASE WHEN exigido_total = 0 THEN 0 ELSE least(1.0, aceitas_total::numeric / exigido_total) END DESC,
		          curso_nome COLLATE "pt-BR-x-icu" ASC
		 LIMIT 10`
	if err := sqlx.SelectContext(ctx, ex, &ranking, consultaRanking, args...); err != nil {
		return port.ResultadoDesempenhoPorCurso{}, err
	}

	// Cursos sem coordenador que bateram no filtro — fora do ranking, só
	// os nomes (ux.md: "Consulte-os na tabela abaixo"). Sem DISTINCT: a
	// CTE `agregado` já tem grão de curso (GROUP BY curso_id) — cada curso
	// já aparece uma única vez, não há cardinalidade a deduplicar aqui.
	var vagos []cursoVagoDB
	consultaVagos := cte + `
		SELECT curso_nome
		  FROM agregado
		 WHERE coordenador_id IS NULL
		 ORDER BY curso_nome COLLATE "pt-BR-x-icu" ASC`
	if err := sqlx.SelectContext(ctx, ex, &vagos, consultaVagos, args...); err != nil {
		return port.ResultadoDesempenhoPorCurso{}, err
	}

	// Totais do universo inteiro (não só o top 10). Mesmo motivo acima:
	// `agregado` já está no grão de curso, então count(*) FILTER basta —
	// nenhum DISTINCT precisa entrar aqui (T-120 continua valendo só para
	// resumoSQL, que agrega sobre o grão de ITEM, não sobre esta CTE).
	var totais struct {
		TotalComCoordenador int `db:"total_cursos_com_coordenador"`
		TotalSemCoordenador int `db:"total_cursos_sem_coordenador"`
	}
	consultaTotais := cte + `
		SELECT count(*) FILTER (WHERE coordenador_id IS NOT NULL) AS total_cursos_com_coordenador,
		       count(*) FILTER (WHERE coordenador_id IS NULL) AS total_cursos_sem_coordenador
		  FROM agregado`
	if err := sqlx.GetContext(ctx, ex, &totais, consultaTotais, args...); err != nil {
		return port.ResultadoDesempenhoPorCurso{}, err
	}

	itens := make([]port.ItemDesempenhoPorCurso, 0, len(ranking))
	for _, l := range ranking {
		percentual := 0.0
		if l.ExigidoTotal > 0 {
			percentual = float64(l.AceitasTotal) / float64(l.ExigidoTotal) * 100
			if percentual > 100 {
				percentual = 100
			}
		}
		itens = append(itens, port.ItemDesempenhoPorCurso{
			CursoID: l.CursoID, CursoNome: l.CursoNome, ResponsavelNome: l.ResponsavelNome,
			ExigidoTotal: l.ExigidoTotal, AceitasTotal: l.AceitasTotal, Percentual: percentual,
		})
	}
	nomesVagos := make([]string, 0, len(vagos))
	for _, v := range vagos {
		nomesVagos = append(nomesVagos, v.CursoNome)
	}
	return port.ResultadoDesempenhoPorCurso{
		Itens: itens, CursosSemCoordenador: nomesVagos,
		TotalCursosComCoordenador: totais.TotalComCoordenador, TotalCursosSemCoordenador: totais.TotalSemCoordenador,
	}, nil
}

// ExportarDesempenho — streaming com cursor real (design.md §9.3): cada
// linha lida do *sql.Rows é imediatamente entregue ao callback, nunca
// acumulada num slice. UTF-8 com BOM e separador ";" são responsabilidade
// do use case/handler; aqui só a leitura em fluxo.
func (r *EntregaRepository) ExportarDesempenho(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroRelatorio, linha func(port.LinhaCSVRelatorio) error) (int, error) {
	if filtro.PeriodoID == uuid.Nil {
		return 0, domain.ErrPeriodoObrigatorio
	}
	selectCampos, from, where, args, _, err := r.montarRelatorioBase(escopo, filtro)
	if err != nil {
		return 0, err
	}
	consulta := fmt.Sprintf("SELECT %s %s WHERE %s ORDER BY curso.nome COLLATE \"pt-BR-x-icu\" ASC, meta.nome COLLATE \"pt-BR-x-icu\" ASC",
		selectCampos, from, where)

	linhasSQL, err := r.db.QueryxContext(ctx, consulta, args...)
	if err != nil {
		return 0, err
	}
	defer linhasSQL.Close()

	total := 0
	for linhasSQL.Next() {
		var l linhaRelatorioDB
		if err := linhasSQL.StructScan(&l); err != nil {
			return total, err
		}
		item, err := l.paraLinha()
		if err != nil {
			return total, err
		}
		if err := linha(item); err != nil {
			return total, err
		}
		total++
	}
	return total, linhasSQL.Err()
}
