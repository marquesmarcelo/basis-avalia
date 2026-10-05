package main

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/basis-avalia/backend/internal/domain/indicador"
	"github.com/basis-avalia/backend/internal/domain/meta"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type indicadorPlataformaSeed struct {
	codigo     string
	nome       string
	referencia string
	situacao   valueobject.SituacaoCatalogo
}

// catalogoDoINEP — exatamente o da seção 7 da spec de indicadores. "3.2"
// nasce inativo: sem ele, IE-06 não é observável.
var catalogoDoINEP = []indicadorPlataformaSeed{
	{"1.4", "Núcleo Docente Estruturante", "Instrumento de Avaliação de Cursos de Graduação 2017 — Dimensão 1, indicador 1.4", valueobject.CatalogoAtivo},
	{"1.5", "Coordenação de curso", "Instrumento de Avaliação de Cursos de Graduação 2017 — Dimensão 1, indicador 1.5", valueobject.CatalogoAtivo},
	{"2.1", "Núcleo de apoio ao discente", "Instrumento de Avaliação de Cursos de Graduação 2017 — Dimensão 2, indicador 2.1", valueobject.CatalogoAtivo},
	{"3.2", "Acompanhamento de egressos", "Instrumento de Avaliação de Cursos de Graduação 2017 — Dimensão 3, indicador 3.2", valueobject.CatalogoInativo},
}

type indicadorInstituicaoSeed struct {
	siglaInstituicao string
	codigo           string
	nome             string
	situacao         valueobject.SituacaoCatalogo
}

// indicadoresProprios — FSA tem um "1.4" próprio, código repetido do
// catálogo comum em escopo diferente: sem ele, IE-03 não é observável.
// "GEST-02" nasce inativo: sem ele, IN-04/MC-07 não são observáveis.
var indicadoresProprios = []indicadorInstituicaoSeed{
	{"FSA", "GEST-01", "Reuniões com representação discente", valueobject.CatalogoAtivo},
	{"FSA", "GEST-02", "Painel interno de acompanhamento", valueobject.CatalogoInativo},
	{"FSA", "1.4", "Comissão própria de NDE", valueobject.CatalogoAtivo},
	{"IVV", "GEST-01", "Reuniões com representação discente", valueobject.CatalogoAtivo},
}

type metaSeed struct {
	siglaInstituicao string
	nome             string
	// codigosDoINEP e codigosProprios: pares (código, siglaInstituicao do
	// indicador próprio) — vazio quando não aplicável.
	codigosDoINEP   []string
	codigosProprios []string
}

// metasDeDesenvolvimento — exatamente as cinco da seção 7 da spec. A
// primeira é o caso do dono (MC-02, MC-05, MC-14): sem ela, nenhum
// cenário de "uma entrega atende dois indicadores" tem dado real.
var metasDeDesenvolvimento = []metaSeed{
	{"FSA", "Registrar reuniões de NDE em ata", []string{"1.4", "1.5"}, nil},
	{"FSA", "Relatório de acompanhamento do curso", []string{"1.5"}, nil},
	{"FSA", "Reunião semestral com representantes discentes", []string{"1.4"}, []string{"GEST-01"}},
	{"FSA", "Plano de ensino revisado", []string{"2.1"}, nil},
	{"FSA", "Painel de indicadores do curso", nil, []string{"GEST-02"}},
}

