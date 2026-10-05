package http

import (
	"net/http"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	usuariocmd "github.com/basis-avalia/backend/internal/usecase/command/usuario"
	usuarioquery "github.com/basis-avalia/backend/internal/usecase/query/usuario"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// "perfil" não está na allowlist: com o modelo de conjunto, ordenar por
// perfil deixou de ter um valor único por linha (G-07).
var ordenacaoUsuarioAllowlist = map[string]bool{"nome": true, "email": true, "criado_em": true}

// UsuarioHandler serve /usuarios (PI), /instituicoes/{id}/pesquisadores
// (Administrador) e /administradores (Administrador) — design.md §4.3. A
// diferença entre as três rotas é só o Alcance passado a cada wrapper de
// registro (ver rotas.go em cmd/api/main.go), nunca um `if perfil ==`.
type UsuarioHandler struct {
	criar          *usuariocmd.CriarUsuarioUseCase
	atualizar      *usuariocmd.AtualizarUsuarioUseCase
	excluir        *usuariocmd.ExcluirUsuarioUseCase
	redefinirSenha *usuariocmd.RedefinirSenhaUsuarioUseCase
	listar         *usuarioquery.ListarUsuariosUseCase
	buscar         *usuarioquery.BuscarUsuarioUseCase
}

func NovoUsuarioHandler(
	criar *usuariocmd.CriarUsuarioUseCase,
	atualizar *usuariocmd.AtualizarUsuarioUseCase,
	excluir *usuariocmd.ExcluirUsuarioUseCase,
	redefinirSenha *usuariocmd.RedefinirSenhaUsuarioUseCase,
	listar *usuarioquery.ListarUsuariosUseCase,
	buscar *usuarioquery.BuscarUsuarioUseCase,
) *UsuarioHandler {
	return &UsuarioHandler{criar: criar, atualizar: atualizar, excluir: excluir, redefinirSenha: redefinirSenha, listar: listar, buscar: buscar}
}

func usuarioParaResposta(u usuario.Usuario) UsuarioResponse {
	var atualizadoEm *string
	if u.AtualizadoEm != nil {
		s := u.AtualizadoEm.Format(time.RFC3339)
		atualizadoEm = &s
	}
	perfisOrdenados := u.Perfis.Ordenado()
	perfis := make([]string, len(perfisOrdenados))
	rotulos := make([]string, len(perfisOrdenados))
	for i, p := range perfisOrdenados {
		perfis[i] = string(p)
		rotulos[i] = p.Rotulo()
	}
	return UsuarioResponse{
		ID: u.ID.String(), Nome: u.Nome, Email: u.Email.String(), Perfis: perfis, PerfisRotulos: rotulos,
		SenhaProvisoria: u.SenhaProvisoria,
		CriadoEm:        u.CriadoEm.Format(time.RFC3339), AtualizadoEm: atualizadoEm, Versao: u.Versao,
	}
}

// perfisDoRequest converte a lista bruta do corpo da requisição — nunca
// valida nem coage aqui (isso é regra de ConjuntoDePerfis, design.md
// §15 item 4); só traduz string para o tipo.
func perfisDoRequest(brutos []string) []valueobject.Perfil {
	if brutos == nil {
		return nil
	}
	perfis := make([]valueobject.Perfil, len(brutos))
	for i, p := range brutos {
		perfis[i] = valueobject.Perfil(p)
	}
	return perfis
}

// instituicaoDoCaminho extrai {id} da URL só para o alcance que o exige —
// PesquisadoresDeUmaInstituicao. Os demais alcances ignoram o parâmetro.
func instituicaoDoCaminhoHTTP(c *gin.Context, alcance autorizacao.Alcance) (*uuid.UUID, error) {
	if alcance != autorizacao.PesquisadoresDeUmaInstituicao {
		return nil, nil
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return nil, domain.ErrNaoEncontrado
	}
	return &id, nil
}

// Listar godoc
// @Summary      Listar usuários, pesquisadores institucionais ou administradores
// @Description  Cada rota lista um universo diferente: usuários da própria instituição de quem chama, pesquisadores institucionais de uma instituição específica, ou administradores da plataforma.
// @Tags         usuarios
// @Produce      json
// @Param        id      path      string  false  "id da instituição — só na rota de pesquisadores"
// @Param        busca   query     string  false  "busca por nome ou e-mail"
// @Param        perfil  query     string  false  "filtra por posse do perfil — só na rota de usuários"
// @Success      200 {object} map[string]any
// @Failure      400 {object} erroResposta "PARAMETRO_INVALIDO"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Router       /api/v1/usuarios [get]
// @Router       /api/v1/instituicoes/{id}/pesquisadores [get]
// @Router       /api/v1/administradores [get]
func (h *UsuarioHandler) Listar(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		paginacao, err := ParsePaginacao(c, ordenacaoUsuarioAllowlist, "nome", "asc")
		if err != nil {
			c.Error(err)
			return
		}
		var perfilFiltro *valueobject.Perfil
		if bruto := c.Query("perfil"); bruto != "" {
			p, err := valueobject.NovoPerfil(bruto)
			if err != nil {
				c.Error(&domain.ErrParametro{Nome: "perfil"})
				return
			}
			perfilFiltro = &p
		}

		resultado, err := h.listar.Executar(c.Request.Context(), usuarioquery.ListarUsuariosInput{
			Ator: ator, Alcance: alcance, InstituicaoDoCaminho: instituicaoDoCaminho,
			Busca: c.Query("busca"), Perfil: perfilFiltro,
			Page: paginacao.Page, PageSize: paginacao.PageSize, Sort: paginacao.Sort, Order: paginacao.Order,
		})
		if err != nil {
			c.Error(err)
			return
		}
		itens := make([]UsuarioResponse, 0, len(resultado.Itens))
		for _, u := range resultado.Itens {
			itens = append(itens, usuarioParaResposta(u))
		}
		c.JSON(http.StatusOK, gin.H{"data": itens, "meta": MetaPaginacao{
			Page: paginacao.Page, PageSize: paginacao.PageSize, Total: resultado.Total,
			TotalPages: calcularTotalPages(resultado.Total, paginacao.PageSize),
		}})
	}
}

