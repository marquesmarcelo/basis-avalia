package indicador

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/port"
)

type CriarIndicadorInput struct {
	Ator      autorizacao.Ator
	Codigo    string
	Nome      string
	Descricao string
}

// CriarIndicadorUseCase cobre IN-01, IN-02, IN-03 — escopo institucional
// imposto pela rota, nunca lido do corpo da requisição.
type CriarIndicadorUseCase struct {
	repo  port.IndicadorRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarIndicadorUseCase(repo port.IndicadorRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarIndicadorUseCase {
	return &CriarIndicadorUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarIndicadorUseCase) Executar(ctx context.Context, in CriarIndicadorInput) (indicador.Indicador, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CatalogoDeIndicadores, autorizacao.AcaoCriar, nil)
	if err != nil {
		return indicador.Indicador{}, err
	}

	novo, err := indicador.NovoDaInstituicao(*esc.InstituicaoID(), in.Codigo, in.Nome, in.Descricao)
	if err != nil {
		return indicador.Indicador{}, err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarIndicador, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Indicador"
		evento.RecursoID = &novo.ID
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return indicador.Indicador{}, erro
	}
	return *novo, nil
}
