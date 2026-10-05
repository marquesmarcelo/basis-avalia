package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestReadiness_CredencialInvalida_CorpoNaoVazaDetalheDoDriver prova
// design.md §16 T-103 (achado M-2): com uma dependência indisponível, o
// corpo da resposta nunca carrega usuário, banco, IP, porta nem SQLSTATE
// do driver — só o nome da dependência. O erro real vai para o log do
// servidor, não para quem chamou a rota.
func TestReadiness_CredencialInvalida_CorpoNaoVazaDetalheDoDriver(t *testing.T) {
	bruta := testhelpers.DatabaseURLDeTeste(t)
	dsn, err := url.Parse(bruta)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}
	dsn.User = url.UserPassword(dsn.User.Username(), "senha-errada-de-proposito")

	dbRuim, err := sqlx.Open("pgx", dsn.String())
	if err != nil {
		t.Fatalf("sqlx.Open: %v", err)
	}
	t.Cleanup(func() { dbRuim.Close() })

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}
	opcoesRedis, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("redis.ParseURL: %v", err)
	}
	redisClient := redis.NewClient(opcoesRedis)
	t.Cleanup(func() { redisClient.Close() })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHealthHandler(dbRuim, redisClient, "0.0.0-teste")
	router.GET("/readyz", handler.Readiness)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503, obtido %d (%s)", rec.Code, rec.Body.String())
	}

	corpoBruto := rec.Body.String()
	proibidos := []string{"basisavalia", "senha-errada-de-proposito", "SQLSTATE", "password authentication", ":5432"}
	for _, termo := range proibidos {
		if strings.Contains(corpoBruto, termo) {
			t.Errorf("corpo da resposta vazou detalhe do driver (%q): %s", termo, corpoBruto)
		}
	}

	var corpo struct {
		Status string   `json:"status"`
		Falhas []string `json:"falhas"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("resposta não é JSON válido: %v (%s)", err, corpoBruto)
	}
	if corpo.Status != "indisponivel" {
		t.Errorf(`status = %q, esperava "indisponivel"`, corpo.Status)
	}
	achouPostgres := false
	for _, f := range corpo.Falhas {
		if f == "postgres" {
			achouPostgres = true
		}
	}
	if !achouPostgres {
		t.Errorf(`falhas deveria conter "postgres", obtido %v`, corpo.Falhas)
	}
}
