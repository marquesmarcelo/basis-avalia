package postgres

import (
	"context"

	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type linhaInstituicaoPublica struct {
	ID    uuid.UUID `db:"id"`
	Nome  string    `db:"nome"`
	Sigla string    `db:"sigla"`
}

// InstituicaoPublicaQuery — porta estreita sem Escopo, pública por
// desenho (design.md §4.2, §5.6).
type InstituicaoPublicaQuery struct {
	db *sqlx.DB
}

func NovoInstituicaoPublicaQuery(db *sqlx.DB) *InstituicaoPublicaQuery {
	return &InstituicaoPublicaQuery{db: db}
}

var _ port.InstituicaoPublicaQuery = (*InstituicaoPublicaQuery)(nil)

// ListarParaCombo só devolve instituição ativa E com ao menos um detentor
// ativo do perfil de pesquisador_institucional (3.17, 3.20) — sem coluna
// derivada, sem view materializada, via EXISTS na tabela de vínculo
// (design.md §5.5).
func (q *InstituicaoPublicaQuery) ListarParaCombo(ctx context.Context) ([]port.InstituicaoPublica, error) {
	var linhas []linhaInstituicaoPublica
	err := sqlx.SelectContext(ctx, q.db, &linhas, `
		SELECT i.id, i.nome, i.sigla
		  FROM instituicao i
		 WHERE i.situacao = 'ativa'
		   AND i.excluido_em IS NULL
		   AND EXISTS (SELECT 1
		                 FROM usuario_perfil up
		                 JOIN usuario u ON u.id = up.usuario_id
		                WHERE up.instituicao_id = i.id
		                  AND up.perfil = 'pesquisador_institucional'
		                  AND u.excluido_em IS NULL)
		 ORDER BY i.nome COLLATE "pt-BR-x-icu"`,
	)
	if err != nil {
		return nil, err
	}
	itens := make([]port.InstituicaoPublica, 0, len(linhas))
	for _, l := range linhas {
		itens = append(itens, port.InstituicaoPublica{ID: l.ID, Nome: l.Nome, Sigla: l.Sigla})
	}
	return itens, nil
}
