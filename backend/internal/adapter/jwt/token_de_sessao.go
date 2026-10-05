package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrTokenInvalido = errors.New("token de sessão inválido")

const ttl = 8 * time.Hour

// claimsJWT — só sub, ins (omitido quando nulo), emt, iat, exp (design.md
// §7.1). Nenhum outro campo: perfil, nome e e-mail nunca entram no token.
type claimsJWT struct {
	Instituicao    *string `json:"ins,omitempty"`
	EmitidoEmMicro int64   `json:"emt"`
	jwt.RegisteredClaims
}

// TokenDeSessao — adapter Postgres-free, HS256 puro (design.md §7.1).
type TokenDeSessao struct {
	segredo []byte
	relogio port.Relogio
}

func NovoTokenDeSessao(segredo string, relogio port.Relogio) *TokenDeSessao {
	return &TokenDeSessao{segredo: []byte(segredo), relogio: relogio}
}

var _ port.TokenDeSessao = (*TokenDeSessao)(nil)

func (t *TokenDeSessao) Emitir(ator autorizacao.Ator, emitidoEm time.Time) (string, time.Time, error) {
	expiraEm := emitidoEm.Add(ttl)
	claims := claimsJWT{
		EmitidoEmMicro: emitidoEm.UnixMicro(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   ator.UsuarioID().String(),
			IssuedAt:  jwt.NewNumericDate(emitidoEm),
			ExpiresAt: jwt.NewNumericDate(expiraEm),
		},
	}
	if instituicaoID := ator.InstituicaoID(); instituicaoID != nil {
		s := instituicaoID.String()
		claims.Instituicao = &s
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	assinado, err := token.SignedString(t.segredo)
	if err != nil {
		return "", time.Time{}, err
	}
	return assinado, expiraEm, nil
}

func (t *TokenDeSessao) Validar(tokenBruto string) (port.Claims, error) {
	var claims claimsJWT
	token, err := jwt.ParseWithClaims(tokenBruto, &claims, func(tok *jwt.Token) (any, error) {
		return t.segredo, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithTimeFunc(t.relogio.Agora))
	if err != nil || !token.Valid {
		return port.Claims{}, fmt.Errorf("%w: %v", ErrTokenInvalido, err)
	}

	usuarioID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return port.Claims{}, ErrTokenInvalido
	}

	var instituicaoID *uuid.UUID
	if claims.Instituicao != nil {
		id, err := uuid.Parse(*claims.Instituicao)
		if err != nil {
			return port.Claims{}, ErrTokenInvalido
		}
		instituicaoID = &id
	}

	return port.Claims{
		UsuarioID:      usuarioID,
		InstituicaoID:  instituicaoID,
		EmitidoEmMicro: claims.EmitidoEmMicro,
	}, nil
}
