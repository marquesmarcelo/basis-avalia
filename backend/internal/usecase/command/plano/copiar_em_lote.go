package plano

import (
	"context"
	"errors"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

const limiteCursosPorLote = 100

type CopiarEmLoteInput struct {
	Ator             autorizacao.Ator
	PlanoOrigemID    uuid.UUID
	PeriodoDestinoID uuid.UUID
	Cursos           []uuid.UUID
}

type CursoPulado struct {
	CursoID uuid.UUID
	Motivo  string
}

type CopiarEmLoteResultado struct {
	Criados int
	Pulados []CursoPulado
}

// CopiarEmLoteUseCase cobre 3.6 da spec e design.md §5.3 — a operação de
// maior potencial de estrago da feature. A garantia contra o curso que já
// tem plano NÃO é a pré-verificação: é a violação do índice único
// uq_plano_curso_periodo, tratada abaixo como resultado esperado
// (CP-05). Uma transação por curso — falha em um não desfaz os demais
// (CP-07), e cada plano copiado nasce em rascunho, sem aprovação, sem
// entrega, sem anexo (CP-02).
type CopiarEmLoteUseCase struct {
	repo  port.PlanoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCopiarEmLoteUseCase(repo port.PlanoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CopiarEmLoteUseCase {
	return &CopiarEmLoteUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CopiarEmLoteUseCase) Executar(ctx context.Context, in CopiarEmLoteInput) (CopiarEmLoteResultado, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.PlanosDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return CopiarEmLoteResultado{}, err
	}
	if len(in.Cursos) == 0 {
		return CopiarEmLoteResultado{}, domain.ErrCursosObrigatorios
	}
	if len(in.Cursos) > limiteCursosPorLote {
		return CopiarEmLoteResultado{}, domain.ErrLoteAcimaDoLimite
	}

	origem, err := uc.repo.BuscarPorID(ctx, esc, in.PlanoOrigemID)
	if err != nil {
		return CopiarEmLoteResultado{}, err
	}
	itensOrigem := origem.Itens

	resultado := CopiarEmLoteResultado{}
	for _, cursoID := range in.Cursos {
		erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
			ativo, err := uc.repo.CursoValidoParaPlano(ctx, esc, cursoID)
			if err != nil {
				return err
			}
			if !ativo {
				return domain.ErrCursoInativo
			}
			// CopiarPara: leva descrição, objetivo, resultados e
			// alinhamentos da origem; NUNCA aprovação nem situação — todo
			// plano copiado nasce em rascunho (CP-02). Copiar a aprovação
			// seria o pior defeito possível desta feature.
			novo, err := plano.NovoPlano(*esc.InstituicaoID(), cursoID, in.PeriodoDestinoID, plano.DadosDoPlano{
				Descricao: origem.Plano.Descricao, ObjetivoGeral: origem.Plano.ObjetivoGeral,
				ResultadosEsperados: origem.Plano.ResultadosEsperados,
				AlinhamentoPDI:      origem.Plano.AlinhamentoPDI, AlinhamentoPPC: origem.Plano.AlinhamentoPPC,
			})
			if err != nil {
				return err
			}
			itensParaCopia := make([]port.ItemDoPlanoResponse, len(itensOrigem))
			for i, item := range itensOrigem {
				itensParaCopia[i] = port.ItemDoPlanoResponse{MetaID: item.MetaID, Quantidade: item.Quantidade}
			}
			if err := uc.repo.InserirComItens(ctx, esc, novo, itensParaCopia); err != nil {
				return err
			}
			usuarioID := in.Ator.UsuarioID()
			evento := auditoria.NovoEvento(auditoria.CopiarPlano, auditoria.ResultadoSucesso)
			evento.AtorID = &usuarioID
			evento.InstituicaoID = esc.InstituicaoID()
			evento.RecursoTipo = "Plano"
			evento.RecursoID = &novo.ID
			evento.Detalhes["plano_origem_id"] = origem.Plano.ID.String()
			evento.Detalhes["curso_id"] = cursoID.String()
			if err := uc.audit.Registrar(ctx, evento); err != nil {
				return err
			}
			resultado.Criados++
			return nil
		})
		if erro != nil {
			motivo := "falha ao criar"
			if errors.Is(erro, domain.ErrPlanoDuplicado) {
				motivo = "já tem plano neste período"
			}
			resultado.Pulados = append(resultado.Pulados, CursoPulado{CursoID: cursoID, Motivo: motivo})
			continue
		}
	}
	return resultado, nil
}
