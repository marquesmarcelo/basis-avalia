package entrega

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// DiasDePrazoDeCorrecao — PM-1 da spec: 7 dias corridos.
const DiasDePrazoDeCorrecao = 7

// Entrega — entidade de domínio (specs/metas-coordenacao/design.md §3.2).
// Pende do ITEM do plano, nunca da pessoa (3.2 da spec): é o que faz a
// troca de coordenador não mexer em nada do que já foi comprovado (X11).
type Entrega struct {
	ID, ItemPlanoID, CursoID, InstituicaoID uuid.UUID
	EnviadaPor                              uuid.UUID
	CorrigidaPor                            *uuid.UUID
	Observacao                              string
	Situacao                                valueobject.SituacaoEntrega
	Rodadas                                 valueobject.RodadaDeRecusa
	PrazoCorrecao                           *time.Time
	AvaliadaPor                             *uuid.UUID
	AvaliadaEm                              *time.Time
	Motivo                                  string
	AvaliadorEraCoordenador                 bool
	PendenciaVistaEm                        *time.Time

	NotificacaoEvento      *string
	NotificacaoGeradaEm    *time.Time
	NotificacaoEnviadaEm   *time.Time
	NotificacaoTentativas  int
	NotificacaoUltimoErro  *string

	CriadoEm     time.Time
	AtualizadoEm *time.Time
	ExcluidoEm   *time.Time
	Versao       int
}

