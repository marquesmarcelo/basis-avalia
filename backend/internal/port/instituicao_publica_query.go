package port

import (
	"context"

	"github.com/google/uuid"
)

// InstituicaoPublica é o formato exato da rota pública (R1): id, nome e
// sigla — nunca código e-MEC, usuários ou contagem (3.17).
type InstituicaoPublica struct {
	ID    uuid.UUID
	Nome  string
	Sigla string
}

// InstituicaoPublicaQuery — porta estreita SEM Escopo, pública por desenho
// (design.md §4.2). Método novo aqui exige decisão do arquiteto.
type InstituicaoPublicaQuery interface {
	ListarParaCombo(ctx context.Context) ([]InstituicaoPublica, error)
}
