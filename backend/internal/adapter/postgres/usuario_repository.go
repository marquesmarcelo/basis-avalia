package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// ordenacaoUsuario é a allowlist de defesa em profundidade — o handler já
// valida contra a mesma lista (design.md §6.2), mas o adapter nunca confia
// em string vinda de fora sem checar de novo. "perfil" não está aqui: com
// o modelo de conjunto, "ordenar por perfil" deixou de ter um valor único
// por linha (G-07).
var ordenacaoUsuario = map[string]string{
	"nome":      `nome COLLATE "pt-BR-x-icu"`,
	"email":     "email",
	"criado_em": "criado_em",
}

type linhaUsuario struct {
	ID                      uuid.UUID  `db:"id"`
	InstituicaoID           *uuid.UUID `db:"instituicao_id"`
	Nome                    string     `db:"nome"`
	Email                   string     `db:"email"`
	SenhaHash               *string    `db:"senha_hash"`
	SenhaProvisoria         bool       `db:"senha_provisoria"`
	SessoesValidasAPartirDe time.Time  `db:"sessoes_validas_a_partir_de"`
	ProvedorIdentidade      string     `db:"provedor_identidade"`
	IdentificadorExterno    *string    `db:"identificador_externo"`
	CriadoEm                time.Time  `db:"criado_em"`
	AtualizadoEm            *time.Time `db:"atualizado_em"`
	ExcluidoEm              *time.Time `db:"excluido_em"`
	Versao                  int        `db:"versao"`
}

func (l linhaUsuario) paraDominio(perfis valueobject.ConjuntoDePerfis) (*usuario.Usuario, error) {
	email, err := valueobject.NovoEmail(l.Email)
	if err != nil {
		return nil, err
	}
	var hash *valueobject.SenhaHash
	if l.SenhaHash != nil {
		h, err := valueobject.NovaSenhaHash(*l.SenhaHash)
		if err != nil {
			return nil, err
		}
		hash = &h
	}
	return &usuario.Usuario{
		ID:                      l.ID,
		InstituicaoID:           l.InstituicaoID,
		Nome:                    l.Nome,
		Email:                   email,
		SenhaHash:               hash,
		Perfis:                  perfis,
		SenhaProvisoria:         l.SenhaProvisoria,
		SessoesValidasAPartirDe: l.SessoesValidasAPartirDe,
		ProvedorIdentidade:      valueobject.ProvedorIdentidade(l.ProvedorIdentidade),
		IdentificadorExterno:    l.IdentificadorExterno,
		CriadoEm:                l.CriadoEm,
		AtualizadoEm:            l.AtualizadoEm,
		ExcluidoEm:              l.ExcluidoEm,
		Versao:                  l.Versao,
	}, nil
}

type UsuarioRepository struct {
	db *sqlx.DB
}

func NovoUsuarioRepository(db *sqlx.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

var _ port.UsuarioRepository = (*UsuarioRepository)(nil)

// perfisDoUsuario é a "segunda consulta" do padrão 1+1 (design.md §5.7) —
// usada tanto por BuscarPorID (um id) quanto, com IN, por Listar (até
// page_size ids).
func perfisDoUsuario(ctx context.Context, ex sqlx.ExtContext, id uuid.UUID) (valueobject.ConjuntoDePerfis, error) {
	var brutos []string
	if err := sqlx.SelectContext(ctx, ex, &brutos, `SELECT perfil FROM usuario_perfil WHERE usuario_id = $1`, id); err != nil {
		return valueobject.ConjuntoDePerfis{}, err
	}
	perfis := make([]valueobject.Perfil, len(brutos))
	for i, p := range brutos {
		perfis[i] = valueobject.Perfil(p)
	}
	return valueobject.NovoConjunto(perfis...)
}

func (r *UsuarioRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (*usuario.Usuario, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 2)
	if err != nil {
		return nil, err
	}
	consulta := "SELECT * FROM usuario WHERE id = $1 AND " + clausula
	todosArgs := append([]any{id}, args...)

	ex := Executor(ctx, r.db)
	var linha linhaUsuario
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, todosArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNaoEncontrado
		}
		return nil, err
	}
	perfis, err := perfisDoUsuario(ctx, ex, id)
	if err != nil {
		return nil, err
	}
	return linha.paraDominio(perfis)
}

