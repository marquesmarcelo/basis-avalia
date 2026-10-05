package main

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/google/uuid"
)

type periodoSeed struct {
	nome       string
	dataInicio string
	dataFim    string
}

// periodosDeDesenvolvimento — exatamente os três da seção 7 da spec de
// plano-acao. "Hoje" nos exemplos da spec é 15/03/2026: 2025.2 já
// encerrado, 2026.1 aberto, 2026.2 ainda não iniciado.
var periodosDeDesenvolvimento = []periodoSeed{
	{"2025.2", "2025-08-01", "2025-12-20"},
	{"2026.1", "2026-01-01", "2026-07-30"},
	{"2026.2", "2026-08-01", "2026-12-20"},
}

type itemPlanoSeed struct {
	nomeDaMeta string
	quantidade int
}

type planoSeed struct {
	nomeDoCurso    string
	nomeDoPeriodo  string
	situacao       string // "rascunho" | "vigente"
	aprovacaoData  string // "" = sem aprovação
	aprovacaoOrgao string
	itens          []itemPlanoSeed
}

// planosDeDesenvolvimento — exatamente os cinco planos de "Planos da FSA
// em 2026.1" da seção 7 da spec: sem eles, SI-01 (rascunho não cobra),
// SI-05 (sem aprovação aparece), SI-08/CP-08 (curso vago) e IT-01/IT-02 (a
// mesma meta em quantidades diferentes) não têm dado real.
var planosDeDesenvolvimento = []planoSeed{
	{
		"Engenharia de Software", "2026.1", "vigente", "2026-02-10", "nde",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 4},
			{"Relatório de acompanhamento do curso", 1},
			{"Reunião semestral com representantes discentes", 2},
		},
	},
	{
		"Sistemas de Informação", "2026.1", "vigente", "", "",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 2},
			{"Relatório de acompanhamento do curso", 1},
		},
	},
	{
		"Pedagogia", "2026.1", "vigente", "", "",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 2},
			{"Plano de ensino revisado", 1},
		},
	},
	{
		"Biomedicina", "2026.1", "vigente", "2026-02-20", "colegiado_curso",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 2},
		},
	},
	{
		"Análise e Desenvolvimento de Sistemas", "2026.1", "rascunho", "", "",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 4},
		},
	},

	// Os dois planos abaixo ficam em 2026.2 de propósito (achado de
	// revisão): os planos acima estão todos em 2026.1, um período com
	// datas fixas de calendário (jan-jul/2026) que o relógio real já
	// ultrapassa depois de julho — sem um plano vigente em um período que
	// ainda esteja aberto quando alguém rodar o seed mais tarde, o
	// relatório de desempenho fica sem nenhum item "em andamento" ou
	// "cumprida" observável de verdade, só histórico encerrado.
	{
		"Sistemas de Informação", "2026.2", "vigente", "", "",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 3},
		},
	},
	{
		"Análise e Desenvolvimento de Sistemas", "2026.2", "vigente", "", "",
		[]itemPlanoSeed{
			{"Registrar reuniões de NDE em ata", 3},
			{"Relatório de acompanhamento do curso", 2},
		},
	},
}

// seedarPlanosDeAcao — parte 4 do seed de desenvolvimento
// (specs/plano-acao, seção 7). Roda depois de cursos e metas já
// existirem. Idempotente por (curso_id, periodo_id), a mesma chave do
// índice único uq_plano_curso_periodo.
func seedarPlanosDeAcao(db *sqlx.DB, instituicaoFSAID uuid.UUID) {
	idPeriodoPorNome := map[string]uuid.UUID{}
	for _, p := range periodosDeDesenvolvimento {
		id, err := obterOuCriarPeriodo(db, instituicaoFSAID, p)
		if err != nil {
			log.Fatalf("seed: período %s: %v", p.nome, err)
		}
		idPeriodoPorNome[p.nome] = id
	}

	for _, p := range planosDeDesenvolvimento {
		if err := criarPlanoSeNaoExistir(db, instituicaoFSAID, idPeriodoPorNome[p.nomeDoPeriodo], p); err != nil {
			log.Fatalf("seed: plano de %s em %s: %v", p.nomeDoCurso, p.nomeDoPeriodo, err)
		}
	}

	log.Println("seed: planos de ação verificados/aplicados")
}

