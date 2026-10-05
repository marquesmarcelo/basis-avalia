package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// novoUsuarioDominio monta a entidade via construtor de domínio — os
// testes de U-02/U-03/U-09 exercitam repo.Inserir de verdade, porque a
// garantia é do índice único parcial, não do fixture de setup.
func novoUsuarioDominio(t *testing.T, instituicaoID *uuid.UUID, nome, email string) *usuario.Usuario {
	t.Helper()
	emailVO, err := valueobject.NovoEmail(email)
	if err != nil {
		t.Fatalf("NovoEmail: %v", err)
	}
	hashVO, err := valueobject.NovaSenhaHash(testhelpers.HashDeTeste)
	if err != nil {
		t.Fatalf("NovaSenhaHash: %v", err)
	}
	perfis, err := valueobject.NovoConjunto(valueobject.Professor)
	if err != nil {
		t.Fatalf("NovoConjunto: %v", err)
	}
	u, err := usuario.NovoUsuario(nome, emailVO, hashVO, perfis, instituicaoID)
	if err != nil {
		t.Fatalf("NovoUsuario: %v", err)
	}
	return u
}

func escopoDaInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.UsuariosDaPropriaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

func TestUsuarioRepository_T02_IsolamentoNaoPodeSerRemovido(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAX"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IVVX"})
	renataID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Nome: "Renata Coimbra", Email: "renata.coimbra@ivv.edu.br",
		Perfil: "pesquisador_institucional",
	})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	_, err := repo.BuscarPorID(context.Background(), escopoFSA, renataID)
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obtido %v", err)
	}
}

func TestUsuarioRepository_T01_ListarSoTrazDaInstituicaoDaSessao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAY"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IVVY"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &ivvID, Perfil: "pesquisador_institucional"})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 2 {
		t.Fatalf("esperava total 2 (só FSA), obtido %d", resultado.Total)
	}
	for _, u := range resultado.Itens {
		if u.InstituicaoID == nil || *u.InstituicaoID != fsaID {
			t.Fatalf("linha de outra instituição vazou no escopo da FSA: %v", u.InstituicaoID)
		}
	}
}

func TestUsuarioRepository_E03_AtualizarComVersaoVelhaDevolveConflito(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAZ"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "João Ribeiro", Perfil: "professor"})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	u, err := repo.BuscarPorID(context.Background(), escopoFSA, usuarioID)
	if err != nil {
		t.Fatalf("BuscarPorID: %v", err)
	}

	u.Nome = "João Ribeiro Neto"
	if err := repo.Atualizar(context.Background(), escopoFSA, u, 999); err != domain.ErrConflitoDeVersao {
		t.Fatalf("esperava ErrConflitoDeVersao, obtido %v", err)
	}
}

func TestUsuarioRepository_E05_ExcluirLogicamenteAnulaHashENaoRemoveLinha(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAW"})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	if err := repo.ExcluirLogicamente(context.Background(), escopoFSA, usuarioID, time.Now()); err != nil {
		t.Fatalf("ExcluirLogicamente: %v", err)
	}

	var excluidoEm *string
	var senhaHash *string
	if err := db.QueryRow(`SELECT excluido_em::text, senha_hash FROM usuario WHERE id = $1`, usuarioID).Scan(&excluidoEm, &senhaHash); err != nil {
		t.Fatalf("consulta: %v", err)
	}
	if excluidoEm == nil {
		t.Fatal("excluido_em deveria estar preenchido")
	}
	if senhaHash != nil {
		t.Fatal("senha_hash deveria estar nulo")
	}
}

func TestUsuarioRepository_G17_ExcluidoNaoApareceEmListar(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAV"})
	testhelpers.CriarUsuarioExcluido(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Carlos Pereira", Perfil: "professor"})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 0 {
		t.Fatalf("usuário excluído não deveria ser contado, total=%d", resultado.Total)
	}
}