func (r *UsuarioRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarUsuarios) (port.ResultadoListaUsuarios, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		return port.ResultadoListaUsuarios{}, err
	}
	n := len(args) + 1

	condicoes := []string{clausula}
	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(nome)) LIKE unaccent(lower($%d)) OR unaccent(lower(email)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
	}
	if filtro.Perfil != nil {
		// Filtro de posse (G-03): possui o perfil, não "o perfil é este" —
		// mesma forma da cláusula de exigePerfil de AplicarEscopo, mas de
		// origem diferente (filtro de negócio, não autorização).
		condicoes = append(condicoes, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM usuario_perfil up WHERE up.usuario_id = usuario.id AND up.perfil = $%d)", n))
		args = append(args, string(*filtro.Perfil))
		n++
	}

	where := strings.Join(condicoes, " AND ")

	var total int
	ex := Executor(ctx, r.db)
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM usuario WHERE "+where, args...); err != nil {
		return port.ResultadoListaUsuarios{}, err
	}

	sortColuna, ok := ordenacaoUsuario[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoUsuario["nome"]
	}
	ordem := "ASC"
	if filtro.Order == "desc" {
		ordem = "DESC"
	}
	page, pageSize := filtro.Page, filtro.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	argsPaginados := append(append([]any{}, args...), pageSize, offset)
	consulta := fmt.Sprintf("SELECT * FROM usuario WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortColuna, ordem, n, n+1)

	var linhas []linhaUsuario
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaUsuarios{}, err
	}

	// Segunda consulta do padrão 1+1 (design.md §5.7): nunca N+1, nunca
	// GROUP BY antes do LIMIT.
	ids := make([]uuid.UUID, len(linhas))
	for i, l := range linhas {
		ids[i] = l.ID
	}
	perfisPorUsuario := map[uuid.UUID][]valueobject.Perfil{}
	if len(ids) > 0 {
		type linhaPerfil struct {
			UsuarioID uuid.UUID `db:"usuario_id"`
			Perfil    string    `db:"perfil"`
		}
		var vinculos []linhaPerfil
		consultaVinculos, argsVinculos, err := sqlx.In(`SELECT usuario_id, perfil FROM usuario_perfil WHERE usuario_id IN (?)`, ids)
		if err != nil {
			return port.ResultadoListaUsuarios{}, err
		}
		consultaVinculos = r.db.Rebind(consultaVinculos)
		if err := sqlx.SelectContext(ctx, ex, &vinculos, consultaVinculos, argsVinculos...); err != nil {
			return port.ResultadoListaUsuarios{}, err
		}
		for _, v := range vinculos {
			perfisPorUsuario[v.UsuarioID] = append(perfisPorUsuario[v.UsuarioID], valueobject.Perfil(v.Perfil))
		}
	}

	itens := make([]usuario.Usuario, 0, len(linhas))
	for _, l := range linhas {
		conjunto, err := valueobject.NovoConjunto(perfisPorUsuario[l.ID]...)
		if err != nil {
			return port.ResultadoListaUsuarios{}, err
		}
		u, err := l.paraDominio(conjunto)
		if err != nil {
			return port.ResultadoListaUsuarios{}, err
		}
		itens = append(itens, *u)
	}

	return port.ResultadoListaUsuarios{Itens: itens, Total: total}, nil
}

