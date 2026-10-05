package autorizacao

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

// TestEscopoInconstruivelForaDeAutorizar é teste de MECANISMO (fundacao-
// metas.md §3.4, herdado de design.md §3.2 de autenticacao-usuarios):
// prova, e não apenas documenta, que um Escopo só nasce de Autorizar.
// Duas verificações independentes, cada uma capaz de pegar uma regressão
// que a outra deixaria passar:
//
//  1. Nenhum campo de Escopo é exportado — não dá para montar um Escopo
//     por literal fora do pacote (nem por engano, nem "só para testar").
//  2. Nenhuma função exportada do pacote, além de Autorizar, devolve
//     Escopo — não dá para contornar (1) criando um segundo construtor.
func TestEscopoInconstruivelForaDeAutorizar(t *testing.T) {
	tipo := reflect.TypeOf(Escopo{})
	for i := 0; i < tipo.NumField(); i++ {
		campo := tipo.Field(i)
		if campo.IsExported() {
			t.Fatalf("campo %q de Escopo está exportado — quebra a garantia de só Autorizar construir um Escopo", campo.Name)
		}
	}

	fset := token.NewFileSet()
	pacotes, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	for _, pacote := range pacotes {
		for _, arquivo := range pacote.Files {
			for _, decl := range arquivo.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil { // método, não construtor livre
					continue
				}
				if !fn.Name.IsExported() {
					continue
				}
				if fn.Name.Name == "Autorizar" {
					continue
				}
				if devolveEscopo(fn) {
					t.Fatalf("função exportada %q devolve Escopo — segundo construtor fora de Autorizar", fn.Name.Name)
				}
			}
		}
	}
}

func devolveEscopo(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, campo := range fn.Type.Results.List {
		if ident, ok := campo.Type.(*ast.Ident); ok && ident.Name == "Escopo" {
			return true
		}
	}
	return false
}