func TestUsuarioRepository_G02_BuscaIgnoraCaixaEAcento(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAU"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "João Ribeiro", Email: "joao.ribeiro.g02@ies.edu.br", Perfil: "professor"})

	escopoFSA := escopoDaInstituicao(t, fsaID)

	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Busca: "ri", Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 1 {
		t.Fatalf("esperava encontrar João Ribeiro buscando 'ri', total=%d", resultado.Total)
	}

	resultado2, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Busca: "joao", Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado2.Total != 1 {
		t.Fatalf("esperava que 'joao' encontrasse 'João' (unaccent), total=%d", resultado2.Total)
	}
}

// TestUsuarioRepository_O4_CuringasDaBuscaSaoEscapados prova design.md §16
// T-108 (achado O-4): "%" e "_" no termo de busca são texto literal, não
// curinga de LIKE. Sem o escape, buscar "100%" perderia o usuário com esse
// literal no nome (o "%" viraria "qualquer coisa depois de 100"), e um "%"
// isolado devolveria a instituição inteira.
func TestUsuarioRepository_O4_CuringasDaBuscaSaoEscapados(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAO4"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Turma 100% Aprovada", Email: "turma100.o4@ies.edu.br", Perfil: "professor"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Beatriz Andrade", Email: "beatriz.o4@ies.edu.br", Perfil: "professor"})

	escopoFSA := escopoDaInstituicao(t, fsaID)

	comLiteral, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Busca: "100%", Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if comLiteral.Total != 1 {
		t.Fatalf(`buscar "100%%" deveria encontrar só "Turma 100%% Aprovada" (literal), total=%d`, comLiteral.Total)
	}

	// "%" isolado, se fosse tratado como curinga de verdade, casaria com
	// QUALQUER linha (inclusive "Beatriz Andrade", que não tem "%" no
	// nome) — o mesmo efeito de "SELECT *" sem filtro. Escapado, "%" é só
	// o caractere literal: só bate com quem o tem de fato no nome.
	soCuringa, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Busca: "%", Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if soCuringa.Total != 1 {
		t.Fatalf(`buscar "%%" isolado deveria encontrar só quem tem "%%" literal no nome (1 usuário), não a instituição inteira — total=%d`, soCuringa.Total)
	}
}

func TestUsuarioRepository_G04_OrdenacaoPadraoRespeitaCollationPtBR(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAT"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Letícia Moraes", Email: "leticia.moraes.g04@fsa.edu.br", Perfil: "aluno"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Ávila Gomes", Email: "avila.gomes.g04@fsa.edu.br", Perfil: "professor"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Nome: "Beatriz Andrade", Email: "beatriz.andrade.g04@fsa.edu.br", Perfil: "pesquisador_institucional"})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(resultado.Itens) != 3 {
		t.Fatalf("esperava 3 usuários, obtido %d", len(resultado.Itens))
	}
	esperado := []string{"Ávila Gomes", "Beatriz Andrade", "Letícia Moraes"}
	for i, nome := range esperado {
		if resultado.Itens[i].Nome != nome {
			t.Fatalf("posição %d: esperado %q, obtido %q", i, nome, resultado.Itens[i].Nome)
		}
	}
}

// TestUsuarioRepository_G03_FiltroDePosseTrazQuemTemOutrosPerfisTambem prova
// o caso que o modelo de perfil único não tinha: filtrar por "professor"
// traz Beatriz mesmo ela também possuindo pesquisador_institucional — o
// filtro é POSSE, não igualdade (design.md §4.3, T-087).
func TestUsuarioRepository_G03_FiltroDePosseTrazQuemTemOutrosPerfisTambem(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "G03A"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Beatriz Andrade", Email: "beatriz.g03@fsa.edu.br",
		Perfis: []string{"professor", "pesquisador_institucional"},
	})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Letícia Moraes", Email: "leticia.g03@fsa.edu.br", Perfil: "aluno",
	})

	escopoFSA := escopoDaInstituicao(t, fsaID)
	professor := valueobject.Professor
	resultado, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{
		Perfil: &professor, Page: 1, PageSize: 20, Sort: "nome", Order: "asc",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 1 || resultado.Itens[0].Nome != "Beatriz Andrade" {
		t.Fatalf("esperava só Beatriz (possui professor entre outros perfis), obtido %d itens", resultado.Total)
	}
	if !resultado.Itens[0].Perfis.Possui(valueobject.PesquisadorInstitucional) {
		t.Fatal("o item devolvido deveria trazer TODOS os perfis de Beatriz, não só o filtrado")
	}
}

