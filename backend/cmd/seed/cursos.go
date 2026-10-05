package main

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type cursoSeed struct {
	nome       string
	codigoEMec string
	grau       string
	modalidade string
	situacao   string
}

// cursosDaFSA — exatamente os seis da seção 7 da spec de cursos.
// "Nutrição" nasce inativo: sem ele, CP-06 (designação vigente em curso
// inativo mantém o perfil) não é observável. "Pedagogia" fica sem
// designação de propósito: é o dado real de CU-05 (curso vago).
var cursosDaFSA = []cursoSeed{
	{"Engenharia de Software", "1122334", "bacharelado", "presencial", "ativo"},
	{"Sistemas de Informação", "1122335", "bacharelado", "presencial", "ativo"},
	{"Pedagogia", "1122336", "licenciatura", "a_distancia", "ativo"},
	{"Análise e Desenvolvimento de Sistemas", "", "tecnologo", "presencial", "ativo"},
	{"Nutrição", "1122338", "bacharelado", "presencial", "inativo"},
	{"Biomedicina", "1122339", "bacharelado", "presencial", "ativo"},
}

type designacaoSeed struct {
	nomeCurso        string
	emailCoordenador string
	portaria         string
	dataInicio       string
	dataFim          *string
	autodesignacao   bool
}

func dataFim(iso string) *string { return &iso }

// designacoesDaFSA — exatamente as seis da seção 7 da spec, com "hoje" =
// 15/03/2026 como referência de leitura da tabela. Situação é sempre
// derivada em tempo de leitura (C-04) — a que aparece "futura" hoje vira
// "vigente" e depois "encerrada" conforme o relógio real avança, e é
// assim que tem de ser.
var designacoesDaFSA = []designacaoSeed{
	{"Engenharia de Software", "ana.lima@fsa.edu.br", "47/2026", "2026-01-01", dataFim("2026-07-31"), false},
	{"Sistemas de Informação", "ana.lima@fsa.edu.br", "47/2026", "2026-01-01", dataFim("2026-12-31"), false},
	{"Análise e Desenvolvimento de Sistemas", "paulo.tavares@fsa.edu.br", "12/2025", "2025-08-01", nil, false},
	{"Nutrição", "diego.nunes@fsa.edu.br", "51/2026", "2026-02-01", dataFim("2026-12-31"), false},
	{"Biomedicina", "beatriz.andrade@fsa.edu.br", "70/2026", "2026-03-01", dataFim("2026-12-31"), true},
	{"Engenharia de Software", "paulo.tavares@fsa.edu.br", "88/2026", "2026-08-01", dataFim("2026-12-31"), false},
}

// seedarCursosEDesignacoes — parte 4 do seed de desenvolvimento
// (specs/cursos, T-159). Só roda com APP_ENV=development, depois de
// instituições e pessoas já existirem. Idempotente por (instituição, nome
// normalizado) para curso e por (curso, coordenador, início) para
// designação.
func seedarCursosEDesignacoes(db *sqlx.DB, instituicaoFSAID uuid.UUID) {
	idPorNome := map[string]uuid.UUID{}
	for _, c := range cursosDaFSA {
		id, err := obterOuCriarCurso(db, instituicaoFSAID, c)
		if err != nil {
			log.Fatalf("seed: curso %s: %v", c.nome, err)
		}
		idPorNome[c.nome] = id
	}

	for _, d := range designacoesDaFSA {
		cursoID, ok := idPorNome[d.nomeCurso]
		if !ok {
			log.Fatalf("seed: designação: curso %q não existe", d.nomeCurso)
		}
		coordenadorID, err := obterUsuarioPorEmail(db, instituicaoFSAID, d.emailCoordenador)
		if err != nil {
			log.Fatalf("seed: designação de %s: coordenador %s: %v", d.nomeCurso, d.emailCoordenador, err)
		}
		if err := obterOuCriarDesignacao(db, cursoID, instituicaoFSAID, coordenadorID, d); err != nil {
			log.Fatalf("seed: designação de %s para %s: %v", d.nomeCurso, d.emailCoordenador, err)
		}
	}

	log.Println("seed: cursos e designações verificados/aplicados")
}

