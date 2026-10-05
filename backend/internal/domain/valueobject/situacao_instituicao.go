package valueobject

// SituacaoInstituicao — ativa ou inativa (3.19). Distinto de exclusão
// lógica: instituição nunca é excluída, apenas inativada.
type SituacaoInstituicao string

const (
	Ativa   SituacaoInstituicao = "ativa"
	Inativa SituacaoInstituicao = "inativa"
)

func (s SituacaoInstituicao) Alternar() SituacaoInstituicao {
	if s == Ativa {
		return Inativa
	}
	return Ativa
}
