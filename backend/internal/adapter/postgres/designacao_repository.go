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
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ordenacaoDesignacao = map[string]string{
	"data_inicio": "designacao.data_inicio",
	"data_fim":    "designacao.data_fim",
	"coordenador": `coordenador.nome COLLATE "pt-BR-x-icu"`,
	"portaria":    "designacao.portaria",
}

// fragmentoSituacaoDesignacao espelha valueobject.Vigencia.SituacaoEm em
// SQL, para o filtro de "situacao" do grid — a MESMA regra do predicado
// único (FragmentoDesignacaoVigente cobre só "vigente"; aqui as três
// fronteiras, para o WHERE de filtro, não para o campo da resposta, que
// nunca é gravado nem filtrado por si mesmo fora daqui).
func fragmentoSituacaoDesignacao(alias, situacao string, n int) (string, bool) {
	switch situacao {
	case "futura":
		return fmt.Sprintf("%s.data_inicio > $%d", alias, n), true
	case "vigente":
		return FragmentoDesignacaoVigente(alias, n), true
	case "encerrada":
		return fmt.Sprintf("%s.data_fim IS NOT NULL AND %s.data_fim < $%d", alias, alias, n), true
	}
	return "", false
}

type linhaDesignacao struct {
	ID             uuid.UUID  `db:"id"`
	CursoID        uuid.UUID  `db:"curso_id"`
	InstituicaoID  uuid.UUID  `db:"instituicao_id"`
	CoordenadorID  uuid.UUID  `db:"coordenador_id"`
	Portaria       string     `db:"portaria"`
	DataInicio     string     `db:"data_inicio"`
	DataFim        *string    `db:"data_fim"`
	Autodesignacao bool       `db:"autodesignacao"`
	CriadoEm       time.Time  `db:"criado_em"`
	AtualizadoEm   *time.Time `db:"atualizado_em"`
	ExcluidoEm     *time.Time `db:"excluido_em"`
	Versao         int        `db:"versao"`

	CoordenadorNome           string `db:"coordenador_nome"`
	OutrasDesignacoesVigentes int    `db:"outras_vigentes"`
}

func (l linhaDesignacao) paraItem() (port.ItemDesignacao, error) {
	inicio, err := valueobject.DataLocalTexto(l.DataInicio)
	if err != nil {
		return port.ItemDesignacao{}, err
	}
	var fim *valueobject.DataLocal
	if l.DataFim != nil {
		f, err := valueobject.DataLocalTexto(*l.DataFim)
		if err != nil {
			return port.ItemDesignacao{}, err
		}
		fim = &f
	}
	vigencia, err := valueobject.NovaVigencia(inicio, fim)
	if err != nil {
		return port.ItemDesignacao{}, err
	}
	portaria, err := valueobject.NovaPortaria(l.Portaria)
	if err != nil {
		return port.ItemDesignacao{}, err
	}
	return port.ItemDesignacao{
		Designacao: designacao.Designacao{
			ID: l.ID, CursoID: l.CursoID, InstituicaoID: l.InstituicaoID, CoordenadorID: l.CoordenadorID,
			Portaria: portaria, Vigencia: vigencia, Autodesignacao: l.Autodesignacao,
			CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, ExcluidoEm: l.ExcluidoEm, Versao: l.Versao,
		},
		CoordenadorNome:           l.CoordenadorNome,
		OutrasDesignacoesVigentes: l.OutrasDesignacoesVigentes,
	}, nil
}

type DesignacaoRepository struct {
	db *sqlx.DB
}

func NovoDesignacaoRepository(db *sqlx.DB) *DesignacaoRepository {
	return &DesignacaoRepository{db: db}
}

var _ port.DesignacaoRepository = (*DesignacaoRepository)(nil)

// selectDesignacao — SELECT compartilhado por BuscarPorID e ListarDoCurso;
// outras_vigentes é subconsulta correlacionada (nunca junção — juntar
// designacao com ela mesma multiplicaria linhas antes do agrupamento, o
// mesmo defeito de DISTINCT já visto em indicadores/meta).
func selectDesignacao(dataPlaceholder int) string {
	predicado := FragmentoDesignacaoVigente("d2", dataPlaceholder)
	return fmt.Sprintf(`
		SELECT designacao.id, designacao.curso_id, designacao.instituicao_id, designacao.coordenador_id,
		       designacao.portaria, designacao.data_inicio::text AS data_inicio, designacao.data_fim::text AS data_fim,
		       designacao.autodesignacao, designacao.criado_em, designacao.atualizado_em, designacao.excluido_em, designacao.versao,
		       coordenador.nome AS coordenador_nome,
		       (
		         SELECT count(*) FROM designacao d2
		          WHERE d2.coordenador_id = designacao.coordenador_id
		            AND d2.id <> designacao.id AND d2.excluido_em IS NULL
		            AND %s
		       ) AS outras_vigentes
		  FROM designacao
		  JOIN usuario coordenador ON coordenador.id = designacao.coordenador_id`, predicado)
}

