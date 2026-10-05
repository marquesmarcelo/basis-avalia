package postgres

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func escopoIndicadoresDaPlataforma(t *testing.T) autorizacao.Escopo {
	t.Helper()
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

func escopoCatalogoDeIndicadores(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

func escopoMetasDaInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}
