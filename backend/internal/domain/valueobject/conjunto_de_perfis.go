package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// ordemCanonica é a única fonte de ordenação de ConjuntoDePerfis — usada
// tanto por Ordenado() (serialização, auditoria) quanto por Igual()
// (comparação), para que as duas nunca divirjam (design.md §3.1.1).
var ordemCanonica = []Perfil{Aluno, Professor, CoordenadorCurso, PesquisadorInstitucional, AdministradorSistema}

// ConjuntoDePerfis — Value Object imutável sobre a tabela de vínculo
// usuario_perfil. Nunca vazio, nunca com repetição, sempre serializado em
// ordem canônica (design.md §3.1.1).
type ConjuntoDePerfis struct {
	perfis []Perfil
}

func ordenarEDeduplicar(brutos []Perfil) []Perfil {
	presentes := make(map[Perfil]bool, len(brutos))
	for _, p := range brutos {
		presentes[p] = true
	}
	ordenado := make([]Perfil, 0, len(presentes))
	for _, p := range ordemCanonica {
		if presentes[p] {
			ordenado = append(ordenado, p)
		}
	}
	return ordenado
}

// NovoConjunto valida sem coagir — usado onde o conjunto já deve estar
// completo (ex: reconstrução a partir do banco).
func NovoConjunto(p ...Perfil) (ConjuntoDePerfis, error) {
	if len(p) == 0 {
		return ConjuntoDePerfis{}, domain.ErrConjuntoDePerfisVazio
	}
	temAdministrador := false
	for _, perfil := range p {
		switch perfil {
		case AdministradorSistema, PesquisadorInstitucional, CoordenadorCurso, Professor, Aluno:
		default:
			return ConjuntoDePerfis{}, domain.ErrPerfilInvalido
		}
		if perfil == AdministradorSistema {
			temAdministrador = true
		}
	}
	ordenado := ordenarEDeduplicar(p)
	if temAdministrador && len(ordenado) > 1 {
		return ConjuntoDePerfis{}, domain.ErrCombinacaoDePerfisInvalida
	}
	return ConjuntoDePerfis{perfis: ordenado}, nil
}

// ConjuntoInstitucional é o construtor do CRUD de usuários institucionais:
// aplica a coerção de spec 3.6 (E-16) e recusa administrador_sistema por
// completo, porque quem chama este construtor nunca está criando um
// administrador (design.md §3.1.1). Recusa também coordenador_curso
// (DC-4, specs/cursos/design.md C-09): o perfil não é mais atribuível por
// ninguém, nem pelo PI — é sempre derivado de designação, nunca gravado em
// usuario_perfil por este caminho.
//
// Regra que não pode ter uma segunda linha em lugar nenhum do código:
// aluno só entra quando o conjunto de entrada está vazio. Nunca é
// acrescentado a um conjunto não vazio (U-13, 4.4).
func ConjuntoInstitucional(p []Perfil) (ConjuntoDePerfis, error) {
	if len(p) == 0 {
		return ConjuntoDePerfis{perfis: []Perfil{Aluno}}, nil
	}
	for _, perfil := range p {
		if perfil == AdministradorSistema || perfil == CoordenadorCurso {
			return ConjuntoDePerfis{}, domain.ErrPerfilNaoAtribuivel
		}
	}
	return NovoConjunto(p...)
}

// ConjuntoDeAdministrador devolve sempre {administrador_sistema} — nunca
// falha, porque não depende de entrada do cliente.
func ConjuntoDeAdministrador() ConjuntoDePerfis {
	return ConjuntoDePerfis{perfis: []Perfil{AdministradorSistema}}
}

// Pode é a união: verdadeiro se QUALQUER perfil do conjunto tiver a
// permissão (A-07).
func (c ConjuntoDePerfis) Pode(perm Permissao) bool {
	for _, p := range c.perfis {
		if p.Pode(perm) {
			return true
		}
	}
	return false
}

func (c ConjuntoDePerfis) Possui(p Perfil) bool {
	for _, atual := range c.perfis {
		if atual == p {
			return true
		}
	}
	return false
}

// Ordenado devolve uma cópia em ordem canônica — nunca a fatia interna,
// para preservar a imutabilidade do Value Object.
func (c ConjuntoDePerfis) Ordenado() []Perfil {
	copia := make([]Perfil, len(c.perfis))
	copy(copia, c.perfis)
	return copia
}

func (c ConjuntoDePerfis) Igual(o ConjuntoDePerfis) bool {
	if len(c.perfis) != len(o.perfis) {
		return false
	}
	for i, p := range c.perfis {
		if o.perfis[i] != p {
			return false
		}
	}
	return true
}

// PertenceAInstituicao é falso somente quando o conjunto é
// {administrador_sistema} — o único perfil sem vínculo institucional
// (3.16). NovoConjunto já garante que administrador_sistema nunca
// coexiste com outro perfil, então basta checar o primeiro.
func (c ConjuntoDePerfis) PertenceAInstituicao() bool {
	return len(c.perfis) == 0 || c.perfis[0] != AdministradorSistema
}
