package instituicao

import (
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Instituicao — entidade de domínio (design.md §3.3). ExcluidoEm existe
// pelos campos base do CLAUDE.md, mas nenhum método a preenche (3.19):
// a instituição é inativada, nunca excluída.
type Instituicao struct {
	ID           uuid.UUID
	Nome         string
	Sigla        valueobject.Sigla
	CodigoEMec   valueobject.CodigoEMec
	Situacao     valueobject.SituacaoInstituicao
	CriadoEm     time.Time
	AtualizadoEm *time.Time
	ExcluidoEm   *time.Time
	Versao       int
}

func validarNome(bruto string) (string, error) {
	nome := strings.TrimSpace(bruto)
	if nome == "" || len(nome) > 200 {
		return "", &domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."}
	}
	return nome, nil
}

// NovaInstituicao gera o UUIDv7 no domínio e nasce sempre ativa, mesmo
// sem nenhum Pesquisador Institucional ainda (I-01, 3.20).
func NovaInstituicao(nomeBruto string, sigla valueobject.Sigla, codigoEMec valueobject.CodigoEMec) (*Instituicao, error) {
	nome, err := validarNome(nomeBruto)
	if err != nil {
		return nil, err
	}
	return &Instituicao{
		ID:         uuid.Must(uuid.NewV7()),
		Nome:       nome,
		Sigla:      sigla,
		CodigoEMec: codigoEMec,
		Situacao:   valueobject.Ativa,
		CriadoEm:   time.Now(),
		Versao:     1,
	}, nil
}

func (i *Instituicao) Renomear(nomeBruto string, sigla valueobject.Sigla, codigoEMec valueobject.CodigoEMec) error {
	nome, err := validarNome(nomeBruto)
	if err != nil {
		return err
	}
	i.Nome = nome
	i.Sigla = sigla
	i.CodigoEMec = codigoEMec
	return nil
}

func (i *Instituicao) AlterarSituacao(nova valueobject.SituacaoInstituicao) {
	i.Situacao = nova
}
