package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func avaliadorDeTeste(t *testing.T, instituicaoID uuid.UUID) autorizacao.Proprio {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	return ator.Proprio()
}

// TestEntregaRepository_T292_ListarFilaNuncaEnviaParametroNaoReferenciado
// prova o achado de revisão T-292: a consulta de CONTAGEM da fila de
// avaliação não usa o id do avaliador em lugar nenhum do seu texto SQL —
// antes, ele viajava mesmo assim como argumento reservado ($1), e o
// Postgres recusava com "could not determine data type of parameter $1"
// (42P18) porque um parâmetro declarado e nunca referenciado não tem
// contexto para o tipo ser inferido. O teste cobre exatamente os dois
// caminhos que quebravam: filtro de situação vazio ("Todas", o caso que
// a tela realmente dispara ao trocar o select) e o filtro padrão
// ("pendente_avaliacao") — os dois usam a MESMA consulta de contagem.
func TestEntregaRepository_T292_ListarFilaNuncaEnviaParametroNaoReferenciado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoEntregaRepository(db)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	esc := escopoEntregaDaInstituicaoPI(t, instituicaoID)
	avaliador := avaliadorDeTeste(t, instituicaoID)

	casos := []struct {
		nome     string
		situacao string
	}{
		{"situação vazia (Todas)", ""},
		{"situação pendente_avaliacao (padrão da tela)", "pendente_avaliacao"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			resultado, err := repo.ListarFilaDeAvaliacao(context.Background(), esc, avaliador, port.FiltroFilaAvaliacao{
				Situacao: c.situacao, Page: 1, PageSize: 20,
			})
			if err != nil {
				t.Fatalf("esperava sucesso, obtido erro: %v", err)
			}
			if resultado.Total != 0 {
				t.Fatalf("esperava 0 entregas (nenhuma criada neste teste), obtido %d", resultado.Total)
			}
		})
	}
}
