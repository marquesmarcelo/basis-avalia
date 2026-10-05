package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// contagemMetasDaInstituicao implementa design.md §7.2 — a metade que NÃO
// atravessa a fronteira: só conta metas da instituição da sessão (IV-03).
// O placeholder da instituição é sempre o ÚLTIMO argumento da consulta —
// nunca reaproveita o número que AplicarEscopo já usou, para não abrir
// espaço para um placeholder "furado" (número referenciado no SQL sem
// argumento correspondente, ou vice-versa).
const contagemMetasDaInstituicao = `
	LEFT JOIN LATERAL (
	  SELECT count(*) AS total FROM meta_indicador mi
	    JOIN meta m ON m.id = mi.meta_id
	   WHERE mi.indicador_id = indicador.id
	     AND m.excluido_em IS NULL AND m.instituicao_id = $%d
	) c_inst ON TRUE`

type IndicadorRepository struct {
	db *sqlx.DB
}

func NovoIndicadorRepository(db *sqlx.DB) *IndicadorRepository {
	return &IndicadorRepository{db: db}
}

var _ port.IndicadorRepository = (*IndicadorRepository)(nil)

func (r *IndicadorRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemIndicador, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 2)
	if err != nil {
		return port.ItemIndicador{}, err
	}
	args := append([]any{id}, escopoArgs...)
	instParam := len(args) + 1
	args = append(args, escopo.InstituicaoID())

	consulta := fmt.Sprintf("SELECT indicador.*, c_inst.total AS metas_da_instituicao FROM indicador "+
		contagemMetasDaInstituicao+" WHERE indicador.id = $1 AND %s", instParam, clausula)

	ex := Executor(ctx, r.db)
	var linha struct {
		linhaIndicador
		MetasDaInstituicao int `db:"metas_da_instituicao"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemIndicador{}, domain.ErrNaoEncontrado
		}
		return port.ItemIndicador{}, err
	}
	ind, err := linha.paraDominio()
	if err != nil {
		return port.ItemIndicador{}, err
	}
	return port.ItemIndicador{Indicador: ind, MetasDaInstituicao: linha.MetasDaInstituicao}, nil
}

func (r *IndicadorRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarIndicadores) (port.ResultadoListaIndicadores, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 1)
	if err != nil {
		return port.ResultadoListaIndicadores{}, err
	}
	args := append([]any{}, escopoArgs...)
	n := len(args) + 1

	condicoes := []string{clausula}
	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(indicador.codigo)) LIKE unaccent(lower($%d)) OR unaccent(lower(indicador.nome)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
	}
	switch filtro.Origem {
	case "plataforma":
		condicoes = append(condicoes, "indicador.escopo = 'plataforma'")
	case "instituicao":
		condicoes = append(condicoes, "indicador.escopo = 'instituicao'")
	}
	if filtro.Situacao != "" && filtro.Situacao != "todos" {
		condicoes = append(condicoes, fmt.Sprintf("indicador.situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM indicador WHERE "+where, args...); err != nil {
		return port.ResultadoListaIndicadores{}, err
	}

	sortColuna, ok := ordenacaoIndicador[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoIndicador["codigo"]
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

	instParam := n
	argsSelect := append(append([]any{}, args...), escopo.InstituicaoID())
	consulta := fmt.Sprintf("SELECT indicador.*, c_inst.total AS metas_da_instituicao FROM indicador "+
		contagemMetasDaInstituicao+" WHERE %s ORDER BY indicador.%s %s LIMIT $%d OFFSET $%d",
		instParam, where, sortColuna, ordem, instParam+1, instParam+2)
	argsPaginados := append(argsSelect, pageSize, offset)

	var linhas []struct {
		linhaIndicador
		MetasDaInstituicao int `db:"metas_da_instituicao"`
	}
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaIndicadores{}, err
	}

	itens := make([]port.ItemIndicador, 0, len(linhas))
	for _, l := range linhas {
		ind, err := l.paraDominio()
		if err != nil {
			return port.ResultadoListaIndicadores{}, err
		}
		itens = append(itens, port.ItemIndicador{Indicador: ind, MetasDaInstituicao: l.MetasDaInstituicao})
	}
	return port.ResultadoListaIndicadores{Itens: itens, Total: total}, nil
}

// Sugerir alimenta o autocomplete de indicador (ux.md) — só ativos,
// misturando os dois escopos, limitado a 20 (PI-6).
func (r *IndicadorRepository) Sugerir(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemSugestaoIndicador, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoIndicador, 1)
	if err != nil {
		return nil, err
	}
	n := len(args) + 1
	condicoes := []string{clausula, "indicador.situacao = 'ativo'"}
	if busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(indicador.codigo)) LIKE unaccent(lower($%d)) OR unaccent(lower(indicador.nome)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(busca)+"%")
		n++
	}
	where := strings.Join(condicoes, " AND ")

	type linhaSugestao struct {
		ID                    uuid.UUID `db:"id"`
		Codigo                string    `db:"codigo"`
		Nome                  string    `db:"nome"`
		Escopo                string    `db:"escopo"`
		ReferenciaInstrumento *string   `db:"referencia_instrumento"`
	}
	var linhas []linhaSugestao
	consulta := "SELECT indicador.id, indicador.codigo, indicador.nome, indicador.escopo, indicador.referencia_instrumento " +
		"FROM indicador WHERE " + where + " ORDER BY indicador.codigo LIMIT 20"
	ex := Executor(ctx, r.db)
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	itens := make([]port.ItemSugestaoIndicador, 0, len(linhas))
	for _, l := range linhas {
		referencia := ""
		if l.ReferenciaInstrumento != nil {
			referencia = *l.ReferenciaInstrumento
		}
		itens = append(itens, port.ItemSugestaoIndicador{
			ID: l.ID, Codigo: l.Codigo, Nome: l.Nome, Escopo: l.Escopo, ReferenciaInstrumento: referencia,
		})
	}
	return itens, nil
}

func (r *IndicadorRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, descricao, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		i.ID, string(i.Escopo), i.InstituicaoID, i.Codigo.String(), i.Nome.String(), i.Descricao, string(i.Situacao), i.CriadoEm, i.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_indicador_instituicao_codigo") {
			return domain.ErrCodigoIndicadorDuplicado
		}
		return err
	}
	return nil
}

func (r *IndicadorRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 7)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{i.Codigo.String(), i.Nome.String(), i.Descricao, agora, i.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE indicador SET codigo=$1, nome=$2, descricao=$3, atualizado_em=$4, versao=versao+1 "+
			"WHERE id=$5 AND versao=$6 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "uq_indicador_instituicao_codigo") {
			return domain.ErrCodigoIndicadorDuplicado
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
	i.AtualizadoEm = &agora
	i.Versao = versaoEsperada + 1
	return nil
}

func (r *IndicadorRepository) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 5)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{string(nova), agora, id, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE indicador SET situacao=$1, atualizado_em=$2, versao=versao+1 WHERE id=$3 AND versao=$4 AND "+clausula,
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

// ExcluirSeSemUso implementa IN-05 — mesma mecânica de IE-07 (FOR UPDATE +
// contagem antes de decidir), mas restrita à instituição da sessão.
func (r *IndicadorRepository) ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)

	var existe bool
	args := append([]any{id}, escopoArgs...)
	consultaTrava := "SELECT EXISTS (SELECT 1 FROM indicador WHERE id = $1 AND " + clausula + " FOR UPDATE)"
	if err := sqlx.GetContext(ctx, ex, &existe, consultaTrava, args...); err != nil {
		return err
	}
	if !existe {
		return domain.ErrNaoEncontrado
	}

	var total int
	if err := sqlx.GetContext(ctx, ex, &total, `
		SELECT count(*) FROM meta_indicador mi JOIN meta m ON m.id = mi.meta_id
		 WHERE mi.indicador_id = $1 AND m.excluido_em IS NULL AND m.instituicao_id = $2`, id, escopo.InstituicaoID()); err != nil {
		return err
	}
	if total > 0 {
		return &domain.ErrIndicadorComMeta{Total: total}
	}

	_, err = ex.ExecContext(ctx, `UPDATE indicador SET excluido_em = now() WHERE id = $1`, id)
	return err
}
