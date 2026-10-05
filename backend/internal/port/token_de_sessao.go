package port

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/google/uuid"
)

// Claims — o que o middleware de sessão recupera de um token válido.
// Perfil, nome, e-mail e situação nunca entram aqui: são lidos do banco a
// cada requisição (design.md §3.5, §7.1).
type Claims struct {
	UsuarioID      uuid.UUID
	InstituicaoID  *uuid.UUID
	EmitidoEmMicro int64
}

// TokenDeSessao — adapter JWT HS256 (design.md §4.2, §7.1).
type TokenDeSessao interface {
	Emitir(ator autorizacao.Ator, emitidoEm time.Time) (token string, expiraEm time.Time, err error)
	Validar(token string) (Claims, error)
}
