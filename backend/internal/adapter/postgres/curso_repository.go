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
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// subConsultaTemVinculo é o MESMO teste que CursoRepository.Excluir usa
// antes de recusar a exclusão — repetido aqui como coluna calculada para
// o grid já esconder o botão sem convidar a um 409 garantido.
const subConsultaTemVinculo = `(EXISTS (SELECT 1 FROM designacao WHERE curso_id = curso.id AND excluido_em IS NULL)
	    OR EXISTS (SELECT 1 FROM plano WHERE curso_id = curso.id AND excluido_em IS NULL)) AS tem_vinculo`

var ordenacaoCurso = map[string]string{
	"nome":        `curso.nome COLLATE "pt-BR-x-icu"`,
	"codigo_emec": "curso.codigo_emec",
	"grau":        "curso.grau",
	"modalidade":  "curso.modalidade",
	"coordenador": "coord.nome",
	"criado_em":   "curso.criado_em",
}

// lateralCoordenadorDoCurso é a ÚNICA fonte do campo computado
// "coordenador" (design.md §5.5, V-6: uma consulta, não uma por linha).
// Não junta situacao do curso — o coordenador aparece mesmo com o curso
// inativo (CP-06 é sobre o perfil da pessoa; aqui é sobre a mesma
// designação vigente, mostrada na linha, sem relação com a situação do
// curso que a exibe).
func lateralCoordenadorDoCurso(proximoPlaceholder int) (string, []any, int) {
	predicado := FragmentoDesignacaoVigente("d", proximoPlaceholder)
	sqlText := fmt.Sprintf(`
	LEFT JOIN LATERAL (
	  SELECT u.id, u.nome, d.data_fim::text AS data_fim,
	         EXISTS (
	           SELECT 1 FROM usuario_perfil up2
	            WHERE up2.usuario_id = u.id AND up2.perfil = 'pesquisador_institucional'
	         ) AS tambem_pi
	    FROM designacao d
	    JOIN usuario u ON u.id = d.coordenador_id
	   WHERE d.curso_id = curso.id AND d.excluido_em IS NULL AND %s
	   LIMIT 1
	) coord ON TRUE`, predicado)
	return sqlText, []any{}, proximoPlaceholder + 1
}

type linhaCurso struct {
	ID            uuid.UUID  `db:"id"`
	InstituicaoID uuid.UUID  `db:"instituicao_id"`
	Nome          string     `db:"nome"`
	CodigoEMec    *string    `db:"codigo_emec"`
	Grau          string     `db:"grau"`
	Modalidade    string     `db:"modalidade"`
	Situacao      string     `db:"situacao"`
	CriadoEm      time.Time  `db:"criado_em"`
	AtualizadoEm  *time.Time `db:"atualizado_em"`
	ExcluidoEm    *time.Time `db:"excluido_em"`
	Versao        int        `db:"versao"`

	CoordenadorID       *uuid.UUID `db:"coord_id"`
	CoordenadorNome     *string    `db:"coord_nome"`
	CoordenadorDataFim  *string    `db:"coord_data_fim"`
	CoordenadorTambemPI *bool      `db:"coord_tambem_pi"`
	TemVinculo          bool       `db:"tem_vinculo"`
}

func (l linhaCurso) paraItem() (port.ItemCurso, error) {
	nome, err := valueobject.NovoNomeCatalogo(l.Nome)
	if err != nil {
		return port.ItemCurso{}, err
	}
	var codigoEMec *valueobject.CodigoEMec
	if l.CodigoEMec != nil {
		c, err := valueobject.NovoCodigoEMec(*l.CodigoEMec)
		if err != nil {
			return port.ItemCurso{}, err
		}
		codigoEMec = &c
	}
	grau, err := valueobject.NovoGrauDeCurso(l.Grau)
	if err != nil {
		return port.ItemCurso{}, err
	}
	modalidade, err := valueobject.NovaModalidadeDeCurso(l.Modalidade)
	if err != nil {
		return port.ItemCurso{}, err
	}
	situacao := valueobject.SituacaoCurso(l.Situacao)

	item := port.ItemCurso{
		Curso: curso.Curso{
			ID: l.ID, InstituicaoID: l.InstituicaoID, Nome: nome, CodigoEMec: codigoEMec,
			Grau: grau, Modalidade: modalidade, Situacao: situacao,
			CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
		},
		TemVinculo: l.TemVinculo,
	}
	if l.CoordenadorID != nil {
		coordenador := &port.CoordenadorDoCurso{ID: *l.CoordenadorID, Nome: *l.CoordenadorNome}
		if l.CoordenadorTambemPI != nil {
			coordenador.TambemPesquisadorInstitucional = *l.CoordenadorTambemPI
		}
		if l.CoordenadorDataFim != nil {
			d, err := valueobject.DataLocalTexto(*l.CoordenadorDataFim)
			if err != nil {
				return port.ItemCurso{}, err
			}
			coordenador.DataFim = &d
		}
		item.Coordenador = coordenador
	}
	return item, nil
}