// Buscar godoc
// @Summary      Buscar usuário, pesquisador institucional ou administrador por id
// @Tags         usuarios
// @Produce      json
// @Param        id          path      string  false  "id da instituição — só na rota de pesquisadores"
// @Param        usuario_id  path      string  true   "id do usuário"
// @Success      200 {object} UsuarioResponse
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Router       /api/v1/usuarios/{usuario_id} [get]
// @Router       /api/v1/instituicoes/{id}/pesquisadores/{usuario_id} [get]
// @Router       /api/v1/administradores/{usuario_id} [get]
func (h *UsuarioHandler) Buscar(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		usuarioID, err := uuid.Parse(c.Param("usuario_id"))
		if err != nil {
			c.Error(domain.ErrNaoEncontrado)
			return
		}
		u, err := h.buscar.Executar(c.Request.Context(), ator, alcance, instituicaoDoCaminho, usuarioID)
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, usuarioParaResposta(u))
	}
}

// CriarUsuario godoc
// @Summary      Criar usuário na própria instituição
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        body  body  UsuarioRequest  true  "Dados do novo usuário"
// @Success      201 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO, PERFIL_INVALIDO, COMBINACAO_DE_PERFIS_INVALIDA, SENHA_OBRIGATORIA ou SENHA_ACIMA_DO_LIMITE"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      409 {object} erroResposta "EMAIL_DUPLICADO"
// @Router       /api/v1/usuarios [post]
func (h *UsuarioHandler) CriarUsuario() gin.HandlerFunc {
	return h.Criar(autorizacao.UsuariosDaPropriaInstituicao)
}

// CriarPesquisador godoc
// @Summary      Criar pesquisador institucional de uma instituição
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        id    path  string                true  "id da instituição"
// @Param        body  body  PesquisadorRequest     true  "Dados do novo pesquisador"
// @Success      201 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO, SENHA_OBRIGATORIA ou SENHA_ACIMA_DO_LIMITE"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      409 {object} erroResposta "EMAIL_DUPLICADO"
// @Router       /api/v1/instituicoes/{id}/pesquisadores [post]
func (h *UsuarioHandler) CriarPesquisador() gin.HandlerFunc {
	return h.Criar(autorizacao.PesquisadoresDeUmaInstituicao)
}

// CriarAdministrador godoc
// @Summary      Criar administrador do sistema
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        body  body  AdministradorRequest  true  "Dados do novo administrador"
// @Success      201 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO, SENHA_OBRIGATORIA ou SENHA_ACIMA_DO_LIMITE"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      409 {object} erroResposta "EMAIL_DUPLICADO"
// @Router       /api/v1/administradores [post]
func (h *UsuarioHandler) CriarAdministrador() gin.HandlerFunc {
	return h.Criar(autorizacao.AdministradoresDaPlataforma)
}

