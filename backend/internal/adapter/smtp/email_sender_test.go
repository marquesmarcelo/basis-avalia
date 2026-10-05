package smtp

import (
	"context"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

// NT-03/design.md §8.4: depois de falhas seguidas, o circuito abre e
// falha rápido — sem tentar a rede — até a janela de reabertura passar.
func TestEmailSender_AbreOCircuitoDepoisDeFalhasSeguidas(t *testing.T) {
	// Porta sem listener: a conexão falha rápido (connection refused),
	// sem depender de rede externa nem de timeout longo.
	remetente := Novo("127.0.0.1", "1", "", "", "nao-responda@basis-avalia.example", false)
	destino, err := valueobject.NovoEmail("coordenador@example.com")
	if err != nil {
		t.Fatalf("montando e-mail de teste: %v", err)
	}

	for i := 0; i < limiteDeFalhasParaAbrir; i++ {
		if err := remetente.Enviar(context.Background(), destino, "assunto", "corpo"); err == nil {
			t.Fatalf("tentativa %d: esperava falha de conexão, obteve sucesso", i+1)
		}
	}

	if !remetente.circuitoAberto() {
		t.Fatal("esperava circuito aberto depois do limite de falhas")
	}

	err = remetente.Enviar(context.Background(), destino, "assunto", "corpo")
	if err != ErrCircuitoAberto {
		t.Fatalf("esperava ErrCircuitoAberto (falha rápida, sem tentar a rede), obteve %v", err)
	}
}

// T-123 (fundacao-metas.md §4.7): um assunto com \r\n tentaria injetar um
// cabeçalho novo (ex: "Bcc: atacante@..."). O adapter recusa — nunca
// sanitiza em silêncio — mesmo que o Value Object já devesse ter barrado
// isso na origem: segunda camada, independente da primeira.
func TestMontarMensagem_AssuntoComCRLFERecusado(t *testing.T) {
	_, err := montarMensagem("de@example.com", "para@example.com", "assunto\r\nBcc: atacante@example.com", "corpo")
	if err != ErrCabecalhoComQuebraDeLinha {
		t.Fatalf("esperava ErrCabecalhoComQuebraDeLinha, obteve %v", err)
	}
}

func TestMontarMensagem_AssuntoComApenasLFERecusado(t *testing.T) {
	_, err := montarMensagem("de@example.com", "para@example.com", "assunto\nX-Injetado: 1", "corpo")
	if err != ErrCabecalhoComQuebraDeLinha {
		t.Fatalf("esperava ErrCabecalhoComQuebraDeLinha, obteve %v", err)
	}
}

// O travessão só sobrevive à volta pelo cliente de e-mail se o cabeçalho
// for RFC 2047 — sem encoded-word, o "—" (UTF-8 multibyte) quebra a
// gramática ASCII do cabeçalho e a maioria dos clientes exibe corrompido.
func TestMontarMensagem_AssuntoComTravessaoEhCodificadoRFC2047(t *testing.T) {
	mensagem, err := montarMensagem("de@example.com", "para@example.com", "Entrega recusada — Engenharia", "corpo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	texto := string(mensagem)
	if !strings.Contains(texto, "Subject: =?UTF-8?q?") && !strings.Contains(texto, "Subject: =?UTF-8?b?") {
		t.Fatalf("esperava cabeçalho Subject em encoded-word RFC 2047, obteve: %s", texto)
	}
	if strings.Contains(texto, "Entrega recusada — Engenharia") {
		t.Fatal("o travessão não deveria aparecer cru no cabeçalho")
	}
}

func TestMontarMensagem_AssuntoASCIIPuroNaoEhCodificado(t *testing.T) {
	mensagem, err := montarMensagem("de@example.com", "para@example.com", "assunto simples", "corpo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !strings.Contains(string(mensagem), "Subject: assunto simples\r\n") {
		t.Fatalf("assunto ASCII puro não deveria ser codificado, obteve: %s", string(mensagem))
	}
}
