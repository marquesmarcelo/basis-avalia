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
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ordenacaoPeriodo = map[string]string{
	"nome":        `nome COLLATE "pt-BR-x-icu"`,
	"data_inicio": "data_inicio",
	"data_fim":    "data_fim",
}

type linhaPeriodo struct {
	ID            uuid.UUID  `db:"id"`
	InstituicaoID uuid.UUID  `db:"instituicao_id"`
	Nome          string     `db:"nome"`
	DataInicio    time.Time  `db:"data_inicio"`
	DataFim       time.Time  `db:"data_fim"`
	CriadoEm      time.Time  `db:"criado_em"`
	AtualizadoEm  *time.Time `db:"atualizado_em"`
	ExcluidoEm    *time.Time `db:"excluido_em"`
	Versao        int        `db:"versao"`
	Planos        int        `db:"planos"`
}

func (l linhaPeriodo) paraItem() (port.ItemPeriodo, error) {
	nome, err := valueobject.NovoNomeCatalogo(l.Nome)
	if err != nil {
		return port.ItemPeriodo{}, err
	}
	inicio := valueobject.DataLocalDe(l.DataInicio, time.UTC)
	fim := valueobject.DataLocalDe(l.DataFim, time.UTC)
	vigencia, err := valueobject.NovaVigencia(inicio, &fim)
	if err != nil {
		return port.ItemPeriodo{}, err
	}
	return port.ItemPeriodo{
		Periodo: periodo.Periodo{
			ID: l.ID, InstituicaoID: l.InstituicaoID, Nome: nome, Vigencia: vigencia,
			CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
		},
		Planos: l.Planos,
	}, nil
}

type PeriodoRepository struct{ db *sqlx.DB }

func NovoPeriodoRepository(db *sqlx.DB) *PeriodoRepository { return &PeriodoRepository{db: db} }

var _ port.PeriodoRepository = (*PeriodoRepository)(nil)

const subConsultaPlanosDoPeriodo = `(SELECT count(*) FROM plano WHERE plano.periodo_id = periodo.id AND plano.excluido_em IS NULL)`

func (r *PeriodoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemPeriodo, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoPeriodo, 2)
	if err != nil {
		return port.ItemPeriodo{}, err
	}
	args = append([]any{id}, args...)
	consulta := "SELECT periodo.*, " + subConsultaPlanosDoPeriodo + " AS planos FROM periodo WHERE periodo.id = $1 AND " + clausula
	ex := Executor(ctx, r.db)
	var linha linhaPeriodo
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemPeriodo{}, domain.ErrNaoEncontrado
		}
		return port.ItemPeriodo{}, err
	}
	return linha.paraItem()
}

func (r *PeriodoRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarPeriodos) (port.ResultadoListaPeriodos, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPeriodo, 1)
	if err != nil {
		return port.ResultadoListaPeriodos{}, err
	}
	args := append([]any{}, escopoArgs...)
	n := len(args) + 1
	condicoes := []string{clausula}

	if filtro.Nome != "" {
		condicoes = append(condicoes, fmt.Sprintf("unaccent(lower(periodo.nome)) LIKE unaccent(lower($%d))", n))
		args = append(args, "%"+escaparCuringasLike(filtro.Nome)+"%")
		n++
	}
	if filtro.Situacao != "" && filtro.Situacao != "todas" {
		hoje := escopo.DataDeReferencia().String()
		switch filtro.Situacao {
		case "nao_iniciado":
			condicoes = append(condicoes, fmt.Sprintf("periodo.data_inicio > $%d", n))
		case "aberto":
			condicoes = append(condicoes, fmt.Sprintf("periodo.data_inicio <= $%d AND periodo.data_fim >= $%d", n, n))
		case "encerrado":
			condicoes = append(condicoes, fmt.Sprintf("periodo.data_fim < $%d", n))
		default:
			return port.ResultadoListaPeriodos{}, &domain.ErrParametro{Nome: "situacao"}
		}
		args = append(args, hoje)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM periodo WHERE "+where, args...); err != nil {
		return port.ResultadoListaPeriodos{}, err
	}

	sortColuna, ok := ordenacaoPeriodo[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoPeriodo["data_inicio"]
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

	consulta := fmt.Sprintf(
		"SELECT periodo.*, %s AS planos FROM periodo WHERE %s ORDER BY periodo.%s %s LIMIT $%d OFFSET $%d",
		subConsultaPlanosDoPeriodo, where, sortColuna, ordem, n, n+1)
	argsPaginados := append(append([]any{}, args...), pageSize, offset)

	var linhas []linhaPeriodo
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaPeriodos{}, err
	}
	itens := make([]port.ItemPeriodo, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return port.ResultadoListaPeriodos{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaPeriodos{Itens: itens, Total: total}, nil
}

func (r *PeriodoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	var dataFimTexto any
	if fim := p.Vigencia.Fim(); fim != nil {
		dataFimTexto = fim.String()
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO periodo (id, instituicao_id, nome, data_inicio, data_fim, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.InstituicaoID, p.Nome.String(), p.Vigencia.Inicio().String(), dataFimTexto, p.CriadoEm, p.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_periodo_instituicao_nome") {
			return domain.ErrNomePeriodoDuplicado
		}
		if violaIndice(err, "ck_periodo_datas") {
			return domain.ErrPeriodoDatasInvalidas
		}
		return err
	}
	return nil
}

func (r *PeriodoRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, p *periodo.Periodo, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPeriodo, 6)
	if err != nil {
		return err
	}
	var dataFimTexto any
	if fim := p.Vigencia.Fim(); fim != nil {
		dataFimTexto = fim.String()
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{p.Nome.String(), p.Vigencia.Inicio().String(), dataFimTexto, agora, p.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE periodo SET nome=$1, data_inicio=$2, data_fim=$3, atualizado_em=$4, versao=versao+1 "+
			"WHERE id=$5 AND versao=$6 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "uq_periodo_instituicao_nome") {
			return domain.ErrNomePeriodoDuplicado
		}
		if violaIndice(err, "ck_periodo_datas") {
			return domain.ErrPeriodoDatasInvalidas
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
	p.AtualizadoEm = &agora
	p.Versao = versaoEsperada + 1
	return nil
}

func (r *PeriodoRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoPeriodo, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx, "UPDATE periodo SET excluido_em = now() WHERE id = $1 AND "+clausula, args...)
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

func (r *PeriodoRepository) TemPlano(ctx context.Context, escopo autorizacao.Escopo, periodoID uuid.UUID) (bool, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoPlano, 2)
	if err != nil {
		return false, err
	}
	args = append([]any{periodoID}, args...)
	ex := Executor(ctx, r.db)
	var existe bool
	consulta := "SELECT EXISTS (SELECT 1 FROM plano WHERE plano.periodo_id = $1 AND " + clausula + ")"
	if err := sqlx.GetContext(ctx, ex, &existe, consulta, args...); err != nil {
		return false, err
	}
	return existe, nil
}