// Criar é a fábrica compartilhada pelos três alcances — a diferença entre
// eles é só o Alcance recebido, nunca um `if perfil ==` (design.md §4.3).
// "perfis" no corpo só é lido para UsuariosDaPropriaInstituicao; os demais
// alcances têm DTO de requisição próprio (CriarPesquisador,
// CriarAdministrador) sem esse campo — o corpo continua aceito com campo
// desconhecido e ignorado se enviado mesmo assim, como U-11 exige.
func (h *UsuarioHandler) Criar(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		var req UsuarioRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
			return
		}
		if req.Nome == "" {
			c.Error(&domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."})
			return
		}
		email, err := valueobject.NovoEmail(req.Email)
		if err != nil {
			c.Error(&domain.ErrValidacao{Campo: "email", Mensagem: "Informe um e-mail válido."})
			return
		}
		senha, err := valueobject.NovaSenhaEmTexto(req.Senha)
		if err != nil {
			c.Error(err)
			return
		}

		// perfis só é lido no alcance do PI (U-12): a rota decide a coerção
		// para {aluno} quando vazio, nunca este handler. Nos demais
		// alcances o conjunto é fixo pelo Alcance — o corpo é ignorado
		// (design.md §6.2).
		var perfis []valueobject.Perfil
		if alcance == autorizacao.UsuariosDaPropriaInstituicao {
			perfis = perfisDoRequest(req.Perfis)
		}

		novo, err := h.criar.Executar(c.Request.Context(), usuariocmd.CriarUsuarioInput{
			Ator: ator, Alcance: alcance, InstituicaoDoCaminho: instituicaoDoCaminho,
			Nome: req.Nome, Email: email, Perfis: perfis, Senha: senha,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, usuarioParaResposta(novo))
	}
}

// AtualizarUsuario godoc
// @Summary      Atualizar usuário da própria instituição
// @Description  Concorrência otimista via "versao": envie o valor recebido no GET mais recente.
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        usuario_id  path  string                   true  "id do usuário"
// @Param        body        body  UsuarioAtualizarRequest  true  "Dados atualizados"
// @Success      200 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO, PERFIL_INVALIDO ou COMBINACAO_DE_PERFIS_INVALIDA"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA, ALTERACAO_DOS_PROPRIOS_PERFIS_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Failure      409 {object} erroResposta "CONFLITO_DE_VERSAO, EMAIL_DUPLICADO, ULTIMO_PESQUISADOR_INSTITUCIONAL ou ULTIMO_ADMINISTRADOR_SISTEMA"
// @Router       /api/v1/usuarios/{usuario_id} [put]
func (h *UsuarioHandler) AtualizarUsuario() gin.HandlerFunc {
	return h.Atualizar(autorizacao.UsuariosDaPropriaInstituicao)
}

// AtualizarPesquisador godoc
// @Summary      Atualizar pesquisador institucional de uma instituição
// @Description  Concorrência otimista via "versao": envie o valor recebido no GET mais recente.
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        id          path  string                       true  "id da instituição"
// @Param        usuario_id  path  string                       true  "id do usuário"
// @Param        body        body  PesquisadorAtualizarRequest  true  "Dados atualizados"
// @Success      200 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Failure      409 {object} erroResposta "CONFLITO_DE_VERSAO, EMAIL_DUPLICADO ou ULTIMO_PESQUISADOR_INSTITUCIONAL"
// @Router       /api/v1/instituicoes/{id}/pesquisadores/{usuario_id} [put]
func (h *UsuarioHandler) AtualizarPesquisador() gin.HandlerFunc {
	return h.Atualizar(autorizacao.PesquisadoresDeUmaInstituicao)
}

// AtualizarAdministrador godoc
// @Summary      Atualizar administrador do sistema
// @Description  Concorrência otimista via "versao": envie o valor recebido no GET mais recente.
// @Tags         usuarios
// @Accept       json
// @Produce      json
// @Param        usuario_id  path  string                         true  "id do usuário"
// @Param        body        body  AdministradorAtualizarRequest  true  "Dados atualizados"
// @Success      200 {object} UsuarioResponse
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Failure      409 {object} erroResposta "CONFLITO_DE_VERSAO, EMAIL_DUPLICADO ou ULTIMO_ADMINISTRADOR_SISTEMA"
// @Router       /api/v1/administradores/{usuario_id} [put]
func (h *UsuarioHandler) AtualizarAdministrador() gin.HandlerFunc {
	return h.Atualizar(autorizacao.AdministradoresDaPlataforma)
}

