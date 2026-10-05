package auditoria

import (
	"context"
	"errors"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
)

type localFake struct {
	eventos []auditoria.Evento
	erro    error
}

func (l *localFake) Registrar(ctx context.Context, e auditoria.Evento) error {
	if l.erro != nil {
		return l.erro
	}
	l.eventos = append(l.eventos, e)
	return nil
}

func TestComposto_FalhaNoLocalNaoRegistraEDevolveErro(t *testing.T) {
	local := &localFake{erro: errors.New("falha simulada de banco")}
	sys, _ := NovoSyslog("", "tcp", "basis-avalia")
	composto := NovoComposto(local, sys)

	err := composto.Registrar(context.Background(), auditoria.NovoEvento(auditoria.CriarUsuario, auditoria.ResultadoSucesso))
	if err == nil {
		t.Fatal("esperava erro propagado do canal local")
	}
}

func TestComposto_SucessoNoLocalRegistraMesmoComSyslogNoOp(t *testing.T) {
	local := &localFake{}
	sys, _ := NovoSyslog("", "tcp", "basis-avalia")
	composto := NovoComposto(local, sys)

	err := composto.Registrar(context.Background(), auditoria.NovoEvento(auditoria.AlterarSenhaPropria, auditoria.ResultadoSucesso))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(local.eventos) != 1 {
		t.Fatalf("esperava 1 evento local, obtido %d", len(local.eventos))
	}
}

func TestComposto_DeveIrParaSyslog_AtualizarUsuarioSoQuandoPerfilMuda(t *testing.T) {
	composto := &Composto{}

	semMudancaDePerfil := auditoria.NovoEvento(auditoria.AtualizarUsuario, auditoria.ResultadoSucesso)
	semMudancaDePerfil.Detalhes["campos_alterados"] = []string{"nome"}
	if composto.deveIrParaSyslog(semMudancaDePerfil) {
		t.Fatal("atualizar_usuario sem mudança de perfil não deveria ir ao syslog")
	}

	comMudancaDePerfil := auditoria.NovoEvento(auditoria.AtualizarUsuario, auditoria.ResultadoSucesso)
	comMudancaDePerfil.Detalhes["perfis_anterior"] = []string{"professor"}
	comMudancaDePerfil.Detalhes["perfis_novo"] = []string{"coordenador_curso"}
	if !composto.deveIrParaSyslog(comMudancaDePerfil) {
		t.Fatal("atualizar_usuario com mudança de perfis deveria ir ao syslog (E-02)")
	}

}
