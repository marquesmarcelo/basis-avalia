package usuario

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// BuscarUsuarioUseCase cobre T-02, T-05: a permissão (403) é avaliada por
// Autorizar antes de qualquer consulta; o isolamento (404) nasce do Escopo
// aplicado dentro do repositório.
type BuscarUsuarioUseCase struct {
	repo port.UsuarioRepository
}

func NovoBuscarUsuarioUseCase(repo port.UsuarioRepository) *BuscarUsuarioUseCase {
	return &BuscarUsuarioUseCase{repo: repo}
}

func (uc *BuscarUsuarioUseCase) Executar(ctx context.Context, ator autorizacao.Ator, alcance autorizacao.Alcance, instituicaoDoCaminho *uuid.UUID, usuarioID uuid.UUID) (usuario.Usuario, error) {
	esc, err := autorizacao.Autorizar(ator, alcance, autorizacao.AcaoBuscar, instituicaoDoCaminho)
	if err != nil {
		return usuario.Usuario{}, err
	}
	u, err := uc.repo.BuscarPorID(ctx, esc, usuarioID)
	if err != nil {
		return usuario.Usuario{}, err
	}
	return *u, nil
}