func TestUsuarioRepository_AS02_ListaDePIsTrazQuemAcumulaOutroPerfilEIsolaOutraInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "AS02A"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "AS02B"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Beatriz Andrade", Email: "beatriz.as02@fsa.edu.br",
		Perfis: []string{"professor", "pesquisador_institucional"},
	})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Letícia Moraes", Email: "leticia.as02@fsa.edu.br", Perfil: "aluno",
	})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &ivvID, Nome: "Renata Coimbra", Email: "renata.as02@ivv.edu.br",
		Perfil: "pesquisador_institucional",
	})

	conjunto, err := valueobject.NovoConjunto(valueobject.AdministradorSistema)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	atorAdmin, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopoPIsDaFSA, err := autorizacao.Autorizar(atorAdmin, autorizacao.PesquisadoresDeUmaInstituicao, autorizacao.AcaoListar, &fsaID)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	resultado, err := repo.Listar(context.Background(), escopoPIsDaFSA, port.FiltroListarUsuarios{
		Page: 1, PageSize: 20, Sort: "nome", Order: "asc",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Total != 1 || resultado.Itens[0].Nome != "Beatriz Andrade" {
		t.Fatalf("esperava só Beatriz (única detentora de PI na FSA), obtido %d itens", resultado.Total)
	}
	if !resultado.Itens[0].Perfis.Possui(valueobject.Professor) {
		t.Fatal("o item devolvido deveria trazer também o perfil de professor de Beatriz")
	}
}

func TestUsuarioRepository_G09_Paginacao45Usuarios(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAS"})
	for i := 0; i < 45; i++ {
		testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
			InstituicaoID: &fsaID,
			Nome:          fmt.Sprintf("Usuario Paginacao %02d", i),
			Email:         fmt.Sprintf("paginacao-%02d@fsa.edu.br", i),
			Perfil:        "professor",
		})
	}

	escopoFSA := escopoDaInstituicao(t, fsaID)
	pagina1, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if pagina1.Total != 45 {
		t.Fatalf("esperava total 45, obtido %d", pagina1.Total)
	}
	if len(pagina1.Itens) != 20 {
		t.Fatalf("esperava 20 itens na página 1, obtido %d", len(pagina1.Itens))
	}

	pagina2, err := repo.Listar(context.Background(), escopoFSA, port.FiltroListarUsuarios{Page: 2, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(pagina2.Itens) != 20 {
		t.Fatalf("esperava 20 itens na página 2, obtido %d", len(pagina2.Itens))
	}
	if pagina1.Itens[0].Nome == pagina2.Itens[0].Nome {
		t.Fatal("página 2 deveria trazer registros diferentes da página 1")
	}
}

// TestUsuarioRepository_U02_EmailDuplicadoNaMesmaInstituicao prova a
// fronteira de integridade que o índice único parcial garante — não a
// checagem de aplicação, o índice em si (specs/autenticacao-usuarios/
// spec.md §15.1, "Unicidade").
func TestUsuarioRepository_U02_EmailDuplicadoNaMesmaInstituicao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAU2"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "João Ribeiro", Email: "joao.ribeiro@ies.edu.br", Perfil: "professor",
	})

	duplicado := novoUsuarioDominio(t, &fsaID, "João Ribeiro Segundo", "joao.ribeiro@ies.edu.br")
	escopoFSA := escopoDaInstituicao(t, fsaID)
	if err := repo.Inserir(context.Background(), escopoFSA, duplicado); err != domain.ErrEmailDuplicado {
		t.Fatalf("esperava ErrEmailDuplicado, obtido %v", err)
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM usuario WHERE instituicao_id = $1 AND email = $2`, fsaID, "joao.ribeiro@ies.edu.br"); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	if total != 1 {
		t.Fatalf("esperava 1 registro (nenhum criado pela tentativa duplicada), obtido %d", total)
	}
}

// TestUsuarioRepository_U03_EmailDeExcluidoVoltaAFicarLivre prova que o
// índice é parcial (WHERE excluido_em IS NULL, não UNIQUE simples): um
// e-mail de usuário excluído logicamente não bloqueia um cadastro novo com
// o mesmo e-mail na mesma instituição.
func TestUsuarioRepository_U03_EmailDeExcluidoVoltaAFicarLivre(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAU3"})
	testhelpers.CriarUsuarioExcluido(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Carlos Pereira", Email: "carlos.pereira@fsa.edu.br", Perfil: "professor",
	})

	novo := novoUsuarioDominio(t, &fsaID, "Carlos Pereira Neto", "carlos.pereira@fsa.edu.br")
	escopoFSA := escopoDaInstituicao(t, fsaID)
	if err := repo.Inserir(context.Background(), escopoFSA, novo); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM usuario WHERE id = $1`, novo.ID) })

	var totalNaoExcluidos int
	if err := db.Get(&totalNaoExcluidos,
		`SELECT count(*) FROM usuario WHERE instituicao_id = $1 AND email = $2 AND excluido_em IS NULL`,
		fsaID, "carlos.pereira@fsa.edu.br"); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	if totalNaoExcluidos != 1 {
		t.Fatalf("esperava 1 registro ativo com o e-mail, obtido %d", totalNaoExcluidos)
	}
	var totalExcluidos int
	if err := db.Get(&totalExcluidos,
		`SELECT count(*) FROM usuario WHERE instituicao_id = $1 AND email = $2 AND excluido_em IS NOT NULL`,
		fsaID, "carlos.pereira@fsa.edu.br"); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	if totalExcluidos != 1 {
		t.Fatalf("esperava o registro antigo continuar marcado como excluído, obtido %d excluídos", totalExcluidos)
	}
}

