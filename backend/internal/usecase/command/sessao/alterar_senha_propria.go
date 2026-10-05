package sessao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

type AlterarSenhaPropriaInput struct {
	Ator       autorizacao.Ator
	SenhaAtual valueobject.SenhaEmTexto
	SenhaNova  valueobject.SenhaEmTexto
}

type AlterarSenhaPropriaOutput struct {
	Instante time.Time
}

// autenticacaoDeSenhaPropria — interface estreita (Interface Segregation):
// este use case só precisa de dois dos cinco métodos de
// port.AutenticacaoRepository, ambos recebendo autorizacao.Proprio, nunca
// Escopo (design.md §4.4). *postgres.AutenticacaoRepository já a satisfaz.
type autenticacaoDeSenhaPropria interface {
	BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*port.CredencialUsuario, error)
	DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error
}

// AlterarSenhaPropriaUseCase cobre S-01..S-12 (design.md §6.2 R5). A ordem
// importa: o cookie novo (emitido pelo handler) precisa nascer com
// emt == o instante gravado aqui, exatamente igual — nunca depois.
type AlterarSenhaPropriaUseCase struct {
	autenticacao autenticacaoDeSenhaPropria
	hash         port.HashDeSenha
	audit        port.AuditLogger
	relogio      port.Relogio
	uow          port.UnidadeDeTrabalho
}

func NovoAlterarSenhaPropriaUseCase(
	autenticacao autenticacaoDeSenhaPropria,
	hash port.HashDeSenha,
	audit port.AuditLogger,
	relogio port.Relogio,
	uow port.UnidadeDeTrabalho,
) *AlterarSenhaPropriaUseCase {
	return &AlterarSenhaPropriaUseCase{autenticacao: autenticacao, hash: hash, audit: audit, relogio: relogio, uow: uow}
}

func (uc *AlterarSenhaPropriaUseCase) Executar(ctx context.Context, in AlterarSenhaPropriaInput) (AlterarSenhaPropriaOutput, error) {
	proprio := in.Ator.Proprio()
	usuarioID := in.Ator.UsuarioID()

	cred, err := uc.autenticacao.BuscarCredencialPropria(ctx, proprio)
	if err != nil {
		return AlterarSenhaPropriaOutput{}, err
	}
	if cred == nil {
		return AlterarSenhaPropriaOutput{}, domain.ErrNaoEncontrado
	}

	confere, err := uc.hash.Conferir(ctx, in.SenhaAtual, cred.SenhaHash)
	if err != nil {
		return AlterarSenhaPropriaOutput{}, err
	}
	if !confere {
		// Fora da transação de mutação: nada foi alterado, então não há o
		// que desfazer — só o registro de que a tentativa falhou (S-03).
		evento := auditoria.NovoEvento(auditoria.AlterarSenhaPropria, auditoria.ResultadoFalha)
		evento.AtorID = &usuarioID
		_ = uc.audit.Registrar(ctx, evento)
		return AlterarSenhaPropriaOutput{}, domain.ErrSenhaAtualIncorreta
	}

	novoHash, err := uc.hash.Gerar(ctx, in.SenhaNova)
	if err != nil {
		return AlterarSenhaPropriaOutput{}, err
	}

	var instante time.Time
	erroExecucao := uc.uow.Executar(ctx, func(ctx context.Context) error {
		instante = uc.relogio.Agora()
		if err := uc.autenticacao.DefinirSenhaPropria(ctx, proprio, novoHash, instante); err != nil {
			return err
		}
		evento := auditoria.NovoEvento(auditoria.AlterarSenhaPropria, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		return uc.audit.Registrar(ctx, evento)
	})
	if erroExecucao != nil {
		return AlterarSenhaPropriaOutput{}, erroExecucao
	}

	return AlterarSenhaPropriaOutput{Instante: instante}, nil
}
