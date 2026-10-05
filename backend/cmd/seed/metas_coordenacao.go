package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/google/uuid"
)

// namespaceEntregasSeed — namespace fixo para gerar IDs determinísticos
// por posição na lista (uuid.NewMD5): é o que torna o seed idempotente
// mesmo quando duas entradas da lista têm o mesmo (curso, meta,
// enviada_por, situação) — caso legítimo aqui, porque EN-01 diz
// explicitamente que entregas idênticas são registros distintos.
var namespaceEntregasSeed = uuid.MustParse("6f6e1f0a-0000-4000-8000-000000000001")

type entregaSeed struct {
	nomeDoCurso      string
	nomeDaMeta       string
	// nomeDoPeriodo desambigua qual item_plano recebe a entrega — achado
	// de revisão: sem isto, um curso com a MESMA meta em dois períodos
	// (ex: Sistemas de Informação em 2026.1 e 2026.2) deixava o SELECT de
	// item_plano com LIMIT 1 e sem ORDER BY escolher uma das duas linhas
	// por acaso, podendo prender a entrega no plano errado (inclusive um
	// rascunho, onde a API real nunca permitiria registrar entrega).
	nomeDoPeriodo           string
	emailEnviadaPor         string
	situacao                string // "pendente_avaliacao" | "aceita" | "recusada"
	rodadas                 int
	prazoCorrecaoAte        *string // instante ISO completo, "" = sem prazo
	emailAvaliadaPor        string  // "" = não avaliada
	motivo                  string
	avaliadorEraCoordenador bool
}

func prazo(iso string) *string { return &iso }

// prazoDaquiA calcula um prazo de correção genuinamente futuro a partir
// do relógio real do momento em que o seed roda — nunca uma data de
// calendário fixa. Achado de revisão: a entrega recusada de Engenharia de
// Software abaixo tinha prazo fixo em 22/03/2026, uma referência válida
// só enquanto "hoje" ficasse perto de 15/03/2026 (a data dos exemplos da
// spec) — o relógio real já passou disso, e "prazo ainda em curso" virou
// prazo expirado sem ninguém mudar uma linha de código.
func prazoDaquiA(dias int) *string {
	iso := time.Now().AddDate(0, 0, dias).Format("2006-01-02") + "T23:59:59-03:00"
	return &iso
}

