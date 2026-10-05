package autorizacao

import "github.com/basis-avalia/backend/internal/domain/valueobject"

// MontarConjuntoEfetivo é a ÚNICA função do projeto que acrescenta
// CoordenadorCurso a um conjunto de perfis (specs/cursos/design.md §5,
// specs/_fundacao-metas.md §4). O perfil nunca é gravado em
// usuario_perfil — é sempre derivado de designação vigente, e este é o
// único ponto de acréscimo. Dois chamadores sancionados: o middleware de
// sessão (constrói o Ator que a autorização usa) e ObterContextoDeSessao
// (a resposta de GET /auth/eu). Nenhum outro lugar do código acrescenta
// CoordenadorCurso a um ConjuntoDePerfis — TestConjuntoEfetivo_
// UnicaFonteDoPerfilDerivado prende isso por grep.
func MontarConjuntoEfetivo(atribuidos valueobject.ConjuntoDePerfis, coordenaHoje bool) (valueobject.ConjuntoDePerfis, error) {
	if !coordenaHoje {
		return atribuidos, nil
	}
	comCoordenador := append(atribuidos.Ordenado(), valueobject.CoordenadorCurso)
	return valueobject.NovoConjunto(comCoordenador...)
}
