package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// Perfil — Value Object com os cinco valores técnicos fixos da spec (3.6).
type Perfil string

const (
	AdministradorSistema     Perfil = "administrador_sistema"
	PesquisadorInstitucional Perfil = "pesquisador_institucional"
	CoordenadorCurso         Perfil = "coordenador_curso"
	Professor                Perfil = "professor"
	Aluno                    Perfil = "aluno"
)

// Permissao — string tipada. Vive aqui, não em internal/domain/autorizacao,
// para que Perfil possa conferir a matriz 3.9 sem criar import cycle (o
// pacote autorizacao já importa valueobject para Ator e Escopo). O pacote
// autorizacao reexporta este tipo via alias — ver autorizacao/permissao.go.
type Permissao string

const (
	InstituicaoListar      Permissao = "instituicao.listar"
	InstituicaoCriar       Permissao = "instituicao.criar"
	InstituicaoEditar      Permissao = "instituicao.editar"
	InstituicaoInativar    Permissao = "instituicao.inativar"
	PIGerenciar            Permissao = "pi.gerenciar"
	UsuarioListar          Permissao = "usuario.listar"
	UsuarioCriar           Permissao = "usuario.criar"
	UsuarioEditar          Permissao = "usuario.editar"
	UsuarioExcluir         Permissao = "usuario.excluir"
	UsuarioRedefinirSenha  Permissao = "usuario.redefinir_senha"
	AdministradorGerenciar Permissao = "administrador.gerenciar"

	// Acrescentadas por fundacao-metas.md §6.1 — permissões das quatro
	// features de metas. Administrador do Sistema não ganha nenhuma
	// permissão de conteúdo (ele administra a plataforma), com a única
	// exceção de IndicadorPlataformaGerenciar.
	IndicadorPlataformaGerenciar Permissao = "indicador.plataforma.gerenciar"
	IndicadorListar              Permissao = "indicador.listar"
	IndicadorGerenciar           Permissao = "indicador.gerenciar"
	MetaListar                   Permissao = "meta.listar"
	MetaGerenciar                Permissao = "meta.gerenciar"
	CursoListar                  Permissao = "curso.listar"
	CursoGerenciar               Permissao = "curso.gerenciar"
	CursoLerProprio              Permissao = "curso.ler_proprio"
	DesignacaoGerenciar          Permissao = "designacao.gerenciar"
	PeriodoGerenciar             Permissao = "periodo.gerenciar"
	PlanoListar                  Permissao = "plano.listar"
	PlanoGerenciar               Permissao = "plano.gerenciar"
	PlanoLerProprio              Permissao = "plano.ler_proprio"
	EntregaRegistrar             Permissao = "entrega.registrar"
	EntregaListar                Permissao = "entrega.listar"
	EntregaAvaliar               Permissao = "entrega.avaliar"
	RelatorioLer                 Permissao = "relatorio.ler"
	RelatorioExportar            Permissao = "relatorio.exportar"
	RelatorioLerProprio          Permissao = "relatorio.ler_proprio"

	Nenhuma Permissao = ""

	// EscopoProprio — rótulo de métrica, não permissão concedível (nunca
	// entra em matrizPermissoes). Usado nas rotas em que não há permissão
	// a conferir porque a autorização É o escopo do próprio ator — T-128,
	// design.md §6.1.
	EscopoProprio Permissao = "escopo_proprio"
)

func NovoPerfil(bruto string) (Perfil, error) {
	p := Perfil(bruto)
	switch p {
	case AdministradorSistema, PesquisadorInstitucional, CoordenadorCurso, Professor, Aluno:
		return p, nil
	default:
		return "", &domain.ErrValidacao{Campo: "perfil", Mensagem: "Perfil inválido."}
	}
}

func (p Perfil) Rotulo() string {
	switch p {
	case AdministradorSistema:
		return "Administrador do Sistema"
	case PesquisadorInstitucional:
		return "Pesquisador Institucional"
	case CoordenadorCurso:
		return "Coordenador de Curso"
	case Professor:
		return "Professor"
	case Aluno:
		return "Aluno"
	default:
		return ""
	}
}

// PertenceAInstituicao — falso apenas para o Administrador do Sistema, o
// único perfil com vínculo institucional nulo (3.16).
func (p Perfil) PertenceAInstituicao() bool {
	return p != AdministradorSistema
}

// matrizPermissoes implementa a tabela 3.9 da spec, mais administrador.gerenciar
// (acréscimo do arquiteto, design.md §3.2).
var matrizPermissoes = map[Perfil]map[Permissao]bool{
	AdministradorSistema: {
		InstituicaoListar:            true,
		InstituicaoCriar:             true,
		InstituicaoEditar:            true,
		InstituicaoInativar:          true,
		PIGerenciar:                  true,
		AdministradorGerenciar:       true,
		IndicadorPlataformaGerenciar: true,
	},
	PesquisadorInstitucional: {
		UsuarioListar:         true,
		UsuarioCriar:          true,
		UsuarioEditar:         true,
		UsuarioExcluir:        true,
		UsuarioRedefinirSenha: true,
		IndicadorListar:       true,
		IndicadorGerenciar:    true,
		MetaListar:            true,
		MetaGerenciar:         true,
		CursoListar:           true,
		CursoGerenciar:        true,
		DesignacaoGerenciar:   true,
		PeriodoGerenciar:      true,
		PlanoListar:           true,
		PlanoGerenciar:        true,
		EntregaListar:         true,
		EntregaAvaliar:        true,
		RelatorioLer:          true,
		RelatorioExportar:     true,
	},
	CoordenadorCurso: {
		PlanoLerProprio:     true,
		EntregaRegistrar:    true,
		RelatorioLerProprio: true,
		CursoLerProprio:     true,
	},
	Professor: {},
	Aluno:     {},
}

func (p Perfil) Pode(permissao Permissao) bool {
	return matrizPermissoes[p][permissao]
}

// PodeAtribuir impede escalada de privilégio na criação de usuário (3.6):
// o PI nunca atribui administrador_sistema, e o Administrador do Sistema
// nunca atribui coordenador, professor nem aluno. coordenador_curso saiu
// da lista do PI (DC-4, specs/cursos/design.md C-09): o perfil deixou de
// ser atribuível por qualquer um — é sempre derivado de designação.
func (p Perfil) PodeAtribuir(outro Perfil) bool {
	switch p {
	case AdministradorSistema:
		return outro == PesquisadorInstitucional || outro == AdministradorSistema
	case PesquisadorInstitucional:
		switch outro {
		case PesquisadorInstitucional, Professor, Aluno:
			return true
		}
		return false
	default:
		return false
	}
}
