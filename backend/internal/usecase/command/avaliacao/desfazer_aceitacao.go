package avaliacao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type DesfazerAceitacaoInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
	Motivo    string
	Versao    int
}

// DesfazerAceitacaoUseCase cobre AV-11 a AV-14: qualquer PI da instituição
// desfaz, não só quem aceitou (decisão do dono, 3.4) — a trilha de quem
// aceitou e quem desfez vive na AUDITORIA (dois eventos distintos), nunca
// numa segunda coluna na entrega.
type DesfazerAceitacaoUseCase struct {
	repo           port.EntregaRepository
	audit          port.AuditLogger
	uow            port.UnidadeDeTrabalho
	relogio        port.Relogio
	fusoDeExibicao *time.Location
}

func NovoDesfazerAceitacaoUseCase(repo port.EntregaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho, relogio port.Relogio, fuso *time.Location) *DesfazerAceitacaoUseCase {
	return &DesfazerAceitacaoUseCase{repo: repo, audit: audit, uow: uow, relogio: relogio, fusoDeExibicao: fuso}
}

func (uc *DesfazerAceitacaoUseCase) Executar(ctx context.Context, in DesfazerAceitacaoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoAvaliar, nil)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
		if err != nil {
			return err
		}
		quemDesfezID := in.Ator.UsuarioID()
		coordena, err := uc.repo.CoordenaCursoHoje(ctx, in.Ator.Proprio(), detalhe.CursoID, esc.DataDeReferencia())
		if err != nil {
			return err
		}

		atual := entregaDeDetalhe(detalhe)
		if err := atual.DesfazerAceitacao(in.Motivo, quemDesfezID, coordena, uc.relogio.Agora(), uc.fusoDeExibicao); err != nil {
			return err
		}

		ok, err := uc.repo.DesfazerAceitacao(ctx, esc, atual, in.Versao)
		if err != nil {
			return err
		}
		if !ok {
			relida, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
			if err != nil {
				return err
			}
			if relida.Versao != in.Versao {
				return domain.ErrConflitoDeVersao
			}
			return domain.ErrEntregaNaoEstaAceita
		}

		evento := auditoria.NovoEvento(auditoria.DesfazerAceitacao, auditoria.ResultadoSucesso)
		evento.AtorID = &quemDesfezID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &in.EntregaID
		evento.Detalhes["motivo"] = in.Motivo
		evento.Detalhes["avaliador_e_coordenador_do_curso"] = coordena
		return uc.audit.Registrar(ctx, evento)
	})
}
