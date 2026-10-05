package relatorio

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

type ExportarDesempenhoInput struct {
	Ator   autorizacao.Ator
	Filtro port.FiltroRelatorio
	Linha  func(port.LinhaCSVRelatorio) error
}

// ExportarDesempenhoUseCase serve GET /api/v1/relatorios/desempenho/exportacao
// — SEM fallback para o alcance de carteira (RD-12): a exportação é
// indisponível ao coordenador, 403 direto.
type ExportarDesempenhoUseCase struct {
	repo  port.EntregaRepository
	audit port.AuditLogger
}

func NovoExportarDesempenhoUseCase(repo port.EntregaRepository, audit port.AuditLogger) *ExportarDesempenhoUseCase {
	return &ExportarDesempenhoUseCase{repo: repo, audit: audit}
}

func (uc *ExportarDesempenhoUseCase) Executar(ctx context.Context, in ExportarDesempenhoInput) (int, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.DesempenhoDaInstituicao, autorizacao.AcaoExportar, nil)
	if err != nil {
		return 0, err
	}
	total, err := uc.repo.ExportarDesempenho(ctx, esc, in.Filtro, in.Linha)
	if err != nil {
		return total, err
	}
	usuarioID := in.Ator.UsuarioID()
	evento := auditoria.NovoEvento(auditoria.ExportarRelatorioDesempenho, auditoria.ResultadoSucesso)
	evento.AtorID = &usuarioID
	evento.InstituicaoID = esc.InstituicaoID()
	evento.RecursoTipo = "RelatorioDesempenho"
	evento.Detalhes["linhas"] = total
	evento.Detalhes["periodo_id"] = in.Filtro.PeriodoID.String()
	_ = uc.audit.Registrar(ctx, evento)
	return total, nil
}