// Atualizar é a fábrica compartilhada pelos três alcances. "perfis" no
// corpo só é lido para UsuariosDaPropriaInstituicao; os demais alcances
// têm DTO de requisição próprio sem esse campo — o corpo continua aceito
// com campo desconhecido e ignorado se enviado mesmo assim (U-11).
func (h *UsuarioHandler) Atualizar(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		usuarioID, err := uuid.Parse(c.Param("usuario_id"))
		if err != nil {
			c.Error(domain.ErrNaoEncontrado)
			return
		}
		var req UsuarioAtualizarRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
			return
		}
		if req.Versao <= 0 {
			c.Error(&domain.ErrValidacao{Campo: "versao", Mensagem: "Requisição inválida."})
			return
		}
		if req.Nome == "" {
			c.Error(&domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."})
			return
		}
		email, err := valueobject.NovoEmail(req.Email)
		if err != nil {
			c.Error(&domain.ErrValidacao{Campo: "email", Mensagem: "Informe um e-mail válido."})
			return
		}

		// O conjunto final vem sempre daqui, nunca de uma diferença
		// (design.md §6.2, T-056) — nos demais alcances o conjunto é fixo
		// pelo Alcance e o corpo é ignorado.
		var perfis []valueobject.Perfil
		if alcance == autorizacao.UsuariosDaPropriaInstituicao {
			perfis = perfisDoRequest(req.Perfis)
		}

		atualizado, err := h.atualizar.Executar(c.Request.Context(), usuariocmd.AtualizarUsuarioInput{
			Ator: ator, Alcance: alcance, InstituicaoDoCaminho: instituicaoDoCaminho,
			UsuarioID: usuarioID, Nome: req.Nome, Email: email, Perfis: perfis, Versao: req.Versao,
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, usuarioParaResposta(atualizado))
	}
}

// Excluir godoc
// @Summary      Excluir usuário, pesquisador institucional ou administrador
// @Description  Deleção lógica (excluido_em) — nunca DELETE real. Recusada se for o último Pesquisador Institucional ou o último Administrador do Sistema.
// @Tags         usuarios
// @Param        id          path  string  false  "id da instituição — só na rota de pesquisadores"
// @Param        usuario_id  path  string  true   "id do usuário"
// @Success      204
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA, AUTO_EXCLUSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Failure      409 {object} erroResposta "ULTIMO_PESQUISADOR_INSTITUCIONAL ou ULTIMO_ADMINISTRADOR_SISTEMA"
// @Router       /api/v1/usuarios/{usuario_id} [delete]
// @Router       /api/v1/instituicoes/{id}/pesquisadores/{usuario_id} [delete]
// @Router       /api/v1/administradores/{usuario_id} [delete]
func (h *UsuarioHandler) Excluir(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		usuarioID, err := uuid.Parse(c.Param("usuario_id"))
		if err != nil {
			c.Error(domain.ErrNaoEncontrado)
			return
		}
		if err := h.excluir.Executar(c.Request.Context(), usuariocmd.ExcluirUsuarioInput{
			Ator: ator, Alcance: alcance, InstituicaoDoCaminho: instituicaoDoCaminho, UsuarioID: usuarioID,
		}); err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// RedefinirSenha godoc
// @Summary      Redefinir a senha de outro usuário
// @Description  A senha nasce provisória (troca obrigatória no próximo acesso). Redefinir a própria senha por este caminho é negado — use "Alterar minha senha".
// @Tags         usuarios
// @Accept       json
// @Param        id          path  string                  false  "id da instituição — só na rota de pesquisadores"
// @Param        usuario_id  path  string                  true   "id do usuário"
// @Param        body        body  RedefinirSenhaRequest   true   "Nova senha provisória"
// @Success      204
// @Failure      400 {object} erroResposta "PAYLOAD_INVALIDO, SENHA_OBRIGATORIA ou SENHA_ACIMA_DO_LIMITE"
// @Failure      401 {object} erroResposta "SESSAO_EXPIRADA, CONTA_EXCLUIDA ou INSTITUICAO_INATIVA"
// @Failure      403 {object} erroResposta "PERMISSAO_NEGADA ou SENHA_PROVISORIA"
// @Failure      404 {object} erroResposta "NAO_ENCONTRADO"
// @Router       /api/v1/usuarios/{usuario_id}/senha [post]
// @Router       /api/v1/instituicoes/{id}/pesquisadores/{usuario_id}/senha [post]
// @Router       /api/v1/administradores/{usuario_id}/senha [post]
func (h *UsuarioHandler) RedefinirSenha(alcance autorizacao.Alcance) gin.HandlerFunc {
	return func(c *gin.Context) {
		ator, ok := AtorDoContexto(c)
		if !ok {
			c.Error(domain.ErrEscopoInvalido)
			return
		}
		instituicaoDoCaminho, err := instituicaoDoCaminhoHTTP(c, alcance)
		if err != nil {
			c.Error(err)
			return
		}
		usuarioID, err := uuid.Parse(c.Param("usuario_id"))
		if err != nil {
			c.Error(domain.ErrNaoEncontrado)
			return
		}
		var req RedefinirSenhaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(&domain.ErrValidacao{Mensagem: "Requisição inválida."})
			return
		}
		senhaNova, err := valueobject.NovaSenhaEmTexto(req.SenhaNova)
		if err != nil {
			c.Error(err)
			return
		}
		if err := h.redefinirSenha.Executar(c.Request.Context(), usuariocmd.RedefinirSenhaUsuarioInput{
			Ator: ator, Alcance: alcance, InstituicaoDoCaminho: instituicaoDoCaminho, UsuarioID: usuarioID, SenhaNova: senhaNova,
		}); err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
