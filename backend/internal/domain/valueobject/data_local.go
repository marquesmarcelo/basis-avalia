package valueobject

import (
	"fmt"
	"time"
)

// DataLocal é data pura, sem hora e sem fuso embutido — "o dia no fuso de
// exibição, inteiro" (fundacao-metas.md §5.1: DG-06, PE-04, SI-11, AV-03,
// VG-06 são a mesma regra vista de cinco ângulos). Tratar data pura como
// instante é o que faz vencimento aparecer um dia antes.
type DataLocal struct {
	ano int
	mes time.Month
	dia int
}

// DataLocalDe extrai o dia a partir de um instante, no fuso informado —
// nunca no fuso do instante em si. O processo do backend roda em UTC; o
// dia de exibição é America/Sao_Paulo. As duas coisas divergem perto da
// meia-noite, e é exatamente esse caso que este construtor resolve.
func DataLocalDe(instante time.Time, local *time.Location) DataLocal {
	i := instante.In(local)
	ano, mes, dia := i.Date()
	return DataLocal{ano: ano, mes: mes, dia: dia}
}

// DataLocalTexto parseia "2026-03-15" — o formato de data pura da API e
// do banco (coluna DATE).
func DataLocalTexto(iso string) (DataLocal, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return DataLocal{}, fmt.Errorf("data inválida: %w", err)
	}
	ano, mes, dia := t.Date()
	return DataLocal{ano: ano, mes: mes, dia: dia}, nil
}

// AntesDe é estritamente "antes" — datas iguais não são AntesDe uma da
// outra.
func (d DataLocal) AntesDe(o DataLocal) bool {
	return d.chave() < o.chave()
}

func (d DataLocal) chave() int {
	return d.ano*10000 + int(d.mes)*100 + d.dia
}

// MaisDias soma n dias corridos, atravessando mês e ano corretamente. O
// fuso não importa aqui — DataLocal não guarda hora, só a data resultante
// da aritmética de calendário.
func (d DataLocal) MaisDias(n int) DataLocal {
	t := time.Date(d.ano, d.mes, d.dia, 0, 0, 0, 0, time.UTC).AddDate(0, 0, n)
	ano, mes, dia := t.Date()
	return DataLocal{ano: ano, mes: mes, dia: dia}
}

// FimDoDia devolve o último instante deste dia, no fuso informado —
// 23:59:59.999999 (precisão de microssegundo, a do TIMESTAMPTZ), nunca a
// meia-noite do dia seguinte.
func (d DataLocal) FimDoDia(local *time.Location) time.Time {
	return time.Date(d.ano, d.mes, d.dia, 23, 59, 59, 999999000, local)
}

// String devolve o formato ISO de data pura — para SQL (coluna DATE) e
// para a API. Nunca instante, nunca formato brasileiro.
func (d DataLocal) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.ano, int(d.mes), d.dia)
}
