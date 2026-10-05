package valueobject

// SituacaoPlano — rascunho, vigente ou encerrado (specs/plano-acao/
// design.md §3.1). Só as duas primeiras são persistíveis
// (situacao_publicacao); "encerrado" é sempre derivada de
// Plano.SituacaoEfetiva e nunca gravada no banco. O tipo carrega o
// terceiro valor porque é o que a API expõe ao cliente.
type SituacaoPlano string

const (
	PlanoRascunho  SituacaoPlano = "rascunho"
	PlanoVigente   SituacaoPlano = "vigente"
	PlanoEncerrado SituacaoPlano = "encerrado"
)
