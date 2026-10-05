package postgres

import (
	"context"
	"encoding/json"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/jmoiron/sqlx"
)

// AuditoriaRepository — canal local (autoritativo), participa da mesma
// transação da mutação via UnidadeDeTrabalho (design.md §8.1). Tabela
// append-only: sem versao, sem atualizado_em, sem excluido_em.
type AuditoriaRepository struct {
	db *sqlx.DB
}

func NovoAuditoriaRepository(db *sqlx.DB) *AuditoriaRepository {
	return &AuditoriaRepository{db: db}
}

func (r *AuditoriaRepository) Registrar(ctx context.Context, e auditoria.Evento) error {
	ex := Executor(ctx, r.db)
	detalhes, err := json.Marshal(e.Detalhes)
	if err != nil {
		return err
	}
	var ipOrigem *string
	if e.IPOrigem.IsValid() {
		s := e.IPOrigem.String()
		ipOrigem = &s
	}
	var recursoTipo *string
	if e.RecursoTipo != "" {
		recursoTipo = &e.RecursoTipo
	}
	_, err = ex.ExecContext(ctx,
		`INSERT INTO auditoria (id, acao, resultado, ator_id, instituicao_id, recurso_tipo, recurso_id, detalhes, ip_origem, executado_em)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, string(e.Acao), string(e.Resultado), e.AtorID, e.InstituicaoID, recursoTipo, e.RecursoID, detalhes, ipOrigem, e.ExecutadoEm,
	)
	return err
}
