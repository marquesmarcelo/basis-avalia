package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// formasDeRegistroPermitidas — os únicos métodos de Registro (adapter/http)
// que podem registrar rota sob /api/v1 (design.md §6.1, T-128). Lista
// fechada: um quinto método em Registro não entra aqui sozinho — precisa
// de decisão explícita, porque é exatamente essa quinta forma
// não-descoberta que este guarda existe para impedir (D-26).
var formasDeRegistroPermitidas = map[string]bool{
	"Publica":                     true,
	"Autenticada":                 true,
	"AutenticadaSemPermissao":     true,
	"AutenticadaPorEscopoProprio": true,
}

// bypassSancionado — a única chamada em main.go que registra rota sob
// /api/v1 sem passar por Registro, e por quê: POST /auth/login não tem
// sessão a validar antes de logar, então não é nem Publica (reservada à
// rota do combo, 3.17) nem nenhuma das formas autenticadas (T-006).
// Qualquer OUTRA chamada direta em grupoAPI é bypass não sancionado.
const (
	metodoDoBypassSancionado  = "POST"
	caminhoDoBypassSancionado = "/auth/login"
)

// caminhoLiteralDoArgumento devolve o valor de uma string literal usada
// como argumento — "", false se o argumento não for uma constante de
// string (ex.: variável ou expressão). Nesse caso o guarda não consegue
// verificar o caminho e trata isso como falha, nunca como passagem
// silenciosa.
func caminhoLiteralDoArgumento(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	valor, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return valor, true
}

// TestGuardaRotas_QuatroFormasEBypassUnico verifica, por análise estática
// do próprio cmd/api/main.go (mesma técnica de TestGuardaC em
// internal/port, aqui com go/ast sobre um único arquivo), design.md §6.1
// (T-128):
//
//   - toda chamada em `registro.*` usa um dos quatro métodos da lista
//     fechada acima — método novo em Registro que ninguém cadastrar aqui
//     quebra o build, em vez de passar despercebido;
//   - `registro.Publica` tem exatamente uma ocorrência (reservada à rota
//     do combo público de instituições, spec 3.17);
//   - `registro.AutenticadaSemPermissao` só é chamada com caminho sob
//     `/auth/` ou `/version` — regra derivada do próprio literal de
//     caminho passado na chamada, não de uma lista de rotas aprovadas;
//   - nenhuma chamada direta em `grupoAPI` (bypass de Registro) existe
//     além da única exceção sancionada e documentada em main.go
//     (POST /auth/login, T-006).
//
// Limitação assumida: o guarda reconhece os bypasses e chamadas válidas
// pelo NOME do identificador receptor (`registro`, `grupoAPI`) tal como
// main.go os declara hoje. Não é uma prova independente da variável —
// é checagem do arquivo real, com o mesmo espírito de TestGuardaC. Se
// main.go um dia renomear essas variáveis, o guarda para de enxergar as
// chamadas (silêncio, não falha) — é um risco aceito, documentado aqui
// como o tripwire de T-118 documenta o seu.
func TestGuardaRotas_QuatroFormasEBypassUnico(t *testing.T) {
	fset := token.NewFileSet()
	arquivo, err := parser.ParseFile(fset, "main.go", nil, 0)
	if err != nil {
		t.Fatalf("ParseFile(main.go): %v", err)
	}

	contagemPublica := 0
	contagemBypassSancionado := 0

	ast.Inspect(arquivo, func(n ast.Node) bool {
		chamada, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := chamada.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receptor, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		switch receptor.Name {
		case "registro":
			if !formasDeRegistroPermitidas[sel.Sel.Name] {
				t.Errorf("registro.%s: método fora da lista fechada de formas de registro de rota — "+
					"toda rota sob /api/v1 usa exatamente uma das quatro formas (design.md §6.1, T-128)", sel.Sel.Name)
				return true
			}
			if sel.Sel.Name == "Publica" {
				contagemPublica++
				return true
			}
			if sel.Sel.Name != "AutenticadaSemPermissao" {
				return true
			}
			if len(chamada.Args) < 2 {
				t.Errorf("registro.AutenticadaSemPermissao com %d argumento(s) — não dá para verificar o caminho", len(chamada.Args))
				return true
			}
			caminho, ok := caminhoLiteralDoArgumento(chamada.Args[1])
			if !ok {
				t.Errorf("registro.AutenticadaSemPermissao: segundo argumento não é uma string literal — " +
					"o guarda não consegue verificar o caminho, e isso conta como falha (T-128)")
				return true
			}
			if !strings.HasPrefix(caminho, "/auth/") && !strings.HasPrefix(caminho, "/version") {
				t.Errorf("registro.AutenticadaSemPermissao(%q): caminho fora de /auth/ e /version — "+
					"rota de negócio não pode dispensar permissão nem escopo (design.md §6.1, T-128). "+
					"Use registro.Autenticada com a permissão, ou registro.AutenticadaPorEscopoProprio "+
					"quando a autorização É o escopo.", caminho)
			}

		case "grupoAPI":
			if len(chamada.Args) < 1 {
				t.Errorf("grupoAPI.%s com %d argumento(s) — bypass de Registro não verificável", sel.Sel.Name, len(chamada.Args))
				return true
			}
			caminho, ok := caminhoLiteralDoArgumento(chamada.Args[0])
			if !ok {
				t.Errorf("grupoAPI.%s: primeiro argumento não é string literal — bypass de Registro não verificável (T-128)", sel.Sel.Name)
				return true
			}
			if sel.Sel.Name == metodoDoBypassSancionado && caminho == caminhoDoBypassSancionado {
				contagemBypassSancionado++
				return true
			}
			t.Errorf("grupoAPI.%s(%q): bypassa Registro — toda rota sob /api/v1 usa exatamente uma das quatro formas, "+
				"exceto a única exceção sancionada e documentada em main.go (%s %s, T-006). Registre via registro.<Forma> (T-128).",
				sel.Sel.Name, caminho, metodoDoBypassSancionado, caminhoDoBypassSancionado)
		}
		return true
	})

	if contagemPublica != 1 {
		t.Errorf("registro.Publica: %d ocorrência(s) — a spec reserva Publica a uma única rota (3.17); esperado exatamente 1", contagemPublica)
	}
	if contagemBypassSancionado != 1 {
		t.Errorf("bypass sancionado de Registro (%s %s): %d ocorrência(s) — esperado exatamente 1 (T-006)",
			metodoDoBypassSancionado, caminhoDoBypassSancionado, contagemBypassSancionado)
	}
	t.Logf("Guarda de rotas: Publica=%d, bypass sancionado=%d — todas as demais chamadas de registro.* e grupoAPI.* validadas contra a lista fechada de quatro formas",
		contagemPublica, contagemBypassSancionado)
}
