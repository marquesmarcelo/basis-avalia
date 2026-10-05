package autorizacao

import (
	"errors"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

var ErrAtorInvalido = errors.New("ator inválido")

// Ator — Value Object imutável. NovoAtor recusa a combinação impossível:
// conjunto institucional com vínculo nulo, ou {administrador_sistema} com
// vínculo preenchido (design.md §3.2).
//
// dataDeReferencia (fundacao-metas.md §5.2) nunca é preenchida por
// NovoAtor — ela é fixada uma única vez por requisição, no middleware de
// sessão, via ComDataDeReferencia. Não entra no construtor porque a
// maioria dos chamadores de NovoAtor (login, testes que não tocam datas)
// não tem nem precisa de uma leitura de relógio; exigi-la ali obrigaria
// todo call site existente a fornecer uma data que não usa.
type Ator struct {
	usuarioID        uuid.UUID
	perfis           valueobject.ConjuntoDePerfis
	instituicaoID    *uuid.UUID
	dataDeReferencia valueobject.DataLocal
}

func NovoAtor(usuarioID uuid.UUID, perfis valueobject.ConjuntoDePerfis, instituicaoID *uuid.UUID) (Ator, error) {
	institucional := perfis.PertenceAInstituicao()
	if institucional && instituicaoID == nil {
		return Ator{}, ErrAtorInvalido
	}
	if !institucional && instituicaoID != nil {
		return Ator{}, ErrAtorInvalido
	}
	return Ator{usuarioID: usuarioID, perfis: perfis, instituicaoID: instituicaoID}, nil
}

func (a Ator) UsuarioID() uuid.UUID                    { return a.usuarioID }
func (a Ator) Perfis() valueobject.ConjuntoDePerfis    { return a.perfis }
func (a Ator) InstituicaoID() *uuid.UUID               { return a.instituicaoID }
func (a Ator) DataDeReferencia() valueobject.DataLocal { return a.dataDeReferencia }

// ComDataDeReferencia devolve uma cópia do Ator com a data de referência
// fixada — nunca muta o receptor (Value Object imutável). Único chamador
// sancionado: o middleware de sessão, uma vez por requisição
// (fundacao-metas.md §5.2).
func (a Ator) ComDataDeReferencia(d valueobject.DataLocal) Ator {
	a.dataDeReferencia = d
	return a
}

// Proprio devolve a prova de que o identificador a ser usado veio da
// identidade já estabelecida (o próprio Ator), nunca de entrada do
// cliente — é a única forma de obter um Proprio (design.md §4.4).
func (a Ator) Proprio() Proprio {
	return Proprio{usuarioID: a.usuarioID, valido: true}
}
