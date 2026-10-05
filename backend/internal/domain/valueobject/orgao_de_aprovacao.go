package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// OrgaoDeAprovacao — lista fechada (specs/plano-acao/spec.md PP-4): NDE ou
// Colegiado de curso. Texto livre perderia a padronização que o campo
// existe para dar.
type OrgaoDeAprovacao string

const (
	OrgaoNDE              OrgaoDeAprovacao = "nde"
	OrgaoColegiadoDeCurso OrgaoDeAprovacao = "colegiado_curso"
)

func NovoOrgaoDeAprovacao(bruto string) (OrgaoDeAprovacao, error) {
	o := OrgaoDeAprovacao(bruto)
	switch o {
	case OrgaoNDE, OrgaoColegiadoDeCurso:
		return o, nil
	default:
		return "", domain.ErrValorInvalido
	}
}
