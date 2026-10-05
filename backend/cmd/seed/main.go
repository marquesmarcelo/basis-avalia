package main

import (
	"context"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/basis-avalia/backend/internal/adapter/argon2"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
)

func main() {
	databaseURL := requireEnv("DATABASE_URL")
	appEnv := getEnv("APP_ENV", "production")

	db, err := sqlx.Connect("pgx", databaseURL)
	if err != nil {
		log.Fatalf("seed: conectando ao postgres: %v", err)
	}
	defer db.Close()

	hashDeSenha := argon2.NovoHashDeSenha(4)

	seedarAdministradorInicial(db, hashDeSenha)

	if appEnv == "development" {
		seedarMassaDeDesenvolvimento(db, hashDeSenha)
	}

	log.Println("seed: concluído")
}

// seedarAdministradorInicial — parte 1 (design.md §9, 3.13). Roda em
// qualquer ambiente, idempotente: só cria se não existir nenhum
// administrador_sistema ativo. Não cria instituição nem PI.
func seedarAdministradorInicial(db *sqlx.DB, hashDeSenha *argon2.HashDeSenha) {
	email := os.Getenv("SEED_ADMIN_EMAIL")
	senha := os.Getenv("SEED_ADMIN_SENHA")
	if email == "" || senha == "" {
		log.Fatal("seed: SEED_ADMIN_EMAIL e SEED_ADMIN_SENHA são obrigatórias e não têm valor padrão")
	}
	nome := getEnv("SEED_ADMIN_NOME", "Administrador do Sistema")

	var existeAdminAtivo bool
	err := db.Get(&existeAdminAtivo,
		`SELECT EXISTS (SELECT 1 FROM usuario_perfil up
		                  JOIN usuario u ON u.id = up.usuario_id
		                 WHERE up.perfil = 'administrador_sistema' AND u.excluido_em IS NULL)`)
	if err != nil {
		log.Fatalf("seed: verificando administrador existente: %v", err)
	}
	if existeAdminAtivo {
		log.Println("seed: já existe um Administrador do Sistema ativo — nada a fazer")
		return
	}

	emailVO, err := valueobject.NovoEmail(email)
	if err != nil {
		log.Fatalf("seed: SEED_ADMIN_EMAIL inválido: %v", err)
	}
	senhaVO, err := valueobject.NovaSenhaEmTexto(senha)
	if err != nil {
		log.Fatalf("seed: SEED_ADMIN_SENHA inválida: %v", err)
	}
	hashVO, err := hashDeSenha.Gerar(context.Background(), senhaVO)
	if err != nil {
		log.Fatalf("seed: gerando hash do administrador inicial: %v", err)
	}

	admin, err := usuario.NovoUsuario(nome, emailVO, hashVO, valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		log.Fatalf("seed: montando administrador inicial: %v", err)
	}

	if err := inserirUsuario(db, admin); err != nil {
		log.Fatalf("seed: inserindo administrador inicial: %v", err)
	}
	log.Println("seed: administrador inicial criado — senha provisória, troca obrigatória no primeiro acesso")
}

func inserirUsuario(db *sqlx.DB, u *usuario.Usuario) error {
	var senhaHash *string
	if u.SenhaHash != nil {
		codificado := u.SenhaHash.Codificado()
		senhaHash = &codificado
	}
	_, err := db.Exec(
		`INSERT INTO usuario (id, instituicao_id, nome, email, senha_hash, senha_provisoria,
		                      sessoes_validas_a_partir_de, provedor_identidade, identificador_externo,
		                      criado_em, excluido_em, versao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		u.ID, u.InstituicaoID, u.Nome, u.Email.String(), senhaHash, u.SenhaProvisoria,
		u.SessoesValidasAPartirDe, string(u.ProvedorIdentidade), u.IdentificadorExterno,
		u.CriadoEm, u.ExcluidoEm, u.Versao,
	)
	if err != nil {
		return err
	}
	for _, perfil := range u.Perfis.Ordenado() {
		var instituicaoDoVinculo *uuid.UUID
		if perfil != valueobject.AdministradorSistema {
			instituicaoDoVinculo = u.InstituicaoID
		}
		if _, err := db.Exec(
			`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,$2,$3)`,
			u.ID, string(perfil), instituicaoDoVinculo,
		); err != nil {
			return err
		}
	}
	return nil
}

func requireEnv(chave string) string {
	valor := os.Getenv(chave)
	if valor == "" {
		log.Fatalf("seed: variável de ambiente obrigatória ausente: %s", chave)
	}
	return valor
}

func getEnv(chave, padrao string) string {
	if valor := os.Getenv(chave); valor != "" {
		return valor
	}
	return padrao
}
