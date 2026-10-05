package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type linhaCredencial struct {
	ID              uuid.UUID  `db:"id"`
	InstituicaoID   *uuid.UUID `db:"instituicao_id"`
	Nome            string     `db:"nome"`
	Email           string     `db:"email"`
	SenhaHash       string     `db:"senha_hash"`
	SenhaProvisoria bool       `db:"senha_provisoria"`
}

func (l linhaCredencial) paraCredencial(perfis valueobject.ConjuntoDePerfis) (*port.CredencialUsuario, error) {
	emailVO, err := valueobject.NovoEmail(l.Email)
	if err != nil {
		return nil, err
	}
	hash, err := valueobject.NovaSenhaHash(l.SenhaHash)
	if err != nil {
		return nil, err
	}
	return &port.CredencialUsuario{
		ID: l.ID, InstituicaoID: l.InstituicaoID, Nome: l.Nome, Email: emailVO,
		SenhaHash: hash, Perfis: perfis, SenhaProvisoria: l.SenhaProvisoria,
	}, nil
}

type linhaContextoDeSessao struct {
	ID                      uuid.UUID  `db:"id"`
	Nome                    string     `db:"nome"`
	Email                   string     `db:"email"`
	Perfis                  string     `db:"perfis"`
	SenhaProvisoria         bool       `db:"senha_provisoria"`
	InstituicaoID           *uuid.UUID `db:"instituicao_id"`
	ExcluidoEm              *time.Time `db:"excluido_em"`
	SessoesValidasAPartirDe time.Time  `db:"sessoes_validas_a_partir_de"`
	InstituicaoNome         *string    `db:"instituicao_nome"`
	InstituicaoSigla        *string    `db:"instituicao_sigla"`
	InstituicaoSituacao     *string    `db:"instituicao_situacao"`
	CursosCoordenados       int        `db:"cursos_coordenados"`
}

// AutenticacaoRepository — porta estreita sem Escopo (design.md §4.2, §7.3).
type AutenticacaoRepository struct {
	db *sqlx.DB
}

func NovoAutenticacaoRepository(db *sqlx.DB) *AutenticacaoRepository {
	return &AutenticacaoRepository{db: db}
}

var _ port.AutenticacaoRepository = (*AutenticacaoRepository)(nil)