func (r *UsuarioRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario) error {
	ex := Executor(ctx, r.db)
	var senhaHash *string
	if u.SenhaHash != nil {
		codificado := u.SenhaHash.Codificado()
		senhaHash = &codificado
	}
	_, err := ex.ExecContext(ctx,
		`INSERT INTO usuario (id, instituicao_id, nome, email, senha_hash, senha_provisoria,
		                       sessoes_validas_a_partir_de, provedor_identidade, identificador_externo, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		u.ID, u.InstituicaoID, u.Nome, u.Email.String(), senhaHash, u.SenhaProvisoria,
		u.SessoesValidasAPartirDe, string(u.ProvedorIdentidade), u.IdentificadorExterno, u.CriadoEm, u.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_usuario_instituicao_email_ativo") {
			return domain.ErrEmailDuplicado
		}
		return err
	}
	return inserirVinculosDePerfil(ctx, ex, u.ID, u.InstituicaoID, u.Perfis)
}

func inserirVinculosDePerfil(ctx context.Context, ex sqlx.ExtContext, usuarioID uuid.UUID, instituicaoID *uuid.UUID, perfis valueobject.ConjuntoDePerfis) error {
	for _, perfil := range perfis.Ordenado() {
		var instituicaoDoVinculo *uuid.UUID
		if perfil != valueobject.AdministradorSistema {
			instituicaoDoVinculo = instituicaoID
		}
		if _, err := ex.ExecContext(ctx,
			`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,$2,$3)`,
			usuarioID, string(perfil), instituicaoDoVinculo,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *UsuarioRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, u *usuario.Usuario, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoUsuario, 6)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{u.Nome, u.Email.String(), agora, u.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE usuario SET nome=$1, email=$2, atualizado_em=$3, versao=versao+1 "+
			"WHERE id=$4 AND versao=$5 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "uq_usuario_instituicao_email_ativo") {
			return domain.ErrEmailDuplicado
		}
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrConflitoDeVersao
	}
	u.AtualizadoEm = &agora
	u.Versao = versaoEsperada + 1
	return nil
}

// SubstituirPerfis reescreve o conjunto inteiro do usuário — nunca aplica
// diferença. Confere o escopo antes de mexer no vínculo, exatamente como
// as demais mutações (design.md §4.2).
func (r *UsuarioRepository) SubstituirPerfis(ctx context.Context, escopo autorizacao.Escopo, usuarioID uuid.UUID, perfis valueobject.ConjuntoDePerfis) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoUsuario, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)

	var existe bool
	args := append([]any{usuarioID}, escopoArgs...)
	if err := sqlx.GetContext(ctx, ex, &existe, "SELECT EXISTS (SELECT 1 FROM usuario WHERE id = $1 AND "+clausula+")", args...); err != nil {
		return err
	}
	if !existe {
		return domain.ErrNaoEncontrado
	}

	if _, err := ex.ExecContext(ctx, `DELETE FROM usuario_perfil WHERE usuario_id = $1`, usuarioID); err != nil {
		return err
	}

	var instituicaoID *uuid.UUID
	if !escopo.Plataforma() {
		instituicaoID = escopo.InstituicaoID()
	}
	return inserirVinculosDePerfil(ctx, ex, usuarioID, instituicaoID, perfis)
}

func (r *UsuarioRepository) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoUsuario, 3)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{instante, id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE usuario SET excluido_em=$1, senha_hash=NULL WHERE id=$2 AND "+clausula,
		args...,
	)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *UsuarioRepository) DefinirSenha(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, hash valueobject.SenhaHash, senhaProvisoria bool, instante time.Time) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoUsuario, 5)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{hash.Codificado(), senhaProvisoria, instante, id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE usuario SET senha_hash=$1, senha_provisoria=$2, sessoes_validas_a_partir_de=$3, atualizado_em=$3 "+
			"WHERE id=$4 AND "+clausula,
		args...,
	)
	if err != nil {
		return err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if linhas == 0 {
		return domain.ErrNaoEncontrado
	}
	return nil
}

func (r *UsuarioRepository) InvalidarSessoes(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, instante time.Time) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoUsuario, 3)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{instante, id}, escopoArgs...)
	_, err = ex.ExecContext(ctx,
		"UPDATE usuario SET sessoes_validas_a_partir_de=$1 WHERE id=$2 AND "+clausula,
		args...,
	)
	return err
}

// ContarDetentoresDoPerfil implementa a contagem de design.md §5.8 —
// atravessa a tabela de vínculo, nunca a antiga coluna perfil. Chamada
// sempre depois de TravarPopulacao, dentro da mesma transação (§5.4).
func (r *UsuarioRepository) ContarDetentoresDoPerfil(ctx context.Context, escopo autorizacao.Escopo, perfil valueobject.Perfil) (int, error) {
	ex := Executor(ctx, r.db)
	var total int
	var err error
	if escopo.Plataforma() {
		err = sqlx.GetContext(ctx, ex, &total, `
			SELECT count(*)
			  FROM usuario_perfil up
			  JOIN usuario u ON u.id = up.usuario_id
			 WHERE up.instituicao_id IS NULL AND up.perfil = $1 AND u.excluido_em IS NULL`,
			string(perfil))
		return total, err
	}
	err = sqlx.GetContext(ctx, ex, &total, `
		SELECT count(*)
		  FROM usuario_perfil up
		  JOIN usuario u ON u.id = up.usuario_id
		 WHERE up.instituicao_id = $1 AND up.perfil = $2 AND u.excluido_em IS NULL`,
		escopo.InstituicaoID(), string(perfil))
	return total, err
}

// TravarPopulacao implementa §5.4: serializa por instituição (lock de
// linha) no alcance institucional, ou por advisory lock no de plataforma —
// onde não há linha para travar.
func (r *UsuarioRepository) TravarPopulacao(ctx context.Context, escopo autorizacao.Escopo) error {
	ex := Executor(ctx, r.db)
	if escopo.Plataforma() {
		_, err := ex.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('populacao_administrador_sistema'))`)
		return err
	}
	if escopo.InstituicaoID() == nil {
		return domain.ErrEscopoInvalido
	}
	_, err := ex.ExecContext(ctx, `SELECT 1 FROM instituicao WHERE id = $1 FOR UPDATE`, escopo.InstituicaoID())
	return err
}

func violaIndice(err error, nomeIndice string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName == nomeIndice
	}
	return false
}
