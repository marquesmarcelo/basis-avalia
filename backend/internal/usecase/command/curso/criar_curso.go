package curso

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

type CriarCursoInput struct {
	Ator       autorizacao.Ator
	Nome       string
	CodigoEMec string
	Grau       string
	Modalidade string
	// Situacao — cobre o caso raro de cadastrar já como inativo (ex:
	// migração de curso histórico, ux.md tela "Cursos"). Vazio ou
	// "ativo" mantém o padrão de NovoCurso; só "inativo" desvia.
	Situacao string
}

// CriarCursoUseCase cobre CU-01 a CU-04. Nenhum campo de coordenador
// aqui — a entidade nasce sem ele, por desenho (design.md §3.2).
type CriarCursoUseCase struct {
	repo  port.CursoRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarCursoUseCase(repo port.CursoRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarCursoUseCase {
	return &CriarCursoUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *CriarCursoUseCase) Executar(ctx context.Context, in CriarCursoInput) (curso.Curso, error) {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return curso.Curso{}, err
	}

	novo, err := curso.NovoCurso(*esc.InstituicaoID(), in.Nome, in.CodigoEMec, in.Grau, in.Modalidade)
	if err != nil {
		return curso.Curso{}, err
	}
	if in.Situacao == "inativo" {
		novo.AlterarSituacao(valueobject.CursoInativo)
	}

	var resultado curso.Curso
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarCurso, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = esc.InstituicaoID()
		evento.RecursoTipo = "Curso"
		evento.RecursoID = &novo.ID
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *novo
		return nil
	})
	if erro != nil {
		return curso.Curso{}, erro
	}
	return resultado, nil
}
