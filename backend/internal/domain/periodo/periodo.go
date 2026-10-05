package periodo

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Periodo — entidade de domínio (specs/plano-acao/design.md §3.2). Sem
// coluna de situação: reusa valueobject.Vigencia, de cursos — "o dia da
// data de fim entra inteiro" é a mesma regra, e duas cópias divergem na
// primeira correção (P-01).
type Periodo struct {
	ID            uuid.UUID
	InstituicaoID uuid.UUID
	Nome          valueobject.NomeCatalogo
	Vigencia      valueobject.Vigencia
	CriadoEm      time.Time
	AtualizadoEm  *time.Time
	ExcluidoEm    *time.Time
	Versao        int
}

func NovoPeriodo(instituicaoID uuid.UUID, nomeBruto string, inicio valueobject.DataLocal, fim valueobject.DataLocal) (*Periodo, error) {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return nil, err
	}
	vigencia, err := valueobject.NovaVigencia(inicio, &fim)
	if err != nil {
		return nil, err
	}
	return &Periodo{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: instituicaoID, Nome: nome,
		Vigencia: vigencia, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

func (p *Periodo) Atualizar(nomeBruto string, inicio, fim valueobject.DataLocal) error {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return err
	}
	vigencia, err := valueobject.NovaVigencia(inicio, &fim)
	if err != nil {
		return err
	}
	p.Nome, p.Vigencia = nome, vigencia
	return nil
}

// SituacaoEm devolve "nao_iniciado" / "aberto" / "encerrado" — mesmo
// mapeamento de futura/vigente/encerrada de Vigencia.SituacaoEm
// (design.md §3.1), com o rótulo que a API de período usa.
func (p *Periodo) SituacaoEm(hoje valueobject.DataLocal) string {
	switch p.Vigencia.SituacaoEm(hoje) {
	case valueobject.Futura:
		return "nao_iniciado"
	case valueobject.Vigente:
		return "aberto"
	default:
		return "encerrado"
	}
}
