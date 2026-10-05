package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// indicadoresEmbutidosDeJSON reproduz a mesma decodificação de
// linhaMeta.paraItem (meta_repository.go), a partir da coluna agregada
// que lateralIndicadoresDaMeta produz — reusada aqui para os itens do
// plano exibirem os indicadores da meta (IT-03), sem reimplementar a
// junção lateral em si.
func indicadoresEmbutidosDeJSON(bruto *string) ([]port.IndicadorDaMeta, error) {
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
		referencia := ""
		if e.ReferenciaInstrumento != nil {
			referencia = *e.ReferenciaInstrumento
		}
		ricos = append(ricos, port.IndicadorDaMeta{
			ID: id, Codigo: e.Codigo, Nome: e.Nome, Escopo: e.Escopo,
			ReferenciaInstrumento: referencia, Situacao: e.Situacao,
		})
	}
	return ricos, nil
}

var ordenacaoPlano = map[string]string{
	"curso":     `curso.nome COLLATE "pt-BR-x-icu"`,
	"periodo":   `periodo.data_inicio`,
	"situacao":  "ordem_situacao",
	"criado_em": "plano.criado_em",
}

// clausulaVago é o predicado "sem designação vigente na data de
// referência" (fundacao-metas.md §5.3) — reusa o único fragmento de
// vigência, nunca reescrito à mão aqui.
func clausulaVagoDoCurso(aliasCurso string, n int) string {
	return fmt.Sprintf(
		"NOT EXISTS (SELECT 1 FROM designacao d WHERE d.curso_id = %s.id AND d.excluido_em IS NULL AND %s)",
		aliasCurso, FragmentoDesignacaoVigente("d", n))
}

// clausulaOrdemSituacao só ordena a apresentação — nunca decide quais
// linhas voltam (isso é WHERE, sempre via AplicarEscopo). Espelha
// plano.SituacaoEfetiva em SQL só para o ORDER BY: rascunho, vigente,
// encerrado, nessa ordem.
func clausulaOrdemSituacao(n int) string {
	return fmt.Sprintf(`CASE
		WHEN plano.situacao_publicacao = 'rascunho' THEN 0
		WHEN plano.encerrado_em IS NOT NULL OR periodo.data_fim < $%d THEN 2
		ELSE 1
	END`, n)
}

func traduzirFiltroSituacao(situacao string, n int) (string, bool) {
	switch situacao {
	case "", "todas":
		return "", false
	case "rascunho":
		return "plano.situacao_publicacao = 'rascunho'", true
	case "vigente":
		return fmt.Sprintf("plano.situacao_publicacao = 'vigente' AND plano.encerrado_em IS NULL AND periodo.data_fim >= $%d", n), true
	case "encerrado":
		return fmt.Sprintf("plano.situacao_publicacao = 'vigente' AND (plano.encerrado_em IS NOT NULL OR periodo.data_fim < $%d)", n), true
	default:
		return "", false
	}
}

type linhaPlanoDB struct {
	ID                  uuid.UUID  `db:"id"`
	InstituicaoID       uuid.UUID  `db:"instituicao_id"`
	CursoID             uuid.UUID  `db:"curso_id"`
	PeriodoID           uuid.UUID  `db:"periodo_id"`
	Descricao           string     `db:"descricao"`
	ObjetivoGeral       string     `db:"objetivo_geral"`
	ResultadosEsperados string     `db:"resultados_esperados"`
	AlinhamentoPDI      string     `db:"alinhamento_pdi"`
	AlinhamentoPPC      string     `db:"alinhamento_ppc"`
	AprovacaoData       *time.Time `db:"aprovacao_data"`
	AprovacaoOrgao      *string    `db:"aprovacao_orgao"`
	SituacaoPublicacao  string     `db:"situacao_publicacao"`
	EncerradoEm         *time.Time `db:"encerrado_em"`
	EncerramentoMotivo  *string    `db:"encerramento_motivo"`
	CriadoEm            time.Time  `db:"criado_em"`
	AtualizadoEm        *time.Time `db:"atualizado_em"`
	ExcluidoEm          *time.Time `db:"excluido_em"`
	Versao              int        `db:"versao"`

	PeriodoNome     string    `db:"periodo_nome"`
	PeriodoInicio   time.Time `db:"periodo_data_inicio"`
	PeriodoFim      time.Time `db:"periodo_data_fim"`
	CursoNome       string    `db:"curso_nome"`
	CursoGrau       string    `db:"curso_grau"`
	CursoModalidade string    `db:"curso_modalidade"`
	CursoCodigoEMec *string   `db:"curso_codigo_emec"`
	Vago            bool      `db:"vago"`
	TotalItens      int       `db:"total_itens"`
	TotalExigido    int       `db:"total_exigido"`
}

