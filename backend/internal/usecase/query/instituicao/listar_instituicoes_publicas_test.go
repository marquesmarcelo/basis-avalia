package instituicao

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/port"
)

type instituicaoPublicaQueryMock struct {
	itens []port.InstituicaoPublica
	erro  error
}

func (m *instituicaoPublicaQueryMock) ListarParaCombo(ctx context.Context) ([]port.InstituicaoPublica, error) {
	return m.itens, m.erro
}

func TestListarInstituicoesPublicas_L11_DevolveApenasIDNomeESigla(t *testing.T) {
	mock := &instituicaoPublicaQueryMock{itens: []port.InstituicaoPublica{
		{Nome: "Faculdade Serra Azul", Sigla: "FSA"},
		{Nome: "Instituto Vale Verde", Sigla: "IVV"},
	}}
	uc := NovoListarInstituicoesPublicasUseCase(mock)

	out, err := uc.Executar(context.Background())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("esperava 2 itens, obtido %d", len(out))
	}
	if out[0].Sigla != "FSA" || out[1].Sigla != "IVV" {
		t.Fatalf("ordem/sigla incorreta: %+v", out)
	}
}
