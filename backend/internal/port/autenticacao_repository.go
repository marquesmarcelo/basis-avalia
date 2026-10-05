package port

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// CredencialUsuario é o suficiente para o use case Autenticar conferir a
// senha e montar o Ator — não é a entidade completa.
type CredencialUsuario struct {
	ID              uuid.UUID
	InstituicaoID   *uuid.UUID
	Nome            string
	Email           valueobject.Email
	SenhaHash       valueobject.SenhaHash
	Perfis          valueobject.ConjuntoDePerfis
	SenhaProvisoria bool
}

// ContextoDeSessao é o que o middleware de sessão precisa a cada
// requisição (design.md §7.3) — sem filtro de excluido_em de propósito,
// para distinguir "linha ausente" (SESSAO_EXPIRADA) de "linha excluída"
// (CONTA_EXCLUIDA).
//
// CoordenaHoje e CursosCoordenados vêm da MESMA consulta (specs/cursos/
// design.md §4.2, specs/_fundacao-metas.md §4.1): designação vigente na
// data de referência recebida, sem juntar curso — designação vigente de
// curso inativo continua contando (C-07, CP-06). Perfis aqui é sempre o
// conjunto ATRIBUÍDO, nunca o efetivo — quem acrescenta CoordenadorCurso
// é autorizacao.MontarConjuntoEfetivo, no chamador, nunca aqui.
type ContextoDeSessao struct {
	UsuarioID               uuid.UUID
	Nome                    string
	Email                   valueobject.Email
	Perfis                  valueobject.ConjuntoDePerfis
	SenhaProvisoria         bool
	InstituicaoID           *uuid.UUID
	ExcluidoEm              *time.Time
	SessoesValidasAPartirDe time.Time
	InstituicaoNome         string
	InstituicaoSigla        string
	InstituicaoSituacao     string
	CoordenaHoje            bool
	CursosCoordenados       int
}

// AutenticacaoRepository — porta estreita sem Escopo, por desenho
// (design.md §4.4). Regra de admissão: um método só pode viver aqui se o
// único identificador que recebe vier (a) do par de credenciais oferecido
// para autenticação (instituicaoID + email, no corpo do login), ou (b) de
// um autorizacao.Proprio — nunca de um identificador escolhido pelo
// cliente via corpo, caminho ou query string. Se recebe esse tipo de
// identificador, é operação de negócio e exige Escopo em UsuarioRepository
// ou InstituicaoRepository, sem exceção. Lista fechada e congelada pelos
// testes de reflexão em internal/port/portas_test.go — método novo aqui
// exige decisão do arquiteto (design.md §4.4).
type AutenticacaoRepository interface {
	BuscarCredencial(ctx context.Context, instituicaoID *uuid.UUID, email valueobject.Email) (*CredencialUsuario, error)
	CarregarContextoDeSessao(ctx context.Context, usuarioID uuid.UUID, hoje valueobject.DataLocal) (*ContextoDeSessao, error)
	BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*CredencialUsuario, error)
	DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error
	InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error
}
