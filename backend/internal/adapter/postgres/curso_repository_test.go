package postgres

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func escopoCursosDaInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, err := valueobject.DataLocalTexto("2026-03-15")
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	ator = ator.ComDataDeReferencia(hoje)
	escopo, err := autorizacao.Autorizar(ator, autorizacao.CursosDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

// TestCursoRepository_T178_CoordenadorEmUmaUnicaConsultaVagoOuPreenchido
// prova design.md §5.5/V-6: o curso vago devolve Coordenador nil; o curso
// com designação vigente devolve nome, data_fim e tambem_pi — tudo na
// MESMA consulta de Listar (não uma ida ao banco por linha).
func TestCursoRepository_T178_CoordenadorEmUmaUnicaConsultaVagoOuPreenchido(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR01"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{
		InstituicaoID: &fsaID, Nome: "Ana Lima", Perfis: []string{"professor", "pesquisador_institucional"},
	})
	cursoVagoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Pedagogia CR01"})
	cursoComCoordenadorID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Engenharia CR01"})
	fim := "2026-07-31"
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoComCoordenadorID, InstituicaoID: fsaID, CoordenadorID: anaID,
		Portaria: "47/2026", DataInicio: "2026-01-01", DataFim: &fim,
	})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopo, port.FiltroListarCursos{Page: 1, PageSize: 20, Sort: "nome", Order: "asc"})
	if err != nil {
		t.Fatalf("Listar: %v", err)
	}

	var vago, comCoordenador *port.ItemCurso
	for i := range resultado.Itens {
		switch resultado.Itens[i].Curso.ID {
		case cursoVagoID:
			vago = &resultado.Itens[i]
		case cursoComCoordenadorID:
			comCoordenador = &resultado.Itens[i]
		}
	}
	if vago == nil || comCoordenador == nil {
		t.Fatalf("esperava os dois cursos na listagem, obtido %d itens", len(resultado.Itens))
	}
	if vago.Coordenador != nil {
		t.Fatalf("curso vago deveria ter Coordenador nil, obtido %+v", vago.Coordenador)
	}
	if comCoordenador.Coordenador == nil {
		t.Fatal("esperava Coordenador preenchido")
	}
	if comCoordenador.Coordenador.Nome != "Ana Lima" {
		t.Fatalf("nome do coordenador incorreto: %s", comCoordenador.Coordenador.Nome)
	}
	if comCoordenador.Coordenador.DataFim == nil || comCoordenador.Coordenador.DataFim.String() != "2026-07-31" {
		t.Fatalf("data_fim do coordenador incorreta: %v", comCoordenador.Coordenador.DataFim)
	}
	if !comCoordenador.Coordenador.TambemPesquisadorInstitucional {
		t.Fatal("esperava tambem_pesquisador_institucional verdadeiro para Ana")
	}
	if vago.TemVinculo {
		t.Fatal("curso vago sem nenhuma designação não deveria ter tem_vinculo")
	}
	if !comCoordenador.TemVinculo {
		t.Fatal("curso com designação vigente deveria ter tem_vinculo")
	}
}

// TestCursoRepository_TemVinculoComDesignacaoEncerradaSemVigente prova que
// tem_vinculo olha para QUALQUER designação não excluída (mesmo situação
// encerrada, mesmo Coordenador vindo nil por não haver vigente) — é o
// mesmo EXISTS de Excluir, não um espelho de "coordenador != nil".
func TestCursoRepository_TemVinculoComDesignacaoEncerradaSemVigente(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR07"})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Encerrada CR07"})
	fimPassado := "2025-12-31"
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: coordenadorID, DataInicio: "2025-01-01", DataFim: &fimPassado,
	})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	item, err := repo.BuscarPorID(context.Background(), escopo, cursoID)
	if err != nil {
		t.Fatalf("BuscarPorID: %v", err)
	}
	if item.Coordenador != nil {
		t.Fatal("sem designação vigente, Coordenador deveria ser nil (a designação já encerrou)")
	}
	if !item.TemVinculo {
		t.Fatal("designação encerrada ainda é vínculo — tem_vinculo deveria ser true")
	}
}

