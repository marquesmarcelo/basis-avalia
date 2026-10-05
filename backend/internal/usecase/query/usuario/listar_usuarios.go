package usuario

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type ListarUsuariosInput struct {
	Ator                 autorizacao.Ator
	Alcance              autorizacao.Alcance
	InstituicaoDoCaminho *uuid.UUID
	Busca                string
	Perfil               *valueobject.Perfil
	Page                 int
	PageSize             int
	Sort                 string
	Order                string
}

// ListarUsuariosUseCase cobre G-01..G-18, T-01, AS-02.
type ListarUsuariosUseCase struct {
	repo port.UsuarioRepository
}

func NovoListarUsuariosUseCase(repo port.UsuarioRepository) *ListarUsuariosUseCase {
	return &ListarUsuariosUseCase{repo: repo}
}

func (uc *ListarUsuariosUseCase) Executar(ctx context.Context, in ListarUsuariosInput) (port.ResultadoListaUsuarios, error) {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoListar, in.InstituicaoDoCaminho)
	if err != nil {
		return port.ResultadoListaUsuarios{}, err
	}
	return uc.repo.Listar(ctx, esc, port.FiltroListarUsuarios{
		Busca: in.Busca, Perfil: in.Perfil, Page: in.Page, PageSize: in.PageSize, Sort: in.Sort, Order: in.Order,
	})
}
