package usuario

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type CriarUsuarioInput struct {
	Ator                 autorizacao.Ator
	Alcance              autorizacao.Alcance
	InstituicaoDoCaminho *uuid.UUID
	Nome                 string
	Email                valueobject.Email
	Perfis               []valueobject.Perfil // ignorado quando o alcance fixa o conjunto (design.md §6.2)
	Senha                valueobject.SenhaEmTexto
}

// instituicaoDoAlcance decide a instituição do novo usuário a partir do
// alcance — nunca do payload do cliente (U-11).
func instituicaoDoAlcance(alcance autorizacao.Alcance, ator autorizacao.Ator, instituicaoDoCaminho *uuid.UUID) *uuid.UUID {
	switch alcance {
	case autorizacao.UsuariosDaPropriaInstituicao:
		return ator.InstituicaoID()
	case autorizacao.PesquisadoresDeUmaInstituicao:
		return instituicaoDoCaminho
	default:
		return nil
	}
}

// conjuntoDoAlcance decide o conjunto de perfis do novo/atualizado usuário:
// o alcance impõe um conjunto fixo (pesquisadores, administradores), ou o
// conjunto vem do payload, coagido por ConjuntoInstitucional — nunca
// validado sem coerção, porque quem chama este caminho nunca está criando
// um administrador (design.md §3.1.1, §6.2).
func conjuntoDoAlcance(alcance autorizacao.Alcance, perfisDoPayload []valueobject.Perfil) (valueobject.ConjuntoDePerfis, error) {
	if fixo := alcance.PerfisFixos(); fixo != nil {
		return *fixo, nil
	}
	return valueobject.ConjuntoInstitucional(perfisDoPayload)
}

// CriarUsuarioUseCase atende três alcances (PI, pesquisadores de uma
// instituição, administradores da plataforma) — a diferença está
// inteiramente no Alcance recebido, nunca em `if perfil ==` (design.md §4.1).
// Cobre U-01..U-15, AS-01, AS-07.
type CriarUsuarioUseCase struct {
	repo  port.UsuarioRepository
	hash  port.HashDeSenha
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoCriarUsuarioUseCase(repo port.UsuarioRepository, hash port.HashDeSenha, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *CriarUsuarioUseCase {
	return &CriarUsuarioUseCase{repo: repo, hash: hash, audit: audit, uow: uow}
}

func (uc *CriarUsuarioUseCase) Executar(ctx context.Context, in CriarUsuarioInput) (usuario.Usuario, error) {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoCriar, in.InstituicaoDoCaminho)
	if err != nil {
		return usuario.Usuario{}, err
	}

	// U-10: administrador_sistema no payload do PI é recusado dentro de
	// ConjuntoInstitucional (403 PERFIL_NAO_ATRIBUIVEL) — nenhuma checagem
	// duplicada aqui.
	perfis, err := conjuntoDoAlcance(in.Alcance, in.Perfis)
	if err != nil {
		return usuario.Usuario{}, err
	}

	instituicaoID := instituicaoDoAlcance(in.Alcance, in.Ator, in.InstituicaoDoCaminho)

	hash, err := uc.hash.Gerar(ctx, in.Senha)
	if err != nil {
		return usuario.Usuario{}, err
	}

	novo, err := usuario.NovoUsuario(in.Nome, in.Email, hash, perfis, instituicaoID)
	if err != nil {
		return usuario.Usuario{}, err
	}

	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		if err := uc.repo.Inserir(ctx, esc, novo); err != nil {
			return err
		}
		acao := auditoria.CriarUsuario
		if in.Alcance == autorizacao.AdministradoresDaPlataforma {
			acao = auditoria.CriarAdministrador
		}
		atorID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &atorID
		evento.InstituicaoID = instituicaoID
		evento.RecursoTipo = "Usuario"
		evento.RecursoID = &novo.ID
		return uc.audit.Registrar(ctx, evento)
	})
	if erro != nil {
		return usuario.Usuario{}, erro
	}
	return *novo, nil
}
