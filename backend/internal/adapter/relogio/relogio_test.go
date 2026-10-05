package relogio

import (
	"testing"
	"time"
)

// TestRelogio_NuncaRetrocedeMesmoComRelogioDoSistemaRetrocedendo prova a
// causa raiz encontrada em produção de teste (R5, systematic-debugging):
// o relógio de parede do container (WSL2/Docker Desktop) pode retroceder
// sob carga de CPU — comprovado empiricamente com retrocesso real de até
// ~550ms numa amostragem de 20s sob contenção. Duas chamadas de Agora()
// em pontos diferentes do fluxo (login e troca de senha) precisam manter
// ordem crescente mesmo se a fonte de tempo subjacente não garantir isso.
func TestRelogio_NuncaRetrocedeMesmoComRelogioDoSistemaRetrocedendo(t *testing.T) {
	base := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	retrocedido := base.Add(-500 * time.Millisecond)

	chamadas := []time.Time{base, retrocedido}
	i := 0
	r := &Relogio{agora: func() time.Time {
		v := chamadas[i]
		if i < len(chamadas)-1 {
			i++
		}
		return v
	}}

	primeiro := r.Agora()
	segundo := r.Agora()

	if !primeiro.Equal(base) {
		t.Fatalf("primeira chamada deveria refletir a fonte real: esperado %s, obtido %s", base, primeiro)
	}
	if !segundo.After(primeiro) {
		t.Fatalf("segunda chamada retrocedeu: primeiro=%s segundo=%s (fonte subjacente retrocedeu %s)", primeiro, segundo, primeiro.Sub(retrocedido))
	}
}

// TestRelogio_TruncaParaMicrossegundo mantém a garantia original: sem
// isso, o valor gravado no banco (TIMESTAMPTZ, precisão de microssegundo,
// que ARREDONDA o resto) diverge do emt do JWT (UnixMicro, que trunca).
func TestRelogio_TruncaParaMicrossegundo(t *testing.T) {
	comNanos := time.Date(2026, 9, 29, 12, 0, 0, 123456789, time.UTC)
	r := &Relogio{agora: func() time.Time { return comNanos }}

	got := r.Agora()

	esperado := comNanos.Truncate(time.Microsecond)
	if !got.Equal(esperado) {
		t.Fatalf("esperava truncagem para microssegundo: esperado %s, obtido %s", esperado, got)
	}
}

// TestRelogio_ChamadasCrescentesRetornamValorRealDaFonte prova que a
// proteção contra retrocesso não distorce o caso normal — quando a fonte
// avança de verdade, o valor devolvido é exatamente o da fonte (truncado),
// nunca artificialmente adiantado.
func TestRelogio_ChamadasCrescentesRetornamValorRealDaFonte(t *testing.T) {
	t1 := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(300 * time.Millisecond)

	chamadas := []time.Time{t1, t2}
	i := 0
	r := &Relogio{agora: func() time.Time {
		v := chamadas[i]
		if i < len(chamadas)-1 {
			i++
		}
		return v
	}}

	primeiro := r.Agora()
	segundo := r.Agora()

	if !primeiro.Equal(t1) {
		t.Fatalf("esperado %s, obtido %s", t1, primeiro)
	}
	if !segundo.Equal(t2) {
		t.Fatalf("fonte avançou de verdade — não deveria ser alterada: esperado %s, obtido %s", t2, segundo)
	}
}

// TestNovo_ImplementaPortRelogio garante que a fábrica pública continua
// satisfazendo port.Relogio após a mudança de Relogio{} (valor) para
// *Relogio (ponteiro, necessário para o mutex de proteção).
func TestNovo_ImplementaPortRelogio(t *testing.T) {
	r := Novo()
	agora := r.Agora()
	if agora.IsZero() {
		t.Fatal("Agora() não deveria devolver o zero value")
	}
}

// TestNovo_DevolveSempreAMesmaInstancia prova a causa raiz encontrada
// depois de corrigir só R5: sem uma instância única por processo, a
// fixture de teste (que chama Novo() para entrar na mesma sequência
// monotônica do login) e o app de cada teste (que também chama Novo())
// acabavam com relógios independentes — cada um "começando do zero",
// sem se proteger um do outro (T02, T05, AS03, AS10, G07 continuavam
// intermitentes mesmo com Agora() já monotônico por instância).
func TestNovo_DevolveSempreAMesmaInstancia(t *testing.T) {
	a := Novo()
	b := Novo()
	if a != b {
		t.Fatal("Novo() deveria devolver sempre o mesmo ponteiro dentro do processo")
	}
}