func (r *DesignacaoRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) (port.ItemDesignacao, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoDesignacao, 2)
	if err != nil {
		return port.ItemDesignacao{}, err
	}
	whereArgs := append([]any{id}, escopoArgs...)
	// O placeholder da subconsulta de outras_vigentes vem DEPOIS de todos
	// os placeholders do WHERE — nunca um número fixo: um "$2" hardcoded
	// que o WHERE não referencia é exatamente "could not determine data
	// type of parameter" (achado desta rodada, comprovado reproduzindo).
	dataPlaceholder := len(whereArgs) + 1
	todosArgs := append(append([]any{}, whereArgs...), escopo.DataDeReferencia().String())

	consulta := selectDesignacao(dataPlaceholder) + " WHERE designacao.id = $1 AND " + clausula
	ex := Executor(ctx, r.db)
	var linha linhaDesignacao
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, todosArgs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.ItemDesignacao{}, domain.ErrNaoEncontrado
		}
		return port.ItemDesignacao{}, err
	}
	return linha.paraItem()
}

func (r *DesignacaoRepository) ListarDoCurso(ctx context.Context, escopo autorizacao.Escopo, cursoID uuid.UUID, filtro port.FiltroListarDesignacoes) (port.ResultadoListaDesignacoes, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoDesignacao, 2)
	if err != nil {
		return port.ResultadoListaDesignacoes{}, err
	}
	condicoes := []string{clausula, "designacao.curso_id = $1"}
	whereArgs := append([]any{cursoID}, escopoArgs...)
	n := len(whereArgs) + 1

	if filtro.Situacao != "" && filtro.Situacao != "todas" {
		fragmento, ok := fragmentoSituacaoDesignacao("designacao", filtro.Situacao, n)
		if !ok {
			return port.ResultadoListaDesignacoes{}, domain.ErrValorInvalido
		}
		condicoes = append(condicoes, fragmento)
		whereArgs = append(whereArgs, escopo.DataDeReferencia().String())
		n++
	}
	if filtro.CoordenadorID != nil {
		condicoes = append(condicoes, fmt.Sprintf("designacao.coordenador_id = $%d", n))
		whereArgs = append(whereArgs, *filtro.CoordenadorID)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM designacao WHERE "+where, whereArgs...); err != nil {
		return port.ResultadoListaDesignacoes{}, err
	}

	sortColuna, ok := ordenacaoDesignacao[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoDesignacao["data_inicio"]
	}
	ordem := "DESC"
	if filtro.Order == "asc" {
		ordem = "ASC"
	}
	page, pageSize := filtro.Page, filtro.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Placeholder da subconsulta de outras_vigentes: logo depois de todos
	// os placeholders já usados pelo WHERE (n), nunca um número fixo.
	consulta := fmt.Sprintf("%s WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		selectDesignacao(n), where, sortColuna, ordem, n+1, n+2)
	argsSelect := append(append([]any{}, whereArgs...), escopo.DataDeReferencia().String())
	argsPaginados := append(argsSelect, pageSize, offset)

	var linhas []linhaDesignacao
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, argsPaginados...); err != nil {
		return port.ResultadoListaDesignacoes{}, err
	}
	itens := make([]port.ItemDesignacao, 0, len(linhas))
	for _, l := range linhas {
		item, err := l.paraItem()
		if err != nil {
			return port.ResultadoListaDesignacoes{}, err
		}
		itens = append(itens, item)
	}
	return port.ResultadoListaDesignacoes{Itens: itens, Total: total}, nil
}

