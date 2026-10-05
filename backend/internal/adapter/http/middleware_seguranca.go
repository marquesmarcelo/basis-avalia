package http

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// MiddlewareSeguranca aplica os cabeçalhos de segurança da borda HTTP da
// API (design.md §4.5, §16 T-098) — registrado antes de tudo, na porta
// pública. CSP e Cache-Control têm exceção de rota; HSTS ramifica por
// ambiente; os demais são incondicionais.
func MiddlewareSeguranca(appEnv string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		if !strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			c.Header("Cache-Control", "no-store")
		}
		if appEnv == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
