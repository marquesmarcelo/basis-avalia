package entrega

import (
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func saoPaulo(t *testing.T) *time.Location {
	t.Helper()
	local, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("carregando fuso: %v", err)
	}
	return local
}

func novaEntregaDeTeste() *Entrega {
	e, _ := NovaEntrega(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), "")
	return e
}

// AV-03: recusa em 29/07/2026 às 16h40 (horário de Brasília) abre prazo
// até 05/08/2026 às 23h59min59s (-03:00).
func TestEntrega_Avaliar_Recusar_AbrePrazoDeSeteDias(t *testing.T) {
	local := saoPaulo(t)
	e := novaEntregaDeTeste()
	recusaEm := time.Date(2026, 7, 29, 16, 40, 0, 0, local)

	resultado, err := valueobject.NovoResultadoDeAvaliacao("recusada", "A lista de presença não corresponde à data da ata.")
	if err != nil {
		t.Fatalf("erro montando resultado: %v", err)
	}
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, recusaEm, local); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if e.Situacao != valueobject.Recusada {
		t.Fatalf("esperava recusada, obteve %s", e.Situacao)
	}
	if e.Rodadas.Int() != 1 {
		t.Fatalf("esperava rodada 1, obteve %d", e.Rodadas.Int())
	}
	if e.PrazoCorrecao == nil {
		t.Fatal("esperava prazo de correção aberto")
	}
	esperado := time.Date(2026, 8, 5, 23, 59, 59, 999999000, local)
	if !e.PrazoCorrecao.Equal(esperado) {
		t.Fatalf("esperava prazo %v, obteve %v", esperado, *e.PrazoCorrecao)
	}
}

// AV-10: a terceira recusa encerra a entrega em definitivo, sem novo
// prazo — reenviar depois disso responde LIMITE_DE_RODADAS_ATINGIDO.
func TestEntrega_Avaliar_TerceiraRecusa_SemNovoPrazo(t *testing.T) {
	local := saoPaulo(t)
	e := novaEntregaDeTeste()
	agora := time.Date(2026, 1, 1, 12, 0, 0, 0, local)
	avaliador := uuid.Must(uuid.NewV7())

	for i := 0; i < 2; i++ {
		resultado, _ := valueobject.NovoResultadoDeAvaliacao("recusada", "motivo")
		if err := e.Avaliar(resultado, avaliador, false, agora, local); err != nil {
			t.Fatalf("recusa %d: erro inesperado: %v", i+1, err)
		}
		e.Corrigir("corrigido", uuid.Must(uuid.NewV7()))
	}

	resultado, _ := valueobject.NovoResultadoDeAvaliacao("recusada", "terceira recusa")
	if err := e.Avaliar(resultado, avaliador, false, agora, local); err != nil {
		t.Fatalf("erro inesperado na terceira recusa: %v", err)
	}
	if e.Rodadas.Int() != 3 {
		t.Fatalf("esperava rodada 3, obteve %d", e.Rodadas.Int())
	}
	if e.PrazoCorrecao != nil {
		t.Fatal("esperava NENHUM prazo de correção na terceira recusa")
	}

	err := e.PodeCorrigir(agora)
	if err != domain.ErrLimiteDeRodadasAtingido {
		t.Fatalf("esperava ErrLimiteDeRodadasAtingido, obteve %v", err)
	}
}

// AV-05: correção depois do prazo responde PRAZO_DE_CORRECAO_EXPIRADO.
func TestEntrega_PodeCorrigir_ForaDoPrazo(t *testing.T) {
	local := saoPaulo(t)
	e := novaEntregaDeTeste()
	recusaEm := time.Date(2026, 7, 29, 16, 40, 0, 0, local)
	resultado, _ := valueobject.NovoResultadoDeAvaliacao("recusada", "motivo")
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, recusaEm, local); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	depoisDoPrazo := time.Date(2026, 8, 6, 0, 0, 0, 0, local)
	if err := e.PodeCorrigir(depoisDoPrazo); err != domain.ErrPrazoDeCorrecaoExpirado {
		t.Fatalf("esperava ErrPrazoDeCorrecaoExpirado, obteve %v", err)
	}

	// AV-04: dentro do prazo, mesmo com o período encerrado (o período não
	// entra em PodeCorrigir — é responsabilidade do use case ignorá-lo na
	// correção).
	dentroDoPrazo := time.Date(2026, 8, 3, 0, 0, 0, 0, local)
	if err := e.PodeCorrigir(dentroDoPrazo); err != nil {
		t.Fatalf("esperava nil dentro do prazo, obteve %v", err)
	}
}

