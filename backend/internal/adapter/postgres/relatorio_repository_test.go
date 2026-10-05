package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

func escopoRelatorioDeInstituicao(t *testing.T, instituicaoID uuid.UUID) autorizacao.Escopo {
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
	esc, err := autorizacao.Autorizar(ator, autorizacao.DesempenhoDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return esc
}

// M-10/M-11, T-288: a consulta de LISTAGEM/TOTAL do relatório (o que
// `montarRelatorioBase` produz — a base de `RelatorioDesempenho` e
// `ExportarDesempenho`, grão de item de plano) NUNCA contém DISTINCT —
// teste de mecanismo puro, sem banco. DISTINCT deduplica linhas, não
// agregados: introduzi-lo aqui faz este teste falhar, o que é o objetivo.
//
// Este teste cobre SÓ esta consulta — a do RESUMO (grão de curso, que
// converte grão de propósito) tem seu próprio teste,
// TestResumoSQL_DistinctSancionadoUsaIdentificadorDoCurso, com a exceção
// numa allowlist explícita (T-120): nome de teste que promete mais do que
// mede é defeito.
func TestMontarRelatorioBase_SemDistinct(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)
	repo := NovoEntregaRepository(nil)

	selectCampos, from, where, _, _, err := repo.montarRelatorioBase(esc, port.FiltroRelatorio{PeriodoID: uuid.Must(uuid.NewV7())})
	if err != nil {
		t.Fatalf("montarRelatorioBase: %v", err)
	}
	consulta := strings.ToUpper(selectCampos + from + where)
	if strings.Contains(consulta, "DISTINCT") {
		t.Fatalf("a consulta do relatório NÃO PODE conter DISTINCT — mascararia a multiplicação nos agregados:\n%s", consulta)
	}
}

// allowlistDeDistinctSancionado — a lista fechada de ocorrências de
// DISTINCT permitidas em agregado nesta feature (decisão do arquiteto,
// T-120): sancionado quando CONVERTE GRÃO (conta uma entidade de grão mais
// grosso sobre um conjunto de grão mais fino), proibido quando esconde
// cardinalidade de junção. Cada entrada documenta por que a conversão de
// grão é legítima ali — acrescentar uma entrada nova é decisão do
// arquiteto, nunca silenciosa.
var allowlistDeDistinctSancionado = []struct {
	consulta string
	trecho   string
	motivo   string
}{
	{
		consulta: "resumoSQL (resumo do relatório de desempenho)",
		trecho:   "count(distinct case when coordenador_id is null then curso_id end)",
		motivo:   "converte grão de item de plano (curso × meta) para grão de curso — conta CURSOS sobre um conjunto de linhas de item, nunca esconde multiplicação de junção",
	},
}

// T-120: a consulta de RESUMO (grão de curso) tem exatamente UM DISTINCT,
// sancionado pela allowlist acima — e a chave é o IDENTIFICADOR do curso
// (`curso_id`), nunca um atributo textual (`curso_nome`). Nome não é
// identidade: dois cursos homônimos (dois campi da mesma instituição, por
// exemplo) contam como um se a chave for o nome, e o número que chega ao
// avaliador externo sai errado em silêncio — exatamente a classe de
// defeito que este teste existe para prender.
func TestResumoSQL_DistinctSancionadoUsaIdentificadorDoCurso(t *testing.T) {
	consulta := resumoSQL("selectCampos", "from", "where")
	consultaMin := strings.ToLower(consulta)

	ocorrencias := strings.Count(consultaMin, "distinct")
	if ocorrencias != len(allowlistDeDistinctSancionado) {
		t.Fatalf("esperava exatamente %d ocorrência(s) sancionada(s) de DISTINCT na consulta de resumo, encontrou %d:\n%s",
			len(allowlistDeDistinctSancionado), ocorrencias, consulta)
	}

	for _, sancionado := range allowlistDeDistinctSancionado {
		if !strings.Contains(consultaMin, sancionado.trecho) {
			t.Fatalf("esperava a ocorrência sancionada exata de %q (%s), não encontrada — se o DISTINCT mudou de forma, revise a allowlist com o arquiteto:\n%s",
				sancionado.consulta, sancionado.motivo, consulta)
		}
	}

	if strings.Contains(consultaMin, "distinct case when coordenador_id is null then curso_nome") {
		t.Fatal("regressão de T-120: a chave do DISTINCT voltou a ser curso_nome — nome não é identidade, dois cursos homônimos contariam como um")
	}
}

