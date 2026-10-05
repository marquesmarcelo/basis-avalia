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
	"github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EntregaRepository struct{ db *sqlx.DB }

func NovoEntregaRepository(db *sqlx.DB) *EntregaRepository { return &EntregaRepository{db: db} }

var _ port.EntregaRepository = (*EntregaRepository)(nil)

// --- Leitura de uma entrega e seus anexos ----------------------------------

type linhaEntregaDB struct {
	ID                      uuid.UUID  `db:"id"`
	ItemPlanoID             uuid.UUID  `db:"item_plano_id"`
	CursoID                 uuid.UUID  `db:"curso_id"`
	InstituicaoID           uuid.UUID  `db:"instituicao_id"`
	EnviadaPor              uuid.UUID  `db:"enviada_por"`
	CorrigidaPor            *uuid.UUID `db:"corrigida_por"`
	Observacao              string     `db:"observacao"`
	Situacao                string     `db:"situacao"`
	Rodadas                 int        `db:"rodadas_de_recusa"`
	PrazoCorrecao           *time.Time `db:"prazo_correcao_ate"`
	AvaliadaPor             *uuid.UUID `db:"avaliada_por"`
	AvaliadaEm              *time.Time `db:"avaliada_em"`
	Motivo                  string     `db:"motivo"`
	AvaliadorEraCoordenador bool       `db:"avaliador_era_coordenador"`
	PendenciaVistaEm        *time.Time `db:"pendencia_vista_em"`

	NotificacaoEvento     *string    `db:"notificacao_evento"`
	NotificacaoGeradaEm   *time.Time `db:"notificacao_gerada_em"`
	NotificacaoEnviadaEm  *time.Time `db:"notificacao_enviada_em"`
	NotificacaoTentativas int        `db:"notificacao_tentativas"`
	NotificacaoUltimoErro *string    `db:"notificacao_ultimo_erro"`
	// NotificacaoReivindicadaEm — T-127 (design.md §8.4): nunca lido por
	// esta struct, mas `entrega.*` traz a coluna, e sqlx.StructScan falha
	// sem um campo correspondente (destino não encontrado).
	NotificacaoReivindicadaEm *time.Time `db:"notificacao_reivindicada_em"`

	CriadoEm     time.Time  `db:"criado_em"`
	AtualizadoEm *time.Time `db:"atualizado_em"`
	ExcluidoEm   *time.Time `db:"excluido_em"`
	Versao       int        `db:"versao"`

	MetaID           uuid.UUID `db:"meta_id"`
	MetaNome         string    `db:"meta_nome"`
	CursoNome        string    `db:"curso_nome"`
	Quantidade       int       `db:"quantidade"`
	EnviadaPorNome   string    `db:"enviada_por_nome"`
	CorrigidaPorNome *string   `db:"corrigida_por_nome"`
	AvaliadaPorNome  *string   `db:"avaliada_por_nome"`
}

func (l linhaEntregaDB) paraLinha() port.LinhaEntrega {
	return port.LinhaEntrega{
		ID: l.ID, ItemPlanoID: l.ItemPlanoID, CursoID: l.CursoID, CursoNome: l.CursoNome,
		MetaID: l.MetaID, MetaNome: l.MetaNome, Situacao: l.Situacao, Rodadas: l.Rodadas,
		PrazoCorrecao: l.PrazoCorrecao, EnviadaPorID: l.EnviadaPor, EnviadaPorNome: l.EnviadaPorNome,
		CorrigidaPorNome: l.CorrigidaPorNome, Observacao: l.Observacao, Motivo: l.Motivo,
		AvaliadaPorNome: l.AvaliadaPorNome, AvaliadaEm: l.AvaliadaEm, AvaliadorEraCoordenador: l.AvaliadorEraCoordenador,
		PendenciaVistaEm: l.PendenciaVistaEm, CriadoEm: l.CriadoEm, AtualizadoEm: l.AtualizadoEm, Versao: l.Versao,
	}
}

