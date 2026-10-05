package indicadorplataforma

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/port"
)

type CriarIndicadorPlataformaInput struct {
	Ator                  autorizacao.Ator
	Codigo                string
	Nome                  string
	Descricao             string
	ReferenciaInstrumento string
}

// CriarIndicadorPlataformaUseCase cobre IE-01, IE-02, IE-03 (design.md §5).
type CriarIndicadorPlataformaUseCase struct {
	repo  port.IndicadorPlataformaRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarIndicadorPlataformaUseCase(repo port.IndicadorPlataformaRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarIndicadorPlataformaUseCase {
	return &CriarIndicadorPlataformaUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarIndicadorPlataformaUseCase) Executar(ctx context.Context, in CriarIndicadorPlataformaInput) (indicador.Indicador, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.IndicadoresDaPlataforma, autorizacao.AcaoCriar, nil)
	if err != nil {
		return indicador.Indicador{}, err
	}

	novo, err := indicador.NovoDaPlataforma(in.Codigo, in.Nome, in.Descricao, in.ReferenciaInstrumento)
	if err != nil {
		return indicador.Indicador{}, err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarIndicadorInep, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &novo.ID
		evento.Detalhes["codigo"] = novo.Codigo.String()
		evento.Detalhes["referencia_instrumento"] = novo.ReferenciaInstrumento.String()
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return indicador.Indicador{}, erro
	}
	return *novo, nil
}