// seedarIndicadoresEMetas — parte 3 do seed de desenvolvimento
// (specs/indicadores, DI-6). Só roda com APP_ENV=development, depois de
// instituições e pessoas já existirem. Idempotente por (escopo, código)
// para indicador e por (instituição, nome normalizado) para meta.
func seedarIndicadoresEMetas(db *sqlx.DB, idPorSigla map[string]uuid.UUID) {
	idIndicadorINEP := map[string]uuid.UUID{}
	for _, s := range catalogoDoINEP {
		id, err := obterOuCriarIndicadorPlataforma(db, s)
		if err != nil {
			log.Fatalf("seed: indicador do INEP %s: %v", s.codigo, err)
		}
		idIndicadorINEP[s.codigo] = id
	}

	idIndicadorProprio := map[string]uuid.UUID{} // chave: sigla+"/"+codigo
	for _, s := range indicadoresProprios {
		instituicaoID, ok := idPorSigla[s.siglaInstituicao]
		if !ok {
			log.Fatalf("seed: indicador próprio %s: instituição %s não existe", s.codigo, s.siglaInstituicao)
		}
		id, err := obterOuCriarIndicadorProprio(db, instituicaoID, s)
		if err != nil {
			log.Fatalf("seed: indicador próprio %s/%s: %v", s.siglaInstituicao, s.codigo, err)
		}
		idIndicadorProprio[s.siglaInstituicao+"/"+s.codigo] = id
	}

	for _, s := range metasDeDesenvolvimento {
		instituicaoID, ok := idPorSigla[s.siglaInstituicao]
		if !ok {
			log.Fatalf("seed: meta %s: instituição %s não existe", s.nome, s.siglaInstituicao)
		}
		var indicadores []uuid.UUID
		for _, codigo := range s.codigosDoINEP {
			id, ok := idIndicadorINEP[codigo]
			if !ok {
				log.Fatalf("seed: meta %s: indicador do INEP %s não existe", s.nome, codigo)
			}
			indicadores = append(indicadores, id)
		}
		for _, codigo := range s.codigosProprios {
			id, ok := idIndicadorProprio[s.siglaInstituicao+"/"+codigo]
			if !ok {
				log.Fatalf("seed: meta %s: indicador próprio %s não existe", s.nome, codigo)
			}
			indicadores = append(indicadores, id)
		}
		if err := obterOuCriarMeta(db, instituicaoID, s.nome, indicadores); err != nil {
			log.Fatalf("seed: meta %s: %v", s.nome, err)
		}
	}

	log.Println("seed: catálogo de indicadores e metas verificado/aplicado")
}

func obterOuCriarIndicadorPlataforma(db *sqlx.DB, s indicadorPlataformaSeed) (uuid.UUID, error) {
	var idExistente uuid.UUID
	err := db.Get(&idExistente, `SELECT id FROM indicador WHERE instituicao_id IS NULL AND codigo = $1 AND excluido_em IS NULL`, s.codigo)
	if err == nil {
		return idExistente, nil
	}

	novo, err := indicador.NovoDaPlataforma(s.codigo, s.nome, "", s.referencia)
	if err != nil {
		return uuid.UUID{}, err
	}
	novo.Situacao = s.situacao
	_, err = db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, descricao, referencia_instrumento, situacao, criado_em, versao)
		 VALUES ($1,'plataforma',NULL,$2,$3,$4,$5,$6,$7,$8)`,
		novo.ID, novo.Codigo.String(), novo.Nome.String(), novo.Descricao, novo.ReferenciaInstrumento.String(),
		string(novo.Situacao), novo.CriadoEm, novo.Versao,
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return novo.ID, nil
}

func obterOuCriarIndicadorProprio(db *sqlx.DB, instituicaoID uuid.UUID, s indicadorInstituicaoSeed) (uuid.UUID, error) {
	var idExistente uuid.UUID
	err := db.Get(&idExistente, `SELECT id FROM indicador WHERE instituicao_id = $1 AND codigo = $2 AND excluido_em IS NULL`, instituicaoID, s.codigo)
	if err == nil {
		return idExistente, nil
	}

	novo, err := indicador.NovoDaInstituicao(instituicaoID, s.codigo, s.nome, "")
	if err != nil {
		return uuid.UUID{}, err
	}
	novo.Situacao = s.situacao
	_, err = db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, descricao, situacao, criado_em, versao)
		 VALUES ($1,'instituicao',$2,$3,$4,$5,$6,$7,$8)`,
		novo.ID, instituicaoID, novo.Codigo.String(), novo.Nome.String(), novo.Descricao, string(novo.Situacao), novo.CriadoEm, novo.Versao,
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return novo.ID, nil
}

func obterOuCriarMeta(db *sqlx.DB, instituicaoID uuid.UUID, nome string, indicadores []uuid.UUID) error {
	var jaExiste bool
	err := db.Get(&jaExiste,
		`SELECT EXISTS (SELECT 1 FROM meta WHERE instituicao_id = $1 AND lower(btrim(nome)) = lower(btrim($2)) AND excluido_em IS NULL)`,
		instituicaoID, nome,
	)
	if err != nil {
		return err
	}
	if jaExiste {
		return nil
	}

	nova, err := meta.NovaMeta(instituicaoID, nome, "", indicadores)
	if err != nil {
		return err
	}
	if _, err := db.Exec(
		`INSERT INTO meta (id, instituicao_id, nome, descricao, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		nova.ID, nova.InstituicaoID, nova.Nome.String(), nova.Descricao, string(nova.Situacao), nova.CriadoEm, nova.Versao,
	); err != nil {
		return err
	}
	for _, indicadorID := range nova.Indicadores {
		if _, err := db.Exec(`INSERT INTO meta_indicador (meta_id, indicador_id) VALUES ($1,$2)`, nova.ID, indicadorID); err != nil {
			return err
		}
	}
	return nil
}