// TestUsuarioRepository_U09_MesmoEmailEmDuasInstituicoes prova a premissa
// do modelo multi-institucional: o índice único é por (instituicao_id,
// email), nunca por email isolado — duas contas independentes, e alterar
// uma não afeta a outra.
func TestUsuarioRepository_U09_MesmoEmailEmDuasInstituicoes(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoUsuarioRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "FSAU9"})
	ivvID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "IVVU9"})
	testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "João Ribeiro", Email: "joao.ribeiro@ies.edu.br", Perfil: "professor",
	})

	naIVV := novoUsuarioDominio(t, &ivvID, "João Ribeiro", "joao.ribeiro@ies.edu.br")
	escopoIVV := escopoDaInstituicao(t, ivvID)
	if err := repo.Inserir(context.Background(), escopoIVV, naIVV); err != nil {
		t.Fatalf("esperava sucesso — mesmo e-mail em outra instituição é conta independente: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM usuario WHERE id = $1`, naIVV.ID) })

	var totalContas int
	if err := db.Get(&totalContas, `SELECT count(*) FROM usuario WHERE email = $1`, "joao.ribeiro@ies.edu.br"); err != nil {
		t.Fatalf("contagem: %v", err)
	}
	if totalContas != 2 {
		t.Fatalf("esperava 2 contas independentes com o mesmo e-mail, obtido %d", totalContas)
	}

	naIVV.Nome = "João Ribeiro (IVV)"
	if err := repo.Atualizar(context.Background(), escopoIVV, naIVV, naIVV.Versao); err != nil {
		t.Fatalf("Atualizar na IVV: %v", err)
	}

	var nomeNaFSA string
	if err := db.Get(&nomeNaFSA, `SELECT nome FROM usuario WHERE instituicao_id = $1 AND email = $2`, fsaID, "joao.ribeiro@ies.edu.br"); err != nil {
		t.Fatalf("ler conta da FSA: %v", err)
	}
	if nomeNaFSA != "João Ribeiro" {
		t.Fatalf("alterar a conta do IVV vazou para a conta da FSA: nome ficou %q", nomeNaFSA)
	}
}