const colunasEntregaJunta = `
	entrega.*, item_plano.meta_id, item_plano.quantidade, meta.nome AS meta_nome, curso.nome AS curso_nome,
	enviou.nome AS enviada_por_nome, corrigiu.nome AS corrigida_por_nome, avaliou.nome AS avaliada_por_nome
	FROM entrega
	JOIN item_plano ON item_plano.id = entrega.item_plano_id
	JOIN meta ON meta.id = item_plano.meta_id
	JOIN curso ON curso.id = entrega.curso_id
	JOIN usuario enviou ON enviou.id = entrega.enviada_por
	LEFT JOIN usuario corrigiu ON corrigiu.id = entrega.corrigida_por
	LEFT JOIN usuario avaliou ON avaliou.id = entrega.avaliada_por
`

func (r *EntregaRepository) listarAnexos(ctx context.Context, entregaID uuid.UUID) ([]port.AnexoResponse, error) {
	ex := Executor(ctx, r.db)
	var linhas []struct {
		ID           uuid.UUID `db:"id"`
		NomeOriginal string    `db:"nome_original"`
		Tipo         string    `db:"tipo"`
		TamanhoBytes int64     `db:"tamanho_bytes"`
		HashSHA256   string    `db:"hash_sha256"`
		CriadoEm     time.Time `db:"criado_em"`
	}
	consulta := `SELECT id, nome_original, tipo, tamanho_bytes, hash_sha256, criado_em
	               FROM anexo WHERE entrega_id = $1 AND excluido_em IS NULL ORDER BY criado_em`
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, entregaID); err != nil {
		return nil, err
	}
	anexos := make([]port.AnexoResponse, 0, len(linhas))
	for _, l := range linhas {
		anexos = append(anexos, port.AnexoResponse{
			ID: l.ID, NomeOriginal: l.NomeOriginal, Tipo: l.Tipo, TamanhoBytes: l.TamanhoBytes,
			HashSHA256: l.HashSHA256, CriadoEm: l.CriadoEm,
		})
	}
	return anexos, nil
}

// indicadoresDaMeta usa SemCarteira porque indicador (AlvoIndicador) não
// tem coluna de curso — a carteira já foi satisfeita um nível acima, pelo
// Alvo que a possui (fundacao-metas.md §3.4, comentário de Escopo.SemCarteira).
func (r *EntregaRepository) indicadoresDaMeta(ctx context.Context, escopo autorizacao.Escopo, metaID uuid.UUID) ([]port.IndicadorDaMeta, error) {
	clausula, args, err := AplicarEscopo(escopo.SemCarteira(), AlvoIndicador, 2)
	if err != nil {
		return nil, err
	}
	args = append([]any{metaID}, args...)
	ex := Executor(ctx, r.db)
	var linhas []struct {
		ID                    uuid.UUID `db:"id"`
		Codigo                string    `db:"codigo"`
		Nome                  string    `db:"nome"`
		Escopo                string    `db:"escopo"`
		ReferenciaInstrumento *string   `db:"referencia_instrumento"`
		Situacao              string    `db:"situacao"`
	}
	consulta := `SELECT indicador.id, indicador.codigo, indicador.nome, indicador.escopo,
	                    indicador.referencia_instrumento, indicador.situacao
	               FROM meta_indicador
	               JOIN indicador ON indicador.id = meta_indicador.indicador_id
	              WHERE meta_indicador.meta_id = $1 AND ` + clausula + ` ORDER BY indicador.codigo`
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return nil, err
	}
	indicadores := make([]port.IndicadorDaMeta, 0, len(linhas))
	for _, l := range linhas {
		ref := ""
		if l.ReferenciaInstrumento != nil {
			ref = *l.ReferenciaInstrumento
		}
		indicadores = append(indicadores, port.IndicadorDaMeta{
			ID: l.ID, Codigo: l.Codigo, Nome: l.Nome, Escopo: l.Escopo,
			ReferenciaInstrumento: ref, Situacao: l.Situacao,
		})
	}
	return indicadores, nil
}