// TestCursoRepository_FiltroCoordenadorID prova que o filtro por
// coordenador usa o placeholder correto do lateral (achado de revisão: um
// off-by-N aqui devolveria zero linhas ou um erro de placeholder fora de
// faixa, silenciosamente).
func TestCursoRepository_FiltroCoordenadorID(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR05"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	paulinhoID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoDaAnaID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Da Ana CR05"})
	cursoDoOutroID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Do Outro CR05"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoDaAnaID, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2026-01-01",
	})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoDoOutroID, InstituicaoID: fsaID, CoordenadorID: paulinhoID, DataInicio: "2026-01-01",
	})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopo, port.FiltroListarCursos{
		CoordenadorID: &anaID, Page: 1, PageSize: 20, Sort: "nome", Order: "asc",
	})
	if err != nil {
		t.Fatalf("Listar: %v", err)
	}
	if len(resultado.Itens) != 1 || resultado.Itens[0].Curso.ID != cursoDaAnaID {
		t.Fatalf("esperava só o curso da Ana, obtido %d itens", len(resultado.Itens))
	}
	if resultado.Total != 1 {
		t.Fatalf("achado de revisão: total devia contar só as linhas do filtro de coordenador, obtido %d", resultado.Total)
	}
}

// TestCursoRepository_FiltroVago prova a opção "Vago" do filtro
// Coordenador (spec.md §7, ux.md): só cursos sem designação vigente.
func TestCursoRepository_FiltroVago(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR06"})
	anaID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoComCoordenadorID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Com Coordenador CR06"})
	cursoVagoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Vago CR06"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoComCoordenadorID, InstituicaoID: fsaID, CoordenadorID: anaID, DataInicio: "2026-01-01",
	})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	resultado, err := repo.Listar(context.Background(), escopo, port.FiltroListarCursos{
		Vago: true, Page: 1, PageSize: 20, Sort: "nome", Order: "asc",
	})
	if err != nil {
		t.Fatalf("Listar: %v", err)
	}
	if len(resultado.Itens) != 1 || resultado.Itens[0].Curso.ID != cursoVagoID {
		t.Fatalf("esperava só o curso vago, obtido %d itens", len(resultado.Itens))
	}
	if resultado.Total != 1 {
		t.Fatalf("esperava total 1, obtido %d", resultado.Total)
	}
	if resultado.Itens[0].Coordenador != nil {
		t.Fatal("curso vago não deveria ter coordenador na resposta")
	}
}

// TestCursoRepository_ExcluirComDesignacaoDevolveCursoComVinculo prova
// design.md §5.4: curso com designação (mesmo excluída logicamente a
// designação não bloquearia — aqui a designação está ativa) não pode ser
// excluído.
func TestCursoRepository_ExcluirComDesignacaoDevolveCursoComVinculo(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR02"})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Com Vinculo CR02"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	err := repo.Excluir(context.Background(), escopo, cursoID)
	if err != domain.ErrCursoComVinculo {
		t.Fatalf("esperava ErrCursoComVinculo, obtido %v", err)
	}
}

// TestCursoRepository_ExcluirSemVinculoFunciona é o caminho oposto —
// curso sem nenhuma designação exclui normalmente.
func TestCursoRepository_ExcluirSemVinculoFunciona(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR03"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Sem Vinculo CR03"})

	escopo := escopoCursosDaInstituicao(t, fsaID)
	if err := repo.Excluir(context.Background(), escopo, cursoID); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if _, err := repo.BuscarPorID(context.Background(), escopo, cursoID); err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado após excluir, obtido %v", err)
	}
}

// TestCursoRepository_AlterarSituacaoNaoTocaDesignacao (T-172): inativar
// um curso não altera nenhuma linha de designacao.
func TestCursoRepository_AlterarSituacaoNaoTocaDesignacao(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	repo := NovoCursoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CR04"})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Inativar CR04"})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoID, InstituicaoID: fsaID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})

	var versaoAntes int
	if err := db.Get(&versaoAntes, `SELECT versao FROM designacao WHERE curso_id = $1`, cursoID); err != nil {
		t.Fatalf("ler versão antes: %v", err)
	}

	escopo := escopoCursosDaInstituicao(t, fsaID)
	if err := repo.AlterarSituacao(context.Background(), escopo, cursoID, valueobject.CursoInativo, 1); err != nil {
		t.Fatalf("AlterarSituacao: %v", err)
	}

	var versaoDepois int
	if err := db.Get(&versaoDepois, `SELECT versao FROM designacao WHERE curso_id = $1`, cursoID); err != nil {
		t.Fatalf("ler versão depois: %v", err)
	}
	if versaoAntes != versaoDepois {
		t.Fatalf("inativar o curso não deveria tocar a designação: versão foi de %d para %d", versaoAntes, versaoDepois)
	}
}
