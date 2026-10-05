package autorizacao_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConjuntoEfetivo_UnicaFonteDoPerfilDerivado é o guarda mecânico de
// T-161 (specs/cursos/tasks.md): varre todo o backend (fora de _test.go)
// atrás do idioma que acrescenta CoordenadorCurso a um ConjuntoDePerfis, e
// exige que ele exista em EXATAMENTE um lugar — MontarConjuntoEfetivo.
// Alguém que reintroduza esse acréscimo em outro ponto (ex: direto no
// middleware, ou num use case) quebra este teste, mesmo que o resto
// compile e passe.
func TestConjuntoEfetivo_UnicaFonteDoPerfilDerivado(t *testing.T) {
	raiz := filepath.Join("..", "..") // backend/internal/domain/autorizacao -> backend/internal
	var achados []string

	err := filepath.WalkDir(raiz, func(caminho string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(caminho, ".go") || strings.HasSuffix(caminho, "_test.go") {
			return nil
		}
		conteudo, err := os.ReadFile(caminho)
		if err != nil {
			return err
		}
		// Idioma exato de acréscimo: append(...) na mesma linha que
		// CoordenadorCurso — e não o que só o COMPARA ou o DECLARA
		// (matrizPermissoes, ordemCanonica, o switch de NovoConjunto,
		// PodeAtribuir): essas não constroem um conjunto novo com o perfil
		// acrescentado, só reconhecem que ele existe.
		for _, linha := range strings.Split(string(conteudo), "\n") {
			if strings.Contains(linha, "append(") && strings.Contains(linha, "CoordenadorCurso") {
				achados = append(achados, caminho+": "+strings.TrimSpace(linha))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("varredura: %v", err)
	}

	if len(achados) != 1 {
		t.Fatalf("esperava exatamente 1 lugar que acrescenta CoordenadorCurso a um conjunto, encontrei %d: %v", len(achados), achados)
	}
	if !strings.Contains(achados[0], "conjunto_efetivo.go") {
		t.Fatalf("o único acréscimo deveria estar em conjunto_efetivo.go (MontarConjuntoEfetivo), encontrei em %s", achados[0])
	}
}
