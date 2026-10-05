package valueobject

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
)

func construirZip(t *testing.T, arquivos map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for nome, conteudo := range arquivos {
		f, err := w.Create(nome)
		if err != nil {
			t.Fatalf("criando entrada de zip: %v", err)
		}
		if _, err := f.Write([]byte(conteudo)); err != nil {
			t.Fatalf("escrevendo entrada de zip: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("fechando zip: %v", err)
	}
	return buf.Bytes()
}

// AN-02, nível 1: executável renomeado para PDF é recusado pela
// assinatura de conteúdo (design.md §6.2).
func TestDetectarTipoDeAnexo_ExecutavelRenomeadoParaPDF(t *testing.T) {
	falso := []byte("MZ\x90\x00\x03\x00\x00\x00conteudo de executavel qualquer")
	_, err := DetectarTipoDeAnexo(falso)
	if err != domain.ErrAnexoTipoNaoPermitido {
		t.Fatalf("esperava ErrAnexoTipoNaoPermitido, obteve %v", err)
	}
}

// AN-02, nível 2, a armadilha do design (§6.2): um ZIP qualquer (não DOCX,
// não ODT) renomeado para .docx é recusado só depois de abrir o ZIP —
// http.DetectContentType sozinho aceitaria, porque devolve
// "application/zip" para qualquer ZIP, inclusive um .jar.
func TestDetectarTipoDeAnexo_ZipQualquerRenomeadoParaDocx(t *testing.T) {
	zipGenerico := construirZip(t, map[string]string{
		"META-INF/MANIFEST.MF": "Manifest-Version: 1.0\n",
		"com/exemplo/Main.class": "conteudo binario qualquer",
	})
	_, err := DetectarTipoDeAnexo(zipGenerico)
	if err != domain.ErrAnexoTipoNaoPermitido {
		t.Fatalf("esperava ErrAnexoTipoNaoPermitido, obteve %v", err)
	}
}

func TestDetectarTipoDeAnexo_PDFValido(t *testing.T) {
	pdf := []byte("%PDF-1.4\n%conteudo de pdf de teste")
	tipo, err := DetectarTipoDeAnexo(pdf)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if tipo != AnexoPDF {
		t.Fatalf("esperava pdf, obteve %s", tipo)
	}
}

func TestDetectarTipoDeAnexo_DocxValido(t *testing.T) {
	docx := construirZip(t, map[string]string{
		"[Content_Types].xml": "<Types/>",
		"word/document.xml":   "<document/>",
	})
	tipo, err := DetectarTipoDeAnexo(docx)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if tipo != AnexoDOCX {
		t.Fatalf("esperava docx, obteve %s", tipo)
	}
}

func TestDetectarTipoDeAnexo_OdtValido(t *testing.T) {
	odt := construirZip(t, map[string]string{
		"mimetype":  "application/vnd.oasis.opendocument.text",
		"content.xml": "<office/>",
	})
	tipo, err := DetectarTipoDeAnexo(odt)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if tipo != AnexoODT {
		t.Fatalf("esperava odt, obteve %s", tipo)
	}
}
