package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func atorEntregaDeTeste(t *testing.T, instituicaoID, usuarioID uuid.UUID) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.CoordenadorCurso)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(usuarioID, conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	return ator.ComDataDeReferencia(hoje)
}

func escopoEntregaDaCarteira(t *testing.T, instituicaoID, usuarioID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	ator := atorEntregaDeTeste(t, instituicaoID, usuarioID)
	esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoCriar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return esc
}

// EN-06, V-6 (design.md §4.1, §7): duas transações concorrentes com a
// MESMA (usuario_id, rota, chave) produzem UMA entrega — a garantia é a
// chave primária de `idempotencia`, dentro da transação, nunca uma
// verificação prévia (que seria a própria corrida).
func TestEntregaRepository_EN06_IdempotenciaSobCorridaReal(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	itemID := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 4})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	repo := NovoEntregaRepository(db)
	uow := NovaUnidadeDeTrabalho(db)
	esc := escopoEntregaDaCarteira(t, instituicaoID, usuarioID)
	const chave = "chave-de-corrida-en06"
	const rota = "registrar_entrega"

	tentar := func() (uuid.UUID, error) {
		nova, err := entrega.NovaEntrega(itemID, cursoID, instituicaoID, usuarioID, "")
		if err != nil {
			return uuid.UUID{}, err
		}
		erroTx := uow.Executar(context.Background(), func(ctx context.Context) error {
			if err := repo.InserirIdempotencia(ctx, atorEntregaDeTeste(t, instituicaoID, usuarioID).Proprio(), rota, chave, nova.ID); err != nil {
				return err
			}
			return repo.InserirEntregaComAnexos(ctx, esc, nova, nil)
		})
		if erroTx != nil {
			return uuid.UUID{}, erroTx
		}
		return nova.ID, nil
	}

	const tentativas = 5
	var wg sync.WaitGroup
	ids := make([]uuid.UUID, tentativas)
	erros := make([]error, tentativas)
	for i := 0; i < tentativas; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id, err := tentar()
			ids[i] = id
			erros[i] = err
		}(i)
	}
	wg.Wait()

	sucessos := 0
	duplicadas := 0
	for _, err := range erros {
		switch {
		case err == nil:
			sucessos++
		case err == domain.ErrChaveDuplicada:
			duplicadas++
		default:
			t.Fatalf("erro inesperado sob corrida: %v", err)
		}
	}
	if sucessos != 1 {
		t.Fatalf("esperava exatamente 1 sucesso sob corrida, obteve %d (duplicadas=%d)", sucessos, duplicadas)
	}
	if duplicadas != tentativas-1 {
		t.Fatalf("esperava %d chaves duplicadas, obteve %d", tentativas-1, duplicadas)
	}

	var total int
	if err := db.Get(&total, `SELECT count(*) FROM entrega WHERE item_plano_id = $1`, itemID); err != nil {
		t.Fatalf("contando entregas: %v", err)
	}
	if total != 1 {
		t.Fatalf("V-6: esperava exatamente UMA entrega gravada no banco, encontrou %d", total)
	}

	entregaID, err := repo.BuscarEntregaPorChaveIdempotencia(context.Background(), atorEntregaDeTeste(t, instituicaoID, usuarioID).Proprio(), rota, chave)
	if err != nil {
		t.Fatalf("BuscarEntregaPorChaveIdempotencia: %v", err)
	}
	achouOriginal := false
	for i, id := range ids {
		if erros[i] == nil && id == entregaID {
			achouOriginal = true
		}
	}
	if !achouOriginal {
		t.Fatal("a chave de idempotência deveria apontar para a entrega que realmente foi inserida")
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM idempotencia WHERE usuario_id = $1 AND rota = $2 AND chave = $3`, usuarioID, rota, chave)
	})
}

// VI-06: entrega de outra instituição responde não encontrado — o
// isolamento é o predicado de AplicarEscopo, sempre.
func TestEntregaRepository_VI06_OutraInstituicaoNaoEncontrada(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoA := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	instituicaoB := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cursoA := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoA})
	periodoA := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoA})
	planoA := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoA, CursoID: cursoA, PeriodoID: periodoA, SituacaoPublicacao: "vigente"})
	metaA := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoA})
	itemA := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoA, CursoID: cursoA, InstituicaoID: instituicaoA, MetaID: metaA, Quantidade: 1})
	usuarioA := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoA})
	entregaA := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoA, EnviadaPor: usuarioA})

	usuarioB := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoB})
	repo := NovoEntregaRepository(db)
	escB := escopoEntregaDaInstituicaoPI(t, instituicaoB)

	_, err := repo.BuscarPorID(context.Background(), escB, entregaA, atorEntregaDeTeste(t, instituicaoB, usuarioB).Proprio())
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obteve %v", err)
	}
}

func escopoEntregaDaInstituicaoPI(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.PesquisadorInstitucional)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	hoje, _ := valueobject.DataLocalTexto("2026-03-15")
	ator = ator.ComDataDeReferencia(hoje)
	esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return esc
}

// AN-05 (fronteira de isolamento — testada agora, não adiada para a
// Release, por decisão do arquiteto): anexo de uma entrega de outra
// instituição responde não encontrado. A chave do objeto nunca chega a
// ser lida — BuscarAnexoParaDownload falha antes disso.
func TestEntregaRepository_AN05_AnexoDeOutraInstituicaoNaoEncontrado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoA := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	instituicaoB := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	cursoA := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoA})
	periodoA := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoA})
	planoA := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoA, CursoID: cursoA, PeriodoID: periodoA, SituacaoPublicacao: "vigente"})
	metaA := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoA})
	itemA := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoA, CursoID: cursoA, InstituicaoID: instituicaoA, MetaID: metaA, Quantidade: 1})
	usuarioA := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoA})
	entregaA := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoA, EnviadaPor: usuarioA})
	anexoA := testhelpers.CriarAnexo(t, db, testhelpers.OpcoesAnexo{EntregaID: entregaA, CursoID: cursoA, InstituicaoID: instituicaoA})

	repo := NovoEntregaRepository(db)
	escB := escopoEntregaDaInstituicaoPI(t, instituicaoB)

	_, _, _, err := repo.BuscarAnexoParaDownload(context.Background(), escB, anexoA)
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obteve %v", err)
	}
}

// AN-06 (fronteira de isolamento, mesmo tratamento de AN-05): anexo de um
// curso fora da carteira do coordenador responde não encontrado, nunca
// 403 — dentro da própria instituição, "existe mas não é seu" e "não
// existe" são indistinguíveis de propósito.
func TestEntregaRepository_AN06_AnexoDeCursoForaDaCarteiraNaoEncontrado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})

	cursoAlheio := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	planoAlheio := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoAlheio, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	itemAlheio := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoAlheio, CursoID: cursoAlheio, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1})
	dono := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	entregaAlheia := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemAlheio, CursoID: cursoAlheio, InstituicaoID: instituicaoID, EnviadaPor: dono})
	anexoAlheio := testhelpers.CriarAnexo(t, db, testhelpers.OpcoesAnexo{EntregaID: entregaAlheia, CursoID: cursoAlheio, InstituicaoID: instituicaoID})

	// O coordenador tem designação vigente num curso DIFERENTE — a
	// carteira dele nunca inclui cursoAlheio.
	cursoDaCarteira := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})

	repo := NovoEntregaRepository(db)
	esc := escopoEntregaDaCarteira(t, instituicaoID, coordenadorID)

	_, _, _, err := repo.BuscarAnexoParaDownload(context.Background(), esc, anexoAlheio)
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obteve %v", err)
	}
}

// AN-04 (caminho positivo, contraparte de AN-05/AN-06): o coordenador
// baixa o anexo de uma entrega do PRÓPRIO curso — a chave do objeto é
// devolvida só depois de passar por Escopo.
func TestEntregaRepository_AN04_AnexoDoProprioCursoEncontrado(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})

	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	cursoDaCarteira := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID, DataInicio: "2026-01-01",
	})
	plano := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoDaCarteira, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	item := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: plano, CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 1})
	entrega := testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: item, CursoID: cursoDaCarteira, InstituicaoID: instituicaoID, EnviadaPor: coordenadorID})
	anexo := testhelpers.CriarAnexo(t, db, testhelpers.OpcoesAnexo{
		EntregaID: entrega, CursoID: cursoDaCarteira, InstituicaoID: instituicaoID,
		NomeOriginal: "ata.pdf", ChaveObjeto: "chave-de-teste-an04",
	})

	repo := NovoEntregaRepository(db)
	esc := escopoEntregaDaCarteira(t, instituicaoID, coordenadorID)

	nome, tipo, chave, err := repo.BuscarAnexoParaDownload(context.Background(), esc, anexo)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if nome != "ata.pdf" {
		t.Fatalf("esperava nome_original 'ata.pdf', obteve %q", nome)
	}
	if chave != "chave-de-teste-an04" {
		t.Fatalf("esperava a chave do objeto, obteve %q", chave)
	}
	if tipo == "" {
		t.Fatal("esperava tipo não vazio")
	}
}