func obterOuCriarPeriodo(db *sqlx.DB, instituicaoID uuid.UUID, s periodoSeed) (uuid.UUID, error) {
	var idExistente uuid.UUID
	err := db.Get(&idExistente,
		`SELECT id FROM periodo WHERE instituicao_id = $1 AND lower(btrim(nome)) = lower(btrim($2)) AND excluido_em IS NULL`,
		instituicaoID, s.nome)
	if err == nil {
		return idExistente, nil
	}

	id := uuid.Must(uuid.NewV7())
	_, err = db.Exec(
		`INSERT INTO periodo (id, instituicao_id, nome, data_inicio, data_fim) VALUES ($1,$2,$3,$4,$5)`,
		id, instituicaoID, s.nome, s.dataInicio, s.dataFim,
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

func buscarCursoIDPorNome(db *sqlx.DB, instituicaoID uuid.UUID, nome string) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.Get(&id,
		`SELECT id FROM curso WHERE instituicao_id = $1 AND lower(btrim(nome)) = lower(btrim($2)) AND excluido_em IS NULL`,
		instituicaoID, nome)
	return id, err
}

func buscarMetaIDPorNome(db *sqlx.DB, instituicaoID uuid.UUID, nome string) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.Get(&id,
		`SELECT id FROM meta WHERE instituicao_id = $1 AND lower(btrim(nome)) = lower(btrim($2)) AND excluido_em IS NULL`,
		instituicaoID, nome)
	return id, err
}

func criarPlanoSeNaoExistir(db *sqlx.DB, instituicaoID, periodoID uuid.UUID, s planoSeed) error {
	cursoID, err := buscarCursoIDPorNome(db, instituicaoID, s.nomeDoCurso)
	if err != nil {
		return err
	}

	var idExistente uuid.UUID
	err = db.Get(&idExistente, `SELECT id FROM plano WHERE curso_id = $1 AND periodo_id = $2 AND excluido_em IS NULL`, cursoID, periodoID)
	if err == nil {
		return nil
	}

	var aprovacaoData, aprovacaoOrgao any
	if s.aprovacaoData != "" {
		aprovacaoData, aprovacaoOrgao = s.aprovacaoData, s.aprovacaoOrgao
	}

	planoID := uuid.Must(uuid.NewV7())
	_, err = db.Exec(
		`INSERT INTO plano (id, instituicao_id, curso_id, periodo_id, descricao, objetivo_geral, resultados_esperados,
		                     aprovacao_data, aprovacao_orgao, situacao_publicacao, criado_em)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		planoID, instituicaoID, cursoID, periodoID,
		"Plano de ação do curso "+s.nomeDoCurso+" para o período "+s.nomeDoPeriodo+".",
		"Elevar a maturidade de gestão do curso perante o instrumento de avaliação do INEP.",
		"Metas cumpridas e evidências registradas para a próxima avaliação externa.",
		aprovacaoData, aprovacaoOrgao, s.situacao, time.Now(),
	)
	if err != nil {
		return err
	}

	for _, item := range s.itens {
		metaID, err := buscarMetaIDPorNome(db, instituicaoID, item.nomeDaMeta)
		if err != nil {
			return err
		}
		if _, err := db.Exec(
			`INSERT INTO item_plano (id, plano_id, curso_id, instituicao_id, meta_id, quantidade) VALUES ($1,$2,$3,$4,$5,$6)`,
			uuid.Must(uuid.NewV7()), planoID, cursoID, instituicaoID, metaID, item.quantidade,
		); err != nil {
			return err
		}
	}
	return nil
}
