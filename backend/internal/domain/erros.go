package domain

import "errors"

// Erros de domínio compartilhados entre use cases e adapters. Traduzidos
// para HTTP por internal/adapter/http/middleware_erro.go (design.md §6.3).
var (
	ErrCredenciaisInvalidas             = errors.New("credenciais inválidas")
	ErrSenhaAtualIncorreta              = errors.New("a senha atual está incorreta")
	ErrPermissaoNegada                  = errors.New("permissão negada")
	ErrPerfilNaoAtribuivel              = errors.New("perfil não atribuível")
	ErrAutoExclusaoNegada               = errors.New("autoexclusão negada")
	ErrAlteracaoDosPropriosPerfisNegada = errors.New("alteração dos próprios perfis negada")
	ErrPerfilInvalido                   = errors.New("perfil inválido")
	ErrCombinacaoDePerfisInvalida       = errors.New("combinação de perfis inválida")
	ErrConjuntoDePerfisVazio            = errors.New("conjunto de perfis vazio")
	ErrNaoEncontrado                    = errors.New("não encontrado")
	ErrConflitoDeVersao                 = errors.New("conflito de versão")
	ErrEmailDuplicado                   = errors.New("e-mail duplicado")
	ErrSiglaDuplicada                   = errors.New("sigla duplicada")
	ErrCodigoEMecDuplicado              = errors.New("código e-mec duplicado")
	ErrUltimoPesquisadorInstitucional   = errors.New("último pesquisador institucional")
	ErrUltimoAdministradorSistema       = errors.New("último administrador do sistema")
	ErrSenhaObrigatoria                 = errors.New("senha obrigatória")
	ErrSenhaAcimaDoLimite               = errors.New("senha acima do limite")
	ErrEscopoInvalido                   = errors.New("escopo inválido")
	// ErrRedefinirPropriaSenhaNegada — mesmo código HTTP de ErrPermissaoNegada
	// (403 PERMISSAO_NEGADA), mensagem distinta (E-12): "Use Alterar minha
	// senha para trocar a sua própria senha."
	ErrRedefinirPropriaSenhaNegada = errors.New("redefinir a própria senha por este caminho é negado")

	// Erros de indicadores/indicadores/erros.go (specs/indicadores/design.md
	// §6) — catálogo do INEP e catálogo de metas.
	ErrValorInvalido                 = errors.New("valor inválido")
	ErrEscopoImutavel                = errors.New("escopo imutável")
	ErrReferenciaInstrumentoInvalida = errors.New("referência do instrumento inválida")
	ErrCodigoIndicadorDuplicado      = errors.New("código de indicador duplicado")
	ErrNomeMetaDuplicado             = errors.New("nome de meta duplicado")
	ErrIndicadorObrigatorio          = errors.New("indicador obrigatório")
	ErrIndicadoresAcimaDoLimite      = errors.New("indicadores acima do limite")
	ErrIndicadorDuplicadoNaMeta      = errors.New("indicador duplicado na meta")
	ErrIndicadorInativo              = errors.New("indicador inativo")
	ErrMetaEmPlano                   = errors.New("meta em plano")

	// Erros de cursos (specs/cursos/design.md §6) — curso e designação de
	// coordenação.
	ErrNomeCursoDuplicado       = errors.New("nome de curso duplicado")
	ErrCodigoEMecCursoDuplicado = errors.New("código e-mec de curso duplicado")
	ErrCursoComVinculo          = errors.New("curso com vínculo")
	ErrDesignacaoDatasInvalidas = errors.New("datas da designação inválidas")
	ErrDesignacaoSobreposta     = errors.New("designação sobreposta")
	ErrCoordenadorInvalido      = errors.New("coordenador inválido")
	ErrDesignacaoComEfeito      = errors.New("designação com efeito")

	// Erros de plano-acao (specs/plano-acao/design.md §6) — período, plano
	// de ação curso/coordenador, itens e documento.
	ErrPeriodoDatasInvalidas    = errors.New("datas do período inválidas")
	ErrNomePeriodoDuplicado     = errors.New("nome de período duplicado")
	ErrPeriodoComPlano          = errors.New("período com plano")
	ErrPlanoDuplicado           = errors.New("plano duplicado")
	ErrAprovacaoIncompleta      = errors.New("aprovação incompleta")
	ErrCursoInativo             = errors.New("curso inativo")
	ErrPlanoSemItem             = errors.New("plano sem item")
	ErrPeriodoEncerrado         = errors.New("período encerrado")
	ErrPlanoComEntrega          = errors.New("plano com entrega")
	ErrMotivoObrigatorio        = errors.New("motivo obrigatório")
	ErrPlanoVigenteNaoExcluivel = errors.New("plano vigente não excluível")
	ErrQuantidadeInvalida       = errors.New("quantidade inválida")
	ErrMetaDuplicadaNoPlano     = errors.New("meta duplicada no plano")
	ErrMetaInativa              = errors.New("meta inativa")
	ErrPlanoEncerradoParaEdicao = errors.New("plano encerrado para edição")
	ErrItemComEntrega           = errors.New("item com entrega")
	ErrCursosObrigatorios       = errors.New("cursos obrigatórios")
	ErrLoteAcimaDoLimite        = errors.New("lote acima do limite")
	ErrMarcadorNaoEncontrado    = errors.New("marcador não encontrado no modelo do documento")

	// Erros de metas-coordenacao (specs/metas-coordenacao/design.md §11) —
	// entrega, anexo, avaliação e relatório de desempenho.
	ErrPlanoNaoVigente           = errors.New("plano não vigente")
	ErrPeriodoNaoIniciado        = errors.New("período não iniciado")
	ErrCursoSemCoordenador       = errors.New("curso sem coordenador")
	ErrEntregaSemAnexo           = errors.New("entrega sem anexo")
	ErrAnexoTipoNaoPermitido     = errors.New("anexo de tipo não permitido")
	ErrAnexoAcimaDoLimite        = errors.New("anexo acima do limite")
	ErrAnexosAcimaDoLimite       = errors.New("anexos acima do limite")
	ErrIdempotencyKeyObrigatoria = errors.New("idempotency-key obrigatória")
	ErrEntregaAceitaNaoEditavel  = errors.New("entrega aceita não editável")
	ErrEntregaAceitaNaoExcluivel = errors.New("entrega aceita não excluível")
	ErrExclusaoDeEntregaAlheia   = errors.New("exclusão de entrega alheia")
	ErrPrazoDeCorrecaoExpirado   = errors.New("prazo de correção expirado")
	ErrLimiteDeRodadasAtingido   = errors.New("limite de rodadas atingido")
	ErrEntregaJaAvaliada         = errors.New("entrega já avaliada")
	ErrEntregaNaoEstaAceita      = errors.New("entrega não está aceita")
	ErrPeriodoObrigatorio        = errors.New("período obrigatório")
	// ErrChaveDuplicada — violação da PK de `idempotencia` (design.md §7,
	// M-04): nunca traduzido para HTTP diretamente, sempre capturado pelo
	// use case de registrar entrega para reler a entrega original.
	ErrChaveDuplicada = errors.New("chave de idempotência duplicada")
	// ErrCaractereDeControleNaoPermitido — fundacao-metas.md §4.7: texto
	// livre do usuário nunca contém caractere de controle (\r, \n e os
	// demais C0/DEL) — validar na entrada não basta porque quem injeta usa
	// a gramática do destino (cabeçalho de e-mail, célula de planilha).
	ErrCaractereDeControleNaoPermitido = errors.New("caractere de controle não permitido")
)

// ErrIndicadorComMeta — 409 INDICADOR_COM_META. Total é a contagem de metas
// que ainda referenciam o indicador — total agregado na família de
// plataforma (PI-5, nunca por instituição), da instituição na família
// própria (design.md §7.2).
type ErrIndicadorComMeta struct{ Total int }

func (e *ErrIndicadorComMeta) Error() string { return "indicador com meta em uso" }

// ErrValidacao carrega o campo e a mensagem exata que a spec fixa para
// problemas de corpo de requisição (400 PAYLOAD_INVALIDO).
type ErrValidacao struct {
	Campo    string
	Mensagem string
}

func (e *ErrValidacao) Error() string { return e.Mensagem }

// ErrParametro cobre problema de query string (400 PARAMETRO_INVALIDO) —
// allowlist de ordenação, paginação, filtros.
type ErrParametro struct {
	Nome string
}

func (e *ErrParametro) Error() string { return "parâmetro inválido: " + e.Nome }
