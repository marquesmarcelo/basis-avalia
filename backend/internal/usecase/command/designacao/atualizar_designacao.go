package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	designacaodomain "github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarDesignacaoInput struct {
	Ator          autorizacao.Ator
	DesignacaoID  uuid.UUID
	CoordenadorID uuid.UUID
	Portaria      string
	DataInicio    string
	DataFim       *string
	Versao        int
}

// AtualizarDesignacaoUseCase implementa a tabela de design.md §5.3: só
// `futura` aceita trocar coordenador/início; `vigente`/`encerrada` só
// portaria/fim — e o backend RECUSA (409 DESIGNACAO_COM_EFEITO), não só
// desabilita na tela (Reconciliação 2 do dono).
type AtualizarDesignacaoUseCase struct {
	repo  port.DesignacaoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarDesignacaoUseCase(repo port.DesignacaoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarDesignacaoUseCase {
	return &AtualizarDesignacaoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarDesignacaoUseCase) Executar(ctx context.Context, in AtualizarDesignacaoInput) (designacaodomain.Designacao, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoEditar, nil)
	if err != nil {
		return designacaodomain.Designacao{}, err
	}

	inicio, err := valueobject.DataLocalTexto(in.DataInicio)
	if err != nil {
		return designacaodomain.Designacao{}, err
	}
	var fim *valueobject.DataLocal
	if in.DataFim != nil {
		f, err := valueobject.DataLocalTexto(*in.DataFim)
		if err != nil {
			return designacaodomain.Designacao{}, err
		}
		fim = &f
	}

	var resultado designacaodomain.Designacao
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		item, err := uc.repo.BuscarPorID(ctx, esc, in.DesignacaoID)
		if err != nil {
			return err
		}
		atual := item.Designacao
		situacaoAtual := atual.SituacaoEm(esc.DataDeReferencia())

		antes := map[string]any{
			"coordenador_id": atual.CoordenadorID.String(),
			"data_inicio":    atual.Vigencia.Inicio().String(),
		}
		if atual.Vigencia.Fim() != nil {
			antes["data_fim"] = atual.Vigencia.Fim().String()
		}

		situacaoEraFutura := situacaoAtual == valueobject.Futura
		if situacaoEraFutura {
			if err := atual.AtualizarCompleta(in.CoordenadorID, in.Portaria, inicio, fim); err != nil {
				return err
			}
		} else {
			mudouCoordenador := in.CoordenadorID != atual.CoordenadorID
			mudouInicio := inicio.String() != atual.Vigencia.Inicio().String()
			if mudouCoordenador || mudouInicio {
				return designacaodomain.ErrSeNaoForFutura(situacaoAtual)
			}
			if err := atual.AtualizarFimEPortaria(in.Portaria, fim); err != nil {
				return err
			}
		}

		// O repositório espelha a mesma divisão do domínio (design.md
		// §5.3) — nunca AtualizarCompleta fora da situação futura, mesmo
		// que os valores coincidam por acaso.
		if situacaoEraFutura {
			if err := uc.repo.AtualizarCompleta(ctx, esc, &atual, in.Versao); err != nil {
				return err
			}
		} else {
			if err := uc.repo.AtualizarFimEPortaria(ctx, esc, &atual, in.Versao); err != nil {
				return err
			}
		}

		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.AtualizarDesignacao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Designacao"
		evento.RecursoID = &atual.ID
		// C-05: coordenador e vigência registram o par antes/depois
		// (spec.md §10); os demais campos só o nome de quem mudou —
		// mas esta é a única linha que muda algo além disso.
		evento.Detalhes["antes"] = antes
		depois := map[string]any{
			"coordenador_id": atual.CoordenadorID.String(),
			"data_inicio":    atual.Vigencia.Inicio().String(),
		}
		if atual.Vigencia.Fim() != nil {
			depois["data_fim"] = atual.Vigencia.Fim().String()
		}
		evento.Detalhes["depois"] = depois
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = atual
		return nil
	})
	if erro != nil {
		return designacaodomain.Designacao{}, erro
	}
	return resultado, nil
}
