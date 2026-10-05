package port

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

// HashDeSenha — adapter argon2id (design.md §4.2, D-01, D-02).
// ConferirDescartavel existe só para consumir tempo equivalente ao de uma
// conferência real quando não há conta/hash a comparar (L-08): o resultado
// nunca é usado pelo use case, só o efeito colateral de duração.
type HashDeSenha interface {
	Gerar(ctx context.Context, senha valueobject.SenhaEmTexto) (valueobject.SenhaHash, error)
	Conferir(ctx context.Context, senha valueobject.SenhaEmTexto, hash valueobject.SenhaHash) (bool, error)
	ConferirDescartavel(ctx context.Context, senha valueobject.SenhaEmTexto)
}
