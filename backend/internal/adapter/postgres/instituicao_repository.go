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
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ordenacaoInstituicao = map[string]string{
	"sigla":       "sigla",
	"nome":        `nome COLLATE "pt-BR-x-icu"`,
	"codigo_emec": "codigo_emec",
	"criado_em":   "criado_em",
}

type linhaInstituicao struct {
	ID                  uuid.UUID  `db:"id"`
	Nome                string     `db:"nome"`
	Sigla               string     `db:"sigla"`
	CodigoEMec          *string    `db:"codigo_emec"`
	Situacao            string     `db:"situacao"`
	CriadoEm            time.Time  `db:"criado_em"`
	AtualizadoEm        *time.Time `db:"atualizado_em"`
	ExcluidoEm          *time.Time `db:"excluido_em"`
	Versao              int        `db:"versao"`
	PesquisadoresAtivos int        `db:"pesquisadores_ativos"`
}

func (l linhaInstituicao) paraDominio() (port.ItemInstituicao, error) {
	sigla, err := valueobject.NovaSigla(l.Sigla)
	if err != nil {
		return port.ItemInstituicao{}, err
	}
	codigoBruto := ""
	if l.CodigoEMec != nil {
		codigoBruto = *l.CodigoEMec
	}
	codigo, err := valueobject.NovoCodigoEMec(codigoBruto)
	if err != nil {
		return port.ItemInstituicao{}, err
	}
	return port.ItemInstituicao{
		Instituicao: instituicao.Instituicao{
			ID:           l.ID,
			Nome:         l.Nome,
			Sigla:        sigla,
			CodigoEMec:   codigo,
			Situacao:     valueobject.SituacaoInstituicao(l.Situacao),
			CriadoEm:     l.CriadoEm,
			AtualizadoEm: l.AtualizadoEm,
			ExcluidoEm:   l.ExcluidoEm,
			Versao:       l.Versao,
		},
		PesquisadoresAtivos: l.PesquisadoresAtivos,
	}, nil
}

const selectInstituicaoComContagem = `
	SELECT i.*,
	       (SELECT count(*) FROM usuario_perfil up
	         JOIN usuario u ON u.id = up.usuario_id
	         WHERE up.instituicao_id = i.id
	           AND up.perfil = 'pesquisador_institucional'
	           AND u.excluido_em IS NULL) AS pesquisadores_ativos
	  FROM instituicao i`

type InstituicaoRepository struct {
	db *sqlx.DB
}

func NovoInstituicaoRepository(db *sqlx.DB) *InstituicaoRepository {
	return &InstituicaoRepository{db: db}
}

var _ port.InstituicaoRepository = (*InstituicaoRepository)(nil)

func (r *InstituicaoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemInstituicao, error) {
	if !escopo.Plataforma() {
		return port.ItemInstituicao{}, domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var linha linhaInstituicao
	err := sqlx.GetContext(ctx, ex, &linha, selectInstituicaoComContagem+" WHERE i.id = $1 AND i.excluido_em IS NULL", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemInstituicao{}, domain.ErrNaoEncontrado
		}
		return port.ItemInstituicao{}, err
	}
	return linha.paraDominio()
}

func (r *InstituicaoRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarInstituicoes) (port.ResultadoListaInstituicoes, error) {
	if !escopo.Plataforma() {
		return port.ResultadoListaInstituicoes{}, domain.ErrEscopoInvalido
	}

	condicoes := []string{"excluido_em IS NULL"}
	var args []any
	n := 1

	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(nome)) LIKE unaccent(lower($%d)) OR unaccent(lower(sigla)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
	}
	if filtro.Situacao != "" && filtro.Situacao != "todas" {
		condicoes = append(condicoes, fmt.Sprintf("situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM instituicao WHERE "+where, args...); err != nil {
		return port.ResultadoListaInstituicoes{}, err
	}

	sortColuna, ok := ordenacaoInstituicao[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoInstituicao["nome"]
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
	consulta := fmt.Sprintf("%s WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		selectInstituicaoComContagem, where, sortColuna, ordem, n, n+1)

	var linhas []linhaInstituicao
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaInstituicoes{}, err
	}

	itens := make([]port.ItemInstituicao, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraDominio()
		if err != nil {
			return port.ResultadoListaInstituicoes{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaInstituicoes{Itens: itens, Total: total}, nil
}

func (r *InstituicaoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *instituicao.Instituicao) error {
	if !escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	var codigo any
	if !i.CodigoEMec.Nulo() {
		codigo = i.CodigoEMec.String()
	}
	_, err := ex.ExecContext(ctx,
		`INSERT INTO instituicao (id, nome, sigla, codigo_emec, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		i.ID, i.Nome, i.Sigla.String(), codigo, string(i.Situacao), i.CriadoEm, i.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_instituicao_sigla") {
			return domain.ErrSiglaDuplicada
		}
		if violaIndice(err, "uq_instituicao_codigo_emec") {
			return domain.ErrCodigoEMecDuplicado
		}
		return err
	}
	return nil
}

func (r *InstituicaoRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, i *instituicao.Instituicao, versaoEsperada int) error {
	if !escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	var codigo any
	if !i.CodigoEMec.Nulo() {
		codigo = i.CodigoEMec.String()
	}
	resultado, err := ex.ExecContext(ctx,
		`UPDATE instituicao SET nome=$1, sigla=$2, codigo_emec=$3, atualizado_em=$4, versao=versao+1
		  WHERE id=$5 AND versao=$6 AND excluido_em IS NULL`,
		i.Nome, i.Sigla.String(), codigo, agora, i.ID, versaoEsperada,
	)
	if err != nil {
		if violaIndice(err, "uq_instituicao_sigla") {
			return domain.ErrSiglaDuplicada
		}
		if violaIndice(err, "uq_instituicao_codigo_emec") {
			return domain.ErrCodigoEMecDuplicado
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

func (r *InstituicaoRepository) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoInstituicao, versaoEsperada int) error {
	if !escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	// A situação nunca conflita com uq_instituicao_codigo_emec: o UPDATE
	// não toca codigo_emec, e o índice agora vale entre TODAS as
	// instituições não excluídas (não só as ativas) — duas instituições
	// com o mesmo código já são barradas no cadastro/edição, então a
	// reativação nunca encontra o código "livre" para colidir.
	resultado, err := ex.ExecContext(ctx,
		`UPDATE instituicao SET situacao=$1, atualizado_em=$2, versao=versao+1
		  WHERE id=$3 AND versao=$4 AND excluido_em IS NULL`,
		string(nova), agora, id, versaoEsperada,
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
