package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func montarRouterComSeguranca(appEnv string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MiddlewareSeguranca(appEnv))
	router.GET("/api/v1/publico/instituicoes", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/swagger/index.html", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/version", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func TestMiddlewareSeguranca_CincoCabecalhosIncondicionais(t *testing.T) {
	router := montarRouterComSeguranca("development")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/publico/instituicoes", nil))

	esperados := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
		"Permissions-Policy":     "camera=(), microphone=(), geolocation=()",
	}
	for cabecalho, valor := range esperados {
		if got := rec.Header().Get(cabecalho); got != valor {
			t.Errorf("%s = %q, esperava %q", cabecalho, got, valor)
		}
	}
}

func TestMiddlewareSeguranca_CSPPresenteForaDoSwagger(t *testing.T) {
	router := montarRouterComSeguranca("development")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/publico/instituicoes", nil))

	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'none'; frame-ancestors 'none'" {
		t.Errorf("CSP = %q", got)
	}
}

func TestMiddlewareSeguranca_CSPAusenteNoSwagger(t *testing.T) {
	router := montarRouterComSeguranca("development")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))

	if got := rec.Header().Get("Content-Security-Policy"); got != "" {
		t.Errorf("CSP não deveria estar presente em /swagger/*, obtido %q", got)
	}
}

func TestMiddlewareSeguranca_CacheControlSoEmApiV1(t *testing.T) {
	router := montarRouterComSeguranca("development")

	recApi := httptest.NewRecorder()
	router.ServeHTTP(recApi, httptest.NewRequest(http.MethodGet, "/api/v1/publico/instituicoes", nil))
	if got := recApi.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control em /api/v1/*: %q, esperava no-store", got)
	}

	recVersion := httptest.NewRecorder()
	router.ServeHTTP(recVersion, httptest.NewRequest(http.MethodGet, "/version", nil))
	if got := recVersion.Header().Get("Cache-Control"); got != "" {
		t.Errorf("Cache-Control não deveria estar presente em /version, obtido %q", got)
	}
}

func TestMiddlewareSeguranca_HSTSSoEmProducao(t *testing.T) {
	recDev := httptest.NewRecorder()
	montarRouterComSeguranca("development").ServeHTTP(recDev, httptest.NewRequest(http.MethodGet, "/version", nil))
	if got := recDev.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS não deveria aparecer em development, obtido %q", got)
	}

	recProd := httptest.NewRecorder()
	montarRouterComSeguranca("production").ServeHTTP(recProd, httptest.NewRequest(http.MethodGet, "/version", nil))
	if got := recProd.Header().Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains" {
		t.Errorf("HSTS em production = %q", got)
	}
}
