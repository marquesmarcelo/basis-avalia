package testhelpers

import (
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// DatabaseURLDeTeste devolve DATABASE_URL_TEST com um gate único: sem a
// variável e RUN_TESTS=true (a suíte "rodando para valer", serviço `test`
// do docker-compose.dev.yml), t.Fatal — pular um teste de integração em
// silêncio produz "ok" verde sem ter executado nada, e foi exatamente
// assim que 137 chamadas a BancoDeTeste em 33 arquivos ficaram invisíveis
// quando alguém rodava pelo serviço `backend` (sem DATABASE_URL_TEST) em
// vez de `test`. Sem RUN_TESTS=true (desenvolvedor rodando à mão, sem
// banco de teste local), t.Skip continua sendo o comportamento útil.
// Único ponto de verdade da regra — todo teste que precisa só do DSN cru
// (não de uma conexão já aberta) chama isto em vez de duplicar a checagem.
func DatabaseURLDeTeste(t *testing.T) string {
	t.Helper()

	url := os.Getenv("DATABASE_URL_TEST")
	if url == "" {
		if os.Getenv("RUN_TESTS") == "true" {
			t.Fatal("DATABASE_URL_TEST não definida com RUN_TESTS=true — a suíte de integração não pode ser pulada em silêncio; rode pelo serviço `test` do docker-compose.dev.yml")
		}
		t.Skip("DATABASE_URL_TEST não definida — pulando teste de integração")
	}
	return url
}

// BancoDeTeste conecta no banco de teste isolado (DATABASE_URL_TEST, ver
// CLAUDE.md) e fecha a conexão ao final via t.Cleanup.
func BancoDeTeste(t *testing.T) *sqlx.DB {
	t.Helper()

	url := DatabaseURLDeTeste(t)

	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		t.Fatalf("BancoDeTeste: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
