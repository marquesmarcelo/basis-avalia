package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// GrauDeCurso — lista fechada (specs/cursos/design.md §3.1).
type GrauDeCurso string

const (
	Bacharelado  GrauDeCurso = "bacharelado"
	Licenciatura GrauDeCurso = "licenciatura"
	Tecnologo    GrauDeCurso = "tecnologo"
)

func NovoGrauDeCurso(bruto string) (GrauDeCurso, error) {
	g := GrauDeCurso(bruto)
	switch g {
	case Bacharelado, Licenciatura, Tecnologo:
		return g, nil
	default:
		return "", domain.ErrValorInvalido
	}
}