func (l linhaPlanoDB) paraLinha(hoje valueobject.DataLocal) (port.LinhaPlano, error) {
	var aprovacao *valueobject.Aprovacao
	if l.AprovacaoData != nil && l.AprovacaoOrgao != nil {
		a, _, err := valueobject.NovaAprovacao(
			valueobject.DataLocalDe(*l.AprovacaoData, time.UTC).String(), *l.AprovacaoOrgao)
		if err != nil {
			return port.LinhaPlano{}, err
		}
		aprovacao = &a
	}
	situacaoPub, err := situacaoPublicacaoDeTexto(l.SituacaoPublicacao)
	if err != nil {
		return port.LinhaPlano{}, err
	}
	encerramentoMotivo := ""
	if l.EncerramentoMotivo != nil {
		encerramentoMotivo = *l.EncerramentoMotivo
	}
	pl := plano.Plano{
		ID: l.ID, InstituicaoID: l.InstituicaoID, CursoID: l.CursoID, PeriodoID: l.PeriodoID,
		Descricao: l.Descricao, ObjetivoGeral: l.ObjetivoGeral, ResultadosEsperados: l.ResultadosEsperados,
		AlinhamentoPDI: l.AlinhamentoPDI, AlinhamentoPPC: l.AlinhamentoPPC, Aprovacao: aprovacao,
		SituacaoPublicacao: situacaoPub, EncerradoEm: l.EncerradoEm, EncerramentoMotivo: encerramentoMotivo,
		CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
	}
	fim := valueobject.DataLocalDe(l.PeriodoFim, time.UTC)
	inicio := valueobject.DataLocalDe(l.PeriodoInicio, time.UTC)
	vigencia, err := valueobject.NovaVigencia(inicio, &fim)
	if err != nil {
		return port.LinhaPlano{}, err
	}
	per := periodo.Periodo{Vigencia: vigencia}
	codigoEMec := ""
	if l.CursoCodigoEMec != nil {
		codigoEMec = *l.CursoCodigoEMec
	}
	return port.LinhaPlano{
		Plano: pl, CursoNome: l.CursoNome, CursoGrau: l.CursoGrau, CursoModalidade: l.CursoModalidade,
		CursoCodigoEMec: codigoEMec, CursoVago: l.Vago, PeriodoNome: l.PeriodoNome,
		Situacao: string(pl.SituacaoEfetiva(per, hoje)), SemAprovacao: pl.SemAprovacao(per, hoje),
		TotalItens: l.TotalItens, TotalExigido: l.TotalExigido,
		// TemEntrega: sempre false até specs/metas-coordenacao criar a
		// tabela entrega — ver port/plano_repository.go e
		// testes-pendentes.md.
		TemEntrega: false,
	}, nil
}

func situacaoPublicacaoDeTexto(bruta string) (valueobject.SituacaoPlano, error) {
	switch valueobject.SituacaoPlano(bruta) {
	case valueobject.PlanoRascunho:
		return valueobject.PlanoRascunho, nil
	case valueobject.PlanoVigente:
		return valueobject.PlanoVigente, nil
	default:
		return "", domain.ErrValorInvalido
	}
}

