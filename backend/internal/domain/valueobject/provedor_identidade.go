package valueobject

// ProvedorIdentidade — único valor nesta entrega (3.11). Ponto de extensão
// para login federado (Google), sem ramificação de comportamento hoje.
type ProvedorIdentidade string

const CredencialLocal ProvedorIdentidade = "credencial_local"