// RD-01, RD-02, RD-03, RD-04: o exigido vem da linha do item, nunca
// dividido entre cursos nem multiplicado pelo número de indicadores da
// meta; pendentes e recusadas não contam no cumprimento.
func TestRelatorioDesempenho_RD01_NaoDivideNemMultiplica(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})

	ind1 := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: instituicaoID})
	ind2 := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: instituicaoID})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID, Indicadores: []uuid.UUID{ind1, ind2}})

	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	montarCurso := func(exigido int) (cursoID, itemID uuid.UUID) {
		cursoID = testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
		planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{
			InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente",
		})
		itemID = testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{
			PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: exigido,
		})
		return
	}

	cursoA, itemA := montarCurso(4)
	_, itemB := montarCurso(2)
	_, itemC := montarCurso(2)

	// Curso A: 2 aceitas, 1 pendente, 1 recusada — cumprimento 50%, nunca
	// 100% (o que aconteceria se a junção com os 2 indicadores dobrasse a
	// contagem de aceitas para 4).
	testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "aceita"})
	testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "aceita"})
	testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "pendente_avaliacao"})
	testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemA, CursoID: cursoA, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "recusada", Rodadas: 1, Motivo: "motivo"})

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	resultado, err := repo.RelatorioDesempenho(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID, PageSize: 100})
	if err != nil {
		t.Fatalf("RelatorioDesempenho: %v", err)
	}

	porItem := map[uuid.UUID]port.LinhaRelatorio{}
	for _, l := range resultado.Itens {
		porItem[l.ItemPlanoID] = l
	}
	if len(porItem) != 3 {
		t.Fatalf("esperava 3 linhas (uma por item), obteve %d", len(porItem))
	}

	linhaA := porItem[itemA]
	if linhaA.Exigido != 4 {
		t.Fatalf("curso A: esperava exigido 4, obteve %d", linhaA.Exigido)
	}
	if linhaA.Aceitas != 2 {
		t.Fatalf("curso A: esperava 2 aceitas (NUNCA 4 — a junção com 2 indicadores não pode dobrar a contagem), obteve %d", linhaA.Aceitas)
	}
	if linhaA.Pendentes != 1 || linhaA.EmCorrecao != 1 {
		t.Fatalf("curso A: esperava 1 pendente e 1 em correção, obteve %d/%d", linhaA.Pendentes, linhaA.EmCorrecao)
	}
	if linhaA.Cumprimento != 50 {
		t.Fatalf("curso A: esperava cumprimento 50%%, obteve %.2f", linhaA.Cumprimento)
	}
	if len(linhaA.Indicadores) != 2 {
		t.Fatalf("curso A: esperava 2 indicadores listados, obteve %d", len(linhaA.Indicadores))
	}

	if porItem[itemB].Exigido != 2 || porItem[itemC].Exigido != 2 {
		t.Fatalf("cursos B/C: esperava exigido 2 em cada, obteve %d/%d — a quantidade NUNCA é dividida entre cursos", porItem[itemB].Exigido, porItem[itemC].Exigido)
	}
}