const colunasPlanoJunta = `
	plano.*,
	periodo.nome AS periodo_nome, periodo.data_inicio AS periodo_data_inicio, periodo.data_fim AS periodo_data_fim,
	curso.nome AS curso_nome, curso.grau AS curso_grau, curso.modalidade AS curso_modalidade, curso.codigo_emec AS curso_codigo_emec,
	%s AS vago,
	(SELECT count(*) FROM item_plano WHERE item_plano.plano_id = plano.id AND item_plano.excluido_em IS NULL) AS total_itens,
	(SELECT coalesce(sum(quantidade), 0) FROM item_plano WHERE item_plano.plano_id = plano.id AND item_plano.excluido_em IS NULL) AS total_exigido
	FROM plano
	JOIN periodo ON periodo.id = plano.periodo_id
	JOIN curso ON curso.id = plano.curso_id
`

type PlanoRepository struct{ db *sqlx.DB }

func NovoPlanoRepository(db *sqlx.DB) *PlanoRepository { return &PlanoRepository{db: db} }

var _ port.PlanoRepository = (*PlanoRepository)(nil)

func (r *PlanoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.DetalhePlano, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPlano, 3)
	if err != nil {
		return port.DetalhePlano{}, err
	}
	hojeTexto := escopo.DataDeReferencia().String()
	args := append([]any{id, hojeTexto}, escopoArgs...)
	consulta := "SELECT " + fmt.Sprintf(colunasPlanoJunta, clausulaVagoDoCurso("curso", 2)) +
		" WHERE plano.id = $1 AND " + clausula
	ex := Executor(ctx, r.db)
	var linha linhaPlanoDB
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.DetalhePlano{}, domain.ErrNaoEncontrado
		}
		return port.DetalhePlano{}, err
	}
	base, err := linha.paraLinha(escopo.DataDeReferencia())
	if err != nil {
		return port.DetalhePlano{}, err
	}

	itens, err := r.listarItens(ctx, escopo, id)
	if err != nil {
		return port.DetalhePlano{}, err
	}

	coordenadorNome, coordenadorPortaria, err := r.coordenadorNaData(ctx, linha.CursoID, escopo.DataDeReferencia())
	if err != nil {
		return port.DetalhePlano{}, err
	}

	ultimoDocumento, err := r.ultimoDocumento(ctx, escopo, id)
	if err != nil {
		return port.DetalhePlano{}, err
	}
	base.UltimoDocumento = ultimoDocumento

	return port.DetalhePlano{LinhaPlano: base, CoordenadorNome: coordenadorNome, CoordenadorPortaria: coordenadorPortaria, Itens: itens}, nil
}

// ultimoDocumento é a lateral de "último documento" que substitui
// plano.documento_id (P-11, design.md §7.4) — sustentada pelo índice
// idx_documento_plano (plano_id, gerado_em DESC), sem varredura
// sequencial (V-9).
func (r *PlanoRepository) ultimoDocumento(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) (*port.ResumoDocumento, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoDocumento, 2)
	if err != nil {
		return nil, err
	}
	args = append([]any{planoID}, args...)
	ex := Executor(ctx, r.db)
	var linha struct {
		ID       uuid.UUID `db:"id"`
		GeradoEm time.Time `db:"gerado_em"`
	}
	consulta := "SELECT documento.id, documento.gerado_em FROM documento WHERE documento.plano_id = $1 AND " +
		clausula + " ORDER BY documento.gerado_em DESC LIMIT 1"
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &port.ResumoDocumento{ID: linha.ID, GeradoEm: linha.GeradoEm}, nil
}

