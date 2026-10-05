package plano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarPlanoInput struct {
	Ator    autorizacao.Ator
	PlanoID uuid.UUID
	Dados   plano.DadosDoPlano
	Versao  int
}

// AtualizarPlanoUseCase edita os textos do plano e registra/atualiza a
// aprovação (SI-06) — uma única seção de formulário, um único "Salvar"
// (ux.md, "Dados do plano" + "Aprovação"), por isso não existe um
// use case separado de "registrar_aprovacao".
type AtualizarPlanoUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarPlanoUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarPlanoUseCase {
	return &AtualizarPlanoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarPlanoUseCase) Executar(ctx context.Context, in AtualizarPlanoInput) (plano.Plano, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return plano.Plano{}, err
	}

	var resultado plano.Plano
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		atual := detalhe.Plano
		tinhaAprovacao := atual.Aprovacao != nil
		if err := atual.AtualizarDados(in.Dados); err != nil {
			return err
		}
		if err := uc.repo.AtualizarDados(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Plano"
		evento.RecursoID = &atual.ID
		evento.Detalhes["tinha_aprovacao_antes"] = tinhaAprovacao
		evento.Detalhes["tem_aprovacao_depois"] = atual.Aprovacao != nil
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return plano.Plano{}, erro
	}
	return resultado, nil
}
