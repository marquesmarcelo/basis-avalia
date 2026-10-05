package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// ModalidadeDeCurso — lista fechada (specs/cursos/design.md §3.1).
type ModalidadeDeCurso string

const (
	Presencial ModalidadeDeCurso = "presencial"
	ADistancia ModalidadeDeCurso = "a_distancia"
)

func NovaModalidadeDeCurso(bruto string) (ModalidadeDeCurso, error) {
	m := ModalidadeDeCurso(bruto)
	switch m {
	case Presencial, ADistancia:
		return m, nil
	default:
		return "", domain.ErrValorInvalido
	}
}