// coordenadorNaData devolve nil, nil quando o curso está vago — "Sem
// responsável" é decidido pelo handler/gerador, nunca aqui.
func (r *PlanoRepository) coordenadorNaData(ctx context.Context, cursoID uuid.UUID, hoje valueobject.DataLocal) (*string, *string, error) {
	ex := Executor(ctx, r.db)
	consulta := fmt.Sprintf(
		`SELECT u.nome, d.portaria FROM designacao d
		 JOIN usuario u ON u.id = d.coordenador_id
		 WHERE d.curso_id = $1 AND d.excluido_em IS NULL AND %s
		 LIMIT 1`, FragmentoDesignacaoVigente("d", 2))
	var linha struct {
		Nome     string `db:"nome"`
		Portaria string `db:"portaria"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, cursoID, hoje.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	return &linha.Nome, &linha.Portaria, nil
}

type linhaItemDB struct {
	ID          uuid.UUID `db:"id"`
	MetaID      uuid.UUID `db:"meta_id"`
	MetaNome    string    `db:"meta_nome"`
	Quantidade  int       `db:"quantidade"`
	Versao      int       `db:"versao"`
	Indicadores *string   `db:"indicadores"`
}

func (r *PlanoRepository) listarItens(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) ([]port.ItemDoPlanoResponse, error) {
	// SemCarteira: o plano_id já foi confirmado acessível por
	// AplicarEscopo(AlvoPlano) em BuscarPorID, e o indicador é catálogo da
	// instituição, não do curso — AlvoIndicador não tem coluna de curso
	// (ver Escopo.SemCarteira).
	lateralSQL, lateralArgs, _, err := lateralIndicadoresDaMeta(escopo.SemCarteira(), 2)
	if err != nil {
		return nil, err
	}
	args := append([]any{planoID}, lateralArgs...)
	consulta := `SELECT item_plano.id, item_plano.meta_id, meta.nome AS meta_nome, item_plano.quantidade,
	                    item_plano.versao, ind.indicadores
	               FROM item_plano
	               JOIN meta ON meta.id = item_plano.meta_id ` + lateralSQL + `
	              WHERE item_plano.plano_id = $1 AND item_plano.excluido_em IS NULL
	              ORDER BY meta.nome COLLATE "pt-BR-x-icu"`
	ex := Executor(ctx, r.db)
	var linhas []linhaItemDB
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	itens := make([]port.ItemDoPlanoResponse, 0, len(linhas))
	for _, l := range linhas {
		indicadores, err := indicadoresEmbutidosDeJSON(l.Indicadores)
		if err != nil {
			return nil, err
		}
		itens = append(itens, port.ItemDoPlanoResponse{
			ID: l.ID, MetaID: l.MetaID, MetaNome: l.MetaNome, Indicadores: indicadores,
			Quantidade: l.Quantidade, Versao: l.Versao,
			// TemEntrega: sempre false — ver comentário de LinhaPlano.TemEntrega.
			TemEntrega: false,
		})
	}
	return itens, nil
}

func (r *PlanoRepository) montarWhere(escopo autorizacao.Escopo, filtro port.FiltroListarPlanos, n int) (string, []any, int, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoPlano, n)
	if err != nil {
		return "", nil, 0, err
	}
	n += len(args)
	condicoes := []string{clausula}
	hojeUsado := false
	hojeTexto := escopo.DataDeReferencia().String()

	if filtro.PeriodoID != nil {
		condicoes = append(condicoes, fmt.Sprintf("plano.periodo_id = $%d", n))
		args = append(args, *filtro.PeriodoID)
		n++
	}
	if filtro.CursoID != nil {
		condicoes = append(condicoes, fmt.Sprintf("plano.curso_id = $%d", n))
		args = append(args, *filtro.CursoID)
		n++
	}
	if situacaoSQL, ok := traduzirFiltroSituacao(filtro.Situacao, n); ok {
		condicoes = append(condicoes, situacaoSQL)
		if strings.Contains(situacaoSQL, "$"+fmt.Sprint(n)) {
			args = append(args, hojeTexto)
			n++
			hojeUsado = true
		}
	}
	switch filtro.Aprovacao {
	case "aprovados":
		condicoes = append(condicoes, "plano.aprovacao_data IS NOT NULL")
	case "sem_aprovacao":
		condicoes = append(condicoes, "plano.aprovacao_data IS NULL AND plano.situacao_publicacao = 'vigente'")
	}
	_ = hojeUsado
	return strings.Join(condicoes, " AND "), args, n, nil
}

func (r *PlanoRepository) listarComWhere(ctx context.Context, escopo autorizacao.Escopo, where string, args []any, n int, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	ex := Executor(ctx, r.db)
	consultaJoins := " FROM plano JOIN periodo ON periodo.id = plano.periodo_id JOIN curso ON curso.id = plano.curso_id "
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*)"+consultaJoins+"WHERE "+where, args...); err != nil {
		return port.ResultadoListaPlanos{}, err
	}

	sortColuna, ok := ordenacaoPlano[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoPlano["curso"]
	}
	ordem := "ASC"
	if filtro.Order == "desc" {
		ordem = "DESC"
	}
	page, pageSize := filtro.Page, filtro.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	hojeTexto := escopo.DataDeReferencia().String()
	argsSelect := append(append([]any{}, args...), hojeTexto)
	nHoje := n
	n++

	selectColunas := fmt.Sprintf(colunasPlanoJunta, clausulaVagoDoCurso("curso", nHoje))
	ordemExpr := sortColuna
	if filtro.Sort == "situacao" {
		ordemExpr = clausulaOrdemSituacao(nHoje)
	}
	consulta := fmt.Sprintf("SELECT %s WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		selectColunas, where, ordemExpr, ordem, n, n+1)
	argsPaginados := append(argsSelect, pageSize, offset)

	var linhas []linhaPlanoDB
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaPlanos{}, err
	}
	itens := make([]port.LinhaPlano, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraLinha(escopo.DataDeReferencia())
		if err != nil {
			return port.ResultadoListaPlanos{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaPlanos{Itens: itens, Total: total}, nil
}

func (r *PlanoRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	where, args, n, err := r.montarWhere(escopo, filtro, 1)
	if err != nil {
		return port.ResultadoListaPlanos{}, err
	}
	return r.listarComWhere(ctx, escopo, where, args, n, filtro)
}

// ListarMeusPlanos usa o MESMO Escopo restrito à carteira (P-10,
// PlanosDaCarteira) — nenhum filtro de situação é aplicado aqui além do
// que o cliente pedir: rascunho aparece, em somente leitura (a escrita é
// bloqueada pela ausência de permissão, não por este filtro).
func (r *PlanoRepository) ListarMeusPlanos(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPlanos) (port.ResultadoListaPlanos, error) {
	return r.Listar(ctx, escopo, filtro)
}

func (r *PlanoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	if err := inserirPlanoRaw(ctx, ex, p); err != nil {
		if violaIndice(err, "uq_plano_curso_periodo") {
			return domain.ErrPlanoDuplicado
		}
		return err
	}
	return nil
}

func inserirPlanoRaw(ctx context.Context, ex ExecutorSQL, p *plano.Plano) error {
	var aprovacaoData, aprovacaoOrgao any
	if p.Aprovacao != nil {
		aprovacaoData = p.Aprovacao.Data().String()
		aprovacaoOrgao = string(p.Aprovacao.Orgao())
	}
	_, err := ex.ExecContext(ctx,
		`INSERT INTO plano (id, instituicao_id, curso_id, periodo_id, descricao, objetivo_geral, resultados_esperados,
		                     alinhamento_pdi, alinhamento_ppc, aprovacao_data, aprovacao_orgao, situacao_publicacao,
		                     criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.InstituicaoID, p.CursoID, p.PeriodoID, p.Descricao, p.ObjetivoGeral, p.ResultadosEsperados,
		p.AlinhamentoPDI, p.AlinhamentoPPC, aprovacaoData, aprovacaoOrgao, string(p.SituacaoPublicacao),
		p.CriadoEm, p.Versao,
	)
	return err
}

func (r *PlanoRepository) AtualizarDados(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPlano, 11)
	if err != nil {
		return err
	}
	var aprovacaoData, aprovacaoOrgao any
	if p.Aprovacao != nil {
		aprovacaoData = p.Aprovacao.Data().String()
		aprovacaoOrgao = string(p.Aprovacao.Orgao())
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{
		p.Descricao, p.ObjetivoGeral, p.ResultadosEsperados, p.AlinhamentoPDI, p.AlinhamentoPPC,
		aprovacaoData, aprovacaoOrgao, agora, p.ID, versaoEsperada,
	}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		`UPDATE plano SET descricao=$1, objetivo_geral=$2, resultados_esperados=$3, alinhamento_pdi=$4,
		        alinhamento_ppc=$5, aprovacao_data=$6, aprovacao_orgao=$7, atualizado_em=$8, versao=versao+1
		  WHERE id=$9 AND versao=$10 AND `+clausula,
		args...,
	)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrConflitoDeVersao
	}
	p.AtualizadoEm = &agora
	p.Versao = versaoEsperada + 1
	return nil
}

