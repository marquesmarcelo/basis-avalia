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

var ordenacaoIndicador = map[string]string{
	"codigo":    "codigo",
	"nome":      `nome COLLATE "pt-BR-x-icu"`,
	"escopo":    "escopo",
	"criado_em": "criado_em",
}

type linhaIndicador struct {
	ID                    uuid.UUID  `db:"id"`
	Escopo                string     `db:"escopo"`
	InstituicaoID         *uuid.UUID `db:"instituicao_id"`
	Codigo                string     `db:"codigo"`
	Nome                  string     `db:"nome"`
	Descricao             string     `db:"descricao"`
	ReferenciaInstrumento *string    `db:"referencia_instrumento"`
	Situacao              string     `db:"situacao"`
	CriadoEm              time.Time  `db:"criado_em"`
	AtualizadoEm          *time.Time `db:"atualizado_em"`
	ExcluidoEm            *time.Time `db:"excluido_em"`
	Versao                int        `db:"versao"`
}

func (l linhaIndicador) paraDominio() (indicador.Indicador, error) {
	escopo, err := valueobject.NovoEscopoIndicador(l.Escopo)
	if err != nil {
		return indicador.Indicador{}, err
	}
	codigo, err := valueobject.NovoCodigoIndicador(l.Codigo)
	if err != nil {
		return indicador.Indicador{}, err
	}
	nome, err := valueobject.NovoNomeCatalogo(l.Nome)
	if err != nil {
		return indicador.Indicador{}, err
	}
	situacao, err := valueobject.NovaSituacaoCatalogo(l.Situacao)
	if err != nil {
		return indicador.Indicador{}, err
	}
	var referencia *valueobject.ReferenciaInstrumento
	if l.ReferenciaInstrumento != nil {
		r, err := valueobject.NovaReferenciaInstrumento(*l.ReferenciaInstrumento)
		if err != nil {
			return indicador.Indicador{}, err
		}
		referencia = &r
	}
	return indicador.Indicador{
		ID: l.ID, Escopo: escopo, InstituicaoID: l.InstituicaoID, Codigo: codigo, Nome: nome,
		Descricao: l.Descricao, ReferenciaInstrumento: referencia, Situacao: situacao,
		CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
	}, nil
}

// contagemMetasTotal implementa design.md §7.2 — atravessa a fronteira
// institucional de propósito: é um número agregado sobre todas as
// instituições, que não identifica nenhuma (PI-5).
const contagemMetasTotal = `
	LEFT JOIN LATERAL (
	  SELECT count(*) AS total FROM meta_indicador mi
	    JOIN meta m ON m.id = mi.meta_id
	   WHERE mi.indicador_id = indicador.id AND m.excluido_em IS NULL
	) c_total ON TRUE`

type IndicadorPlataformaRepository struct {
	db *sqlx.DB
}

func NovoIndicadorPlataformaRepository(db *sqlx.DB) *IndicadorPlataformaRepository {
	return &IndicadorPlataformaRepository{db: db}
}

var _ port.IndicadorPlataformaRepository = (*IndicadorPlataformaRepository)(nil)

func (r *IndicadorPlataformaRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemIndicadorPlataforma, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoIndicador, 2)
	if err != nil {
		return port.ItemIndicadorPlataforma{}, err
	}
	consulta := "SELECT indicador.*, c_total.total AS metas_total FROM indicador " +
		contagemMetasTotal + " WHERE indicador.id = $1 AND " + clausula
	todosArgs := append([]any{id}, args...)

	ex := Executor(ctx, r.db)
	var linha struct {
		linhaIndicador
		MetasTotal int `db:"metas_total"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, todosArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemIndicadorPlataforma{}, domain.ErrNaoEncontrado
		}
		return port.ItemIndicadorPlataforma{}, err
	}
	ind, err := linha.paraDominio()
	if err != nil {
		return port.ItemIndicadorPlataforma{}, err
	}
	return port.ItemIndicadorPlataforma{Indicador: ind, MetasTotal: linha.MetasTotal}, nil
}

func (r *IndicadorPlataformaRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarIndicadoresPlataforma) (port.ResultadoListaIndicadoresPlataforma, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoIndicador, 1)
	if err != nil {
		return port.ResultadoListaIndicadoresPlataforma{}, err
	}
	n := len(args) + 1
	condicoes := []string{clausula}

	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(indicador.codigo)) LIKE unaccent(lower($%d)) OR unaccent(lower(indicador.nome)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
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
		return port.ResultadoListaIndicadoresPlataforma{}, err
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

	argsPaginados := append(append([]any{}, args...), pageSize, offset)
	consulta := fmt.Sprintf("SELECT indicador.*, c_total.total AS metas_total FROM indicador %s WHERE %s ORDER BY indicador.%s %s LIMIT $%d OFFSET $%d",
		contagemMetasTotal, where, sortColuna, ordem, n, n+1)

	var linhas []struct {
		linhaIndicador
		MetasTotal int `db:"metas_total"`
	}
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaIndicadoresPlataforma{}, err
	}

	itens := make([]port.ItemIndicadorPlataforma, 0, len(linhas))
	for _, l := range linhas {
		ind, err := l.paraDominio()
		if err != nil {
			return port.ResultadoListaIndicadoresPlataforma{}, err
		}
		itens = append(itens, port.ItemIndicadorPlataforma{Indicador: ind, MetasTotal: l.MetasTotal})
	}
	return port.ResultadoListaIndicadoresPlataforma{Itens: itens, Total: total}, nil
}

func (r *IndicadorPlataformaRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador) error {
	if !escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var referencia any
	if i.ReferenciaInstrumento != nil {
		referencia = i.ReferenciaInstrumento.String()
	}
	_, err := ex.ExecContext(ctx,
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, descricao, referencia_instrumento, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		i.ID, string(i.Escopo), i.InstituicaoID, i.Codigo.String(), i.Nome.String(), i.Descricao, referencia, string(i.Situacao), i.CriadoEm, i.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_indicador_instituicao_codigo") {
			return domain.ErrCodigoIndicadorDuplicado
		}
		return err
	}
	return nil
}

func (r *IndicadorPlataformaRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *indicador.Indicador, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoIndicador, 8)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	var referencia any
	if i.ReferenciaInstrumento != nil {
		referencia = i.ReferenciaInstrumento.String()
	}
	args := append([]any{i.Codigo.String(), i.Nome.String(), i.Descricao, referencia, agora, i.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE indicador SET codigo=$1, nome=$2, descricao=$3, referencia_instrumento=$4, atualizado_em=$5, versao=versao+1 "+
			"WHERE id=$6 AND versao=$7 AND "+clausula,
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

func (r *IndicadorPlataformaRepository) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCatalogo, versaoEsperada int) error {
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

// ExcluirSeSemUso implementa IE-07: trava a linha (FOR UPDATE), conta o
// uso TOTAL (todas as instituições, PI-5) e só então decide — dentro da
// mesma transação, sem janela entre contar e excluir.
func (r *IndicadorPlataformaRepository) ExcluirSeSemUso(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
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
		 WHERE mi.indicador_id = $1 AND m.excluido_em IS NULL`, id); err != nil {
		return err
	}
	if total > 0 {
		return &domain.ErrIndicadorComMeta{Total: total}
	}

	_, err = ex.ExecContext(ctx, `UPDATE indicador SET excluido_em = now() WHERE id = $1`, id)
	return err
}