type CursoRepository struct {
	db *sqlx.DB
}

func NovoCursoRepository(db *sqlx.DB) *CursoRepository {
	return &CursoRepository{db: db}
}

var _ port.CursoRepository = (*CursoRepository)(nil)

func (r *CursoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemCurso, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoCurso, 2)
	if err != nil {
		return port.ItemCurso{}, err
	}
	todosArgs := append([]any{id}, args...)
	lateralSQL, lateralArgs, _ := lateralCoordenadorDoCurso(len(todosArgs) + 1)
	todosArgs = append(todosArgs, lateralArgs...)
	todosArgs = append(todosArgs, escopo.DataDeReferencia().String())

	consulta := `SELECT curso.*, coord.id AS coord_id, coord.nome AS coord_nome,
	                    coord.data_fim AS coord_data_fim, coord.tambem_pi AS coord_tambem_pi,
	                    ` + subConsultaTemVinculo + `
	               FROM curso ` + lateralSQL + `
	              WHERE curso.id = $1 AND ` + clausula

	ex := Executor(ctx, r.db)
	var linha linhaCurso
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, todosArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemCurso{}, domain.ErrNaoEncontrado
		}
		return port.ItemCurso{}, err
	}
	return linha.paraItem()
}

func (r *CursoRepository) Listar(ctx context.Context, escopo autorizacao.Escopo, filtro port.FiltroListarCursos) (port.ResultadoListaCursos, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoCurso, 1)
	if err != nil {
		return port.ResultadoListaCursos{}, err
	}
	condicoes := []string{clausula}
	n := len(args) + 1

	if filtro.Busca != "" {
		condicoes = append(condicoes, fmt.Sprintf("unaccent(lower(curso.nome)) LIKE unaccent(lower($%d))", n))
		args = append(args, "%"+escaparCuringasLike(filtro.Busca)+"%")
		n++
	}
	if filtro.Grau != "" {
		condicoes = append(condicoes, fmt.Sprintf("curso.grau = $%d", n))
		args = append(args, filtro.Grau)
		n++
	}
	if filtro.Modalidade != "" {
		condicoes = append(condicoes, fmt.Sprintf("curso.modalidade = $%d", n))
		args = append(args, filtro.Modalidade)
		n++
	}
	if filtro.Situacao != "" && filtro.Situacao != "todas" {
		condicoes = append(condicoes, fmt.Sprintf("curso.situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	// O lateral entra ANTES do count também: filtro.CoordenadorID e
	// filtro.Vago dependem do resultado do LEFT JOIN LATERAL (quem
	// coordena hoje), e o total precisa contar as mesmas linhas que a
	// página mostra — achado de revisão: contar só sobre "curso" sem o
	// lateral ignorava os dois filtros no total, mesmo filtrando a página.
	lateralSQL, _, proximo := lateralCoordenadorDoCurso(n)
	argsComData := append(append([]any{}, args...), escopo.DataDeReferencia().String())

	whereFinal := where
	if filtro.CoordenadorID != nil {
		whereFinal += fmt.Sprintf(" AND coord.id = $%d", proximo)
		argsComData = append(argsComData, *filtro.CoordenadorID)
	} else if filtro.Vago {
		whereFinal += " AND coord.id IS NULL"
	}

	ex := Executor(ctx, r.db)
	var total int
	totalConsulta := fmt.Sprintf("SELECT count(*) FROM curso %s WHERE %s", lateralSQL, whereFinal)
	if err := sqlx.GetContext(ctx, ex, &total, totalConsulta, argsComData...); err != nil {
		return port.ResultadoListaCursos{}, err
	}

	sortColuna, ok := ordenacaoCurso[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoCurso["nome"]
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

	consulta := fmt.Sprintf(`SELECT curso.*, coord.id AS coord_id, coord.nome AS coord_nome,
	                    coord.data_fim AS coord_data_fim, coord.tambem_pi AS coord_tambem_pi,
	                    `+subConsultaTemVinculo+`
	               FROM curso %s
	              WHERE %s
	              ORDER BY %s %s
	              LIMIT $%d OFFSET $%d`,
		lateralSQL, whereFinal, sortColuna, ordem, len(argsComData)+1, len(argsComData)+2)
	argsPaginados := append(append([]any{}, argsComData...), pageSize, offset)

	var linhas []linhaCurso
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaCursos{}, err
	}
	itens := make([]port.ItemCurso, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return port.ResultadoListaCursos{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaCursos{Itens: itens, Total: total}, nil
}

// ListarMeusCursos — CV-01: cursos ativos com designação vigente do ator.
// O EXISTS de carteira já vem embutido em AplicarEscopo (RestritoACarteiraDe),
// aqui só some situacao='ativo'.
func (r *CursoRepository) ListarMeusCursos(ctx context.Context, escopo autorizacao.Escopo) ([]curso.Curso, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoCurso, 1)
	if err != nil {
		return nil, err
	}
	consulta := `SELECT * FROM curso WHERE ` + clausula + ` AND curso.situacao = 'ativo' ORDER BY curso.nome COLLATE "pt-BR-x-icu"`

	ex := Executor(ctx, r.db)
	var linhas []linhaCurso
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	cursos := make([]curso.Curso, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return nil, err
		}
		cursos = append(cursos, item.Curso)
	}
	return cursos, nil
}

func (r *CursoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	var codigoArg any
	if c.CodigoEMec != nil {
		codigoArg = c.CodigoEMec.String()
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO curso (id, instituicao_id, nome, codigo_emec, grau, modalidade, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.InstituicaoID, c.Nome.String(), codigoArg, string(c.Grau), string(c.Modalidade), string(c.Situacao), c.CriadoEm, c.Versao,
	)
	if err != nil {
		if violaIndice(err, "uq_curso_instituicao_nome") {
			return domain.ErrNomeCursoDuplicado
		}
		if violaIndice(err, "uq_curso_instituicao_emec") {
			return domain.ErrCodigoEMecCursoDuplicado
		}
		return err
	}
	return nil
}

func (r *CursoRepository) Atualizar(ctx context.Context, escopo autorizacao.Escopo, c *curso.Curso, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoCurso, 8)
	if err != nil {
		return err
	}
	var codigoArg any
	if c.CodigoEMec != nil {
		codigoArg = c.CodigoEMec.String()
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{c.Nome.String(), codigoArg, string(c.Grau), string(c.Modalidade), agora, c.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE curso SET nome=$1, codigo_emec=$2, grau=$3, modalidade=$4, atualizado_em=$5, versao=versao+1 "+
			"WHERE id=$6 AND versao=$7 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "uq_curso_instituicao_nome") {
			return domain.ErrNomeCursoDuplicado
		}
		if violaIndice(err, "uq_curso_instituicao_emec") {
			return domain.ErrCodigoEMecCursoDuplicado
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
	c.AtualizadoEm = &agora
	c.Versao = versaoEsperada + 1
	return nil
}

func (r *CursoRepository) AlterarSituacao(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, nova valueobject.SituacaoCurso, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoCurso, 5)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{string(nova), agora, id, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE curso SET situacao=$1, atualizado_em=$2, versao=versao+1 WHERE id=$3 AND versao=$4 AND "+clausula,
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
		return domain.ErrConflitoDeVersao
	}
	return nil
}

// Excluir — 409 CURSO_COM_VINCULO se houver designação ou plano
// (design.md §5.4). A trava é SELECT ... FOR UPDATE na linha do curso,
// dentro da MESMA transação da verificação — fecha a janela entre
// checar e apagar (mesma técnica de design.md §5.8 de autenticação).
// `entrega` ainda não existe (specs/metas-coordenacao não começou);
// quando existir, junta-se a esta mesma verificação.
func (r *CursoRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoCurso, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)

	var trava uuid.UUID
	if err := sqlx.GetContext(ctx, ex, &trava, "SELECT curso.id FROM curso WHERE curso.id = $1 AND "+clausula+" FOR UPDATE", args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNaoEncontrado
		}
		return err
	}

	var temVinculo bool
	if err := sqlx.GetContext(ctx, ex, &temVinculo, `
		SELECT EXISTS (SELECT 1 FROM designacao WHERE curso_id = $1 AND excluido_em IS NULL)
		    OR EXISTS (SELECT 1 FROM plano WHERE curso_id = $1 AND excluido_em IS NULL)`,
		id,
	); err != nil {
		return err
	}
	if temVinculo {
		return domain.ErrCursoComVinculo
	}

	resultado, err := ex.ExecContext(ctx, "UPDATE curso SET excluido_em = now() WHERE id = $1", id)
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

func (r *CursoRepository) TravarSeAtivo(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoCurso, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)

	var trava uuid.UUID
	if err := sqlx.GetContext(ctx, ex, &trava, "SELECT curso.id FROM curso WHERE curso.id = $1 AND "+clausula+" FOR SHARE", args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNaoEncontrado
		}
		return err
	}
	return nil
}
