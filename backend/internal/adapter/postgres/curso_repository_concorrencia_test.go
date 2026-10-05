package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/designacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
)

// TestCursoRepository_T171_ExcluirSobFORUPDATENaoDeixaDesignacaoOrfa é o
// teste de integração de T-171 (specs/cursos/tasks.md): dispara, ao mesmo
// tempo, a exclusão do curso e a criação de uma designação para ele — a
// trava SELECT ... FOR UPDATE na linha do curso (design.md §5.4) precisa
// impedir que as duas terminem de um jeito que deixe uma designação
// apontando para um curso já excluído.
func TestCursoRepository_T171_ExcluirSobFORUPDATENaoDeixaDesignacaoOrfa(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	cursoRepo := NovoCursoRepository(db)
	designacaoRepo := NovoDesignacaoRepository(db)

	fsaID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{Sigla: "CC01"})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &fsaID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: fsaID, Nome: "Curso Corrida CC01"})
	t.Cleanup(func() { db.Exec(`DELETE FROM designacao WHERE curso_id = $1`, cursoID) })

	escopo := escopoCursosDaInstituicao(t, fsaID)
	inicio, err := valueobject.DataLocalTexto("2026-01-01")
	if err != nil {
		t.Fatalf("data: %v", err)
	}

	// As duas chamadas rodam dentro de uma unidade de trabalho própria,
	// exatamente como ExcluirCursoUseCase e CriarDesignacaoUseCase sempre
	// fazem em produção (uow.Executar abre a transação real que mantém o
	// SELECT ... FOR UPDATE preso até o commit). Chamar o repositório
	// direto, sem transação, faz cada instrução se autocommitar sozinha —
	// a trava é liberada entre as instruções, e a corrida deixa de ser a
	// mesma que o caminho de produção corre.
	//
	// A trava real (achado de revisão T-171) NÃO é o FK de designacao→
	// curso: a FK só garante que a linha existe, nunca que excluido_em
	// continua nulo — exclusão lógica não dispara FK. Quem fecha a
	// corrida é CriarDesignacaoUseCase chamando TravarSeAtivo (FOR SHARE)
	// antes de inserir, e é isso que este teste replica manualmente —
	// sem essa chamada aqui, o teste provaria uma trava que o caminho de
	// produção não tem mais motivo de usar sozinha (a FK), não a que
	// realmente serializa hoje.
	uow := NovaUnidadeDeTrabalho(db)
	var wg sync.WaitGroup
	var erroExcluir, erroInserir error
	wg.Add(2)
	go func() {
		defer wg.Done()
		erroExcluir = uow.Executar(context.Background(), func(ctx context.Context) error {
			return cursoRepo.Excluir(ctx, escopo, cursoID)
		})
	}()
	go func() {
		defer wg.Done()
		nova, err := designacao.NovaDesignacao(cursoID, fsaID, coordenadorID, "1/2026", inicio, nil, false)
		if err != nil {
			erroInserir = err
			return
		}
		erroInserir = uow.Executar(context.Background(), func(ctx context.Context) error {
			if err := cursoRepo.TravarSeAtivo(ctx, escopo, cursoID); err != nil {
				return err
			}
			return designacaoRepo.Inserir(ctx, escopoDesignacoesDaInstituicao(t, fsaID), nova)
		})
	}()
	wg.Wait()

	// Não importa qual dos dois "ganhou" — o que não pode acontecer é os
	// dois terem sucesso ao mesmo tempo: se o curso foi excluído, a
	// designação tinha que ter sido recusada (o curso já não existe mais
	// para quem consulta com excluido_em IS NULL); se a designação foi
	// criada, o curso não podia ter sido excluído (CURSO_COM_VINCULO).
	cursoFoiExcluido := erroExcluir == nil
	designacaoFoiCriada := erroInserir == nil

	if cursoFoiExcluido && designacaoFoiCriada {
		var excluidoEm *string
		db.Get(&excluidoEm, `SELECT excluido_em::text FROM curso WHERE id = $1`, cursoID)
		t.Fatalf(
			"corrida não serializada: curso excluído (excluido_em=%v) E designação criada ao mesmo tempo — "+
				"a designação ficou órfã de um curso excluído. erroExcluir=%v erroInserir=%v",
			excluidoEm, erroExcluir, erroInserir,
		)
	}

	if !cursoFoiExcluido && erroExcluir != domain.ErrCursoComVinculo {
		t.Fatalf("exclusão recusada por outro motivo além de vínculo: %v", erroExcluir)
	}
}