// RD-10: filtrar por indicador (EXISTS, nunca junção) não duplica a
// linha — uma meta com 2 indicadores casando com o filtro aparece UMA
// vez, nunca duas.
func TestRelatorioDesempenho_RD10_FiltroPorIndicadorNaoDuplica(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})

	ind1 := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: instituicaoID})
	ind2 := testhelpers.CriarIndicadorInstituicao(t, db, testhelpers.OpcoesIndicadorInstituicao{InstituicaoID: instituicaoID})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID, Indicadores: []uuid.UUID{ind1, ind2}})

	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	itemID := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 4})

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	// Filtro por origem "instituicao" — os DOIS indicadores da meta casam.
	resultado, err := repo.RelatorioDesempenho(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID, Origem: "instituicao", PageSize: 100})
	if err != nil {
		t.Fatalf("RelatorioDesempenho (origem): %v", err)
	}
	if n := contarOcorrenciasDoItem(resultado.Itens, itemID); n != 1 {
		t.Fatalf("filtro por origem: esperava a linha UMA vez, apareceu %d vezes — deduplicação explícita, não efeito colateral de DISTINCT", n)
	}

	// Filtro pelo indicador específico.
	resultado, err = repo.RelatorioDesempenho(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID, IndicadorID: &ind1, PageSize: 100})
	if err != nil {
		t.Fatalf("RelatorioDesempenho (indicador): %v", err)
	}
	if n := contarOcorrenciasDoItem(resultado.Itens, itemID); n != 1 {
		t.Fatalf("filtro por indicador: esperava a linha UMA vez, apareceu %d vezes", n)
	}
	if resultado.Total != 1 {
		t.Fatalf("V-4: total da paginação deveria bater com o conjunto filtrado (1), obteve %d", resultado.Total)
	}
}

func contarOcorrenciasDoItem(itens []port.LinhaRelatorio, itemID uuid.UUID) int {
	n := 0
	for _, l := range itens {
		if l.ItemPlanoID == itemID {
			n++
		}
	}
	return n
}

// T-120: o resumo conta cursos sem coordenador pelo IDENTIFICADOR do
// curso, não pelo nome — contra banco real, complementando o teste de
// mecanismo puro em resumoSQL. Dois cursos vagos DIFERENTES (nomes
// distintos, cada um com seu próprio item de plano no mesmo período) têm
// de contar como 2, nunca menos — é a prova de que a contagem funciona
// para o caso comum antes de garantir o caso do nome duplicado (que o
// próprio schema impede via uq_curso_instituicao_nome, e por isso não é
// reproduzível aqui — ver comentário de resumoSQL).
func TestRelatorioDesempenho_T120_ResumoContaCursosVagosPorIdentificador(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})

	montarCursoVago := func() {
		cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
		planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{
			InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente",
		})
		testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{
			PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 2,
		})
		// Nenhuma designação criada — curso nasce vago de propósito.
	}
	montarCursoVago()
	montarCursoVago()

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	resultado, err := repo.RelatorioDesempenho(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID, PageSize: 100})
	if err != nil {
		t.Fatalf("RelatorioDesempenho: %v", err)
	}
	if resultado.Resumo.CursosSemCoordenador != 2 {
		t.Fatalf("esperava 2 cursos sem coordenador no resumo, obteve %d — a contagem por identificador não pode subcontar", resultado.Resumo.CursosSemCoordenador)
	}
}

// design.md §9.4, decisão do dono (ux.md): o caso exato que fundamenta o
// cap por item. Um curso com duas metas de 2 comprovantes cada — 4
// entregues numa, 0 na outra — soma cru daria 4 de 4 (100% com metade das
// metas zerada). Com o cap (min(aceitas, exigido) por item, ANTES de
// somar), o total é 2 de 4 (50%): excesso numa meta não paga dívida em
// outra.
func TestDesempenhoPorCurso_Cap_ItemAcimaDoExigidoNaoElevaPercentualDoCurso(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})
	metaSobrentregue := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	metaZerada := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})

	itemSobrentregue := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaSobrentregue, Quantidade: 2})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaZerada, Quantidade: 2})

	// 4 entregas aceitas para um item que exige só 2 — o excesso NÃO pode
	// compensar a outra meta, zerada.
	for range 4 {
		testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemSobrentregue, CursoID: cursoID, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "aceita"})
	}

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	resultado, err := repo.DesempenhoPorCurso(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID})
	if err != nil {
		t.Fatalf("DesempenhoPorCurso: %v", err)
	}
	if len(resultado.Itens) != 1 {
		t.Fatalf("esperava 1 curso no ranking, obteve %d", len(resultado.Itens))
	}
	item := resultado.Itens[0]
	if item.ExigidoTotal != 4 {
		t.Fatalf("esperava exigido total 4 (2+2), obteve %d", item.ExigidoTotal)
	}
	if item.AceitasTotal != 2 {
		t.Fatalf("esperava aceitas total 2 (CAPADO: min(4,2) + min(0,2) = 2+0), NUNCA 4 (soma crua), obteve %d", item.AceitasTotal)
	}
	if item.Percentual != 50 {
		t.Fatalf("esperava 50%% (2 de 4), NUNCA 100%% (o que a soma crua daria com metade das metas zerada), obteve %.2f", item.Percentual)
	}
}

