package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Nomes de métrica compatíveis com a convenção OpenTelemetry HTTP semantic
// conventions — troca futura de exporter para OTel SDK é troca de client,
// não reescrita de instrumentação.
var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_server_requests_total",
		Help: "Total de requisições HTTP recebidas, por rota, método e status.",
	}, []string{"method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_duration_seconds",
		Help:    "Duração das requisições HTTP, por rota e método.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

// MetricasMiddleware registra contador de requisições e histograma de
// latência. Aplicado uma única vez na inicialização do servidor (middleware
// global) — nunca por handler individual.
func MetricasMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		inicio := time.Now()
		c.Next()

		rota := c.FullPath()
		if rota == "" {
			rota = "nao_encontrada"
		}

		httpRequestDuration.WithLabelValues(c.Request.Method, rota).Observe(time.Since(inicio).Seconds())
		httpRequestsTotal.WithLabelValues(c.Request.Method, rota, strconv.Itoa(c.Writer.Status())).Inc()
	}
}
