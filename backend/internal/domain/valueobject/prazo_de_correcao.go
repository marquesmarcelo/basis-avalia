package valueobject

import "time"

// PrazoDeCorrecao — instante final do prazo de correção (specs/
// metas-coordenacao/design.md §3.1, PM-1): construído por
// DataLocal(recusa).MaisDias(7).FimDoDia(local), nunca calculado inline
// no use case — a fronteira do fim do dia é o tipo de detalhe que uma
// segunda escrita à mão erra silenciosamente.
type PrazoDeCorrecao struct {
	limite time.Time
}

// NovoPrazoDeCorrecao abre um prazo de N dias corridos a partir do
// instante da recusa, contado no fuso de exibição — nunca UTC (X7 da
// spec: "uma entrega do último dia não pode ser recusada por causa do
// fuso horário").
func NovoPrazoDeCorrecao(recusaEm time.Time, dias int, local *time.Location) PrazoDeCorrecao {
	dia := DataLocalDe(recusaEm, local)
	limite := dia.MaisDias(dias).FimDoDia(local)
	return PrazoDeCorrecao{limite: limite}
}

func PrazoDeCorrecaoDeInstante(limite time.Time) PrazoDeCorrecao {
	return PrazoDeCorrecao{limite: limite}
}

func (p PrazoDeCorrecao) Limite() time.Time { return p.limite }

// EmCurso é a ÚNICA comparação que decide se o prazo ainda vale — nunca
// reescrita inline com `.Before`/`.After` trocados por engano em outro
// arquivo.
func (p PrazoDeCorrecao) EmCurso(agora time.Time) bool {
	return agora.Before(p.limite) || agora.Equal(p.limite)
}
