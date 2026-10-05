package port

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type FiltroListarUsuarios struct {
	Busca    string
	Perfil   *valueobject.Perfil // filtro de posse (G-03): "possui este perfil", um valor por vez
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ResultadoListaUsuarios struct {
	Itens []usuario.Usuario
	Total int
}

// UsuarioRepository — todo método recebe Escopo depois do ctx, sem
// exceção (design.md §4.2, §4.4). Operação sobre a própria identidade
// (alterar a própria senha, encerrar a própria sessão) não vive aqui —
// vive em AutenticacaoRepository, com Proprio no lugar de Escopo.
type UsuarioRepository interface {
	BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (*usuario.Usuario, error)
	Listar(ctx context.Context, escopo autorizacao.Escopo, filtro FiltroListarUsuarios) (ResultadoListaUsuarios, error)
	Inserir(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario) error
	Atualizar(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario, versaoEsperada int) error
	SubstituirPerfis(ctx context.Context, escopo autorizacao.Escopo, usuarioID uuid.UUID, perfis valueobject.ConjuntoDePerfis) error
	ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error
	DefinirSenha(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, hash valueobject.SenhaHash, senhaProvisoria bool, instante time.Time) error
	InvalidarSessoes(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error
	ContarDetentoresDoPerfil(ctx context.Context, escopo autorizacao.Escopo, perfil valueobject.Perfil) (int, error)
	TravarPopulacao(ctx context.Context, escopo autorizacao.Escopo) error
}
