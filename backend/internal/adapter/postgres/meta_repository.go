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
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ordenacaoMeta = map[string]string{
	"nome":      `nome COLLATE "pt-BR-x-icu"`,
	"criado_em": "criado_em",
}

type indicadorEmbutidoJSON struct {
	ID                    string  `json:"id"`
	Codigo                string  `json:"codigo"`
	Nome                  string  `json:"nome"`
	Escopo                string  `json:"escopo"`
	ReferenciaInstrumento *string `json:"referencia_instrumento"`
	Situacao              string  `json:"situacao"`
}

// lateralIndicadoresDaMeta monta a junção lateral que agrega os
// indicadores de cada meta como JSON (design.md §7.1) — SEMPRE via
// AplicarEscopo(AlvoIndicador), nunca junção nua (I-09): "a meta já está
// recortada" não é autorização para o indicador.
func lateralIndicadoresDaMeta(escopo autorizacao.Escopo, proximoPlaceholder int) (string, []any, int, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoIndicador, proximoPlaceholder)
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
	   WHERE mi.meta_id = meta.id AND %s
	) ind ON TRUE`, clausula)
	return sqlText, args, proximoPlaceholder + len(args), nil
}

type linhaMeta struct {
	ID                 uuid.UUID  `db:"id"`
	InstituicaoID      uuid.UUID  `db:"instituicao_id"`
	Nome               string     `db:"nome"`
	Descricao          string     `db:"descricao"`
	Situacao           string     `db:"situacao"`
	QuantidadeSugerida *int       `db:"quantidade_sugerida"`
	CriadoEm           time.Time  `db:"criado_em"`
	AtualizadoEm       *time.Time `db:"atualizado_em"`
	ExcluidoEm         *time.Time `db:"excluido_em"`
	Versao             int        `db:"versao"`
	Indicadores        *string    `db:"indicadores"`
}

// paraItem devolve a entidade de domínio (com Indicadores como []uuid.UUID,
// para os use cases de comando reaplicarem DefinirIndicadores) e a lista
// rica para a resposta da API — as duas derivadas da mesma coluna JSON.
func (l linhaMeta) paraItem() (port.ItemMeta, error) {
	situacao, err := valueobject.NovaSituacaoCatalogo(l.Situacao)
	if err != nil {
		return port.ItemMeta{}, err
	}
	nome, err := valueobject.NovoNomeCatalogo(l.Nome)
	if err != nil {
		return port.ItemMeta{}, err
	}
	var quantidadeSugerida *valueobject.Quantidade
	if l.QuantidadeSugerida != nil {
		q, err := valueobject.NovaQuantidade(*l.QuantidadeSugerida)
		if err != nil {
			return port.ItemMeta{}, err
		}
		quantidadeSugerida = &q
	}
	var embutidos []indicadorEmbutidoJSON
	if l.Indicadores != nil {
		if err := json.Unmarshal([]byte(*l.Indicadores), &embutidos); err != nil {
			return port.ItemMeta{}, err
		}
	}
	ids := make([]uuid.UUID, 0, len(embutidos))
	ricos := make([]port.IndicadorDaMeta, 0, len(embutidos))
	for _, e := range embutidos {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			return port.ItemMeta{}, err
		}
		ids = append(ids, id)
		referencia := ""
		if e.ReferenciaInstrumento != nil {
			referencia = *e.ReferenciaInstrumento
		}
		ricos = append(ricos, port.IndicadorDaMeta{
			ID: id, Codigo: e.Codigo, Nome: e.Nome, Escopo: e.Escopo,
			ReferenciaInstrumento: referencia, Situacao: e.Situacao,
		})
	}
	return port.ItemMeta{
		Meta: meta.Meta{
			ID: l.ID, InstituicaoID: l.InstituicaoID, Nome: nome, Descricao: l.Descricao,
			Indicadores: ids, Situacao: situacao, QuantidadeSugerida: quantidadeSugerida,
			CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm,
			ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
		},
		Indicadores: ricos,
		// Planos é sempre 0: item_plano ainda não existe (ver
		// port/meta_repository.go e testes-pendentes.md).
		Planos: 0,
	}, nil
}

type MetaRepository struct {
	db *sqlx.DB
}

func NovoMetaRepository(db *sqlx.DB) *MetaRepository {
	return &MetaRepository{db: db}
}

var _ port.MetaRepository = (*MetaRepository)(nil)

func (r *MetaRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemMeta, error) {
	clausulaMeta, argsMeta, err := AplicarEscopo(escopo, AlvoMeta, 2)
	if err != nil {
		return port.ItemMeta{}, err
	}
	args := append([]any{id}, argsMeta...)
	lateralSQL, lateralArgs, _, err := lateralIndicadoresDaMeta(escopo, len(args)+1)
	if err != nil {
		return port.ItemMeta{}, err
	}
	args = append(args, lateralArgs...)

	consulta := "SELECT meta.*, ind.indicadores FROM meta " + lateralSQL +
		" WHERE meta.id = $1 AND " + clausulaMeta

	ex := Executor(ctx, r.db)
	var linha linhaMeta
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemMeta{}, domain.ErrNaoEncontrado
		}
		return port.ItemMeta{}, err
	}
	return linha.paraItem()
}

func (r *MetaRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarMetas) (port.ResultadoListaMetas, error) {
	clausulaMeta, argsMeta, err := AplicarEscopo(escopo, AlvoMeta, 1)
	if err != nil {
		return port.ResultadoListaMetas{}, err
	}
	args := append([]any{}, argsMeta...)
	n := len(args) + 1
	condicoes := []string{clausulaMeta}

	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf("unaccent(lower(meta.nome)) LIKE unaccent(lower($%d))", n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
	}
	// IndicadorID/Origem: EXISTS direto sobre meta_indicador/indicador — não
	// é leitura de conteúdo do indicador, só verificação de vínculo já
	// legítimo (o indicador entrou na meta validado na escrita). SEMI-JUNÇÃO,
	// nunca junção: junção multiplicaria a meta por indicador casado
	// (MC-14, design.md §7.1) — DISTINCT é proibido porque corrigiria a
	// contagem de linhas e deixaria total/agregados errados.
	if filtro.IndicadorID != nil {
		condicoes = append(condicoes, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM meta_indicador mi WHERE mi.meta_id = meta.id AND mi.indicador_id = $%d)", n))
		args = append(args, *filtro.IndicadorID)
		n++
	}
	switch filtro.Origem {
	case "plataforma":
		condicoes = append(condicoes,
			"EXISTS (SELECT 1 FROM meta_indicador mi JOIN indicador i2 ON i2.id = mi.indicador_id "+
				"WHERE mi.meta_id = meta.id AND i2.escopo = 'plataforma')")
	case "instituicao":
		condicoes = append(condicoes,
			"EXISTS (SELECT 1 FROM meta_indicador mi JOIN indicador i2 ON i2.id = mi.indicador_id "+
				"WHERE mi.meta_id = meta.id AND i2.escopo = 'instituicao')")
	}
	if filtro.Situacao != "" && filtro.Situacao != "todas" {
		condicoes = append(condicoes, fmt.Sprintf("meta.situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM meta WHERE "+where, args...); err != nil {
		return port.ResultadoListaMetas{}, err
	}

	sortColuna, ok := ordenacaoMeta[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoMeta["nome"]
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

	lateralSQL, lateralArgs, proximo, err := lateralIndicadoresDaMeta(escopo, n)
	if err != nil {
		return port.ResultadoListaMetas{}, err
	}
	argsSelect := append(append([]any{}, args...), lateralArgs...)
	consulta := fmt.Sprintf("SELECT meta.*, ind.indicadores FROM meta %s WHERE %s ORDER BY meta.%s %s LIMIT $%d OFFSET $%d",
		lateralSQL, where, sortColuna, ordem, proximo, proximo+1)
	argsPaginados := append(argsSelect, pageSize, offset)

	var linhas []linhaMeta
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaMetas{}, err
	}
	itens := make([]port.ItemMeta, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return port.ResultadoListaMetas{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaMetas{Itens: itens, Total: total}, nil
}

// Sugerir alimenta o autocomplete de meta — só ativas, limitado a 20.
func (r *MetaRepository) Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemSugestaoMeta, error) {
	clausulaMeta, argsMeta, err := AplicarEscopo(escopo, AlvoMeta, 1)
	if err != nil {
		return nil, err
	}
	args := append([]any{}, argsMeta...)
	n := len(args) + 1
	condicoes := []string{clausulaMeta, "meta.situacao = 'ativo'"}
	if busca != "" {
		condicoes = append(condicoes, fmt.Sprintf("unaccent(lower(meta.nome)) LIKE unaccent(lower($%d))", n))
		args = append(args, "%"+escaparCuringasLike(busca)+"%")
		n++
	}
	where := strings.Join(condicoes, " AND ")

	lateralSQL, lateralArgs, _, err := lateralIndicadoresDaMeta(escopo, n)
	if err != nil {
		return nil, err
	}
	args = append(args, lateralArgs...)

	consulta := "SELECT meta.*, ind.indicadores FROM meta " + lateralSQL +
		" WHERE " + where + " ORDER BY meta.nome LIMIT 20"
	ex := Executor(ctx, r.db)
	var linhas []linhaMeta
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	itens := make([]port.ItemSugestaoMeta, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return nil, err
		}
		var quantidadeSugerida *int
		if item.Meta.QuantidadeSugerida != nil {
			v := item.Meta.QuantidadeSugerida.Int()
			quantidadeSugerida = &v
		}
		itens = append(itens, port.ItemSugestaoMeta{
			ID: item.Meta.ID, Nome: item.Meta.Nome.String(), Indicadores: item.Indicadores,
			QuantidadeSugerida: quantidadeSugerida,
		})
	}
	return itens, nil
}

func (r *MetaRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, m *meta.Meta) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var quantidadeSugerida *int
	if m.QuantidadeSugerida != nil {
		v := m.QuantidadeSugerida.Int()
		quantidadeSugerida = &v
	}
	_, err := ex.ExecContext(ctx,
		`INSERT INTO meta (id, instituicao_id, nome, descricao, situacao, quantidade_sugerida, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.InstituicaoID, m.Nome.String(), m.Descricao, string(m.Situacao), quantidadeSugerida, m.CriadoEm, m.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_meta_instituicao_nome") {
			return domain.ErrNomeMetaDuplicado
		}
		return err
	}
	return inserirVinculosDeIndicador(ctx, ex, m.ID, m.Indicadores)
}