func (r *EntregaRepository) CoordenaCursoHoje(ctx context.Context, proprio autorizacao.Proprio, cursoID uuid.UUID, dataDeReferencia valueobject.DataLocal) (bool, error) {
	ex := Executor(ctx, r.db)
	consulta := fmt.Sprintf(
		`SELECT EXISTS (SELECT 1 FROM designacao d
		   WHERE d.curso_id = $1 AND d.coordenador_id = $2 AND d.excluido_em IS NULL AND %s)`,
		FragmentoDesignacaoVigente("d", 3))
	var existe bool
	if err := sqlx.GetContext(ctx, ex, &existe, consulta, cursoID, proprio.UsuarioID(), dataDeReferencia.String()); err != nil {
		return false, err
	}
	return existe, nil
}

func (r *EntregaRepository) BuscarPorID(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID, avaliador autorizacao.Proprio) (port.DetalheEntrega, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoEntrega, 2)
	if err != nil {
		return port.DetalheEntrega{}, err
	}
	args = append([]any{id}, args...)
	ex := Executor(ctx, r.db)
	var linha linhaEntregaDB
	consulta := "SELECT " + colunasEntregaJunta + " WHERE entrega.id = $1 AND " + clausula
	if err := sqlx.GetContext(ctx, ex, &linha, consulta, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return port.DetalheEntrega{}, domain.ErrNaoEncontrado
		}
		return port.DetalheEntrega{}, err
	}

	anexos, err := r.listarAnexos(ctx, id)
	if err != nil {
		return port.DetalheEntrega{}, err
	}
	indicadores, err := r.indicadoresDaMeta(ctx, escopo, linha.MetaID)
	if err != nil {
		return port.DetalheEntrega{}, err
	}
	coordenado, err := r.CoordenaCursoHoje(ctx, avaliador, linha.CursoID, escopo.DataDeReferencia())
	if err != nil {
		return port.DetalheEntrega{}, err
	}

	base := linha.paraLinha()
	base.Anexos = anexos
	base.CoordenadoPeloAvaliador = coordenado
	return port.DetalheEntrega{LinhaEntrega: base, Quantidade: linha.Quantidade, Indicadores: indicadores}, nil
}

var ordenacaoEntrega = map[string]string{
	"criado_em": "entrega.criado_em",
	"curso":     `curso.nome COLLATE "pt-BR-x-icu"`,
	"meta":      `meta.nome COLLATE "pt-BR-x-icu"`,
}

func (r *EntregaRepository) ListarDoItem(ctx context.Context, escopo autorizacao.Escopo, itemPlanoID uuid.UUID, filtro port.FiltroListarEntregas) (port.ResultadoListaEntregas, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 2)
	if err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	condicoes := []string{"entrega.item_plano_id = $1", clausula}
	args := append([]any{itemPlanoID}, escopoArgs...)
	n := len(args) + 1
	if filtro.Situacao != "" {
		condicoes = append(condicoes, fmt.Sprintf("entrega.situacao = $%d", n))
		args = append(args, filtro.Situacao)
		n++
	}
	where := strings.Join(condicoes, " AND ")

	ex := Executor(ctx, r.db)
	var total int
	if err := sqlx.GetContext(ctx, ex, &total, "SELECT count(*) FROM entrega JOIN item_plano ON item_plano.id = entrega.item_plano_id JOIN meta ON meta.id = item_plano.meta_id JOIN curso ON curso.id = entrega.curso_id JOIN usuario enviou ON enviou.id = entrega.enviada_por WHERE "+where, args...); err != nil {
		return port.ResultadoListaEntregas{}, err
	}

	sortColuna, ok := ordenacaoEntrega[filtro.Sort]
	if !ok {
		sortColuna = ordenacaoEntrega["criado_em"]
	}
	ordem := "ASC"
	if filtro.Order == "desc" {
		ordem = "DESC"
	}
	page, pageSize := paginaEPageSize(filtro.Page, filtro.PageSize)
	offset := (page - 1) * pageSize

	consulta := fmt.Sprintf("SELECT %s WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		colunasEntregaJunta, where, sortColuna, ordem, n, n+1)
	args = append(args, pageSize, offset)

	var linhas []linhaEntregaDB
	if err := sqlx.SelectContext(ctx, ex, &linhas, consulta, args...); err != nil {
		return port.ResultadoListaEntregas{}, err
	}
	itens := make([]port.LinhaEntrega, 0, len(linhas))
	for _, l := range linhas {
		item := l.paraLinha()
		anexos, err := r.listarAnexos(ctx, l.ID)
		if err != nil {
			return port.ResultadoListaEntregas{}, err
		}
		item.Anexos = anexos
		itens = append(itens, item)
	}
	return port.ResultadoListaEntregas{Itens: itens, Total: total}, nil
}

