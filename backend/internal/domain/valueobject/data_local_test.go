package valueobject

import (
	"testing"
	"time"
)

// TestDataLocalDe_FronteiraDeMeiaNoiteEmSaoPaulo prova DG-06/PE-04/SI-11/
// AV-03/VG-06 (fundacao-metas.md §5.1): o dia é o do fuso de exibição,
// mesmo com o processo rodando em UTC. 31/07 23h58 e 01/08 00h02, os dois
// em America/Sao_Paulo, são horários bem próximos em UTC (4 minutos de
// diferença) mas têm de cair em dias DIFERENTES.
func TestDataLocalDe_FronteiraDeMeiaNoiteEmSaoPaulo(t *testing.T) {
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	antes := time.Date(2026, 7, 31, 23, 58, 0, 0, saoPaulo)
	depois := time.Date(2026, 8, 1, 0, 2, 0, 0, saoPaulo)

	dAntes := DataLocalDe(antes, saoPaulo)
	dDepois := DataLocalDe(depois, saoPaulo)

	if dAntes.String() != "2026-07-31" {
		t.Fatalf("esperava 2026-07-31, obtido %s", dAntes.String())
	}
	if dDepois.String() != "2026-08-01" {
		t.Fatalf("esperava 2026-08-01, obtido %s", dDepois.String())
	}
	if !dAntes.AntesDe(dDepois) {
		t.Fatal("31/07 deveria ser AntesDe 01/08")
	}
}

// TestDataLocalDe_MesmoInstanteUTCDiasDiferentesConformeOFuso é o teste
// que prova a armadilha em si: o MESMO instante em UTC pode cair em dois
// dias diferentes dependendo do fuso de leitura — exatamente por isso
// DataLocal nunca deriva do fuso do processo.
func TestDataLocalDe_MesmoInstanteUTCDiasDiferentesConformeOFuso(t *testing.T) {
	saoPaulo, _ := time.LoadLocation("America/Sao_Paulo")
	// 2026-08-01T02:30:00Z == 2026-07-31T23:30:00-03:00
	instante := time.Date(2026, 8, 1, 2, 30, 0, 0, time.UTC)

	emUTC := DataLocalDe(instante, time.UTC)
	emSaoPaulo := DataLocalDe(instante, saoPaulo)

	if emUTC.String() != "2026-08-01" {
		t.Fatalf("em UTC esperava 2026-08-01, obtido %s", emUTC.String())
	}
	if emSaoPaulo.String() != "2026-07-31" {
		t.Fatalf("em America/Sao_Paulo esperava 2026-07-31, obtido %s", emSaoPaulo.String())
	}
}

func TestDataLocalTexto_ParseiaFormatoISO(t *testing.T) {
	d, err := DataLocalTexto("2026-03-15")
	if err != nil {
		t.Fatalf("DataLocalTexto: %v", err)
	}
	if d.String() != "2026-03-15" {
		t.Fatalf("esperava 2026-03-15, obtido %s", d.String())
	}
}

func TestDataLocalTexto_FormatoInvalidoDevolveErro(t *testing.T) {
	if _, err := DataLocalTexto("15/03/2026"); err == nil {
		t.Fatal("formato brasileiro não deveria ser aceito")
	}
	if _, err := DataLocalTexto("2026-13-01"); err == nil {
		t.Fatal("mês 13 não deveria ser aceito")
	}
}

func TestDataLocal_MaisDiasAtravessaMesEAno(t *testing.T) {
	d, _ := DataLocalTexto("2026-07-29")
	prazo := d.MaisDias(7)
	if prazo.String() != "2026-08-05" {
		t.Fatalf("esperava 2026-08-05, obtido %s", prazo.String())
	}

	fimDeAno, _ := DataLocalTexto("2026-12-30")
	viradaDeAno := fimDeAno.MaisDias(5)
	if viradaDeAno.String() != "2027-01-04" {
		t.Fatalf("esperava 2027-01-04, obtido %s", viradaDeAno.String())
	}
}

// TestDataLocal_FimDoDiaProduzOPrazoDeCorrecao é o cenário concreto de
// AV-03: recusa em 29/07 às 16h40, prazo de 7 dias corridos, fim do dia
// do sétimo dia — 2026-08-05T23:59:59.999999-03:00.
func TestDataLocal_FimDoDiaProduzOPrazoDeCorrecao(t *testing.T) {
	saoPaulo, _ := time.LoadLocation("America/Sao_Paulo")
	recusa, _ := DataLocalTexto("2026-07-29")
	prazo := recusa.MaisDias(7).FimDoDia(saoPaulo)

	esperado := time.Date(2026, 8, 5, 23, 59, 59, 999999000, saoPaulo)
	if !prazo.Equal(esperado) {
		t.Fatalf("esperava %s, obtido %s", esperado.Format(time.RFC3339Nano), prazo.Format(time.RFC3339Nano))
	}
}

func TestDataLocal_AntesDeEhEstritoNuncaIgual(t *testing.T) {
	d1, _ := DataLocalTexto("2026-03-15")
	d2, _ := DataLocalTexto("2026-03-15")
	if d1.AntesDe(d2) {
		t.Fatal("mesma data não deveria ser AntesDe ela mesma")
	}
}
