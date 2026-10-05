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

type CriarItemInput struct {
	Ator       autorizacao.Ator
	PlanoID    uuid.UUID
	MetaID     uuid.UUID
	Quantidade int
}

// CriarItemUseCase cobre IT-05, IT-06, IT-07, IT-08, IT-09 (o plano
// encerrado bloqueia item novo pela mesma regra de PLANO_ENCERRADO_PARA_EDICAO).
type CriarItemUseCase struct {
	planoRepo port.PlanoRepository
	itemRepo  port.ItemPlanoRepository
	metaRepo  port.MetaRepository
	audit     port.AuditLogger
	uow       port.UnidadeDeTrabalho
}

func NovoCriarItemUseCase(planoRepo port.PlanoRepository, itemRepo port.ItemPlanoRepository, metaRepo port.MetaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarItemUseCase {
	return &CriarItemUseCase{planoRepo: planoRepo, itemRepo: itemRepo, metaRepo: metaRepo, audit: audit, uow: uow}
}

func (uc *CriarItemUseCase) Executar(ctx context.Context, in CriarItemInput) (itemplano.ItemDoPlano, error) {
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
		metaItem, err := uc.metaRepo.BuscarPorID(ctx, esc, in.MetaID)
		if err != nil {
			return err
		}
		if metaItem.Meta.Situacao != valueobject.CatalogoAtivo {
			return domain.ErrMetaInativa
		}
		jaExiste, err := uc.itemRepo.ExisteMetaNoPlano(ctx, esc, in.PlanoID, in.MetaID)
		if err != nil {
			return err
		}
		if jaExiste {
			return domain.ErrMetaDuplicadaNoPlano
		}
		novo, err := itemplano.NovoItemDoPlano(in.PlanoID, detalhe.Plano.CursoID, detalhe.Plano.InstituicaoID, in.MetaID, in.Quantidade)
		if err != nil {
			return err
		}
		if err := uc.itemRepo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarItemPlano, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "ItemDoPlano"
		evento.RecursoID = &novo.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *novo
		return nil
	})
	if erro != nil {
		return itemplano.ItemDoPlano{}, erro
	}
	return resultado, nil
}