func paginaEPageSize(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

// --- Idempotência e registro -----------------------------------------------

func (r *EntregaRepository) InserirIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string, recursoID uuid.UUID) error {
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO idempotencia (usuario_id, rota, chave, recurso_id) VALUES ($1,$2,$3,$4)`,
		proprio.UsuarioID(), rota, chave, recursoID)
	if err != nil {
		if violaIndice(err, "idempotencia_pkey") {
			return domain.ErrChaveDuplicada
		}
		return err
	}
	return nil
}

func (r *EntregaRepository) BuscarEntregaPorChaveIdempotencia(ctx context.Context, proprio autorizacao.Proprio, rota, chave string) (uuid.UUID, error) {
	var recursoID uuid.UUID
	err := r.db.GetContext(ctx, &recursoID,
		`SELECT recurso_id FROM idempotencia WHERE usuario_id = $1 AND rota = $2 AND chave = $3`,
		proprio.UsuarioID(), rota, chave)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, domain.ErrNaoEncontrado
		}
		return uuid.UUID{}, err
	}
	return recursoID, nil
}

func (r *EntregaRepository) InserirEntregaComAnexos(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, anexos []port.AnexoParaGravar) error {
	if escopo.Plataforma() {
		return domain.ErrEscopoInvalido
	}
	ex := Executor(ctx, r.db)
	_, err := ex.ExecContext(ctx,
		`INSERT INTO entrega (id, item_plano_id, curso_id, instituicao_id, enviada_por, observacao, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.ItemPlanoID, e.CursoID, e.InstituicaoID, e.EnviadaPor, e.Observacao, string(e.Situacao), e.CriadoEm, e.Versao)
	if err != nil {
		return err
	}
	return r.inserirLinhasDeAnexo(ctx, e.ID, e.CursoID, e.InstituicaoID, anexos)
}

