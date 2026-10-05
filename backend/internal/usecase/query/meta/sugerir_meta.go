package meta

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// SugerirMetaUseCase alimenta o autocomplete de meta — sem criação inline.
type SugerirMetaUseCase struct {
	repo port.MetaRepository
}

func NovoSugerirMetaUseCase(repo port.MetaRepository) *SugerirMetaUseCase {
	return &SugerirMetaUseCase{repo: repo}
}

func (uc *SugerirMetaUseCase) Executar(ctx context.Context, ator autorizacao.Ator, busca string) ([]port.ItemSugestaoMeta, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		return nil, err
	}
	return uc.repo.Sugerir(ctx, esc, busca)
}