// design.md §9.4, ux.md "Distinguir sem coordenador / 0% entregue — três
// coisas, nunca a mesma barra": curso COM coordenador que entregou zero é
// desempenho real (0%, barra no ranking); curso vago é ausência de quem
// responda (fora do ranking, lista à parte). Os dois nunca podem colapsar
// no mesmo resultado.
func TestDesempenhoPorCurso_CursoZeroEntregasVersusCursoVago_NaoColapsam(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-07-30"})
	metaID := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	cursoZerado := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: cursoZerado, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID})
	planoZerado := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoZerado, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoZerado, CursoID: cursoZerado, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 2})
	// Nenhuma entrega — 0% real, curso TEM coordenador.

	cursoVago := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	planoVago := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoVago, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
	testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoVago, CursoID: cursoVago, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: 2})
	// Nenhuma designação — curso vago de propósito.

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	resultado, err := repo.DesempenhoPorCurso(context.Background(), esc, port.FiltroRelatorio{PeriodoID: periodoID})
	if err != nil {
		t.Fatalf("DesempenhoPorCurso: %v", err)
	}
	if len(resultado.Itens) != 1 {
		t.Fatalf("esperava 1 curso no RANKING (só o com coordenador), obteve %d", len(resultado.Itens))
	}
	if resultado.Itens[0].Percentual != 0 {
		t.Fatalf("curso com coordenador e zero entregas deveria ter 0%% no ranking, obteve %.2f", resultado.Itens[0].Percentual)
	}
	if resultado.Itens[0].AceitasTotal != 0 || resultado.Itens[0].ExigidoTotal != 2 {
		t.Fatalf("esperava 0 de 2 no curso com coordenador, obteve %d de %d", resultado.Itens[0].AceitasTotal, resultado.Itens[0].ExigidoTotal)
	}
	if len(resultado.CursosSemCoordenador) != 1 {
		t.Fatalf("esperava 1 curso na lista de vagos (FORA do ranking), obteve %d", len(resultado.CursosSemCoordenador))
	}
	if resultado.TotalCursosComCoordenador != 1 || resultado.TotalCursosSemCoordenador != 1 {
		t.Fatalf("esperava 1 com coordenador e 1 sem, obteve %d/%d — os dois casos não podem colapsar no mesmo total",
			resultado.TotalCursosComCoordenador, resultado.TotalCursosSemCoordenador)
	}
}

