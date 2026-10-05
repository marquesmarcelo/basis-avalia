package autorizacao

import "github.com/basis-avalia/backend/internal/domain/valueobject"

// Alcance diz QUEM o ator está administrando. É a rota que o determina —
// e é a única coisa que o handler decide sobre autorização (design.md §3.2).
type Alcance int

const (
	UsuariosDaPropriaInstituicao Alcance = iota
	PesquisadoresDeUmaInstituicao
	AdministradoresDaPlataforma
	InstituicoesDaPlataforma

	// Acrescentados por fundacao-metas.md §6.2 — alcances das quatro
	// features de metas.
	IndicadoresDaPlataforma
	CatalogoDeIndicadores
	MetasDaInstituicao
	CursosDaInstituicao
	DesignacoesDaInstituicao
	PeriodosDaInstituicao
	PlanosDaInstituicao
	EntregasDaInstituicao
	DesempenhoDaInstituicao
	CursosDaCarteira
	PlanosDaCarteira
	EntregasDaCarteira
	DesempenhoDaCarteira
)

// PerfisFixos devolve o conjunto que este alcance impõe ao criar/manter
// usuário, ou nil quando o conjunto vem do payload (coagido por
// valueobject.ConjuntoInstitucional).
func (a Alcance) PerfisFixos() *valueobject.ConjuntoDePerfis {
	switch a {
	case PesquisadoresDeUmaInstituicao:
		c, _ := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
		return &c
	case AdministradoresDaPlataforma:
		c := valueobject.ConjuntoDeAdministrador()
		return &c
	default:
		return nil
	}
}
