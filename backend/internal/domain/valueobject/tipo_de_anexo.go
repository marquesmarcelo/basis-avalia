package valueobject

import (
	"archive/zip"
	"bytes"
	"net/http"
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
)

// TipoDeAnexo — cinco valores fechados (specs/metas-coordenacao/spec.md
// PM-5). Construído SEMPRE a partir do conteúdo real do arquivo, nunca da
// extensão.
type TipoDeAnexo string

const (
	AnexoPDF  TipoDeAnexo = "pdf"
	AnexoJPEG TipoDeAnexo = "jpeg"
	AnexoPNG  TipoDeAnexo = "png"
	AnexoDOCX TipoDeAnexo = "docx"
	AnexoODT  TipoDeAnexo = "odt"
)

// MimeType — usado tanto ao gravar no armazenamento (Content-Type do
// objeto) quanto ao servir o download (Content-Type da resposta).
func (t TipoDeAnexo) MimeType() string {
	switch t {
	case AnexoPDF:
		return "application/pdf"
	case AnexoJPEG:
		return "image/jpeg"
	case AnexoPNG:
		return "image/png"
	case AnexoDOCX:
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case AnexoODT:
		return "application/vnd.oasis.opendocument.text"
	default:
		return "application/octet-stream"
	}
}

func NovoTipoDeAnexo(bruto string) (TipoDeAnexo, error) {
	t := TipoDeAnexo(bruto)
	switch t {
	case AnexoPDF, AnexoJPEG, AnexoPNG, AnexoDOCX, AnexoODT:
		return t, nil
	default:
		return "", domain.ErrAnexoTipoNaoPermitido
	}
}

// DetectarTipoDeAnexo verifica o tipo pelo CONTEÚDO real, em dois níveis
// (specs/metas-coordenacao/design.md §6.2, M-07, AN-02):
//
//  1. http.DetectContentType resolve PDF (assinatura "%PDF-"), PNG e JPEG
//     diretamente.
//  2. Se o resultado for "application/zip", DOCX e ODT são ambos arquivos
//     ZIP — DetectContentType sozinho aceita um .jar renomeado para
//     .docx. O segundo nível abre o ZIP e confere a estrutura interna:
//     DOCX exige "[Content_Types].xml" e uma entrada sob "word/"; ODT
//     exige a entrada "mimetype" começando com
//     "application/vnd.oasis.opendocument.text". Qualquer outro ZIP é
//     recusado.
//
// conteudo é o arquivo inteiro, já limitado a
// anexo.LimiteBytesPorArquivo por http.MaxBytesReader/io.LimitReader
// antes de chegar aqui (M-08) — nunca um stream ilimitado. A extensão do
// nome do arquivo nunca participa desta decisão.
func DetectarTipoDeAnexo(conteudo []byte) (TipoDeAnexo, error) {
	amostra := conteudo
	if len(amostra) > 512 {
		amostra = amostra[:512]
	}
	switch http.DetectContentType(amostra) {
	case "application/pdf":
		return AnexoPDF, nil
	case "image/png":
		return AnexoPNG, nil
	case "image/jpeg":
		return AnexoJPEG, nil
	case "application/zip", "application/x-zip-compressed":
		return detectarTipoDeZip(conteudo)
	default:
		return "", domain.ErrAnexoTipoNaoPermitido
	}
}

func detectarTipoDeZip(conteudo []byte) (TipoDeAnexo, error) {
	leitor, err := zip.NewReader(bytes.NewReader(conteudo), int64(len(conteudo)))
	if err != nil {
		return "", domain.ErrAnexoTipoNaoPermitido
	}

	temContentTypes, temWord := false, false
	for _, arq := range leitor.File {
		switch {
		case arq.Name == "[Content_Types].xml":
			temContentTypes = true
		case strings.HasPrefix(arq.Name, "word/"):
			temWord = true
		case arq.Name == "mimetype":
			if conteudoDoMimetypeODT(arq) {
				return AnexoODT, nil
			}
		}
	}
	if temContentTypes && temWord {
		return AnexoDOCX, nil
	}
	return "", domain.ErrAnexoTipoNaoPermitido
}

func conteudoDoMimetypeODT(arq *zip.File) bool {
	f, err := arq.Open()
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 64)
	n, _ := f.Read(buf)
	return strings.HasPrefix(string(buf[:n]), "application/vnd.oasis.opendocument.text")
}
