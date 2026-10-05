package postgres

import (
	"fmt"
	"strings"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
)

// AplicarEscopo é a ÚNICA função do sistema que monta o fragmento WHERE de
// isolamento por instituição, catálogo comum, posse de perfil e carteira
// de cursos (fundacao-metas.md §3, design.md §3.2/§4.1/§4.3 de
// autenticacao-usuarios). Nenhum outro arquivo deste pacote escreve
// "instituicao_id" numa cláusula WHERE — só aqui, nos INSERT e nas duas
// portas estreitas (AutenticacaoRepository, InstituicaoPublicaQuery).
//
// alvo é a lista FECHADA declarada em alvo.go — quem determina se este
// alvo admite a exceção do catálogo comum, o EXISTS de posse ou o
// fragmento de carteira é o Alvo, nunca o chamador.
//
// Exceção sancionada, e a única:
// UsuarioRepository.ContarDetentoresDoPerfil monta um WHERE
// "up.instituicao_id ..." próprio, sobre a tabela de vínculo
// usuario_perfil. Não é filtro de isolamento — é a contagem da invariante
// de último detentor (design.md §5.8), feita sob a trava de linha, e por
// isso não passa por aqui.
//
// proximoPlaceholder é o número do próximo $N livre na query que está
// sendo montada (para poder concatenar depois de outros filtros, como a
// busca textual). Devolve a cláusula (sem a palavra WHERE) e os argumentos
// na mesma ordem dos placeholders.
func AplicarEscopo(escopo autorizacao.Escopo, alvo Alvo, proximoPlaceholder int) (string, []any, error) {
	if !escopo.Valido() {
		return "", nil, domain.ErrEscopoInvalido
	}
	// Escopo incompatível com o alvo é erro, nunca ramo ausente
	// (fundacao-metas.md §3.6): um Escopo restrito à carteira aplicado a
	// um alvo sem coluna de curso, se o ramo apenas não fosse emitido,
	// devolveria todas as linhas da instituição para um coordenador.
	if escopo.ExigePerfil() != nil && !alvo.admiteExigePerfil {
		return "", nil, domain.ErrEscopoInvalido
	}
	if escopo.RestritoACarteiraDe() != nil && !alvo.temColunaCurso {
		return "", nil, domain.ErrEscopoInvalido
	}

	condicoes := []string{alvo.alias + ".excluido_em IS NULL"}
	var args []any
	n := proximoPlaceholder

	switch {
	case escopo.Plataforma():
		// Não opcional (T-101, achado M-3): sem esta cláusula o recorte
		// de plataforma sai certo só por coincidência. Vale para TODOS os
		// alvos, sempre — nunca só para quem admite o catálogo comum.
		condicoes = append(condicoes, alvo.alias+".instituicao_id IS NULL")
	case alvo.admiteCatalogoComum:
		// A exceção do catálogo comum (fundacao-metas.md §3.3): o
		// parêntese externo do OR não é estilo — sem ele, "a AND b OR c"
		// em SQL é "(a AND b) OR c", e a disjunção escaparia do
		// "excluido_em IS NULL", devolvendo linha excluída logicamente de
		// qualquer instituição. instituicao_id IS NULL carrega a
		// garantia (incapaz por construção de retornar linha de outra
		// instituição); escopo = 'plataforma' carrega a intenção.
		condicoes = append(condicoes, fmt.Sprintf(
			"( %s.instituicao_id = $%d OR ( %s.instituicao_id IS NULL AND %s.escopo = 'plataforma' ) )",
			alvo.alias, n, alvo.alias, alvo.alias))
		args = append(args, escopo.InstituicaoID())
		n++
	default:
		condicoes = append(condicoes, fmt.Sprintf("%s.instituicao_id = $%d", alvo.alias, n))
		args = append(args, escopo.InstituicaoID())
		n++
	}

	// EXISTS de posse: o alvo não TEM um perfil, ele POSSUI vários — com o
	// modelo de conjunto, "o perfil do alvo está nesta lista" deixa de
	// fazer sentido (design.md §3.2, §4.3). Só AlvoUsuario admite isto.
	if perfil := escopo.ExigePerfil(); perfil != nil {
		condicoes = append(condicoes, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM usuario_perfil up WHERE up.usuario_id = %s.id AND up.perfil = $%d)", alvo.alias, n))
		args = append(args, string(*perfil))
		n++
	}

	// Carteira (specs/cursos/design.md §4.3, fundacao-metas.md §3.4): o
	// recorte de quem só alcança o que coordena HOJE. curso.id É a coluna
	// que designacao.curso_id referencia; nos demais alvos (temColunaCurso
	// sem ser o próprio curso), a coluna correlacionada é <alias>.curso_id
	// — "curso" é o único alias que muda a forma, e a decisão fica aqui,
	// nunca no chamador (specs/cursos/design.md §3.2). O predicado de
	// vigência é FragmentoDesignacaoVigente, byte a byte — TestAplicarEscopo_
	// CarteiraUsaOFragmentoUnico prende isso.
	if coordenadorID := escopo.RestritoACarteiraDe(); coordenadorID != nil {
		colunaCurso := alvo.alias + ".curso_id"
		if alvo.alias == "curso" {
			colunaCurso = alvo.alias + ".id"
		}
		condicoes = append(condicoes, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM designacao carteira_d WHERE carteira_d.curso_id = %s AND carteira_d.coordenador_id = $%d AND carteira_d.excluido_em IS NULL AND %s)",
			colunaCurso, n, FragmentoDesignacaoVigente("carteira_d", n+1)))
		args = append(args, *coordenadorID, escopo.DataDeReferencia().String())
		n += 2
	}

	return strings.Join(condicoes, " AND "), args, nil
}