// EN-08: entrega aceita não é editada nem excluída.
func TestEntrega_PodeCorrigir_Aceita(t *testing.T) {
	e := novaEntregaDeTeste()
	resultado, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := e.PodeCorrigir(time.Now()); err != domain.ErrEntregaAceitaNaoEditavel {
		t.Fatalf("esperava ErrEntregaAceitaNaoEditavel, obteve %v", err)
	}
}

func TestEntrega_PodeExcluir(t *testing.T) {
	autor := uuid.Must(uuid.NewV7())
	e, _ := NovaEntrega(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), autor, "")

	if err := e.PodeExcluir(uuid.Must(uuid.NewV7())); err != domain.ErrExclusaoDeEntregaAlheia {
		t.Fatalf("esperava ErrExclusaoDeEntregaAlheia, obteve %v", err)
	}
	if err := e.PodeExcluir(autor); err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}

	resultado, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	_ = e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC)
	if err := e.PodeExcluir(autor); err != domain.ErrEntregaAceitaNaoExcluivel {
		t.Fatalf("esperava ErrEntregaAceitaNaoExcluivel, obteve %v", err)
	}
}

// AV-06: entrega já avaliada não é avaliada de novo.
func TestEntrega_Avaliar_JaAvaliada(t *testing.T) {
	e := novaEntregaDeTeste()
	resultado, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != domain.ErrEntregaJaAvaliada {
		t.Fatalf("esperava ErrEntregaJaAvaliada, obteve %v", err)
	}
}

// AV-11: qualquer PI desfaz a aceitação — consome rodada e abre prazo.
func TestEntrega_DesfazerAceitacao(t *testing.T) {
	local := saoPaulo(t)
	e := novaEntregaDeTeste()
	resultadoAceita, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	quemAceitou := uuid.Must(uuid.NewV7())
	if err := e.Avaliar(resultadoAceita, quemAceitou, false, time.Now(), local); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	quemDesfez := uuid.Must(uuid.NewV7())
	desfazerEm := time.Date(2026, 8, 12, 10, 0, 0, 0, local)
	if err := e.DesfazerAceitacao("a ata anexada é de outra reunião", quemDesfez, false, desfazerEm, local); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if e.Situacao != valueobject.Recusada {
		t.Fatalf("esperava recusada, obteve %s", e.Situacao)
	}
	if e.Rodadas.Int() != 1 {
		t.Fatalf("esperava rodada 1, obteve %d", e.Rodadas.Int())
	}
	esperado := time.Date(2026, 8, 19, 23, 59, 59, 999999000, local)
	if e.PrazoCorrecao == nil || !e.PrazoCorrecao.Equal(esperado) {
		t.Fatalf("esperava prazo %v, obteve %v", esperado, e.PrazoCorrecao)
	}
	if *e.AvaliadaPor != quemDesfez {
		t.Fatal("esperava avaliada_por apontando para quem desfez")
	}
}