func inserirVinculosDeIndicador(ctx context.Context, ex sqlx.ExtContext, metaID uuid.UUID, indicadores []uuid.UUID) error {
	for _, indicadorID := range indicadores {
		if _, err := ex.ExecContext(ctx,
			`INSERT INTO meta_indicador (meta_id, indicador_id) VALUES ($1,$2)`,
			metaID, indicadorID,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *MetaRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, m *meta.Meta, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoMeta, 7)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	var quantidadeSugerida *int
	if m.QuantidadeSugerida != nil {
		v := m.QuantidadeSugerida.Int()
		quantidadeSugerida = &v
	}
	args := append([]any{m.Nome.String(), m.Descricao, quantidadeSugerida, agora, m.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE meta SET nome=$1, descricao=$2, quantidade_sugerida=$3, atualizado_em=$4, versao=versao+1 "+
			"WHERE id=$5 AND versao=$6 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "uq_meta_instituicao_nome") {
			return domain.ErrNomeMetaDuplicado
		}
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrConflitoDeVersao
	}

	// Reescreve o vínculo inteiro — nunca aplica diferença (mesma decisão
	// de usuario_perfil/SubstituirPerfis, design.md §3.3).
	if _, err := ex.ExecContext(ctx, `DELETE FROM meta_indicador WHERE meta_id = $1`, m.ID); err != nil {
		return err
	}
	if err := inserirVinculosDeIndicador(ctx, ex, m.ID, m.Indicadores); err != nil {
		return err
	}

	m.AtualizadoEm = &agora
	m.Versao = versaoEsperada + 1
	return nil
}

func (r *MetaRepository) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoMeta, 5)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{string(nova), agora, id, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE meta SET situacao=$1, atualizado_em=$2, versao=versao+1 WHERE id=$3 AND versao=$4 AND "+clausula,
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
	return nil
}

// Excluir — MC-11 (bloqueio por uso em item_plano) não é verificável
// ainda: a tabela item_plano só nasce em specs/plano-acao. Por ora
// a exclusão lógica não é bloqueada por nenhum uso — ver
// testes-pendentes.md e port/meta_repository.go.
func (r *MetaRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoMeta, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE meta SET excluido_em = now() WHERE id = $1 AND "+clausula,
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
		return domain.ErrNaoEncontrado
	}
	return nil
}
