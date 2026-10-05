package documento

import (
	"time"

	"github.com/google/uuid"
)

// Documento — registro do .docx gerado para um plano (specs/plano-acao/
// design.md §4.1). Imutável por natureza: cada geração cria uma linha
// nova, a anterior nunca é sobrescrita — por isso não tem Versao (a
// exceção declarada em fundacao-metas.md §10, dado escrito uma vez).
type Documento struct {
	ID            uuid.UUID
	PlanoID       uuid.UUID
	CursoID       uuid.UUID
	InstituicaoID uuid.UUID
	ChaveObjeto   string
	NomeArquivo   string
	SituacaoNoAto string
	GeradoPor     uuid.UUID
	GeradoEm      time.Time
	CriadoEm      time.Time
	ExcluidoEm    *time.Time
}

func NovoDocumento(planoID, cursoID, instituicaoID uuid.UUID, chaveObjeto, nomeArquivo, situacaoNoAto string, geradoPor uuid.UUID) *Documento {
	agora := time.Now()
	return &Documento{
		ID: uuid.Must(uuid.NewV7()), PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID,
		ChaveObjeto: chaveObjeto, NomeArquivo: nomeArquivo, SituacaoNoAto: situacaoNoAto,
		GeradoPor: geradoPor, GeradoEm: agora, CriadoEm: agora,
	}
}
