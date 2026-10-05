// Package arquitetura não contém código de produção — só o guarda de
// T-119 (design.md de autenticacao-usuarios §4.6, D-29). Vive fora de
// domain/usecase/adapter de propósito: um guarda que morasse dentro do
// que ele verifica testaria a si mesmo por acidente sempre que o pacote
// hospedeiro mudasse de forma.
package arquitetura

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const (
	moduloRaiz     = "github.com/basis-avalia/backend/..."
	prefixoDomain  = "github.com/basis-avalia/backend/internal/domain"
	prefixoUsecase = "github.com/basis-avalia/backend/internal/usecase"
	prefixoAdapter = "github.com/basis-avalia/backend/internal/adapter"
)

// bibliotecaProibida — biblioteca de infraestrutura que domain e usecase
// nunca importam (design.md §4.6, §15.1). Prefixo, não caminho exato,
// porque pgx e jwt versionam o caminho de import (pgx/v5, jwt/v5) e pgx
// tem subpacotes usados via driver (pgxpool) que também violam a regra.
var bibliotecaProibida = []struct {
	prefixo string
	motivo  string
}{
	{"github.com/gin-gonic/gin", "gin é o framework HTTP — pertence a adapter/http"},
	{"github.com/jmoiron/sqlx", "sqlx é acesso a banco — pertence a adapter/postgres"},
	{"github.com/jackc/pgx", "pgx é o driver Postgres — pertence a adapter/postgres"},
	{"github.com/prometheus/client_golang", "prometheus é métrica técnica — pertence a adapter/metricas"},
	{"github.com/golang-jwt/jwt", "jwt pertence a adapter/jwt"},
	{"golang.org/x/crypto/argon2", "argon2 é hash de senha — pertence a adapter/argon2"},
}

func ehSubPacoteDe(caminho, prefixo string) bool {
	return caminho == prefixo || strings.HasPrefix(caminho, prefixo+"/")
}

func motivoDaProibicao(caminhoImportado string) (string, bool) {
	for _, b := range bibliotecaProibida {
		if ehSubPacoteDe(caminhoImportado, b.prefixo) {
			return b.motivo, true
		}
	}
	return "", false
}

// carregarTodosOsPacotes varre TODO o módulo com "./..." — o mesmo
// módulo, não uma lista de subpacotes escrita à mão. Pacote novo sob
// internal/domain ou internal/usecase é verificado por padrão, sem
// ninguém precisar lembrar de cadastrá-lo aqui (D-26: cobertura que
// depende de lista escrita à mão não conta como mecanismo).
func carregarTodosOsPacotes(t *testing.T) []*packages.Package {
	t.Helper()
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports}
	pkgs, err := packages.Load(cfg, moduloRaiz)
	if err != nil {
		t.Fatalf("falha ao carregar %s: %v", moduloRaiz, err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("packages.Load(%s) não retornou nenhum pacote — guarda não pode varrer o que não carregou", moduloRaiz)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			t.Fatalf("pacote %s carregou com erro (build quebrado invalida o guarda): %v", pkg.PkgPath, pkg.Errors)
		}
	}
	return pkgs
}

// TestGuardaDependencias_DirecaoNuncaApontaParaInfraestrutura verifica
// design.md §4.6 (D-29): domain nunca importa usecase nem adapter;
// usecase nunca importa adapter; nenhum dos dois importa biblioteca de
// infraestrutura (gin, sqlx, pgx, prometheus, jwt, argon2) — essas só
// entram pelo adapter correspondente, nunca por trás dele.
//
// Verificação por IMPORTS DIRETOS de cada pacote, não por lista de
// pacotes vigiados: todo pacote sob internal/domain e internal/usecase é
// inspecionado, sempre, porque a varredura parte do módulo inteiro
// (carregarTodosOsPacotes). Isso também cobre importação indireta por
// transitividade — se A (domain) importa B (domain) e B importa
// adapter, o guarda pega a violação ao inspecionar B diretamente, não
// precisa seguir a cadeia a partir de A.
func TestGuardaDependencias_DirecaoNuncaApontaParaInfraestrutura(t *testing.T) {
	pkgs := carregarTodosOsPacotes(t)

	pacotesDomain, pacotesUsecase := 0, 0
	for _, pkg := range pkgs {
		ehDomain := ehSubPacoteDe(pkg.PkgPath, prefixoDomain)
		ehUsecase := ehSubPacoteDe(pkg.PkgPath, prefixoUsecase)
		if !ehDomain && !ehUsecase {
			continue
		}
		if ehDomain {
			pacotesDomain++
		} else {
			pacotesUsecase++
		}

		for caminhoImportado := range pkg.Imports {
			if ehDomain {
				if ehSubPacoteDe(caminhoImportado, prefixoUsecase) {
					t.Errorf("%s (domain) importa %s (usecase) — domain nunca depende de usecase (design.md §4.6)", pkg.PkgPath, caminhoImportado)
				}
				if ehSubPacoteDe(caminhoImportado, prefixoAdapter) {
					t.Errorf("%s (domain) importa %s (adapter) — domain nunca depende de adapter (design.md §4.6)", pkg.PkgPath, caminhoImportado)
				}
			}
			if ehUsecase && ehSubPacoteDe(caminhoImportado, prefixoAdapter) {
				t.Errorf("%s (usecase) importa %s (adapter) — usecase nunca depende de adapter, só de port (design.md §4.6)", pkg.PkgPath, caminhoImportado)
			}
			if ehDomain || ehUsecase {
				if motivo, proibida := motivoDaProibicao(caminhoImportado); proibida {
					camada := "domain"
					if ehUsecase {
						camada = "usecase"
					}
					t.Errorf("%s (%s) importa %s diretamente — %s (design.md §4.6)", pkg.PkgPath, camada, caminhoImportado, motivo)
				}
			}
		}
	}

	if pacotesDomain == 0 || pacotesUsecase == 0 {
		t.Fatalf("guarda não encontrou pacotes sob internal/domain (%d) ou internal/usecase (%d) — "+
			"módulo carregado errado, ou reestruturação de pastas quebrou o prefixo verificado", pacotesDomain, pacotesUsecase)
	}
	t.Logf("Guarda de dependências: %d pacote(s) de domain e %d pacote(s) de usecase inspecionados no módulo inteiro", pacotesDomain, pacotesUsecase)
}
