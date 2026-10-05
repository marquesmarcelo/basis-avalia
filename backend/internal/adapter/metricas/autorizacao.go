package metricas

import (
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// autorizacaoDecisoesTotal — exigida pelo CLAUDE.md e confirmada na §11 da
// spec: decisões de autorização por permissão e resultado
// (permitido/negado).
var autorizacaoDecisoesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "autorizacao_decisoes_total",
	Help: "Decisões de autorização, por permissão e resultado.",
}, []string{"permissao", "resultado"})

func RegistrarPermitido(p autorizacao.Permissao) {
	autorizacaoDecisoesTotal.WithLabelValues(string(p), "permitido").Inc()
}

func RegistrarNegado(p autorizacao.Permissao) {
	autorizacaoDecisoesTotal.WithLabelValues(string(p), "negado").Inc()
}
