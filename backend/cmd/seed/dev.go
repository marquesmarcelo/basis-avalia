package main

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basis-avalia/backend/internal/adapter/argon2"
	"github.com/basis-avalia/backend/internal/domain/instituicao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type instituicaoSeed struct {
	nome       string
	sigla      string
	codigoEMec string
	situacao   valueobject.SituacaoInstituicao
}

// instituicoesDeDesenvolvimento — exatamente as quatro da §8 da spec.
var instituicoesDeDesenvolvimento = []instituicaoSeed{
	{"Faculdade Serra Azul", "FSA", "12345", valueobject.Ativa},
	{"Instituto Vale Verde", "IVV", "", valueobject.Ativa},
	{"Centro de Ensino Aurora", "CEA", "67890", valueobject.Inativa},
	{"Faculdade Nova Aurora", "FNA", "", valueobject.Ativa},
}

type pessoaSeed struct {
	nome             string
	email            string
	siglaInstituicao string // vazio = Administrador do Sistema, sem instituição
	perfis           []valueobject.Perfil
	senha            string
	excluida         bool
	senhaProvisoria  bool
}

// pessoasDeDesenvolvimento — as da §8 da spec de autenticacao-usuarios, mais
// Paulo Tavares e Diego Nunes (⊕, §7 da spec de cursos). Carlos Pereira
// nasce já excluído e com hash nulo (3.8), para que L-04, U-03 e G-17
// tenham dado real desde o primeiro teste. Só João Ribeiro (FSA) nasce com
// senha provisória — é o dado real de S-01. Os demais já têm senha
// própria, para logar direto em L-01 e G-04. Beatriz (professor +
// pesquisador_institucional) e Letícia (aluno + professor) são o dado
// real de perfis acumulados (A-07, U-14, E-14) — nenhuma pessoa da massa
// tem "aluno" acrescentado por estar em outro perfil (U-13).
//
// coordenador_curso: ninguém tem a atribuição direta — Grupo 5 (design.md
// C-09, T-181, migration 000005_cursos_normalizacao) apagou toda linha do
// tipo do banco. Ana Lima (FSA) e João Ribeiro (IVV) coordenam por
// designação (cursos.go): João precisava de um curso na própria
// instituição para ter onde ser designado, e é esse curso que
// seedarCursoEDesignacaoIVV cria — sem ele, e2e/auth/mesmo-email-duas-
// instituicoes.spec.ts fica sem caminho de derivação para repor o rótulo
// "Coordenador de Curso" que o teste verifica no menu.
var pessoasDeDesenvolvimento = []pessoaSeed{
	{"Rafael Toledo", "rafael.toledo@basis-avalia.local", "", []valueobject.Perfil{valueobject.AdministradorSistema}, "plataforma-2026", false, false},
	{"Maria Souza", "maria.souza@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.PesquisadorInstitucional}, "reuniao-nde-2026", false, false},
	{"Beatriz Andrade", "beatriz.andrade@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor, valueobject.PesquisadorInstitucional}, "avaliacao-inep-2026", false, false},
	{"Ana Lima", "ana.lima@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor, valueobject.Aluno}, "colegiado-terca-14h", false, false},
	{"João Ribeiro", "joao.ribeiro@ies.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor}, "senha-fsa-2026", false, true},
	{"João Ribeiro", "joao.ribeiro@ies.edu.br", "IVV", []valueobject.Perfil{valueobject.Professor}, "senha-ivv-2026", false, false},
	{"Carlos Pereira", "carlos.pereira@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor}, "aposentado-2026", true, false},
	{"Letícia Moraes", "leticia.moraes@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Aluno, valueobject.Professor}, "representacao-discente-26", false, false},
	{"Ávila Gomes", "avila.gomes@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor}, "ata", false, false},
	{"Renata Coimbra", "renata.coimbra@ivv.edu.br", "IVV", []valueobject.Perfil{valueobject.PesquisadorInstitucional}, "nde-vale-verde-26", false, false},
	{"Paulo Tavares", "paulo.tavares@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor}, "portaria-12-2025", false, false},
	{"Diego Nunes", "diego.nunes@fsa.edu.br", "FSA", []valueobject.Perfil{valueobject.Professor}, "designacao-nutricao-26", false, false},
}

// seedarMassaDeDesenvolvimento — parte 2 (design.md §9), só com
// APP_ENV=development. Idempotente por (instituição, e-mail) para pessoas
// e por sigla para instituições — nunca dado real (LGPD, §10 da spec).
func seedarMassaDeDesenvolvimento(db *sqlx.DB, hashDeSenha *argon2.HashDeSenha) {
	idPorSigla := map[string]uuid.UUID{}
	for _, i := range instituicoesDeDesenvolvimento {
		id, err := obterOuCriarInstituicao(db, i)
		if err != nil {
			log.Fatalf("seed: instituição %s: %v", i.sigla, err)
		}
		idPorSigla[i.sigla] = id
	}

	for _, p := range pessoasDeDesenvolvimento {
		var instituicaoID *uuid.UUID
		if p.siglaInstituicao != "" {
			id := idPorSigla[p.siglaInstituicao]
			instituicaoID = &id
		}
		if err := criarPessoaSeNaoExistir(db, hashDeSenha, p, instituicaoID); err != nil {
			log.Fatalf("seed: pessoa %s <%s>: %v", p.nome, p.email, err)
		}
	}

	seedarIndicadoresEMetas(db, idPorSigla)
	seedarCursosEDesignacoes(db, idPorSigla["FSA"])
	seedarCursoEDesignacaoIVV(db, idPorSigla["IVV"])
	seedarPlanosDeAcao(db, idPorSigla["FSA"])
	seedarEntregas(db, idPorSigla["FSA"])

	log.Println("seed: massa de desenvolvimento verificada/aplicada")
}

func obterOuCriarInstituicao(db *sqlx.DB, s instituicaoSeed) (uuid.UUID, error) {
	var idExistente uuid.UUID
	err := db.Get(&idExistente, `SELECT id FROM instituicao WHERE sigla = $1 AND excluido_em IS NULL`, s.sigla)
	if err == nil {
		return idExistente, nil
	}

	sigla, err := valueobject.NovaSigla(s.sigla)
	if err != nil {
		return uuid.UUID{}, err
	}
	codigo, err := valueobject.NovoCodigoEMec(s.codigoEMec)
	if err != nil {
		return uuid.UUID{}, err
	}
	nova, err := instituicao.NovaInstituicao(s.nome, sigla, codigo)
	if err != nil {
		return uuid.UUID{}, err
	}
	nova.Situacao = s.situacao

	var codigoArg any
	if !nova.CodigoEMec.Nulo() {
		codigoArg = nova.CodigoEMec.String()
	}
	_, err = db.Exec(
		`INSERT INTO instituicao (id, nome, sigla, codigo_emec, situacao, criado_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		nova.ID, nova.Nome, nova.Sigla.String(), codigoArg, string(nova.Situacao), nova.CriadoEm, nova.Versao,
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	return nova.ID, nil
}

func criarPessoaSeNaoExistir(db *sqlx.DB, hashDeSenha *argon2.HashDeSenha, p pessoaSeed, instituicaoID *uuid.UUID) error {
	var jaExiste bool
	err := db.Get(&jaExiste,
		`SELECT EXISTS (SELECT 1 FROM usuario WHERE instituicao_id IS NOT DISTINCT FROM $1 AND email = $2)`,
		instituicaoID, p.email,
	)
	if err != nil {
		return err
	}
	if jaExiste {
		return nil
	}

	emailVO, err := valueobject.NovoEmail(p.email)
	if err != nil {
		return err
	}
	senhaVO, err := valueobject.NovaSenhaEmTexto(p.senha)
	if err != nil {
		return err
	}
	hashVO, err := hashDeSenha.Gerar(context.Background(), senhaVO)
	if err != nil {
		return err
	}

	perfis, err := valueobject.NovoConjunto(p.perfis...)
	if err != nil {
		return err
	}
	pessoa, err := usuario.NovoUsuario(p.nome, emailVO, hashVO, perfis, instituicaoID)
	if err != nil {
		return err
	}
	// NovoUsuario sempre nasce com senha provisória (3.4) — a massa de
	// desenvolvimento simula quem já passou pelo primeiro acesso, exceto
	// João Ribeiro (FSA), o dado real de S-01.
	if !p.senhaProvisoria {
		pessoa.DefinirSenhaPropria(hashVO, time.Now())
	}
	if p.excluida {
		pessoa.ExcluirLogicamente(time.Now())
	}

	return inserirUsuario(db, pessoa)
}