func (r *EntregaRepository) inserirLinhasDeAnexo(ctx context.Context, entregaID, cursoID, instituicaoID uuid.UUID, anexos []port.AnexoParaGravar) error {
	if len(anexos) == 0 {
		return nil
	}
	ex := Executor(ctx, r.db)
	consulta := "INSERT INTO anexo (id, entrega_id, curso_id, instituicao_id, nome_original, tipo, tamanho_bytes, chave_objeto, hash_sha256) VALUES "
	linhas := make([]string, 0, len(anexos))
	args := make([]any, 0, len(anexos)*9)
	n := 1
	for _, a := range anexos {
		linhas = append(linhas, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", n, n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8))
		args = append(args, uuid.Must(uuid.NewV7()), entregaID, cursoID, instituicaoID, a.NomeOriginal, a.Tipo, a.TamanhoBytes, a.ChaveObjeto, a.HashSHA256)
		n += 9
	}
	_, err := ex.ExecContext(ctx, consulta+strings.Join(linhas, ","), args...)
	return err
}

func (r *EntregaRepository) AdicionarAnexos(ctx context.Context, escopo autorizacao.Escopo, entregaID uuid.UUID, anexos []port.AnexoParaGravar) ([]port.AnexoResponse, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoEntrega, 2)
	if err != nil {
		return nil, err
	}
	args = append([]any{entregaID}, args...)
	ex := Executor(ctx, r.db)
	var linha struct {
		CursoID       uuid.UUID `db:"curso_id"`
		InstituicaoID uuid.UUID `db:"instituicao_id"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha,
		"SELECT entrega.curso_id, entrega.instituicao_id FROM entrega WHERE entrega.id = $1 AND "+clausula, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNaoEncontrado
		}
		return nil, err
	}
	if err := r.inserirLinhasDeAnexo(ctx, entregaID, linha.CursoID, linha.InstituicaoID, anexos); err != nil {
		return nil, err
	}
	return r.listarAnexos(ctx, entregaID)
}

func (r *EntregaRepository) RemoverAnexo(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID, autor autorizacao.Proprio) (string, error) {
	clausulaAnexo, args, err := AplicarEscopo(escopo, AlvoAnexo, 2)
	if err != nil {
		return "", err
	}
	args = append([]any{anexoID}, args...)
	ex := Executor(ctx, r.db)
	var linha struct {
		ChaveObjeto string    `db:"chave_objeto"`
		EntregaID   uuid.UUID `db:"entrega_id"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha,
		"SELECT anexo.chave_objeto, anexo.entrega_id FROM anexo WHERE anexo.id = $1 AND "+clausulaAnexo, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNaoEncontrado
		}
		return "", err
	}
	var enviadaPor uuid.UUID
	if err := sqlx.GetContext(ctx, ex, &enviadaPor, "SELECT enviada_por FROM entrega WHERE id = $1", linha.EntregaID); err != nil {
		return "", err
	}
	if enviadaPor != autor.UsuarioID() {
		return "", domain.ErrExclusaoDeEntregaAlheia
	}
	if _, err := ex.ExecContext(ctx, "UPDATE anexo SET excluido_em = now() WHERE id = $1", anexoID); err != nil {
		return "", err
	}
	return linha.ChaveObjeto, nil
}

func (r *EntregaRepository) BuscarAnexoParaDownload(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID) (string, string, string, error) {
	clausula, args, err := AplicarEscopo(escopo, AlvoAnexo, 2)
	if err != nil {
		return "", "", "", err
	}
	args = append([]any{anexoID}, args...)
	ex := Executor(ctx, r.db)
	var linha struct {
		NomeOriginal string `db:"nome_original"`
		Tipo         string `db:"tipo"`
		ChaveObjeto  string `db:"chave_objeto"`
	}
	if err := sqlx.GetContext(ctx, ex, &linha,
		"SELECT anexo.nome_original, anexo.tipo, anexo.chave_objeto FROM anexo WHERE anexo.id = $1 AND "+clausula, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", domain.ErrNaoEncontrado
		}
		return "", "", "", err
	}
	return linha.NomeOriginal, linha.Tipo, linha.ChaveObjeto, nil
}

// --- Correção e exclusão ----------------------------------------------------

func (r *EntregaRepository) AtualizarCorrecao(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 6)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	args := append([]any{e.Observacao, string(e.Situacao), e.CorrigidaPor, agora, e.ID, versaoEsperada}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		`UPDATE entrega SET observacao=$1, situacao=$2, corrigida_por=$3, atualizado_em=$4, versao=versao+1
		  WHERE id=$5 AND versao=$6 AND `+clausula, args...)
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
	e.AtualizadoEm = &agora
	e.Versao = versaoEsperada + 1
	return nil
}

func (r *EntregaRepository) ExcluirLogicamente(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx, "UPDATE entrega SET excluido_em = now() WHERE id = $1 AND "+clausula, args...)
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