func (r *PlanoRepository) AtualizarSituacao(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPlano, 6)
	if err != nil {
		return err
	}
	var encerramentoMotivo any
	if p.EncerramentoMotivo != "" {
		encerramentoMotivo = p.EncerramentoMotivo
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{
		string(p.SituacaoPublicacao), p.EncerradoEm, encerramentoMotivo, agora, p.ID, versaoEsperada,
	}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		`UPDATE plano SET situacao_publicacao=$1, encerrado_em=$2, encerramento_motivo=$3, atualizado_em=$4, versao=versao+1
		  WHERE id=$5 AND versao=$6 AND `+clausula,
		args...,
	)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrConflitoDeVersao
	}
	p.AtualizadoEm = &agora
	p.Versao = versaoEsperada + 1
	return nil
}

func (r *PlanoRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPlano, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx, "UPDATE plano SET excluido_em = now() WHERE id = $1 AND "+clausula, args...)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

// PeriodoDoPlano lê nome e vigência do período, filtrado só por
// instituição — ver comentário em port/plano_repository.go sobre por que
// nunca aplica a restrição de carteira aqui.
func (r *PlanoRepository) PeriodoDoPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (periodo.Periodo, error) {
	if escopo.InstituicaoID() == nil {
		return periodo.Periodo{}, domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var linha struct {
		Nome       string    `db:"nome"`
		DataInicio time.Time `db:"data_inicio"`
		DataFim    time.Time `db:"data_fim"`
	}
	consulta := "SELECT nome, data_inicio, data_fim FROM periodo WHERE id = $1 AND instituicao_id = $2 AND excluido_em IS NULL"
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, periodoID, *escopo.InstituicaoID()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return periodo.Periodo{}, domain.ErrNaoEncontrado
		}
		return periodo.Periodo{}, err
	}
	nome, err := valueobject.NovoNomeCatalogo(linha.Nome)
	if err != nil {
		return periodo.Periodo{}, err
	}
	inicio := valueobject.DataLocalDe(linha.DataInicio, time.UTC)
	fim := valueobject.DataLocalDe(linha.DataFim, time.UTC)
	vigencia, err := valueobject.NovaVigencia(inicio, &fim)
	if err != nil {
		return periodo.Periodo{}, err
	}
	return periodo.Periodo{ID: periodoID, InstituicaoID: *escopo.InstituicaoID(), Nome: nome, Vigencia: vigencia}, nil
}