// BuscarCredencial trata "sem conta", "conta excluída" e "instituição
// inativa" como o mesmo resultado (não encontrado) — é o próprio use case
// Autenticar quem responde com a mensagem genérica em qualquer um dos três
// casos (3.1, 3.18).
func (r *AutenticacaoRepository) BuscarCredencial(ctx context.Context, instituicaoID *uuid.UUID, email valueobject.Email) (*port.CredencialUsuario, error) {
	var linha linhaCredencial
	err := sqlx.GetContext(ctx, r.db, &linha, `
		SELECT u.id, u.instituicao_id, u.nome, u.email, u.senha_hash, u.senha_provisoria
		  FROM usuario u
		  LEFT JOIN instituicao i ON i.id = u.instituicao_id
		 WHERE u.instituicao_id IS NOT DISTINCT FROM $1
		   AND u.email = $2
		   AND u.excluido_em IS NULL
		   AND (u.instituicao_id IS NULL OR i.situacao = 'ativa')`,
		instituicaoID, email.String(),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	perfis, err := perfisDoUsuario(ctx, r.db, linha.ID)
	if err != nil {
		return nil, err
	}
	return linha.paraCredencial(perfis)
}

// BuscarCredencialPropria — mesmo shape de BuscarCredencial, mas buscada
// pelo identificador que só pode ter vindo de um Ator (design.md §4.4):
// nunca filtra por instituição porque a pessoa está buscando a si mesma.
func (r *AutenticacaoRepository) BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*port.CredencialUsuario, error) {
	var linha linhaCredencial
	err := sqlx.GetContext(ctx, r.db, &linha, `
		SELECT id, instituicao_id, nome, email, senha_hash, senha_provisoria
		  FROM usuario
		 WHERE id = $1 AND excluido_em IS NULL`,
		p.UsuarioID(),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	perfis, err := perfisDoUsuario(ctx, r.db, linha.ID)
	if err != nil {
		return nil, err
	}
	return linha.paraCredencial(perfis)
}

// CarregarContextoDeSessao é a consulta mais frequente do sistema — uma
// única ida ao banco, com o conjunto de perfis agregado via array_agg
// (design.md §5.6). Sem filtro de excluido_em, de propósito: distingue
// "linha ausente" (SESSAO_EXPIRADA) de "linha excluída" (CONTA_EXCLUIDA).
func (r *AutenticacaoRepository) CarregarContextoDeSessao(ctx context.Context, usuarioID uuid.UUID, hoje valueobject.DataLocal) (*port.ContextoDeSessao, error) {
	var linha linhaContextoDeSessao
	// cursos_coordenados vem de subconsulta correlacionada, não de JOIN: um
	// JOIN com designacao multiplicaria as linhas antes do GROUP BY e
	// exigiria DISTINCT (o mesmo defeito de contagem inflada já visto em
	// indicadores/design.md) — a subconsulta soma uma vez, sem tocar no
	// agrupamento de perfis. Não junta curso (C-07, CP-06): designação
	// vigente de curso inativo continua contando. CoordenaHoje é derivado
	// em Go a partir de CursosCoordenados > 0 — evita repetir a mesma
	// subconsulta como EXISTS, mesma resposta.
	err := sqlx.GetContext(ctx, r.db, &linha, `
		SELECT u.id, u.nome, u.email, u.senha_provisoria, u.instituicao_id,
		       u.excluido_em, u.sessoes_validas_a_partir_de,
		       i.nome     AS instituicao_nome,
		       i.sigla    AS instituicao_sigla,
		       i.situacao AS instituicao_situacao,
		       -- string_agg em vez de array_agg (design.md §5.6): o driver
		       -- pgx/database/sql em uso não escaneia TEXT[] direto para
		       -- []string sem tipo adicional; uma string delimitada por
		       -- vírgula é trivial de dividir em Go e não muda o resultado,
		       -- já que "perfil" nunca contém vírgula.
		       COALESCE(string_agg(up.perfil, ',' ORDER BY up.perfil), '') AS perfis,
		       (
		         SELECT count(*) FROM designacao d
		          WHERE d.coordenador_id = u.id AND d.excluido_em IS NULL
		            AND `+FragmentoDesignacaoVigente("d", 2)+`
		       ) AS cursos_coordenados
		  FROM usuario u
		  LEFT JOIN instituicao i      ON i.id = u.instituicao_id
		  LEFT JOIN usuario_perfil up  ON up.usuario_id = u.id
		 WHERE u.id = $1
		 GROUP BY u.id, i.nome, i.sigla, i.situacao`,
		usuarioID, hoje.String(),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	emailVO, err := valueobject.NovoEmail(linha.Email)
	if err != nil {
		return nil, err
	}

	// Conjunto vazio é estado impossível (design.md §5.6, 3.6): se
	// acontecer, o erro sobe e o middleware responde 500 ERRO_INTERNO com
	// log — nunca SESSAO_EXPIRADA, que mandaria a pessoa para um laço de
	// login sem causa visível.
	var perfisBrutos []valueobject.Perfil
	if linha.Perfis != "" {
		for _, p := range strings.Split(linha.Perfis, ",") {
			perfisBrutos = append(perfisBrutos, valueobject.Perfil(p))
		}
	}
	perfis, err := valueobject.NovoConjunto(perfisBrutos...)
	if err != nil {
		return nil, err
	}

	ctxSessao := &port.ContextoDeSessao{
		UsuarioID:               linha.ID,
		Nome:                    linha.Nome,
		Email:                   emailVO,
		Perfis:                  perfis,
		SenhaProvisoria:         linha.SenhaProvisoria,
		InstituicaoID:           linha.InstituicaoID,
		ExcluidoEm:              linha.ExcluidoEm,
		SessoesValidasAPartirDe: linha.SessoesValidasAPartirDe,
		CoordenaHoje:            linha.CursosCoordenados > 0,
		CursosCoordenados:       linha.CursosCoordenados,
	}
	if linha.InstituicaoNome != nil {
		ctxSessao.InstituicaoNome = *linha.InstituicaoNome
	}
	if linha.InstituicaoSigla != nil {
		ctxSessao.InstituicaoSigla = *linha.InstituicaoSigla
	}
	if linha.InstituicaoSituacao != nil {
		ctxSessao.InstituicaoSituacao = *linha.InstituicaoSituacao
	}
	return ctxSessao, nil
}

// DefinirSenhaPropria grava a nova credencial e, no mesmo UPDATE, invalida
// as sessões anteriores (sessoes_validas_a_partir_de = instante) — é o que
// garante que o cookie novo (emt == instante) é o único que continua
// valendo (§7.1, R5).
func (r *AutenticacaoRepository) DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error {
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`UPDATE usuario SET senha_hash=$1, senha_provisoria=false, sessoes_validas_a_partir_de=$2, atualizado_em=$2
		  WHERE id=$3 AND excluido_em IS NULL`,
		hash.Codificado(), instante, p.UsuarioID(),
	)
	return err
}

func (r *AutenticacaoRepository) InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE usuario SET sessoes_validas_a_partir_de = $1 WHERE id = $2`, instante, p.UsuarioID())
	return err
}
