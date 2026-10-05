package itemplano

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/itemplano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarItemInput struct {
	Ator       autorizacao.Ator
	PlanoID    uuid.UUID
	ItemID     uuid.UUID
	Quantidade int
	Versao     int
}

// AtualizarItemUseCase cobre IT-09: quantidade congela com o plano
// encerrado, por qualquer motivo — período vencido OU encerramento
// antecipado (P-04, reconciliação 2 do dono).
type AtualizarItemUseCase struct {
	planoRepo port.PlanoRepository
	itemRepo  port.ItemPlanoRepository
	audit     port.AuditLogger
	uow       port.UnidadeDeTrabalho
}

func NovoAtualizarItemUseCase(planoRepo port.PlanoRepository, itemRepo port.ItemPlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarItemUseCase {
	return &AtualizarItemUseCase{planoRepo: planoRepo, itemRepo: itemRepo, audit: audit, uow: uow}
}

func (uc *AtualizarItemUseCase) Executar(ctx context.Context, in AtualizarItemInput) (itemplano.ItemDoPlano, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return itemplano.ItemDoPlano{}, err
	}

	var resultado itemplano.ItemDoPlano
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.planoRepo.BuscarPorID(ctx, esc, in.PlanoID)
		if err != nil {
			return err
		}
		if detalhe.Situacao == string(valueobject.PlanoEncerrado) {
			return domain.ErrPlanoEncerradoParaEdicao
		}
		atual, err := uc.itemRepo.BuscarPorID(ctx, esc, in.ItemID)
		if err != nil {
			return err
		}
		if err := atual.AlterarQuantidade(in.Quantidade); err != nil {
			return err
		}
		if err := uc.itemRepo.AtualizarQuantidade(ctx, esc, &atual, in.Versao); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarItemPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "ItemDoPlano"
		evento.RecursoID = &atual.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return itemplano.ItemDoPlano{}, erro
	}
	return resultado, nil
}