// NovaEntrega valida Observacao contra caractere de controle (§4.7):
// opcional, mas se preenchida não pode conter \r, \n ou outro C0/DEL —
// texto livre do usuário que pode circular por saídas futuras (relatório,
// exportação), e a regra vale mesmo sem nenhuma delas existir hoje.
func NovaEntrega(itemPlanoID, cursoID, instituicaoID, enviadaPor uuid.UUID, observacao string) (*Entrega, error) {
	if err := valueobject.ProibirCaractereDeControle(observacao); err != nil {
		return nil, err
	}
	rodadas, _ := valueobject.NovaRodadaDeRecusa(0)
	return &Entrega{
		ID: uuid.Must(uuid.NewV7()), ItemPlanoID: itemPlanoID, CursoID: cursoID, InstituicaoID: instituicaoID,
		EnviadaPor: enviadaPor, Observacao: observacao, Situacao: valueobject.PendenteAvaliacao,
		Rodadas: rodadas, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

// PodeCorrigir — A ÚNICA função que decide se a entrega ainda aceita
// correção (design.md §3.2). Devolve o erro nomeado, não um booleano: o
// chamador nunca precisa redescobrir qual dos três é.
func (e *Entrega) PodeCorrigir(agora time.Time) error {
	if e.Situacao != valueobject.Recusada {
		return domain.ErrEntregaAceitaNaoEditavel
	}
	if e.Rodadas.Esgotada() {
		return domain.ErrLimiteDeRodadasAtingido
	}
	if e.PrazoCorrecao != nil && !valueobject.PrazoDeCorrecaoDeInstante(*e.PrazoCorrecao).EmCurso(agora) {
		return domain.ErrPrazoDeCorrecaoExpirado
	}
	return nil
}

// Corrigir substitui a observação e devolve a entrega a pendente de
// avaliação — chamado só depois de PodeCorrigir(agora) == nil. Anexos são
// tratados à parte (comando adicionar/remover), nunca aqui. Mesma
// validação de NovaEntrega (§4.7): observação sem caractere de controle.
func (e *Entrega) Corrigir(observacao string, corrigidaPor uuid.UUID) error {
	if err := valueobject.ProibirCaractereDeControle(observacao); err != nil {
		return err
	}
	e.Observacao = observacao
	e.CorrigidaPor = &corrigidaPor
	e.Situacao = valueobject.PendenteAvaliacao
	return nil
}

// PodeExcluir — só quem enviou, e só enquanto não está aceita (3.4, EN-10,
// EN-11). A ordem importa: 403 antes de 409, porque "não é seu" é verdade
// mesmo quando a entrega está aceita.
func (e *Entrega) PodeExcluir(autorID uuid.UUID) error {
	if e.EnviadaPor != autorID {
		return domain.ErrExclusaoDeEntregaAlheia
	}
	if e.Situacao == valueobject.Aceita {
		return domain.ErrEntregaAceitaNaoExcluivel
	}
	return nil
}

// Avaliar aplica o resultado (aceitar ou recusar) sobre uma entrega
// pendente — o guarda de concorrência real (UPDATE ... WHERE situacao =
// 'pendente_avaliacao') vive no adapter; este método é a mesma regra em
// forma testável sem banco (design.md §5.2).
func (e *Entrega) Avaliar(resultado valueobject.ResultadoDeAvaliacao, avaliadorID uuid.UUID, avaliadorEraCoordenador bool, agora time.Time, local *time.Location) error {
	if e.Situacao != valueobject.PendenteAvaliacao {
		return domain.ErrEntregaJaAvaliada
	}
	e.AvaliadaPor = &avaliadorID
	e.AvaliadaEm = &agora
	e.AvaliadorEraCoordenador = avaliadorEraCoordenador

	if resultado.Aceita() {
		e.Situacao = valueobject.Aceita
		e.PrazoCorrecao = nil
		return nil
	}
	return e.recusar(resultado.Motivo(), agora, local)
}

// DesfazerAceitacao — qualquer PI da instituição, não só quem aceitou
// (3.4, decisão do dono). Consome uma rodada, exatamente como uma recusa
// direta — combinadas, as duas nunca passam de três (AV-14).
func (e *Entrega) DesfazerAceitacao(motivo string, quemDesfezID uuid.UUID, avaliadorEraCoordenador bool, agora time.Time, local *time.Location) error {
	if e.Situacao != valueobject.Aceita {
		return domain.ErrEntregaNaoEstaAceita
	}
	if trimVazio(motivo) {
		return domain.ErrMotivoObrigatorio
	}
	if err := valueobject.ProibirCaractereDeControle(motivo); err != nil {
		return err
	}
	if e.Rodadas.Esgotada() {
		return domain.ErrLimiteDeRodadasAtingido
	}
	e.AvaliadaPor = &quemDesfezID
	e.AvaliadaEm = &agora
	e.AvaliadorEraCoordenador = avaliadorEraCoordenador
	return e.recusar(motivo, agora, local)
}

// recusar é o núcleo comum de Avaliar(recusada) e DesfazerAceitacao: soma
// uma rodada, grava o motivo, e abre prazo novo — exceto na rodada que
// esgota o limite, que não abre prazo nenhum (PM-2, AV-10).
func (e *Entrega) recusar(motivo string, agora time.Time, local *time.Location) error {
	nova, err := e.Rodadas.MaisUma()
	if err != nil {
		return domain.ErrLimiteDeRodadasAtingido
	}
	e.Rodadas = nova
	e.Situacao = valueobject.Recusada
	e.Motivo = motivo
	if nova.Esgotada() {
		e.PrazoCorrecao = nil
		return nil
	}
	prazo := valueobject.NovoPrazoDeCorrecao(agora, DiasDePrazoDeCorrecao, local).Limite()
	e.PrazoCorrecao = &prazo
	return nil
}

// MarcarPendenciaVista é idempotente: chamar duas vezes não é erro, só a
// segunda não muda nada (NT-04).
func (e *Entrega) MarcarPendenciaVista(agora time.Time) {
	if e.PendenciaVistaEm == nil {
		e.PendenciaVistaEm = &agora
	}
}

// RecusadaDefinitivamente — estado derivado (M-03): NUNCA armazenado,
// sempre calculado a partir de rodadas esgotadas OU prazo vencido.
func (e *Entrega) RecusadaDefinitivamente(agora time.Time) bool {
	if e.Situacao != valueobject.Recusada {
		return false
	}
	if e.Rodadas.Esgotada() {
		return true
	}
	if e.PrazoCorrecao != nil && !valueobject.PrazoDeCorrecaoDeInstante(*e.PrazoCorrecao).EmCurso(agora) {
		return true
	}
	return false
}

func trimVazio(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
