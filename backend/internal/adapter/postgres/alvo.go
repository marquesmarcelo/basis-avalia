package postgres

// Alvo é a lista FECHADA de tabelas que passam pelo filtro de isolamento
// de AplicarEscopo (fundacao-metas.md §3.1, §3.2). Acrescentar um Alvo é
// decisão do arquiteto e exige alterar a fundação no mesmo commit — o
// teste TestAlvos_ExcecaoDoCatalogoEmExatamenteUm percorre esta lista.
type Alvo struct {
	alias               string
	temColunaCurso      bool // entidades abaixo de curso na cadeia (F-05)
	admiteCatalogoComum bool // VERDADEIRO EM EXATAMENTE UM (F-02)
	admiteExigePerfil   bool // só usuario: o EXISTS de posse não existe fora dele
}

var (
	AlvoUsuario   = Alvo{alias: "usuario", admiteExigePerfil: true}
	AlvoIndicador = Alvo{alias: "indicador", admiteCatalogoComum: true}
	AlvoMeta      = Alvo{alias: "meta"}
	// AlvoCurso tem temColunaCurso: true porque a coluna é a própria id —
	// o fragmento de carteira usa curso.id; nos demais, <alias>.curso_id.
	// A diferença fica dentro de AplicarEscopo, não no chamador
	// (specs/cursos/design.md §4.3).
	AlvoCurso      = Alvo{alias: "curso", temColunaCurso: true}
	AlvoDesignacao = Alvo{alias: "designacao", temColunaCurso: true}

	// Acrescentados por specs/plano-acao/design.md §3 — AlvoPeriodo é da
	// instituição, sem coluna de curso (um período vale para todos os
	// cursos dela); os demais estão abaixo de curso na cadeia (F-05) e
	// usam <alias>.curso_id no fragmento de carteira.
	AlvoPeriodo   = Alvo{alias: "periodo"}
	AlvoPlano     = Alvo{alias: "plano", temColunaCurso: true}
	AlvoItemPlano = Alvo{alias: "item_plano", temColunaCurso: true}
	AlvoDocumento = Alvo{alias: "documento", temColunaCurso: true}

	// Acrescentados por specs/metas-coordenacao/design.md §3 — a última
	// feature da cadeia de metas. Entrega e anexo estão abaixo de curso
	// (F-05): usam <alias>.curso_id no fragmento de carteira.
	AlvoEntrega = Alvo{alias: "entrega", temColunaCurso: true}
	AlvoAnexo   = Alvo{alias: "anexo", temColunaCurso: true}
)

// todosOsAlvos existe só para o teste de mecanismo percorrer a lista
// fechada sem precisar saber os nomes das variáveis — acrescentar um Alvo
// sem acrescentá-lo aqui quebra TestAlvos_ExcecaoDoCatalogoEmExatamenteUm
// por omissão, não por falso negativo.
var todosOsAlvos = []Alvo{
	AlvoUsuario, AlvoIndicador, AlvoMeta, AlvoCurso, AlvoDesignacao,
	AlvoPeriodo, AlvoPlano, AlvoItemPlano, AlvoDocumento,
	AlvoEntrega, AlvoAnexo,
}