// T-consistência (pedido do arquiteto): o gráfico (DesempenhoPorCurso) e a
// tabela (RelatorioDesempenho) têm de contar o MESMO universo de cursos
// sob os MESMOS filtros — porque os dois reaproveitam montarRelatorioBase,
// nunca uma segunda consulta com FROM/WHERE próprio. Testado sem filtro E
// com cada um dos três filtros que recortam o conjunto (curso, meta,
// situação) — é sob filtro que uma consulta divergente vazaria primeiro.
func TestDesempenhoPorCurso_MesmoUniversoDaTabelaSobOsMesmosFiltros(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	// Período ENCERRADO antes da data de referência (2026-03-15, fixada em
	// escopoRelatorioDeInstituicao) — item sem entrega aceita cai em
	// "não cumprida" (não em "em_andamento"), o que permite testar o
	// filtro de situação sem depender de prazo em aberto.
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID, DataInicio: "2026-01-01", DataFim: "2026-02-01"})

	metaA := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	metaB := testhelpers.CriarMeta(t, db, testhelpers.OpcoesMeta{InstituicaoID: instituicaoID})
	usuarioID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})

	montarCurso := func(coordenado bool, metaID uuid.UUID, quantidade int, aceitas int) uuid.UUID {
		cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
		if coordenado {
			coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
			testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID})
		}
		planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID, SituacaoPublicacao: "vigente"})
		itemID := testhelpers.CriarItemPlano(t, db, testhelpers.OpcoesItemPlano{PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID, MetaID: metaID, Quantidade: quantidade})
		for range aceitas {
			testhelpers.CriarEntrega(t, db, testhelpers.OpcoesEntrega{ItemPlanoID: itemID, CursoID: cursoID, InstituicaoID: instituicaoID, EnviadaPor: usuarioID, Situacao: "aceita"})
		}
		return cursoID
	}

	// A e B: coordenados, meta A, não cumprem (aceitas < quantidade) —
	// situação "não cumprida". C: coordenado, meta A, cumpre integralmente
	// — situação "cumprida". D: vago, meta B.
	cursoA := montarCurso(true, metaA, 4, 2)
	cursoB := montarCurso(true, metaA, 2, 0)
	cursoC := montarCurso(true, metaA, 1, 1)
	cursoD := montarCurso(false, metaB, 3, 0)
	_ = cursoA

	repo := NovoEntregaRepository(db)
	esc := escopoRelatorioDeInstituicao(t, instituicaoID)

	verificarMesmoUniverso := func(t *testing.T, filtro port.FiltroRelatorio, esperado int) {
		t.Helper()
		filtro.PeriodoID = periodoID
		filtro.PageSize = 100
		relatorio, err := repo.RelatorioDesempenho(context.Background(), esc, filtro)
		if err != nil {
			t.Fatalf("RelatorioDesempenho: %v", err)
		}
		cursosDoRelatorio := map[string]struct{}{}
		for _, l := range relatorio.Itens {
			cursosDoRelatorio[l.CursoNome] = struct{}{}
		}

		porCurso, err := repo.DesempenhoPorCurso(context.Background(), esc, filtro)
		if err != nil {
			t.Fatalf("DesempenhoPorCurso: %v", err)
		}
		totalAgregado := porCurso.TotalCursosComCoordenador + porCurso.TotalCursosSemCoordenador

		if len(cursosDoRelatorio) != esperado {
			t.Fatalf("cenário mal montado: esperava %d cursos distintos na TABELA, obteve %d", esperado, len(cursosDoRelatorio))
		}
		if totalAgregado != esperado {
			t.Fatalf("DIVERGÊNCIA: tabela conta %d cursos distintos, gráfico conta %d (%d com coordenador + %d sem) — os dois têm de contar o MESMO universo sob o mesmo filtro",
				len(cursosDoRelatorio), totalAgregado, porCurso.TotalCursosComCoordenador, porCurso.TotalCursosSemCoordenador)
		}
	}

	t.Run("sem filtro: os 4 cursos (A, B, C, D)", func(t *testing.T) {
		verificarMesmoUniverso(t, port.FiltroRelatorio{}, 4)
	})
	t.Run("filtro por meta: só A, B, C (meta A) — D usa meta B", func(t *testing.T) {
		verificarMesmoUniverso(t, port.FiltroRelatorio{MetaID: &metaA}, 3)
	})
	t.Run("filtro por curso: só D", func(t *testing.T) {
		verificarMesmoUniverso(t, port.FiltroRelatorio{CursoID: &cursoD}, 1)
	})
	t.Run("filtro por situação: só A, B (não cumprida) — C cumpriu, D é sem_responsavel", func(t *testing.T) {
		verificarMesmoUniverso(t, port.FiltroRelatorio{Situacao: "nao_cumprida"}, 2)
	})
	_ = cursoB
	_ = cursoC
}
