package valueobject

import "github.com/basis-avalia/backend/internal/domain"

// SituacaoDesignacao — rótulo derivado de Vigencia.SituacaoEm, nunca
// coluna de banco (specs/cursos/design.md §4.3, C-04).
type SituacaoDesignacao string

const (
	Futura    SituacaoDesignacao = "futura"
	Vigente   SituacaoDesignacao = "vigente"
	Encerrada SituacaoDesignacao = "encerrada"
)

// Vigencia — par {inicio, fim opcional} de uma designação (specs/cursos/
// design.md §3.1). Reusada por Periodo em plano-acao.
type Vigencia struct {
	inicio DataLocal
	fim    *DataLocal
}

func NovaVigencia(inicio DataLocal, fim *DataLocal) (Vigencia, error) {
	if fim != nil && fim.AntesDe(inicio) {
		return Vigencia{}, domain.ErrDesignacaoDatasInvalidas
	}
	return Vigencia{inicio: inicio, fim: fim}, nil
}

func (v Vigencia) Inicio() DataLocal { return v.inicio }
func (v Vigencia) Fim() *DataLocal   { return v.fim }

// SituacaoEm implementa DG-06 (specs/cursos/spec.md §3.2): o dia da data de
// fim entra INTEIRO. Escrita como "!fim.AntesDe(hoje)" — não
// "fim.DepoisDe(hoje)" — de propósito: não existe um "<" ou ">" aqui para
// alguém trocar por "<=" ou ">=" e inverter a fronteira por engano.
func (v Vigencia) SituacaoEm(hoje DataLocal) SituacaoDesignacao {
	if hoje.AntesDe(v.inicio) {
		return Futura
	}
	if v.fim == nil || !v.fim.AntesDe(hoje) {
		return Vigente
	}
	return Encerrada
}
