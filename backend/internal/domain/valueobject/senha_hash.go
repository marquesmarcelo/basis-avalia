package valueobject

import (
	"errors"
	"strings"
)

var ErrHashInvalido = errors.New("hash de senha inválido")

// SenhaHash — Value Object sobre a string PHC produzida pelo adapter argon2id.
type SenhaHash struct {
	valor string
}

func NovaSenhaHash(codificado string) (SenhaHash, error) {
	if codificado == "" || !strings.HasPrefix(codificado, "$argon2id$") {
		return SenhaHash{}, ErrHashInvalido
	}
	return SenhaHash{valor: codificado}, nil
}

// Codificado só é usado pela camada de persistência.
func (h SenhaHash) Codificado() string { return h.valor }

func (h SenhaHash) String() string { return "[hash omitido]" }
