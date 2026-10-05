package avaliacao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AvaliarInput struct {
	Ator      autorizacao.Ator
	EntregaID uuid.UUID
	Resultado string // "aceita" | "recusada"
	Motivo    string
	Versao    int
}

// AvaliarUseCase cobre AV-01, AV-02, AV-06 a AV-10, AV-15 a AV-18. A marca
// avaliador_e_coordenador_do_curso é computada e gravada NO INSTANTE do
// ato, e nunca recalculada depois (M-13, AV-17).
type AvaliarUseCase struct {
	repo           port.EntregaRepository
	audit          port.AuditLogger
	uow            port.UnidadeDeTrabalho
	relogio        port.Relogio
	fusoDeExibicao *time.Location
}

func NovoAvaliarUseCase(repo port.EntregaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho, relogio port.Relogio, fuso *time.Location) *AvaliarUseCase {
	return &AvaliarUseCase{repo: repo, audit: audit, uow: uow, relogio: relogio, fusoDeExibicao: fuso}
}

func (uc *AvaliarUseCase) Executar(ctx context.Context, in AvaliarInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoAvaliar, nil)
	if err != nil {
		return err
	}
	resultado, err := valueobject.NovoResultadoDeAvaliacao(in.Resultado, in.Motivo)
	if err != nil {
		return err
	}

	return uc.uow.Executar(ctx, func(ctx context.Context) error {
		detalhe, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
		if err != nil {
			return err
		}
		avaliadorID := in.Ator.UsuarioID()
		coordena, err := uc.repo.CoordenaCursoHoje(ctx, in.Ator.Proprio(), detalhe.CursoID, esc.DataDeReferencia())
		if err != nil {
			return err
		}

		atual := entregaDeDetalhe(detalhe)
		if err := atual.Avaliar(resultado, avaliadorID, coordena, uc.relogio.Agora(), uc.fusoDeExibicao); err != nil {
			return err
		}

		ok, err := uc.repo.Avaliar(ctx, esc, atual, in.Versao)
		if err != nil {
			return err
		}
		if !ok {
			// Zero linhas afetadas: relê a entrega para escolher entre
			// CONFLITO_DE_VERSAO e ENTREGA_JA_AVALIADA — nunca adivinha
			// (design.md §5.2, AV-07).
			return uc.escolherErroDeConcorrencia(ctx, esc, in)
		}

		evento := auditoria.NovoEvento(auditoria.AvaliarEntrega, auditoria.ResultadoSucesso)
		evento.AtorID = &avaliadorID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Entrega"
		evento.RecursoID = &in.EntregaID
		evento.Detalhes["resultado"] = in.Resultado
		evento.Detalhes["rodada"] = atual.Rodadas.Int()
		evento.Detalhes["avaliador_e_coordenador_do_curso"] = coordena
		return uc.audit.Registrar(ctx, evento)
	})
}

func (uc *AvaliarUseCase) escolherErroDeConcorrencia(ctx context.Context, esc autorizacao.Escopo, in AvaliarInput) error {
	relida, err := uc.repo.BuscarPorID(ctx, esc, in.EntregaID, in.Ator.Proprio())
	if err != nil {
		return err
	}
	if relida.Versao != in.Versao {
		return domain.ErrConflitoDeVersao
	}
	return domain.ErrEntregaJaAvaliada
}
