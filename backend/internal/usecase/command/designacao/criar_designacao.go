package designacao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type CriarDesignacaoInput struct {
	Ator          autorizacao.Ator
	CursoID       uuid.UUID
	CoordenadorID uuid.UUID
	Portaria      string
	DataInicio    string
	DataFim       *string
}

// CriarDesignacaoUseCase cobre CU-*/CP-07/CP-08/CP-09. autodesignacao é
// calculado aqui e nunca recalculado depois (C-06, design.md §5.2).
type CriarDesignacaoUseCase struct {
	repo        port.DesignacaoRepository
	cursoRepo   port.CursoRepository
	usuarioRepo port.UsuarioRepository
	audit       port.AuditLogger
	uow         port.UnidadeDeTrabalho
}

func NovoCriarDesignacaoUseCase(repo port.DesignacaoRepository, cursoRepo port.CursoRepository, usuarioRepo port.UsuarioRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarDesignacaoUseCase {
	return &CriarDesignacaoUseCase{repo: repo, cursoRepo: cursoRepo, usuarioRepo: usuarioRepo, audit: audit, uow: uow}
}

func (uc *CriarDesignacaoUseCase) Executar(ctx context.Context, in CriarDesignacaoInput) (designacao.Designacao, error) {
	escDesignacao, err := autorizacao.Autorizar(in.Ator, autorizacao.DesignacoesDaInstituicao, autorizacao.AcaoCriar, nil)
	if err != nil {
		return designacao.Designacao{}, err
	}

	inicio, err := valueobject.DataLocalTexto(in.DataInicio)
	if err != nil {
		return designacao.Designacao{}, err
	}
	var fim *valueobject.DataLocal
	if in.DataFim != nil {
		f, err := valueobject.DataLocalTexto(*in.DataFim)
		if err != nil {
			return designacao.Designacao{}, err
		}
		fim = &f
	}

	var resultado designacao.Designacao
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		// Curso precisa existir, ativo, na instituição do escopo — 404
		// senão (design.md §5.1, CP-07). FOR SHARE (achado de revisão
		// T-171): o FK de designacao→curso só garante que a LINHA existe,
		// nunca que excluido_em continua nulo — a exclusão lógica não
		// dispara a FK. Sem esta trava, uma exclusão concorrente do curso
		// podia terminar DEPOIS desta leitura mas ANTES do INSERT, e a
		// designação nascia presa a um curso já excluído.
		if err := uc.cursoRepo.TravarSeAtivo(ctx, escDesignacao, in.CursoID); err != nil {
			return err
		}

		// Candidato: possui QUALQUER perfil além de aluno; de outra
		// instituição é 404, não 400 (CP-07).
		escUsuario, err := autorizacao.Autorizar(in.Ator, autorizacao.UsuariosDaPropriaInstituicao, autorizacao.AcaoBuscar, nil)
		if err != nil {
			return err
		}
		candidato, err := uc.usuarioRepo.BuscarPorID(ctx, escUsuario, in.CoordenadorID)
		if err != nil {
			return err
		}
		perfisDoCandido := candidato.Perfis.Ordenado()
		if len(perfisDoCandido) == 1 && candidato.Perfis.Possui(valueobject.Aluno) {
			return domain.ErrCoordenadorInvalido
		}

		autodesignacao := in.Ator.UsuarioID() == in.CoordenadorID
		nova, err := designacao.NovaDesignacao(in.CursoID, *escDesignacao.InstituicaoID(), in.CoordenadorID, in.Portaria, inicio, fim, autodesignacao)
		if err != nil {
			return err
		}

		if err := uc.repo.Inserir(ctx, escDesignacao, nova); err != nil {
			return err
		}
		usuarioID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(auditoria.CriarDesignacao, auditoria.ResultadoSucesso)
		evento.AtorID = &usuarioID
		evento.InstituicaoID = escDesignacao.InstituicaoID()
		evento.RecursoTipo = "Designacao"
		evento.RecursoID = &nova.ID
		evento.Detalhes["coordenador_id"] = nova.CoordenadorID.String()
		evento.Detalhes["portaria"] = nova.Portaria.String()
		evento.Detalhes["data_inicio"] = nova.Vigencia.Inicio().String()
		if nova.Vigencia.Fim() != nil {
			evento.Detalhes["data_fim"] = nova.Vigencia.Fim().String()
		}
		evento.Detalhes["autodesignacao"] = nova.Autodesignacao
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *nova
		return nil
	})
	if erro != nil {
		return designacao.Designacao{}, erro
	}
	return resultado, nil
}
