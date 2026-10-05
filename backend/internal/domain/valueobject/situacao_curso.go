package valueobject

// SituacaoCurso — ativo ou inativo (specs/cursos/design.md §3.1). Distinto
// de exclusão lógica: curso nunca é excluído por inativação.
type SituacaoCurso string

const (
	CursoAtivo   SituacaoCurso = "ativo"
	CursoInativo SituacaoCurso = "inativo"
)

func (s SituacaoCurso) Alternar() SituacaoCurso {
	if s == CursoAtivo {
		return CursoInativo
	}
	return CursoAtivo
}