// ListarCandidatos — CP-07 (design.md §5.1): possui QUALQUER perfil além
// de aluno, na instituição do escopo. O Administrador do Sistema já sai
// pelo isolamento de instituição — não precisa de cláusula própria.
func (r *DesignacaoRepository) ListarCandidatos(ctx context.Context, escopo autorizacao.Escopo, busca string) ([]port.ItemCandidato, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		return nil, err
	}
	condicoes := []string{clausula,
		"EXISTS (SELECT 1 FROM usuario_perfil up WHERE up.usuario_id = usuario.id AND up.perfil <> 'aluno')"}
	n := len(args) + 1
	if busca != "" {
		condicoes = append(condicoes, fmt.Sprintf(
			"(unaccent(lower(usuario.nome)) LIKE unaccent(lower($%d)) OR unaccent(lower(usuario.email)) LIKE unaccent(lower($%d)))", n, n))
		args = append(args, "%"+escaparCuringasLike(busca)+"%")
		n++
	}
	where := strings.Join(condicoes, " AND ")

	consulta := `
		SELECT usuario.id, usuario.nome, usuario.email,
		       COALESCE(string_agg(up.perfil, ',' ORDER BY up.perfil), '') AS perfis
		  FROM usuario
		  LEFT JOIN usuario_perfil up ON up.usuario_id = usuario.id
		 WHERE ` + where + `
		 GROUP BY usuario.id
		 ORDER BY usuario.nome COLLATE "pt-BR-x-icu"
		 LIMIT 20`

	ex := Executor(ctx, r.db)
	type linhaCandidato struct {
		ID     uuid.UUID `db:"id"`
		Nome   string    `db:"nome"`
		Email  string    `db:"email"`
		Perfis string    `db:"perfis"`
	}
	var linhas []linhaCandidato
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	itens := make([]port.ItemCandidato, 0, len(linhas))
	for _, l := range linhas {
		var perfis []string
		if l.Perfis != "" {
			perfis = strings.Split(l.Perfis, ",")
		}
		itens = append(itens, port.ItemCandidato{ID: l.ID, Nome: l.Nome, Email: l.Email, Perfis: perfis})
	}
	return itens, nil
}

func (r *DesignacaoRepository) Inserir(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	var fimArg any
	if d.Vigencia.Fim() != nil {
		fimArg = d.Vigencia.Fim().String()
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO designacao (id, curso_id, instituicao_id, coordenador_id, portaria, data_inicio, data_fim, autodesignacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.CursoID, d.InstituicaoID, d.CoordenadorID, d.Portaria.String(),
		d.Vigencia.Inicio().String(), fimArg, d.Autodesignacao, d.CriadoEm, d.Versao,
	)
	if err != nil {
		if violaIndice(err, "ex_designacao_sem_sobreposicao") {
			return domain.ErrDesignacaoSobreposta
		}
		return err
	}
	return nil
}

func (r *DesignacaoRepository) AtualizarCompleta(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoDesignacao, 8)
	if err != nil {
		return err
	}
	var fimArg any
	if d.Vigencia.Fim() != nil {
		fimArg = d.Vigencia.Fim().String()
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{
		d.CoordenadorID, d.Portaria.String(), d.Vigencia.Inicio().String(), fimArg, agora, d.ID, versaoEsperada,
	}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE designacao SET coordenador_id=$1, portaria=$2, data_inicio=$3, data_fim=$4, atualizado_em=$5, versao=versao+1 "+
			"WHERE id=$6 AND versao=$7 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "ex_designacao_sem_sobreposicao") {
			return domain.ErrDesignacaoSobreposta
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
	d.AtualizadoEm = &agora
	d.Versao = versaoEsperada + 1
	return nil
}

func (r *DesignacaoRepository) AtualizarFimEPortaria(ctx context.Context, escopo autorizacao.Escopo, d *designacao.Designacao, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoDesignacao, 6)
	if err != nil {
		return err
	}
	var fimArg any
	if d.Vigencia.Fim() != nil {
		fimArg = d.Vigencia.Fim().String()
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{d.Portaria.String(), fimArg, agora, d.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE designacao SET portaria=$1, data_fim=$2, atualizado_em=$3, versao=versao+1 "+
			"WHERE id=$4 AND versao=$5 AND "+clausula,
		args...,
	)
	if err != nil {
		if violaIndice(err, "ex_designacao_sem_sobreposicao") {
			return domain.ErrDesignacaoSobreposta
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
	d.AtualizadoEm = &agora
	d.Versao = versaoEsperada + 1
	return nil
}

func (r *DesignacaoRepository) Excluir(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoDesignacao, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE designacao SET excluido_em = now() WHERE id = $1 AND "+clausula,
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
