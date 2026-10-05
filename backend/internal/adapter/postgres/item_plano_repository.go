package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/itemplano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type linhaItemPlanoDB struct {
	ID            uuid.UUID  `db:"id"`
	PlanoID       uuid.UUID  `db:"plano_id"`
	CursoID       uuid.UUID  `db:"curso_id"`
	InstituicaoID uuid.UUID  `db:"instituicao_id"`
	MetaID        uuid.UUID  `db:"meta_id"`
	Quantidade    int        `db:"quantidade"`
	CriadoEm      time.Time  `db:"criado_em"`
	AtualizadoEm  *time.Time `db:"atualizado_em"`
	ExcluidoEm    *time.Time `db:"excluido_em"`
	Versao        int        `db:"versao"`
}

func (l linhaItemPlanoDB) paraDominio() (itemplano.ItemDoPlano, error) {
	quantidade, err := valueobject.NovaQuantidade(l.Quantidade)
	if err != nil {
		return itemplano.ItemDoPlano{}, err
	}
	return itemplano.ItemDoPlano{
		ID: l.ID, PlanoID: l.PlanoID, CursoID: l.CursoID, InstituicaoID: l.InstituicaoID, MetaID: l.MetaID,
		Quantidade: quantidade, CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
	}, nil
}

type ItemPlanoRepository struct{ db *sqlx.DB }

func NovoItemPlanoRepository(db *sqlx.DB) *ItemPlanoRepository { return &ItemPlanoRepository{db: db} }

var _ port.ItemPlanoRepository = (*ItemPlanoRepository)(nil)

func (r *ItemPlanoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (itemplano.ItemDoPlano, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoItemPlano, 2)
	if err != nil {
		return itemplano.ItemDoPlano{}, err
	}
	args = append([]any{id}, args...)
	ex := Executor(ctx, r.db)
	var linha linhaItemPlanoDB
	consulta := "SELECT item_plano.* FROM item_plano WHERE item_plano.id = $1 AND " + clausula
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return itemplano.ItemDoPlano{}, domain.ErrNaoEncontrado
		}
		return itemplano.ItemDoPlano{}, err
	}
	return linha.paraDominio()
}

func (r *ItemPlanoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO item_plano (id, plano_id, curso_id, instituicao_id, meta_id, quantidade, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		i.ID, i.PlanoID, i.CursoID, i.InstituicaoID, i.MetaID, i.Quantidade.Int(), i.CriadoEm, i.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_item_plano_meta") {
			return domain.ErrMetaDuplicadaNoPlano
		}
		return err
	}
	return nil
}

func (r *ItemPlanoRepository) AtualizarQuantidade(ctx context.Context, escopo autorizacao.Escopo, i *itemplano.ItemDoPlano, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoItemPlano, 5)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{i.Quantidade.Int(), agora, i.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE item_plano SET quantidade=$1, atualizado_em=$2, versao=versao+1 WHERE id=$3 AND versao=$4 AND "+clausula,
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
	i.AtualizadoEm = &agora
	i.Versao = versaoEsperada + 1
	return nil
}

func (r *ItemPlanoRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoItemPlano, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx, "UPDATE item_plano SET excluido_em = now() WHERE id = $1 AND "+clausula, args...)
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

func (r *ItemPlanoRepository) ExisteMetaNoPlano(ctx context.Context, escopo autorizacao.Escopo, planoID, metaID uuid.UUID) (bool, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoItemPlano, 3)
	if err != nil {
		return false, err
	}
	args = append([]any{planoID, metaID}, args...)
	ex := Executor(ctx, r.db)
	var existe bool
	consulta := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM item_plano WHERE item_plano.plano_id = $1 AND item_plano.meta_id = $2 AND %s)", clausula)
	if err := sqlx.GetContext(ctx, ex, &existe, consulta, args...); err != nil {
		return false, err
	}
	return existe, nil
}