// entregasDeDesenvolvimento — specs/metas-coordenacao/design.md §4.1 V-8:
// sem a entrega avaliada por Beatriz em Biomedicina, AV-15, RD-11 e a
// métrica de coincidência de papéis não são observáveis. "Hoje" nos
// exemplos originais é 15/03/2026 — mantido como referência de leitura
// para as entradas mais antigas, mas as novas (2026.2) usam prazoDaquiA
// porque dependem de estar corretas em relação ao relógio real.
var entregasDeDesenvolvimento = []entregaSeed{
	// Engenharia de Software — item "Registrar reuniões de NDE em ata"
	// exige 4: 2 aceitas, 1 pendente, 1 recusada. O prazo desta recusada
	// é histórico (22/03/2026) e hoje já expirado — mantido como estava
	// para não mudar o dado que EN-*/AV-* já referenciam; o cenário
	// "recusada com prazo ainda correndo" agora vive em ADS (2026.2, abaixo).
	{"Engenharia de Software", "Registrar reuniões de NDE em ata", "2026.1", "ana.lima@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},
	{"Engenharia de Software", "Registrar reuniões de NDE em ata", "2026.1", "ana.lima@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},
	{"Engenharia de Software", "Registrar reuniões de NDE em ata", "2026.1", "ana.lima@fsa.edu.br", "pendente_avaliacao", 0, nil, "", "", false},
	{"Engenharia de Software", "Registrar reuniões de NDE em ata", "2026.1", "ana.lima@fsa.edu.br", "recusada", 1, prazo("2026-03-22T23:59:59-03:00"), "maria.souza@fsa.edu.br", "A lista de presença não corresponde à data da ata.", false},

	// Biomedicina — Beatriz Andrade é PI e coordenadora vigente do curso
	// (Portaria 70/2026): ela avalia a própria entrega, e a marca de
	// coincidência precisa ficar gravada (AV-15, M-13).
	{"Biomedicina", "Registrar reuniões de NDE em ata", "2026.1", "beatriz.andrade@fsa.edu.br", "aceita", 0, nil, "beatriz.andrade@fsa.edu.br", "", true},

	// Sistemas de Informação (plano 2026.2) — meta CUMPRIDA: as 3 de 3
	// exigidas aceitas. Cenário pedido pelo dono para simular o
	// relatório: curso que cumpre integralmente.
	{"Sistemas de Informação", "Registrar reuniões de NDE em ata", "2026.2", "ana.lima@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},
	{"Sistemas de Informação", "Registrar reuniões de NDE em ata", "2026.2", "ana.lima@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},
	{"Sistemas de Informação", "Registrar reuniões de NDE em ata", "2026.2", "ana.lima@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},

	// Análise e Desenvolvimento de Sistemas (plano 2026.2) — meta
	// PARCIALMENTE cumprida: 1 aceita de 3 exigidas, 1 pendente de
	// avaliação (alimenta a fila do PI) e 1 recusada com prazo
	// genuinamente correndo (prazoDaquiA, nunca data fixa). O segundo
	// item do plano ("Relatório de acompanhamento do curso") fica sem
	// nenhuma entrega de propósito — curso COM coordenador e meta em
	// aberto, mas zero evidência enviada: o contraponto de Pedagogia
	// (curso SEM coordenador, abaixo) — o relatório precisa distinguir
	// "não cumpriu" de "não havia quem cumprisse".
	{"Análise e Desenvolvimento de Sistemas", "Registrar reuniões de NDE em ata", "2026.2", "paulo.tavares@fsa.edu.br", "aceita", 0, nil, "maria.souza@fsa.edu.br", "", false},
	{"Análise e Desenvolvimento de Sistemas", "Registrar reuniões de NDE em ata", "2026.2", "paulo.tavares@fsa.edu.br", "pendente_avaliacao", 0, nil, "", "", false},
	{"Análise e Desenvolvimento de Sistemas", "Registrar reuniões de NDE em ata", "2026.2", "paulo.tavares@fsa.edu.br", "recusada", 1, prazoDaquiA(21), "maria.souza@fsa.edu.br", "O comprovante enviado não corresponde à ata da reunião.", false},

	// Pedagogia (plano 2026.1) já nasce sem designação vigente (curso
	// vago, cmd/seed/cursos.go) e sem nenhuma entrega registrada aqui de
	// propósito — é o cenário "sem coordenador, meta em aberto", distinto
	// de ADS acima (com coordenador, item sem entrega).
}

// seedarEntregas — última parte do seed de desenvolvimento
// (specs/metas-coordenacao, T-257). Roda depois de cursos, metas e planos
// já existirem. Idempotente por (item_plano_id, enviada_por, situacao,
// criado_em do dia) é impraticável sem uma chave natural — a proteção
// real aqui é rodar só uma vez por banco (mesmo padrão de
// seedarCursosEDesignacoes: consultado antes de inserir).
func seedarEntregas(db *sqlx.DB, instituicaoFSAID uuid.UUID) {
	for i, e := range entregasDeDesenvolvimento {
		id := uuid.NewMD5(namespaceEntregasSeed, []byte(fmt.Sprintf("entrega-seed-%d", i)))
		if err := criarEntregaSeNaoExistir(db, instituicaoFSAID, id, e); err != nil {
			log.Fatalf("seed: entrega de %s (%s): %v", e.nomeDoCurso, e.emailEnviadaPor, err)
		}
	}
	log.Println("seed: entregas de desenvolvimento verificadas/aplicadas")
}

func criarEntregaSeNaoExistir(db *sqlx.DB, instituicaoID uuid.UUID, entregaID uuid.UUID, s entregaSeed) error {
	cursoID, err := buscarCursoIDPorNome(db, instituicaoID, s.nomeDoCurso)
	if err != nil {
		return err
	}
	metaID, err := buscarMetaIDPorNome(db, instituicaoID, s.nomeDaMeta)
	if err != nil {
		return err
	}
	// Desambiguado por período (nome) — um curso pode ter a mesma meta em
	// mais de um plano (ex: Sistemas de Informação em 2026.1 e 2026.2);
	// sem esta junção, LIMIT 1 sem ORDER BY escolhia uma linha arbitrária,
	// podendo prender a entrega no plano errado, inclusive um rascunho.
	var itemID uuid.UUID
	err = db.Get(&itemID,
		`SELECT ip.id FROM item_plano ip
		   JOIN plano p ON p.id = ip.plano_id
		   JOIN periodo per ON per.id = p.periodo_id
		  WHERE ip.curso_id = $1 AND ip.meta_id = $2 AND lower(btrim(per.nome)) = lower(btrim($3))
		    AND ip.excluido_em IS NULL AND p.excluido_em IS NULL
		  LIMIT 1`, cursoID, metaID, s.nomeDoPeriodo)
	if err != nil {
		return err
	}
	enviadaPorID, err := obterUsuarioPorEmail(db, instituicaoID, s.emailEnviadaPor)
	if err != nil {
		return err
	}

	var jaExiste bool
	if err := db.Get(&jaExiste, `SELECT EXISTS (SELECT 1 FROM entrega WHERE id = $1)`, entregaID); err != nil {
		return err
	}
	if jaExiste {
		return nil
	}

	var avaliadaPor any
	var avaliadaEm any
	if s.emailAvaliadaPor != "" {
		id, err := obterUsuarioPorEmail(db, instituicaoID, s.emailAvaliadaPor)
		if err != nil {
			return err
		}
		avaliadaPor = id
		avaliadaEm = time.Now()
	}

	_, err = db.Exec(
		`INSERT INTO entrega (id, item_plano_id, curso_id, instituicao_id, enviada_por, situacao,
		                      rodadas_de_recusa, prazo_correcao_ate, avaliada_por, avaliada_em, motivo,
		                      avaliador_era_coordenador, criado_em)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		entregaID, itemID, cursoID, instituicaoID, enviadaPorID, s.situacao,
		s.rodadas, s.prazoCorrecaoAte, avaliadaPor, avaliadaEm, s.motivo, s.avaliadorEraCoordenador, time.Now(),
	)
	if err != nil {
		return err
	}

	// Um anexo por entrega — o suficiente para EN-05 (cada entrega conta
	// 1, qualquer que seja o nº de anexos) ser observável no seed sem
	// depender de upload real: a chave aponta para um objeto que não
	// existe no MinIO, e é por isso que download deste anexo específico
	// não é exercitado pelo seed (só o registro na listagem).
	_, err = db.Exec(
		`INSERT INTO anexo (id, entrega_id, curso_id, instituicao_id, nome_original, tipo, tamanho_bytes, chave_objeto, hash_sha256)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		uuid.Must(uuid.NewV7()), entregaID, cursoID, instituicaoID,
		"comprovante-seed.pdf", "pdf", 102400, "seed/nao-existe/"+entregaID.String(),
		strings.Repeat("0", 64),
	)
	return err
}