// InstituicaoDaSessao lê nome e sigla da instituição do Escopo — contexto
// do próprio ator autenticado, nunca um id vindo do cliente, por isso sem
// AplicarEscopo (não há isolamento a aplicar sobre a própria instituição
// da sessão).
func (r *PlanoRepository) InstituicaoDaSessao(ctx context.Context, escopo autorizacao.Escopo) (string, string, error) {
	if escopo.InstituicaoID() == nil {
		return "", "", domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var linha struct {
		Nome  string `db:"nome"`
		Sigla string `db:"sigla"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha, "SELECT nome, sigla FROM instituicao WHERE id = $1", *escopo.InstituicaoID()); err != nil {
		return "", "", err
	}
	return linha.Nome, linha.Sigla, nil
}

// CursoValidoParaPlano consulta a tabela curso diretamente, sob o mesmo
// Escopo (AlvoCurso) — sem depender de CursoRepository/port, que é
// responsabilidade de specs/cursos e ainda não existe como porta pública
// desta aplicação.
func (r *PlanoRepository) CursoValidoParaPlano(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID) (bool, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoCurso, 2)
	if err != nil {
		return false, err
	}
	args = append([]any{cursoID}, args...)
	ex := Executor(ctx, r.db)
	var situacao string
	consulta := "SELECT curso.situacao FROM curso WHERE curso.id = $1 AND " + clausula
	if err := sqlx.GetContext(ctx, ex, &situacao, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, domain.ErrNaoEncontrado
		}
		return false, err
	}
	return situacao == "ativo", nil
}

// ExisteEntregaNoPlano/ExisteEntregaNoItem — ver comentário em
// port/plano_repository.go: sempre false até a tabela entrega existir.
func (r *PlanoRepository) ExisteEntregaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *PlanoRepository) ExisteEntregaNoItem(ctx context.Context, escopo autorizacao.Escopo, itemID uuid.UUID) (bool, error) {
	return false, nil
}

func (r *PlanoRepository) ListarDestinosDeCopia(ctx context.Context, escopo autorizacao.Escopo, planoOrigemID, periodoDestinoID uuid.UUID, busca string) ([]port.DestinoDeCopia, error) {
	clausulaCurso, escopoArgs, err := AplicarEscopo(escopo, AlvoCurso, 3)
	if err != nil {
		return nil, err
	}
	args := append([]any{periodoDestinoID, escopo.DataDeReferencia().String()}, escopoArgs...)
	n := len(args) + 1
	condicoes := []string{clausulaCurso, "curso.situacao = 'ativo'"}
	if busca != "" {
		condicoes = append(condicoes, fmt.Sprintf("unaccent(lower(curso.nome)) LIKE unaccent(lower($%d))", n))
		args = append(args, "%"+escaparCuringasLike(busca)+"%")
		n++
	}
	where := strings.Join(condicoes, " AND ")

	consulta := fmt.Sprintf(`
		SELECT curso.id AS curso_id, curso.nome AS curso_nome,
		       coord.nome AS coordenador_nome,
		       %s AS vago,
		       EXISTS (SELECT 1 FROM plano p WHERE p.curso_id = curso.id AND p.periodo_id = $1 AND p.excluido_em IS NULL) AS ja_tem_plano
		  FROM curso
		  LEFT JOIN designacao des ON des.curso_id = curso.id AND des.excluido_em IS NULL AND %s
		  LEFT JOIN usuario coord ON coord.id = des.coordenador_id
		 WHERE %s
		 ORDER BY curso.nome COLLATE "pt-BR-x-icu"`,
		clausulaVagoDoCurso("curso", 2), FragmentoDesignacaoVigente("des", 2), where)

	ex := Executor(ctx, r.db)
	var linhas []struct {
		CursoID         uuid.UUID `db:"curso_id"`
		CursoNome       string    `db:"curso_nome"`
		CoordenadorNome *string   `db:"coordenador_nome"`
		Vago            bool      `db:"vago"`
		JaTemPlano      bool      `db:"ja_tem_plano"`
	}
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	destinos := make([]port.DestinoDeCopia, 0, len(linhas))
	for _, l := range linhas {
		destinos = append(destinos, port.DestinoDeCopia{
			CursoID: l.CursoID, CursoNome: l.CursoNome, CoordenadorNome: l.CoordenadorNome,
			Vago: l.Vago, JaTemPlano: l.JaTemPlano,
		})
	}
	return destinos, nil
}

// InserirComItens — usada só pela cópia em lote (T-219, design.md §5.3):
// um INSERT do plano e UM INSERT multi-linha dos itens, dentro da
// transação que o use case já abriu por curso.
func (r *PlanoRepository) InserirComItens(ctx context.Context, escopo autorizacao.Escopo, p *plano.Plano, itensOrigem []port.ItemDoPlanoResponse) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	if err := inserirPlanoRaw(ctx, ex, p); err != nil {
		if violaIndice(err, "uq_plano_curso_periodo") {
			return domain.ErrPlanoDuplicado
		}
		return err
	}
	if len(itensOrigem) == 0 {
		return nil
	}
	valores := make([]string, 0, len(itensOrigem))
	args := make([]any, 0, len(itensOrigem)*6)
	n := 1
	for _, item := range itensOrigem {
		valores = append(valores, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)", n, n+1, n+2, n+3, n+4, n+5))
		args = append(args, uuid.Must(uuid.NewV7()), p.ID, p.CursoID, p.InstituicaoID, item.MetaID, item.Quantidade)
		n += 6
	}
	_, err := ex.ExecContext(ctx,
		"INSERT INTO item_plano (id, plano_id, curso_id, instituicao_id, meta_id, quantidade) VALUES "+strings.Join(valores, ","),
		args...,
	)
	return err
}