// seedarCursoEDesignacaoIVV dá a João Ribeiro (IVV) um curso onde
// coordenar por designação — sem isso, a instituição não tem nenhum
// curso, e o rótulo "Coordenador de Curso" que
// e2e/auth/mesmo-email-duas-instituicoes.spec.ts verifica no menu não tem
// como ser derivado (design.md C-09: coordenador_curso nunca mais é
// atribuição direta). Vigência indeterminada — nunca vence sozinha e
// derruba o teste.
func seedarCursoEDesignacaoIVV(db *sqlx.DB, instituicaoIVVID uuid.UUID) {
	cursoID, err := obterOuCriarCurso(db, instituicaoIVVID, cursoSeed{
		nome: "Administração", codigoEMec: "", grau: "bacharelado", modalidade: "presencial", situacao: "ativo",
	})
	if err != nil {
		log.Fatalf("seed: curso IVV: %v", err)
	}
	coordenadorID, err := obterUsuarioPorEmail(db, instituicaoIVVID, "joao.ribeiro@ies.edu.br")
	if err != nil {
		log.Fatalf("seed: designação IVV: coordenador joao.ribeiro@ies.edu.br: %v", err)
	}
	if err := obterOuCriarDesignacao(db, cursoID, instituicaoIVVID, coordenadorID, designacaoSeed{
		nomeCurso: "Administração", emailCoordenador: "joao.ribeiro@ies.edu.br",
		portaria: "1/2026", dataInicio: "2026-01-01", dataFim: nil, autodesignacao: false,
	}); err != nil {
		log.Fatalf("seed: designação IVV: %v", err)
	}
	log.Println("seed: curso e designação do IVV verificados/aplicados")
}

func obterOuCriarCurso(db *sqlx.DB, instituicaoID uuid.UUID, s cursoSeed) (uuid.UUID, error) {
	var idExistente uuid.UUID
	err := db.Get(&idExistente,
		`SELECT id FROM curso WHERE instituicao_id = $1 AND lower(btrim(nome)) = lower(btrim($2)) AND excluido_em IS NULL`,
		instituicaoID, s.nome,
	)
	if err == nil {
		return idExistente, nil
	}

	novo, err := curso.NovoCurso(instituicaoID, s.nome, s.codigoEMec, s.grau, s.modalidade)
	if err != nil {
		return uuid.UUID{}, err
	}
	if s.situacao == "inativo" {
		novo.AlterarSituacao(valueobject.CursoInativo)
	}

	var codigoArg any
	if novo.CodigoEMec != nil {
		codigoArg = novo.CodigoEMec.String()
	}
	_, err = db.Exec(
		`INSERT INTO curso (id, instituicao_id, nome, codigo_emec, grau, modalidade, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		novo.ID, novo.InstituicaoID, novo.Nome.String(), codigoArg, string(novo.Grau), string(novo.Modalidade),
		string(novo.Situacao), novo.CriadoEm, novo.Versao,
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return novo.ID, nil
}

func obterUsuarioPorEmail(db *sqlx.DB, instituicaoID uuid.UUID, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.Get(&id, `SELECT id FROM usuario WHERE instituicao_id = $1 AND email = $2 AND excluido_em IS NULL`, instituicaoID, email)
	return id, err
}

func obterOuCriarDesignacao(db *sqlx.DB, cursoID, instituicaoID, coordenadorID uuid.UUID, s designacaoSeed) error {
	var jaExiste bool
	err := db.Get(&jaExiste,
		`SELECT EXISTS (SELECT 1 FROM designacao WHERE curso_id = $1 AND coordenador_id = $2 AND data_inicio = $3 AND excluido_em IS NULL)`,
		cursoID, coordenadorID, s.dataInicio,
	)
	if err != nil {
		return err
	}
	if jaExiste {
		return nil
	}

	inicio, err := valueobject.DataLocalTexto(s.dataInicio)
	if err != nil {
		return err
	}
	var fim *valueobject.DataLocal
	if s.dataFim != nil {
		f, err := valueobject.DataLocalTexto(*s.dataFim)
		if err != nil {
			return err
		}
		fim = &f
	}

	nova, err := designacao.NovaDesignacao(cursoID, instituicaoID, coordenadorID, s.portaria, inicio, fim, s.autodesignacao)
	if err != nil {
		return err
	}

	var fimArg any
	if s.dataFim != nil {
		fimArg = *s.dataFim
	}
	_, err = db.Exec(
		`INSERT INTO designacao (id, curso_id, instituicao_id, coordenador_id, portaria, data_inicio, data_fim, autodesignacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		nova.ID, nova.CursoID, nova.InstituicaoID, nova.CoordenadorID, nova.Portaria.String(), s.dataInicio, fimArg,
		nova.Autodesignacao, nova.CriadoEm, nova.Versao,
	)
	return err
}
