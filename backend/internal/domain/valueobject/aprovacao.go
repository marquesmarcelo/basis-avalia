package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// Aprovacao — par {data, órgão}, os dois ou nenhum (specs/plano-acao/
// design.md §3.1, PL-03). Value Object porque a invariante de par não é
// representável em dois campos soltos: não existe construtor que aceite
// data sem órgão.
type Aprovacao struct {
	data  DataLocal
	orgao OrgaoDeAprovacao
}

// NovaAprovacao devolve (Aprovacao{}, false, nil) quando os dois campos
// vêm vazios — o plano existe, e vige, antes de ser aprovado
// (spec.md 3.2). Erro só quando exatamente um dos dois veio preenchido.
func NovaAprovacao(dataBruta, orgaoBruto string) (Aprovacao, bool, error) {
	if dataBruta == "" && orgaoBruto == "" {
		return Aprovacao{}, false, nil
	}
	if dataBruta == "" || orgaoBruto == "" {
		return Aprovacao{}, false, domain.ErrAprovacaoIncompleta
	}
	data, err := DataLocalTexto(dataBruta)
	if err != nil {
		return Aprovacao{}, false, domain.ErrAprovacaoIncompleta
	}
	orgao, err := NovoOrgaoDeAprovacao(orgaoBruto)
	if err != nil {
		return Aprovacao{}, false, err
	}
	return Aprovacao{data: data, orgao: orgao}, true, nil
}

func (a Aprovacao) Data() DataLocal         { return a.data }
func (a Aprovacao) Orgao() OrgaoDeAprovacao { return a.orgao }