func (r *EntregaRepository) MarcarPendenciaVista(ctx context.Context, escopo autorizacao.Escopo, id uuid.UUID) error {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 2)
	if err != nil {
		return err
	}
	ex := Executor(ctx, r.db)
	args := append([]any{id}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		"UPDATE entrega SET pendencia_vista_em = now() WHERE id = $1 AND pendencia_vista_em IS NULL AND "+clausula, args...)
	if err != nil {
		return err
	}
	if _, err := resultado.RowsAffected(); err != nil {
		return err
	}
	return nil
}

// --- Avaliação ---------------------------------------------------------

// notificacaoEventoDe traduz a situação pós-mutação da entrega no evento
// de notificação a gravar (design.md §8.2) — nil para aceita, porque
// aceitação nunca notifica (NT-05).
func notificacaoEventoDe(e *entrega.Entrega, desfazimento bool) any {
	if e.Situacao != valueobject.Recusada {
		return nil
	}
	if desfazimento {
		return "desfazimento"
	}
	return "recusa"
}

// avaliarOuDesfazer é o núcleo comum de Avaliar e DesfazerAceitacao: um
// UPDATE condicional em id + versão + situação-de-partida
// (design.md §5.2, §7.1 de cursos T-273) — zero linhas afetadas é o sinal
// para o chamador reler e escolher entre CONFLITO_DE_VERSAO e
// ENTREGA_JA_AVALIADA/ENTREGA_NAO_ESTA_ACEITA, nunca adivinhado aqui.
func (r *EntregaRepository) avaliarOuDesfazer(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int, situacaoDePartida string, desfazimento bool) (bool, error) {
	clausula, escopoArgs, err := AplicarEscopo(escopo, AlvoEntrega, 12)
	if err != nil {
		return false, err
	}
	ex := Executor(ctx, r.db)
	agora := time.Now()
	notificacaoEvento := notificacaoEventoDe(e, desfazimento)
	var notificacaoGeradaEm any
	if notificacaoEvento != nil {
		notificacaoGeradaEm = agora
	}
	args := append([]any{
		string(e.Situacao), e.Rodadas.Int(), e.PrazoCorrecao, e.AvaliadaPor, e.AvaliadaEm, e.Motivo,
		e.AvaliadorEraCoordenador, notificacaoEvento, notificacaoGeradaEm, agora, e.ID, versaoEsperada,
	}, escopoArgs...)
	resultado, err := ex.ExecContext(ctx,
		`UPDATE entrega SET situacao=$1, rodadas_de_recusa=$2, prazo_correcao_ate=$3, avaliada_por=$4,
		        avaliada_em=$5, motivo=$6, avaliador_era_coordenador=$7,
		        notificacao_evento=$8, notificacao_gerada_em=$9,
		        notificacao_enviada_em=NULL, notificacao_tentativas=0, notificacao_ultimo_erro=NULL,
		        atualizado_em=$10, versao=versao+1
		  WHERE id=$11 AND versao=$12 AND situacao='`+situacaoDePartida+`' AND `+clausula,
		args...)
	if err != nil {
		return false, err
	}
	linhas, err := resultado.RowsAffected()
	if err != nil {
		return false, err
	}
	if linhas == 0 {
		return false, nil
	}
	e.AtualizadoEm = &agora
	e.Versao = versaoEsperada + 1
	return true, nil
}

func (r *EntregaRepository) Avaliar(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) (bool, error) {
	return r.avaliarOuDesfazer(ctx, escopo, e, versaoEsperada, "pendente_avaliacao", false)
}

func (r *EntregaRepository) DesfazerAceitacao(ctx context.Context, escopo autorizacao.Escopo, e *entrega.Entrega, versaoEsperada int) (bool, error) {
	return r.avaliarOuDesfazer(ctx, escopo, e, versaoEsperada, "aceita", true)
}
