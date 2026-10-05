package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func contextoComQuery(query string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/x?"+query, nil)
	return c
}

var allowlistUsuario = map[string]bool{"nome": true, "email": true, "perfil": true, "criado_em": true}

func TestParsePaginacao_G10_PageSizeAcimaDoLimiteEhErro(t *testing.T) {
	c := contextoComQuery("page_size=10000")
	_, err := ParsePaginacao(c, allowlistUsuario, "nome", "asc")
	var pe *domain.ErrParametro
	if err == nil || !isErrParametro(err, &pe) {
		t.Fatalf("esperava ErrParametro, obtido %v", err)
	}
}

func isErrParametro(err error, target **domain.ErrParametro) bool {
	p, ok := err.(*domain.ErrParametro)
	if ok {
		*target = p
	}
	return ok
}

func TestParsePaginacao_G08_SortForaDaAllowlistEhErro(t *testing.T) {
	c := contextoComQuery("sort=senha_hash")
	_, err := ParsePaginacao(c, allowlistUsuario, "nome", "asc")
	if err == nil {
		t.Fatal("esperava erro para sort fora da allowlist")
	}
}

func TestParsePaginacao_SortVazioAplicaOPadrao(t *testing.T) {
	c := contextoComQuery("")
	p, err := ParsePaginacao(c, allowlistUsuario, "nome", "asc")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if p.Sort != "nome" || p.Order != "asc" {
		t.Fatalf("esperado padrão nome/asc, obtido %s/%s", p.Sort, p.Order)
	}
}

func TestParsePaginacao_OrderInvalidoEhErro(t *testing.T) {
	c := contextoComQuery("order=cima")
	_, err := ParsePaginacao(c, allowlistUsuario, "nome", "asc")
	if err == nil {
		t.Fatal("esperava erro para order inválido")
	}
}

func TestParsePaginacao_PageZeroEhErro(t *testing.T) {
	c := contextoComQuery("page=0")
	_, err := ParsePaginacao(c, allowlistUsuario, "nome", "asc")
	if err == nil {
		t.Fatal("esperava erro para page=0")
	}
}
