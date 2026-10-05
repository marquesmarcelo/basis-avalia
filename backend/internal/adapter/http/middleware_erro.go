package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type erroResposta struct {
	Codigo   string `json:"code"`
	Mensagem string `json:"message"`
	Campo    string `json:"campo,omitempty"`
}

// MiddlewareErro traduz erro de domínio para HTTP num único lugar — nenhum
// handler monta gin.H{"error": ...} à mão (design.md §6.3).
func MiddlewareErro() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		erro := c.Errors.Last().Err
		status, resp := traduzirErro(erro)
		// 500 nunca revela detalhe ao cliente (§6.3), mas precisa ficar
		// rastreável no log do servidor — inclui o caso de conjunto de
		// perfis vazio, estado impossível que design.md §5.6 exige logar.
		if status == http.StatusInternalServerError {
			log.Printf("erro interno: %v", erro)
		}
		c.AbortWithStatusJSON(status, gin.H{"error": resp})
	}
}

func traduzirErro(err error) (int, erroResposta) {
	var validacao *domain.ErrValidacao
	if errors.As(err, &validacao) {
		return http.StatusBadRequest, erroResposta{Codigo: "PAYLOAD_INVALIDO", Mensagem: validacao.Mensagem, Campo: validacao.Campo}
	}
	var parametro *domain.ErrParametro
	if errors.As(err, &parametro) {
		return http.StatusBadRequest, erroResposta{Codigo: "PARAMETRO_INVALIDO", Mensagem: "Parâmetro inválido."}
	}
	// IE-07/IN-05: a mensagem informa a contagem — total na instalação
	// para o catálogo comum (PI-5), da instituição para o indicador
	// próprio. Nunca identifica quais instituições usam.
	var comMeta *domain.ErrIndicadorComMeta
	if errors.As(err, &comMeta) {
		return http.StatusConflict, erroResposta{
			Codigo:   "INDICADOR_COM_META",
			Mensagem: fmt.Sprintf("Este indicador é usado por %d meta(s) e não pode ser excluído.", comMeta.Total),
		}
	}

	switch {
	case errors.Is(err, domain.ErrSenhaObrigatoria):
		return http.StatusBadRequest, erroResposta{Codigo: "SENHA_OBRIGATORIA", Mensagem: "Informe a senha.", Campo: "senha"}
	case errors.Is(err, domain.ErrSenhaAcimaDoLimite):
		return http.StatusBadRequest, erroResposta{Codigo: "SENHA_ACIMA_DO_LIMITE", Mensagem: "A senha informada é grande demais.", Campo: "senha"}
	case errors.Is(err, domain.ErrCredenciaisInvalidas):
		return http.StatusUnauthorized, erroResposta{Codigo: "CREDENCIAIS_INVALIDAS", Mensagem: "Instituição, e-mail ou senha inválidos."}
	case errors.Is(err, domain.ErrSenhaAtualIncorreta):
		return http.StatusUnauthorized, erroResposta{Codigo: "SENHA_ATUAL_INCORRETA", Mensagem: "A senha atual está incorreta.", Campo: "senha_atual"}
	case errors.Is(err, domain.ErrRedefinirPropriaSenhaNegada):
		return http.StatusForbidden, erroResposta{Codigo: "PERMISSAO_NEGADA", Mensagem: "Use Alterar minha senha para trocar a sua própria senha."}
	case errors.Is(err, domain.ErrPermissaoNegada):
		return http.StatusForbidden, erroResposta{Codigo: "PERMISSAO_NEGADA", Mensagem: "Você não tem permissão para executar esta ação."}
	case errors.Is(err, domain.ErrPerfilNaoAtribuivel):
		return http.StatusForbidden, erroResposta{Codigo: "PERFIL_NAO_ATRIBUIVEL", Mensagem: "Este perfil não pode ser atribuído."}
	case errors.Is(err, domain.ErrAutoExclusaoNegada):
		return http.StatusForbidden, erroResposta{Codigo: "AUTO_EXCLUSAO_NEGADA", Mensagem: "Você não pode executar esta ação sobre a própria conta."}
	case errors.Is(err, domain.ErrAlteracaoDosPropriosPerfisNegada):
		return http.StatusForbidden, erroResposta{Codigo: "ALTERACAO_DOS_PROPRIOS_PERFIS_NEGADA", Mensagem: "Você não pode alterar os seus próprios perfis."}
	case errors.Is(err, domain.ErrPerfilInvalido):
		return http.StatusBadRequest, erroResposta{Codigo: "PERFIL_INVALIDO", Mensagem: "Perfil inválido.", Campo: "perfis"}
	case errors.Is(err, domain.ErrCombinacaoDePerfisInvalida):
		return http.StatusBadRequest, erroResposta{Codigo: "COMBINACAO_DE_PERFIS_INVALIDA", Mensagem: "Administrador do Sistema não pode ser combinado com outro perfil.", Campo: "perfis"}
	case errors.Is(err, domain.ErrNaoEncontrado):
		return http.StatusNotFound, erroResposta{Codigo: "NAO_ENCONTRADO", Mensagem: "Registro não encontrado."}
	case errors.Is(err, domain.ErrConflitoDeVersao):
		return http.StatusConflict, erroResposta{Codigo: "CONFLITO_DE_VERSAO", Mensagem: "Este registro foi alterado por outro usuário enquanto você editava."}
	case errors.Is(err, domain.ErrEmailDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "EMAIL_DUPLICADO", Mensagem: "Já existe um usuário ativo com este e-mail nesta instituição.", Campo: "email"}
	case errors.Is(err, domain.ErrSiglaDuplicada):
		return http.StatusConflict, erroResposta{Codigo: "SIGLA_DUPLICADA", Mensagem: "Já existe uma instituição com esta sigla.", Campo: "sigla"}
	case errors.Is(err, domain.ErrCodigoEMecDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "CODIGO_EMEC_DUPLICADO", Mensagem: "Já existe uma instituição com este código e-MEC.", Campo: "codigo_emec"}
	case errors.Is(err, domain.ErrUltimoPesquisadorInstitucional):
		return http.StatusConflict, erroResposta{Codigo: "ULTIMO_PESQUISADOR_INSTITUCIONAL", Mensagem: "É necessário manter pelo menos um Pesquisador Institucional ativo nesta instituição."}
	case errors.Is(err, domain.ErrUltimoAdministradorSistema):
		return http.StatusConflict, erroResposta{Codigo: "ULTIMO_ADMINISTRADOR_SISTEMA", Mensagem: "É necessário manter pelo menos um Administrador do Sistema ativo."}
	case errors.Is(err, domain.ErrValorInvalido):
		return http.StatusBadRequest, erroResposta{Codigo: "VALOR_INVALIDO", Mensagem: "Valor inválido."}
	case errors.Is(err, domain.ErrEscopoImutavel):
		return http.StatusBadRequest, erroResposta{Codigo: "ESCOPO_IMUTAVEL", Mensagem: "O escopo do indicador não pode ser alterado.", Campo: "escopo"}
	case errors.Is(err, domain.ErrReferenciaInstrumentoInvalida):
		return http.StatusBadRequest, erroResposta{Codigo: "REFERENCIA_INSTRUMENTO_INVALIDA", Mensagem: "Referência do instrumento inválida para este escopo.", Campo: "referencia_instrumento"}
	case errors.Is(err, domain.ErrCodigoIndicadorDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "CODIGO_INDICADOR_DUPLICADO", Mensagem: "Já existe um indicador com este código.", Campo: "codigo"}
	case errors.Is(err, domain.ErrNomeMetaDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "NOME_META_DUPLICADO", Mensagem: "Já existe uma meta com este nome nesta instituição.", Campo: "nome"}
	case errors.Is(err, domain.ErrIndicadorObrigatorio):
		return http.StatusBadRequest, erroResposta{Codigo: "INDICADOR_OBRIGATORIO", Mensagem: "Selecione pelo menos um indicador.", Campo: "indicadores"}
	case errors.Is(err, domain.ErrIndicadoresAcimaDoLimite):
		return http.StatusBadRequest, erroResposta{Codigo: "INDICADORES_ACIMA_DO_LIMITE", Mensagem: "Uma meta pode ter no máximo 5 indicadores.", Campo: "indicadores"}
	case errors.Is(err, domain.ErrIndicadorDuplicadoNaMeta):
		return http.StatusBadRequest, erroResposta{Codigo: "INDICADOR_DUPLICADO_NA_META", Mensagem: "Este indicador já está na lista.", Campo: "indicadores"}
	case errors.Is(err, domain.ErrIndicadorInativo):
		return http.StatusBadRequest, erroResposta{Codigo: "INDICADOR_INATIVO", Mensagem: "Indicador inativo não pode ser incluído em nova meta.", Campo: "indicadores"}
	case errors.Is(err, domain.ErrMetaEmPlano):
		return http.StatusConflict, erroResposta{Codigo: "META_EM_PLANO", Mensagem: "Esta meta é usada por um plano e não pode ser excluída."}

	// specs/cursos/design.md §6 — curso e designação de coordenação.
	case errors.Is(err, domain.ErrNomeCursoDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "NOME_CURSO_DUPLICADO", Mensagem: "Já existe um curso com este nome nesta instituição.", Campo: "nome"}
	case errors.Is(err, domain.ErrCodigoEMecCursoDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "CODIGO_EMEC_CURSO_DUPLICADO", Mensagem: "Já existe um curso com este código e-MEC nesta instituição.", Campo: "codigo_emec"}
	case errors.Is(err, domain.ErrCursoComVinculo):
		return http.StatusConflict, erroResposta{Codigo: "CURSO_COM_VINCULO", Mensagem: "Este curso tem designação ou plano vinculado e não pode ser excluído."}
	case errors.Is(err, domain.ErrDesignacaoDatasInvalidas):
		return http.StatusBadRequest, erroResposta{Codigo: "DESIGNACAO_DATAS_INVALIDAS", Mensagem: "A data de fim precisa ser igual ou posterior à data de início.", Campo: "data_fim"}
	case errors.Is(err, domain.ErrDesignacaoSobreposta):
		return http.StatusConflict, erroResposta{Codigo: "DESIGNACAO_SOBREPOSTA", Mensagem: "Já existe uma designação vigente para este curso nesse período."}
	case errors.Is(err, domain.ErrCoordenadorInvalido):
		return http.StatusBadRequest, erroResposta{Codigo: "COORDENADOR_INVALIDO", Mensagem: "Este usuário não pode ser designado coordenador.", Campo: "coordenador_id"}
	case errors.Is(err, domain.ErrDesignacaoComEfeito):
		return http.StatusConflict, erroResposta{Codigo: "DESIGNACAO_COM_EFEITO", Mensagem: "Esta designação já produziu efeito — só portaria e data de fim podem ser alteradas."}

	// specs/plano-acao/design.md §6 — período, plano de ação
	// curso/coordenador, itens e documento.
	case errors.Is(err, domain.ErrPeriodoDatasInvalidas):
		return http.StatusBadRequest, erroResposta{Codigo: "PERIODO_DATAS_INVALIDAS", Mensagem: "A data de fim precisa ser igual ou posterior à data de início.", Campo: "data_fim"}
	case errors.Is(err, domain.ErrNomePeriodoDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "NOME_PERIODO_DUPLICADO", Mensagem: "Já existe um período com este nome nesta instituição.", Campo: "nome"}
	case errors.Is(err, domain.ErrPeriodoComPlano):
		return http.StatusConflict, erroResposta{Codigo: "PERIODO_COM_PLANO", Mensagem: "Este período tem planos vinculados e não pode ser excluído."}
	case errors.Is(err, domain.ErrPlanoDuplicado):
		return http.StatusConflict, erroResposta{Codigo: "PLANO_DUPLICADO", Mensagem: "Já existe um plano para este curso neste período.", Campo: "periodo_id"}
	case errors.Is(err, domain.ErrAprovacaoIncompleta):
		return http.StatusBadRequest, erroResposta{Codigo: "APROVACAO_INCOMPLETA", Mensagem: "Informe a data e o órgão de aprovação, ou deixe os dois em branco.", Campo: "aprovacao_data"}
	case errors.Is(err, domain.ErrCursoInativo):
		return http.StatusBadRequest, erroResposta{Codigo: "CURSO_INATIVO", Mensagem: "Curso inativo não pode receber plano.", Campo: "curso_id"}
	case errors.Is(err, domain.ErrPlanoSemItem):
		return http.StatusBadRequest, erroResposta{Codigo: "PLANO_SEM_ITEM", Mensagem: "Adicione pelo menos uma meta antes de publicar este plano."}
	case errors.Is(err, domain.ErrPeriodoEncerrado):
		return http.StatusConflict, erroResposta{Codigo: "PERIODO_ENCERRADO", Mensagem: "O período deste plano já encerrou."}
	case errors.Is(err, domain.ErrPlanoComEntrega):
		return http.StatusConflict, erroResposta{Codigo: "PLANO_COM_ENTREGA", Mensagem: "Não é possível concluir: este plano já tem entregas registradas."}
	case errors.Is(err, domain.ErrMotivoObrigatorio):
		return http.StatusBadRequest, erroResposta{Codigo: "MOTIVO_OBRIGATORIO", Mensagem: "Informe o motivo do encerramento.", Campo: "motivo"}
	case errors.Is(err, domain.ErrPlanoVigenteNaoExcluivel):
		return http.StatusConflict, erroResposta{Codigo: "PLANO_VIGENTE_NAO_EXCLUIVEL", Mensagem: "Este plano não pode ser excluído nesta situação."}
	case errors.Is(err, domain.ErrQuantidadeInvalida):
		return http.StatusBadRequest, erroResposta{Codigo: "QUANTIDADE_INVALIDA", Mensagem: "Informe uma quantidade de 1 ou mais.", Campo: "quantidade"}
	case errors.Is(err, domain.ErrMetaDuplicadaNoPlano):
		return http.StatusConflict, erroResposta{Codigo: "META_DUPLICADA_NO_PLANO", Mensagem: "Esta meta já está neste plano.", Campo: "meta_id"}
	case errors.Is(err, domain.ErrMetaInativa):
		return http.StatusBadRequest, erroResposta{Codigo: "META_INATIVA", Mensagem: "Meta inativa não pode ser incluída em novo item.", Campo: "meta_id"}
	case errors.Is(err, domain.ErrPlanoEncerradoParaEdicao):
		return http.StatusConflict, erroResposta{Codigo: "PLANO_ENCERRADO_PARA_EDICAO", Mensagem: "Alteração bloqueada: este plano já está encerrado."}
	case errors.Is(err, domain.ErrItemComEntrega):
		return http.StatusConflict, erroResposta{Codigo: "ITEM_COM_ENTREGA", Mensagem: "Este item já tem entregas registradas — reduza a quantidade em vez de removê-lo."}
	case errors.Is(err, domain.ErrCursosObrigatorios):
		return http.StatusBadRequest, erroResposta{Codigo: "CURSOS_OBRIGATORIOS", Mensagem: "Selecione ao menos um curso.", Campo: "cursos"}
	case errors.Is(err, domain.ErrLoteAcimaDoLimite):
		return http.StatusBadRequest, erroResposta{Codigo: "LOTE_ACIMA_DO_LIMITE", Mensagem: "O lote aceita no máximo 100 cursos.", Campo: "cursos"}

	// specs/metas-coordenacao/design.md §11 — entrega, anexo, avaliação e
	// relatório de desempenho.
	case errors.Is(err, domain.ErrPlanoNaoVigente):
		return http.StatusConflict, erroResposta{Codigo: "PLANO_NAO_VIGENTE", Mensagem: "O plano deste curso não está vigente."}
	case errors.Is(err, domain.ErrPeriodoNaoIniciado):
		return http.StatusConflict, erroResposta{Codigo: "PERIODO_NAO_INICIADO", Mensagem: "Este período ainda não começou."}
	case errors.Is(err, domain.ErrCursoSemCoordenador):
		return http.StatusConflict, erroResposta{Codigo: "CURSO_SEM_COORDENADOR", Mensagem: "Este curso está sem coordenador designado."}
	case errors.Is(err, domain.ErrEntregaSemAnexo):
		return http.StatusBadRequest, erroResposta{Codigo: "ENTREGA_SEM_ANEXO", Mensagem: "Anexe pelo menos um comprovante.", Campo: "arquivos"}
	case errors.Is(err, domain.ErrAnexoTipoNaoPermitido):
		return http.StatusBadRequest, erroResposta{Codigo: "ANEXO_TIPO_NAO_PERMITIDO", Mensagem: "O conteúdo do arquivo não corresponde a um tipo permitido.", Campo: "arquivos"}
	case errors.Is(err, domain.ErrAnexoAcimaDoLimite):
		return http.StatusBadRequest, erroResposta{Codigo: "ANEXO_ACIMA_DO_LIMITE", Mensagem: "Arquivo acima do limite de 10 MB.", Campo: "arquivos"}
	case errors.Is(err, domain.ErrAnexosAcimaDoLimite):
		return http.StatusBadRequest, erroResposta{Codigo: "ANEXOS_ACIMA_DO_LIMITE", Mensagem: "Limite de 10 arquivos ou 50 MB por entrega atingido.", Campo: "arquivos"}
	case errors.Is(err, domain.ErrIdempotencyKeyObrigatoria):
		return http.StatusBadRequest, erroResposta{Codigo: "IDEMPOTENCY_KEY_OBRIGATORIA", Mensagem: "Cabeçalho Idempotency-Key obrigatório."}
	case errors.Is(err, domain.ErrEntregaAceitaNaoEditavel):
		return http.StatusConflict, erroResposta{Codigo: "ENTREGA_ACEITA_NAO_EDITAVEL", Mensagem: "Esta entrega está aceita e não pode ser editada."}
	case errors.Is(err, domain.ErrEntregaAceitaNaoExcluivel):
		return http.StatusConflict, erroResposta{Codigo: "ENTREGA_ACEITA_NAO_EXCLUIVEL", Mensagem: "Esta entrega está aceita e não pode ser excluída."}
	case errors.Is(err, domain.ErrExclusaoDeEntregaAlheia):
		return http.StatusForbidden, erroResposta{Codigo: "EXCLUSAO_DE_ENTREGA_ALHEIA", Mensagem: "Você só pode excluir entregas enviadas por você."}
	case errors.Is(err, domain.ErrPrazoDeCorrecaoExpirado):
		return http.StatusConflict, erroResposta{Codigo: "PRAZO_DE_CORRECAO_EXPIRADO", Mensagem: "O prazo de correção desta entrega terminou."}
	case errors.Is(err, domain.ErrLimiteDeRodadasAtingido):
		return http.StatusConflict, erroResposta{Codigo: "LIMITE_DE_RODADAS_ATINGIDO", Mensagem: "Esta entrega já teve três recusas e não pode mais ser corrigida."}
	case errors.Is(err, domain.ErrEntregaJaAvaliada):
		return http.StatusConflict, erroResposta{Codigo: "ENTREGA_JA_AVALIADA", Mensagem: "Esta entrega não está mais pendente de avaliação."}
	case errors.Is(err, domain.ErrEntregaNaoEstaAceita):
		return http.StatusConflict, erroResposta{Codigo: "ENTREGA_NAO_ESTA_ACEITA", Mensagem: "Esta entrega não está aceita."}
	case errors.Is(err, domain.ErrPeriodoObrigatorio):
		return http.StatusBadRequest, erroResposta{Codigo: "PERIODO_OBRIGATORIO", Mensagem: "Informe o período.", Campo: "periodo_id"}
	case errors.Is(err, domain.ErrCaractereDeControleNaoPermitido):
		return http.StatusBadRequest, erroResposta{Codigo: "CARACTERE_DE_CONTROLE_NAO_PERMITIDO", Mensagem: "O texto não pode conter quebra de linha ou caractere de controle."}

	// ErrMarcadorNaoEncontrado (P-08) cai no default: 500 genérico ao
	// cliente, mensagem completa só no log do servidor.
	default:
		// Nunca stack trace, nome de tabela, SQL ou caminho de arquivo — em
		// nenhum APP_ENV (design.md §6.3). O detalhe vai para o log do servidor.
		return http.StatusInternalServerError, erroResposta{Codigo: "ERRO_INTERNO", Mensagem: "Não foi possível concluir a operação agora."}
	}
}