// AV-13: desfazer exige motivo e só vale sobre entrega aceita.
func TestEntrega_DesfazerAceitacao_Guardas(t *testing.T) {
	e := novaEntregaDeTeste()
	if err := e.DesfazerAceitacao("motivo", uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != domain.ErrEntregaNaoEstaAceita {
		t.Fatalf("esperava ErrEntregaNaoEstaAceita, obteve %v", err)
	}

	resultado, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	_ = e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC)
	if err := e.DesfazerAceitacao("", uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != domain.ErrMotivoObrigatorio {
		t.Fatalf("esperava ErrMotivoObrigatorio, obteve %v", err)
	}
}

// AV-14: não se desfaz aceitação com as rodadas esgotadas — o caso em que
// a entrega acumulou 3 rodadas (recusa + desfazimentos anteriores) e está
// aceita de novo.
func TestEntrega_DesfazerAceitacao_RodadasEsgotadas(t *testing.T) {
	local := saoPaulo(t)
	e := novaEntregaDeTeste()
	avaliador := uuid.Must(uuid.NewV7())
	agora := time.Now()

	// Consome as três rodadas por ciclos de aceitar -> desfazer.
	for i := 0; i < 3; i++ {
		resultadoAceita, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
		if err := e.Avaliar(resultadoAceita, avaliador, false, agora, local); err != nil {
			t.Fatalf("ciclo %d, aceitar: erro inesperado: %v", i+1, err)
		}
		if err := e.DesfazerAceitacao("motivo", avaliador, false, agora, local); err != nil {
			t.Fatalf("ciclo %d, desfazer: erro inesperado: %v", i+1, err)
		}
		e.Corrigir("corrigido", uuid.Must(uuid.NewV7()))
	}
	if e.Rodadas.Int() != 3 {
		t.Fatalf("esperava rodada 3 antes da última aceitação, obteve %d", e.Rodadas.Int())
	}

	resultadoAceita, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	if err := e.Avaliar(resultadoAceita, avaliador, false, agora, local); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if err := e.DesfazerAceitacao("motivo", avaliador, false, agora, local); err != domain.ErrLimiteDeRodadasAtingido {
		t.Fatalf("esperava ErrLimiteDeRodadasAtingido, obteve %v", err)
	}
}

func TestEntrega_MarcarPendenciaVista_Idempotente(t *testing.T) {
	e := novaEntregaDeTeste()
	primeira := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	segunda := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	e.MarcarPendenciaVista(primeira)
	e.MarcarPendenciaVista(segunda)
	if !e.PendenciaVistaEm.Equal(primeira) {
		t.Fatalf("esperava manter o primeiro instante, obteve %v", *e.PendenciaVistaEm)
	}
}

// fundacao-metas.md §4.7 (T-123/T-125): um nome de curso com \r\n é dado
// válido para o banco e para a tela, e viraria instrução ao atravessar a
// fronteira do SMTP — a mesma classe vale para observação e motivo,
// texto livre digitado pelo usuário. O Value Object recusa ANTES de
// qualquer escrita, independente de qualquer adapter de saída existir.
func TestNovaEntrega_ObservacaoComCaractereDeControleERecusada(t *testing.T) {
	_, err := NovaEntrega(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), "observação\r\nX-Injected: header")
	if err != domain.ErrCaractereDeControleNaoPermitido {
		t.Fatalf("esperava ErrCaractereDeControleNaoPermitido, obteve %v", err)
	}
}

func TestEntrega_Corrigir_ObservacaoComCaractereDeControleERecusada(t *testing.T) {
	e := novaEntregaDeTeste()
	if err := e.Corrigir("nova observação\ncom quebra de linha", uuid.Must(uuid.NewV7())); err != domain.ErrCaractereDeControleNaoPermitido {
		t.Fatalf("esperava ErrCaractereDeControleNaoPermitido, obteve %v", err)
	}
}

func TestEntrega_DesfazerAceitacao_MotivoComCaractereDeControleERecusado(t *testing.T) {
	e := novaEntregaDeTeste()
	resultado, _ := valueobject.NovoResultadoDeAvaliacao("aceita", "")
	if err := e.Avaliar(resultado, uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	err := e.DesfazerAceitacao("motivo\r\nSubject: outro assunto", uuid.Must(uuid.NewV7()), false, time.Now(), time.UTC)
	if err != domain.ErrCaractereDeControleNaoPermitido {
		t.Fatalf("esperava ErrCaractereDeControleNaoPermitido, obteve %v", err)
	}
}
