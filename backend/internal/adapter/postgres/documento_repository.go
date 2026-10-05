package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/documento"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type linhaDocumentoDB struct {
	ID            uuid.UUID  `db:"id"`
	PlanoID       uuid.UUID  `db:"plano_id"`
	CursoID       uuid.UUID  `db:"curso_id"`
	InstituicaoID uuid.UUID  `db:"instituicao_id"`
	ChaveObjeto   string     `db:"chave_objeto"`
	NomeArquivo   string     `db:"nome_arquivo"`
	SituacaoNoAto string     `db:"situacao_no_ato"`
	GeradoPor     uuid.UUID  `db:"gerado_por"`
	GeradoEm      time.Time  `db:"gerado_em"`
	CriadoEm      time.Time  `db:"criado_em"`
	ExcluidoEm    *time.Time `db:"excluido_em"`
}

func (l linhaDocumentoDB) paraDominio() documento.Documento {
	return documento.Documento{
		ID: l.ID, PlanoID: l.PlanoID, CursoID: l.CursoID, InstituicaoID: l.InstituicaoID,
		ChaveObjeto: l.ChaveObjeto, NomeArquivo: l.NomeArquivo, SituacaoNoAto: l.SituacaoNoAto,
		GeradoPor: l.GeradoPor, GeradoEm: l.GeradoEm, CriadoEm: l.CriadoEm, ExcluidoEm: l.ExcluidoEm,
	}
}

type DocumentoRepository struct{ db *sqlx.DB }

func NovoDocumentoRepository(db *sqlx.DB) *DocumentoRepository { return &DocumentoRepository{db: db} }

var _ port.DocumentoRepository = (*DocumentoRepository)(nil)

// Inserir é o ÚNICO caminho de escrita de gerar documento — nunca toca em
// `plano` (P-11, design.md §7.4): sem isso, o coordenador gerando um
// documento derrubaria a versão que o PI está editando em outra aba.
func (r *DocumentoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, d *documento.Documento) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO documento (id, plano_id, curso_id, instituicao_id, chave_objeto, nome_arquivo, situacao_no_ato, gerado_por, gerado_em, criado_em)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.PlanoID, d.CursoID, d.InstituicaoID, d.ChaveObjeto, d.NomeArquivo, d.SituacaoNoAto, d.GeradoPor, d.GeradoEm, d.CriadoEm,
	)
	return err
}

func (r *DocumentoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (documento.Documento, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoDocumento, 2)
	if err != nil {
		return documento.Documento{}, err
	}
	args = append([]any{id}, args...)
	ex := Executor(ctx, r.db)
	var linha linhaDocumentoDB
	consulta := "SELECT documento.* FROM documento WHERE documento.id = $1 AND " + clausula
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return documento.Documento{}, domain.ErrNaoEncontrado
		}
		return documento.Documento{}, err
	}
	return linha.paraDominio(), nil
}
