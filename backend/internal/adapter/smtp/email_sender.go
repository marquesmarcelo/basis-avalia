package smtp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

// ErrCircuitoAberto — o breaker está aberto: falha rápido, sem tentar a
// rede (design.md §8.4, CLAUDE.md "Circuit Breaker em chamadas externas").
var ErrCircuitoAberto = errors.New("circuito do servidor de e-mail está aberto")

// ErrCabecalhoComQuebraDeLinha — T-123 (fundacao-metas.md §4.7): \r ou \n
// dentro de um valor de cabeçalho não é dado, é o próprio delimitador de
// cabeçalho do RFC 5322 — um assunto com "\r\nBcc: atacante@..." vira um
// cabeçalho novo. A validação do Value Object (motivo/observação, camada
// de domínio) já impede isso na origem; esta é a segunda camada,
// independente, que recusa no ponto onde o dado vira protocolo — mesmo
// raciocínio do CHECK no banco. Recusar (nunca sanitizar em silêncio):
// mascarar o \r\n mascararia também um ataque bem-sucedido como envio normal.
var ErrCabecalhoComQuebraDeLinha = errors.New("cabeçalho de e-mail com quebra de linha")

const (
	limiteDeFalhasParaAbrir = 5
	janelaDeReabertura      = 30 * time.Second
)

// EmailSender — adapter de port.EmailSender sobre net/smtp (a stdlib
// resolve: nenhuma dependência nova é necessária para falar com um
// servidor SMTP simples como o Mailpit ou um relay corporativo). O
// circuit breaker vive AQUI, nunca no use case, que não deve saber que
// existe (CLAUDE.md).
type EmailSender struct {
	host, porta, usuario, senha, remetente string
	usaTLS                                 bool

	mu             sync.Mutex
	falhasSeguidas int
	abertoAte      time.Time
}

func Novo(host, porta, usuario, senha, remetente string, usaTLS bool) *EmailSender {
	return &EmailSender{host: host, porta: porta, usuario: usuario, senha: senha, remetente: remetente, usaTLS: usaTLS}
}

func (e *EmailSender) circuitoAberto() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.falhasSeguidas >= limiteDeFalhasParaAbrir && time.Now().Before(e.abertoAte)
}

func (e *EmailSender) registrarSucesso() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.falhasSeguidas = 0
}

func (e *EmailSender) registrarFalha() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.falhasSeguidas++
	if e.falhasSeguidas >= limiteDeFalhasParaAbrir {
		e.abertoAte = time.Now().Add(janelaDeReabertura)
	}
}

func (e *EmailSender) Enviar(ctx context.Context, destino valueobject.Email, assunto, corpo string) error {
	// Validar a mensagem não depende de rede nem do estado do breaker —
	// roda antes de qualquer um dos dois, para nunca abrir/fechar o
	// circuito por causa de um cabeçalho recusado localmente.
	mensagem, err := montarMensagem(e.remetente, destino.String(), assunto, corpo)
	if err != nil {
		return err
	}

	if e.circuitoAberto() {
		return ErrCircuitoAberto
	}

	endereco := fmt.Sprintf("%s:%s", e.host, e.porta)
	var auth smtp.Auth
	if e.usuario != "" {
		auth = smtp.PlainAuth("", e.usuario, e.senha, e.host)
	}

	erro := make(chan error, 1)
	go func() {
		erro <- smtp.SendMail(endereco, auth, e.remetente, []string{destino.String()}, mensagem)
	}()

	select {
	case <-ctx.Done():
		e.registrarFalha()
		return ctx.Err()
	case err := <-erro:
		if err != nil {
			e.registrarFalha()
			return err
		}
		e.registrarSucesso()
		return nil
	}
}

// VerificarDisponibilidade sustenta /readyz (fundacao-metas.md §9) — só
// confirma que o servidor SMTP aceita conexão TCP, nunca devolve o erro
// do driver ao cliente HTTP (isso é feito pelo chamador, que só loga).
func (e *EmailSender) VerificarDisponibilidade(ctx context.Context) error {
	endereco := fmt.Sprintf("%s:%s", e.host, e.porta)
	conn, err := (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "tcp", endereco)
	if err != nil {
		return err
	}
	return conn.Close()
}

// codificarCabecalho aplica RFC 2047 (encoded-word) em qualquer valor de
// cabeçalho — sem isso, um assunto com acentuação ou travessão ("—") vira
// bytes UTF-8 crus no cabeçalho, fora da gramática do RFC 5322, e a
// maioria dos clientes de e-mail exibe corrompido (o bug de exibição do
// travessão no assunto era exatamente isto). Recusa \r/\n ANTES de
// codificar: um encoded-word não escapa o texto original, então a checagem
// tem que vir primeiro, não depois.
func codificarCabecalho(valor string) (string, error) {
	if strings.ContainsAny(valor, "\r\n") {
		return "", ErrCabecalhoComQuebraDeLinha
	}
	return mime.QEncoding.Encode("UTF-8", valor), nil
}

func montarMensagem(remetente, destino, assunto, corpo string) ([]byte, error) {
	assuntoCodificado, err := codificarCabecalho(assunto)
	if err != nil {
		return nil, err
	}
	// From/To não passam por RFC 2047 (são endereço, não texto livre), mas
	// recusam \r/\n pela mesma razão — remetente é configuração fixa,
	// destino vem de valueobject.Email (já valida o formato), e ainda
	// assim a checagem roda aqui: segunda camada, independente da origem.
	if strings.ContainsAny(remetente, "\r\n") || strings.ContainsAny(destino, "\r\n") {
		return nil, ErrCabecalhoComQuebraDeLinha
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		remetente, destino, assuntoCodificado, corpo)
	return b.Bytes(), nil
}
