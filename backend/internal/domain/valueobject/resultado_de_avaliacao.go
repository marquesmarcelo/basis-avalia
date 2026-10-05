package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// ResultadoDeAvaliacao — "aceita" ou "recusada" (specs/metas-coordenacao/
// design.md §3.1). Recusar exige motivo não vazio — validado aqui, no
// construtor, para que nenhuma instância inválida exista (AV-02).
type ResultadoDeAvaliacao struct {
	valor  string
	motivo string
}

func NovoResultadoDeAvaliacao(valorBruto, motivo string) (ResultadoDeAvaliacao, error) {
	switch valorBruto {
	case "aceita":
		return ResultadoDeAvaliacao{valor: "aceita"}, nil
	case "recusada":
		if trimVazioRA(motivo) {
			return ResultadoDeAvaliacao{}, domain.ErrMotivoObrigatorio
		}
		if err := ProibirCaractereDeControle(motivo); err != nil {
			return ResultadoDeAvaliacao{}, err
		}
		return ResultadoDeAvaliacao{valor: "recusada", motivo: motivo}, nil
	default:
		return ResultadoDeAvaliacao{}, domain.ErrValorInvalido
	}
}

func (r ResultadoDeAvaliacao) Aceita() bool    { return r.valor == "aceita" }
func (r ResultadoDeAvaliacao) Recusada() bool  { return r.valor == "recusada" }
func (r ResultadoDeAvaliacao) Motivo() string  { return r.motivo }
func (r ResultadoDeAvaliacao) String() string  { return r.valor }

func trimVazioRA(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
