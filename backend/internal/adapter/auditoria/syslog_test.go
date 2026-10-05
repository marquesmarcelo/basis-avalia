package auditoria

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
)

func TestSyslog_ServidorVazioEhNoOp(t *testing.T) {
	s, err := NovoSyslog("", "tcp", "basis-avalia")
	if err != nil {
		t.Fatalf("servidor vazio não deveria falhar: %v", err)
	}
	if err := s.Enviar(auditoria.NovoEvento(auditoria.CriarInstituicao, auditoria.ResultadoSucesso)); err != nil {
		t.Fatalf("Enviar em modo no-op não deveria falhar: %v", err)
	}
}

func TestSyslog_EnderecoInalcancavelNaoQuebraAInicializacao(t *testing.T) {
	s, err := NovoSyslog("127.0.0.1:1", "tcp", "basis-avalia")
	if err == nil {
		t.Log("aviso: dial não falhou neste ambiente — ainda assim o adapter deve ser utilizável")
	}
	if s == nil {
		t.Fatal("NovoSyslog nunca deveria devolver nil, mesmo com endereço inalcançável")
	}
	if err := s.Enviar(auditoria.NovoEvento(auditoria.CriarInstituicao, auditoria.ResultadoSucesso)); err != nil && s.writer == nil {
		t.Fatalf("Enviar sem writer conectado deveria ser no-op, obtido erro: %v", err)
	}
}
