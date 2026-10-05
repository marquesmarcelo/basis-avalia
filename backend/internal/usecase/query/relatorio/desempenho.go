package relatorio

import (
	"context"
	"errors"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type DesempenhoInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroRelatorio
}

// DesempenhoUseCase serve GET /api/v1/relatorios/desempenho. O
// coordenador vê só os cursos dele, sem filtro de responsável (RD-14); o
// PI vê o relatório completo.
type DesempenhoUseCase struct{ repo port.EntregaRepository }

func NovoDesempenhoUseCase(repo port.EntregaRepository) *DesempenhoUseCase {
	return &DesempenhoUseCase{repo: repo}
}

func (uc *DesempenhoUseCase) Executar(ctx context.Context, in DesempenhoInput) (port.ResultadoRelatorio, error) {
	esc, err := AutorizarRelatorio(in.Ator)
	if err != nil {
		return port.ResultadoRelatorio{}, err
	}
	return uc.repo.RelatorioDesempenho(ctx, esc, in.Filtro)
}

// AutorizarRelatorio tenta DesempenhoDaInstituicao (PI) e, se negado,
// DesempenhoDaCarteira (Coordenador) — mesma união de VI-05.
func AutorizarRelatorio(ator autorizacao.Ator) (autorizacao.Escopo, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.DesempenhoDaInstituicao, autorizacao.AcaoListar, nil)
	if err == nil {
		return esc, nil
	}
	if !errors.Is(err, domain.ErrPermissaoNegada) {
		return autorizacao.Escopo{}, err
	}
	return autorizacao.Autorizar(ator, autorizacao.DesempenhoDaCarteira, autorizacao.AcaoListar, nil)
}
