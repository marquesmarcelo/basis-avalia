package metricas

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// notificacoesPendentes — gauge crescendo é relê parado (fundacao-metas.md
// §7, design.md §8.4): é o que transforma "recusa registrada sem e-mail
// enviado" em alerta, não descoberta.
var notificacoesPendentes = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "notificacoes_pendentes",
	Help: "Notificações de entrega geradas e ainda não enviadas.",
})

var notificacoesEnviadasTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "notificacoes_enviadas_total",
	Help: "Notificações de entrega enviadas com sucesso, por evento.",
}, []string{"evento"})

func RegistrarNotificacoesPendentes(n int) {
	notificacoesPendentes.Set(float64(n))
}

func RegistrarNotificacaoEnviada(evento string) {
	notificacoesEnviadasTotal.WithLabelValues(evento).Inc()
}
