package anexo

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// LimiteBytesPorArquivo, LimiteArquivosPorEntrega e LimiteBytesPorEntrega
// — PM-5 da spec: 10 MB por arquivo, 10 arquivos, 50 MB por entrega.
const (
	LimiteBytesPorArquivo   = 10 * 1024 * 1024
	LimiteArquivosPorEntrega = 10
	LimiteBytesPorEntrega    = 50 * 1024 * 1024
)

// Anexo — entidade de domínio (specs/metas-coordenacao/design.md §3.3).
// Sem Versao: escrito uma vez, removido logicamente só pelo autor da
// entrega (fundacao-metas.md §10). NomeOriginal pode conter dado pessoal
// de terceiros e nunca vai para log, e-mail ou mensagem de erro (§9 da
// spec).
type Anexo struct {
	ID, EntregaID, CursoID, InstituicaoID uuid.UUID
	NomeOriginal                          string
	Tipo                                  valueobject.TipoDeAnexo
	TamanhoBytes                          int64
	ChaveObjeto                           string
	HashSHA256                            string
	CriadoEm                              time.Time
	ExcluidoEm                            *time.Time
}

func NovoAnexo(entregaID, cursoID, instituicaoID uuid.UUID, nomeOriginal string, tipo valueobject.TipoDeAnexo, tamanhoBytes int64, chaveObjeto, hash string) *Anexo {
	return &Anexo{
		ID: uuid.Must(uuid.NewV7()), EntregaID: entregaID, CursoID: cursoID, InstituicaoID: instituicaoID,
		NomeOriginal: nomeOriginal, Tipo: tipo, TamanhoBytes: tamanhoBytes,
		ChaveObjeto: chaveObjeto, HashSHA256: hash, CriadoEm: time.Now(),
	}
}
