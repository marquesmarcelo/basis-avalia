package valueobject

import (
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// CodigoEMec — opcional (IES em credenciamento ainda não o tem, I-02).
// Sem validação de formato: a spec não define um, e inventar recusaria
// código legítimo (design.md §3.1).
type CodigoEMec struct {
	valor string
	nulo  bool
}

func NovoCodigoEMec(bruto string) (CodigoEMec, error) {
	normalizado := strings.TrimSpace(bruto)
	if normalizado == "" {
		return CodigoEMec{nulo: true}, nil
	}
	if len(normalizado) > 20 {
		return CodigoEMec{}, &domain.ErrValidacao{Campo: "codigo_emec", Mensagem: "Código e-MEC inválido."}
	}
	return CodigoEMec{valor: normalizado}, nil
}

func (c CodigoEMec) Nulo() bool { return c.nulo }

// String devolve vazio quando nulo — é assim que a coluna do grid exibe.
func (c CodigoEMec) String() string {
	if c.nulo {
		return ""
	}
	return c.valor
}
