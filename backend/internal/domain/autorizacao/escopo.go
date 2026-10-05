package autorizacao

import (
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Escopo é o recorte de dados que o ator está autorizado a tocar nesta
// operação. Campos não exportados, nenhum construtor exportado além do
// devolvido por Autorizar: é impossível obter um Escopo sem passar pela
// autorização (design.md §3.2).
//
// exigePerfil substitui a antiga lista de perfis alcançáveis: com o
// modelo de conjunto, o alvo não TEM um perfil — ele POSSUI vários — e o
// que os alcances precisam é "o alvo possui este perfil" (um EXISTS),
// nunca "o perfil do alvo está nesta lista". nil significa que nenhum
// perfil específico é exigido do alvo.
type Escopo struct {
	valido              bool
	plataforma          bool
	instituicaoID       *uuid.UUID
	exigePerfil         *valueobject.Perfil
	dataDeReferencia    valueobject.DataLocal
	restritoACarteiraDe *uuid.UUID
}

func (e Escopo) Valido() bool                            { return e.valido }
func (e Escopo) Plataforma() bool                        { return e.plataforma }
func (e Escopo) InstituicaoID() *uuid.UUID               { return e.instituicaoID }
func (e Escopo) ExigePerfil() *valueobject.Perfil        { return e.exigePerfil }
func (e Escopo) DataDeReferencia() valueobject.DataLocal { return e.dataDeReferencia }

// RestritoACarteiraDe é o id do coordenador quando o alcance restringe o
// recorte aos cursos que ele coordena hoje (fundacao-metas.md §3.4) — nil
// para qualquer alcance que não seja de carteira. É regra de autorização,
// não filtro de tela: é o que produz 404 para recurso de curso alheio
// dentro da própria instituição.
func (e Escopo) RestritoACarteiraDe() *uuid.UUID { return e.restritoACarteiraDe }

// SemCarteira devolve uma cópia sem a restrição de carteira, mantendo o
// isolamento por instituição — nunca usada para o recurso de topo que o
// alcance autorizou (plano, curso, designação), só para uma junção
// SEGUINTE a partir dele com um alvo que legitimamente não tem coluna de
// curso (catálogo de indicador, meta, período): esses são recursos da
// instituição, não do curso, e o recurso de topo já aplicou a carteira
// (specs/plano-acao/design.md §5.4). Sem isto, AplicarEscopo recusaria a
// junção com ErrEscopoInvalido (§3.6) mesmo quando a carteira já foi
// satisfeita um nível acima — nunca afrouxa o isolamento por instituição,
// só remove a dimensão que o alvo não representa.
func (e Escopo) SemCarteira() Escopo {
	e.restritoACarteiraDe = nil
	return e
}
