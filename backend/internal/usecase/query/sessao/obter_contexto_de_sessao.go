package sessao

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type InstituicaoResumo struct {
	ID    uuid.UUID
	Nome  string
	Sigla string
}

// ContextoDeSessaoOutput é o formato exato de GET /auth/eu (R4) — mesmo
// corpo devolvido por POST /auth/login (design.md §6.2, §6.4).
//
// Perfis é o conjunto EFETIVO (fundacao-metas.md §4.2) — atribuído mais
// Coordenador de Curso quando a pessoa coordena hoje. PerfisDerivados é o
// subconjunto de Perfis que não está gravado em usuario_perfil — hoje,
// só pode conter "coordenador_curso". CursosCoordenados é o único fato
// exposto sobre a designação em si (specs/cursos/design.md §5.5,
// fundacao-metas.md §4.1) — coordena_hoje nunca vai para a API.
type ContextoDeSessaoOutput struct {
	UsuarioID         uuid.UUID
	Nome              string
	Email             string
	Perfis            []valueobject.Perfil
	PerfisRotulos     []string
	PerfisDerivados   []valueobject.Perfil
	CursosCoordenados int
	SenhaProvisoria   bool
	Instituicao       *InstituicaoResumo // nil para o Administrador do Sistema
	Permissoes        []autorizacao.Permissao
}

// ObterContextoDeSessaoUseCase cobre SH-02, SH-12, SH-13, AS-05, CP-01,
// CP-13.
type ObterContextoDeSessaoUseCase struct {
	repo port.AutenticacaoRepository
}

func NovoObterContextoDeSessaoUseCase(repo port.AutenticacaoRepository) *ObterContextoDeSessaoUseCase {
	return &ObterContextoDeSessaoUseCase{repo: repo}
}

func (uc *ObterContextoDeSessaoUseCase) Executar(ctx context.Context, usuarioID uuid.UUID, hoje valueobject.DataLocal) (ContextoDeSessaoOutput, error) {
	ctxSessao, err := uc.repo.CarregarContextoDeSessao(ctx, usuarioID, hoje)
	if err != nil {
		return ContextoDeSessaoOutput{}, err
	}
	if ctxSessao == nil || ctxSessao.ExcluidoEm != nil {
		return ContextoDeSessaoOutput{}, domain.ErrNaoEncontrado
	}

	var instituicao *InstituicaoResumo
	if ctxSessao.InstituicaoID != nil {
		instituicao = &InstituicaoResumo{
			ID:    *ctxSessao.InstituicaoID,
			Nome:  ctxSessao.InstituicaoNome,
			Sigla: ctxSessao.InstituicaoSigla,
		}
	}

	// CP-01: a resposta de "quem sou eu" é o único lugar onde o conjunto
	// EFETIVO é construído para exibição — o CRUD de usuários continua
	// lendo/gravando só o atribuído (ctxSessao.Perfis).
	efetivo, err := autorizacao.MontarConjuntoEfetivo(ctxSessao.Perfis, ctxSessao.CoordenaHoje)
	if err != nil {
		return ContextoDeSessaoOutput{}, err
	}

	perfis := efetivo.Ordenado()
	rotulos := make([]string, len(perfis))
	for i, p := range perfis {
		rotulos[i] = p.Rotulo()
	}
	var derivados []valueobject.Perfil
	if ctxSessao.CoordenaHoje {
		derivados = []valueobject.Perfil{valueobject.CoordenadorCurso}
	}

	return ContextoDeSessaoOutput{
		UsuarioID:         ctxSessao.UsuarioID,
		Nome:              ctxSessao.Nome,
		Email:             ctxSessao.Email.String(),
		Perfis:            perfis,
		PerfisRotulos:     rotulos,
		PerfisDerivados:   derivados,
		CursosCoordenados: ctxSessao.CursosCoordenados,
		SenhaProvisoria:   ctxSessao.SenhaProvisoria,
		Instituicao:       instituicao,
		// União das permissões dos perfis do conjunto EFETIVO (A-07,
		// §6.4) — o backend reavalia em toda requisição, nunca confia no
		// que o cliente devolveria (A-04).
		Permissoes: permissoesDoConjunto(efetivo),
	}, nil
}

var todasAsPermissoes = []autorizacao.Permissao{
	autorizacao.InstituicaoListar, autorizacao.InstituicaoCriar, autorizacao.InstituicaoEditar, autorizacao.InstituicaoInativar,
	autorizacao.PIGerenciar,
	autorizacao.UsuarioListar, autorizacao.UsuarioCriar, autorizacao.UsuarioEditar, autorizacao.UsuarioExcluir, autorizacao.UsuarioRedefinirSenha,
	autorizacao.AdministradorGerenciar,
	// Acrescentadas por fundacao-metas.md §6.1 — permissões das quatro
	// features de metas (indicadores já em uso; as demais preparadas para
	// as próximas features da cadeia).
	autorizacao.IndicadorPlataformaGerenciar, autorizacao.IndicadorListar, autorizacao.IndicadorGerenciar,
	autorizacao.MetaListar, autorizacao.MetaGerenciar,
	autorizacao.CursoListar, autorizacao.CursoGerenciar, autorizacao.CursoLerProprio,
	autorizacao.DesignacaoGerenciar, autorizacao.PeriodoGerenciar,
	autorizacao.PlanoListar, autorizacao.PlanoGerenciar, autorizacao.PlanoLerProprio,
	autorizacao.EntregaRegistrar, autorizacao.EntregaListar, autorizacao.EntregaAvaliar,
	autorizacao.RelatorioLer, autorizacao.RelatorioExportar, autorizacao.RelatorioLerProprio,
}

// permissoesDoConjunto é a única fonte de verdade que o frontend consulta
// para menu e ações — o backend nunca confia nela de volta, sempre
// reavalia (A-04, design.md §6.2 R4).
func permissoesDoConjunto(c valueobject.ConjuntoDePerfis) []autorizacao.Permissao {
	var minhas []autorizacao.Permissao
	for _, permissao := range todasAsPermissoes {
		if c.Pode(permissao) {
			minhas = append(minhas, permissao)
		}
	}
	return minhas
}
