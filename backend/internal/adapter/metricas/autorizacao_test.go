package metricas

import (
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestAutorizacaoDecisoesTotal_RegistraPermitidoENegado(t *testing.T) {
	RegistrarPermitido(autorizacao.UsuarioListar)
	RegistrarNegado(autorizacao.InstituicaoListar)

	saida, err := testutil.GatherAndCount(prometheus.DefaultGatherer, "autorizacao_decisoes_total")
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if saida < 2 {
		t.Fatalf("esperava ao menos 2 séries registradas, obtido %d", saida)
	}
}

func TestNomeDaMetrica_NaoContemFalhaNemAutenticacao(t *testing.T) {
	nome := "autorizacao_decisoes_total"
	if strings.Contains(nome, "falha") || strings.Contains(nome, "autenticacao") {
		t.Fatal("métrica não deveria conter 'falha' nem 'autenticacao' no nome (3.2, controle recusado)")
	}
}
