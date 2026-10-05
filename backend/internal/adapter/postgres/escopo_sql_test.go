package postgres

import (
	"strings"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

func conjuntoTesteEscopo(t *testing.T, p valueobject.Perfil) valueobject.ConjuntoDePerfis {
	t.Helper()
	c, err := valueobject.NovoConjunto(p)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	return c
}

func TestAplicarEscopo_ValorZeroDevolveErro(t *testing.T) {
	var vazio autorizacao.Escopo
	_, _, err := AplicarEscopo(vazio, AlvoUsuario, 1)
	if err != domain.ErrEscopoInvalido {
		t.Fatalf("esperava ErrEscopoInvalido, obtido %v", err)
	}
}

// TestAplicarEscopo_PlataformaExigeInstituicaoNula prova design.md §4.3/§16
// (T-101, achado M-3): o ramo de plataforma NÃO é opcional. Sem esta
// cláusula, o recorte de plataforma saía certo por coincidência (o único
// alcance de plataforma que chegava aqui sempre trazia exigePerfil) — o
// primeiro alcance de plataforma sem exigePerfil vazaria todos os usuários
// de todas as instituições. Critério de comprovação negativa: remover a
// cláusula "usuario.instituicao_id IS NULL" de AplicarEscopo faz este teste
// falhar.
func TestAplicarEscopo_PlataformaExigeInstituicaoNula(t *testing.T) {
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.InstituicoesDaPlataforma, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "usuario.instituicao_id IS NULL") {
		t.Fatalf("escopo de plataforma precisa emitir instituicao_id IS NULL: %s", clausula)
	}
	if strings.Contains(clausula, "usuario.instituicao_id = $") {
		t.Fatalf("escopo de plataforma não pode emitir instituicao_id = $N (isso é escopo institucional): %s", clausula)
	}
	if len(args) != 0 {
		t.Fatalf("instituicao_id IS NULL não usa placeholder — esperava zero argumentos, obtido %v", args)
	}
}

func TestAplicarEscopo_InstitucionalSempreEmiteInstituicaoID(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjuntoTesteEscopo(t, valueobject.PesquisadorInstitucional), &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.UsuariosDaPropriaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "usuario.instituicao_id = $1") {
		t.Fatalf("escopo institucional deveria emitir instituicao_id: %s", clausula)
	}
	if len(args) < 1 {
		t.Fatal("esperava ao menos um argumento")
	}
}

// TestAplicarEscopo_SemExcecaoForaDoIndicador prova fundacao-metas.md §3.7:
// para qualquer alvo que NÃO seja AlvoIndicador (aqui, AlvoUsuario), um
// escopo institucional emite só "instituicao_id = $N" — nunca o OR da
// exceção do catálogo comum, que é exclusiva de quem declara
// admiteCatalogoComum.
func TestAplicarEscopo_SemExcecaoForaDoIndicador(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjuntoTesteEscopo(t, valueobject.PesquisadorInstitucional), &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.UsuariosDaPropriaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, _, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "usuario.instituicao_id = $1") {
		t.Fatalf("esperava instituicao_id = $1: %s", clausula)
	}
	if strings.Contains(clausula, " OR ") {
		t.Fatalf("AlvoUsuario não admite a exceção do catálogo comum, não deveria ter OR: %s", clausula)
	}
}

// TestAplicarEscopo_ExcecaoDoCatalogoTemIsNullEParenteses prova
// fundacao-metas.md §3.3: no alvo que declara admiteCatalogoComum, a
// cláusula institucional é a UNIÃO de duas condições, disjuntas e entre
// parênteses — nunca "instituicao_id = $N OR instituicao_id IS NULL" sem
// o parêntese, que escaparia do "excluido_em IS NULL".
func TestAplicarEscopo_ExcecaoDoCatalogoTemIsNullEParenteses(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjuntoTesteEscopo(t, valueobject.PesquisadorInstitucional), &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.UsuariosDaPropriaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, args, err := AplicarEscopo(escopo, AlvoIndicador, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "indicador.instituicao_id = $1") {
		t.Fatalf("esperava o ramo institucional: %s", clausula)
	}
	if !strings.Contains(clausula, "indicador.instituicao_id IS NULL") {
		t.Fatalf("esperava o ramo do catálogo comum (IS NULL): %s", clausula)
	}
	if !strings.Contains(clausula, "indicador.escopo = 'plataforma'") {
		t.Fatalf("esperava a condição de intenção escopo = 'plataforma': %s", clausula)
	}
	// O parêntese externo do OR: procura o padrão literal "( ... OR ( ... ) )".
	if !strings.Contains(clausula, "( indicador.instituicao_id = $1 OR ( indicador.instituicao_id IS NULL AND indicador.escopo = 'plataforma' ) )") {
		t.Fatalf("esperava o OR entre parênteses, exatamente: %s", clausula)
	}
	if len(args) != 1 {
		t.Fatalf("esperava 1 argumento (a instituição), obtido %v", args)
	}
}

// TestAplicarEscopo_SemExcecaoParaMeta prova fundacao-metas.md §3.7 e
// specs/indicadores/spec.md IV-05 para o alvo que esta feature acrescenta:
// meta continua com o filtro SEM exceção alguma — a exceção do catálogo
// comum é exclusiva de AlvoIndicador. Falha se alguém ampliar
// admiteCatalogoComum para AlvoMeta.
func TestAplicarEscopo_SemExcecaoParaMeta(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), conjuntoTesteEscopo(t, valueobject.PesquisadorInstitucional), &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.MetasDaInstituicao, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, _, err := AplicarEscopo(escopo, AlvoMeta, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "meta.instituicao_id = $1") {
		t.Fatalf("esperava meta.instituicao_id = $1: %s", clausula)
	}
	if strings.Contains(clausula, " OR ") {
		t.Fatalf("AlvoMeta não admite a exceção do catálogo comum, não deveria ter OR: %s", clausula)
	}
}

// TestAplicarEscopo_ExigePerfilEmiteExistsDePosse prova a cláusula nova de
// design.md §4.3: o alcance de pesquisadores exige que o alvo POSSUA
// pesquisador_institucional via EXISTS sobre a tabela de vínculo, nunca
// "perfil = valor" — a coluna não existe mais.
func TestAplicarEscopo_ExigePerfilEmiteExistsDePosse(t *testing.T) {
	caminhoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.PesquisadoresDeUmaInstituicao, autorizacao.AcaoListar, &caminhoID)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	clausula, args, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "EXISTS (SELECT 1 FROM usuario_perfil up WHERE up.usuario_id = usuario.id AND up.perfil = $2)") {
		t.Fatalf("esperava EXISTS de posse do perfil exigido: %s", clausula)
	}
	if len(args) != 2 || args[1] != "pesquisador_institucional" {
		t.Fatalf("esperava o perfil exigido como último argumento: %v", args)
	}
}

// TestAplicarEscopo_EscopoIncompativelComAlvoFalha prova fundacao-metas.md
// §3.6: Escopo incompatível com o Alvo é erro, nunca ramo ausente — um
// Escopo com exigePerfil aplicado a um alvo que não admite isso (só
// AlvoUsuario admite) precisa falhar, não silenciosamente omitir o EXISTS
// (o que devolveria todo o catálogo, sem filtrar por posse).
func TestAplicarEscopo_EscopoIncompativelComAlvoFalha(t *testing.T) {
	caminhoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(uuid.Must(uuid.NewV7()), valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	// PesquisadoresDeUmaInstituicao liga exigePerfil = pesquisador_institucional.
	escopo, err := autorizacao.Autorizar(ator, autorizacao.PesquisadoresDeUmaInstituicao, autorizacao.AcaoListar, &caminhoID)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}

	// AlvoIndicador não admite exigePerfil (só AlvoUsuario admite).
	_, _, err = AplicarEscopo(escopo, AlvoIndicador, 1)
	if err != domain.ErrEscopoInvalido {
		t.Fatalf("esperava ErrEscopoInvalido para exigePerfil fora de AlvoUsuario, obtido %v", err)
	}
}

// escopoDeCarteiraTeste constrói um Escopo com RestritoACarteiraDe ligado
// — via CursosDaCarteira, o alcance real que o produz (fundacao-metas.md
// §3.4, specs/cursos/design.md §5.1).
func escopoDeCarteiraTeste(t *testing.T, coordenadorID uuid.UUID) autorizacao.Escopo {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(valueobject.CoordenadorCurso)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	instituicaoID := uuid.Must(uuid.NewV7())
	ator, err := autorizacao.NovoAtor(coordenadorID, conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	escopo, err := autorizacao.Autorizar(ator, autorizacao.CursosDaCarteira, autorizacao.AcaoListar, nil)
	if err != nil {
		t.Fatalf("Autorizar: %v", err)
	}
	return escopo
}

// TestAplicarEscopo_CarteiraEmAlvoSemCursoFalha completa o enunciado de
// fundacao-metas.md §3.6 que TestAplicarEscopo_EscopoIncompativelComAlvoFalha
// deixou pendente: um Escopo restrito à carteira aplicado a um alvo sem
// coluna de curso (AlvoUsuario) é erro, nunca ramo ausente — omitir o
// EXISTS devolveria todos os usuários da instituição para um coordenador.
func TestAplicarEscopo_CarteiraEmAlvoSemCursoFalha(t *testing.T) {
	escopo := escopoDeCarteiraTeste(t, uuid.Must(uuid.NewV7()))
	_, _, err := AplicarEscopo(escopo, AlvoUsuario, 1)
	if err != domain.ErrEscopoInvalido {
		t.Fatalf("esperava ErrEscopoInvalido para carteira em alvo sem coluna de curso, obtido %v", err)
	}
}

// TestAplicarEscopo_CarteiraUsaOFragmentoUnico prova T-166 (specs/cursos/
// tasks.md): a cláusula de carteira embute, byte a byte,
// FragmentoDesignacaoVigente — nunca uma reescrita paralela da regra de
// vigência.
func TestAplicarEscopo_CarteiraUsaOFragmentoUnico(t *testing.T) {
	coordenadorID := uuid.Must(uuid.NewV7())
	escopo := escopoDeCarteiraTeste(t, coordenadorID)

	clausula, args, err := AplicarEscopo(escopo, AlvoCurso, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	fragmentoEsperado := FragmentoDesignacaoVigente("carteira_d", 3)
	if !strings.Contains(clausula, fragmentoEsperado) {
		t.Fatalf("esperava o fragmento único de vigência, byte a byte: %s\nfragmento: %s", clausula, fragmentoEsperado)
	}
	if !strings.Contains(clausula, "carteira_d.curso_id = curso.id") {
		t.Fatalf("para AlvoCurso, a coluna correlacionada deveria ser curso.id: %s", clausula)
	}
	if len(args) != 3 { // instituicao_id, coordenador_id, data
		t.Fatalf("esperava 3 argumentos (instituição, coordenador, data), obtido %d: %v", len(args), args)
	}
	if args[1] != coordenadorID {
		t.Fatalf("esperava o id do coordenador como argumento, obtido %v", args[1])
	}
}

// TestAplicarEscopo_CarteiraEmAlvoDescendenteUsaCursoID prova que, fora de
// AlvoCurso, a coluna correlacionada é <alias>.curso_id — a diferença fica
// dentro de AplicarEscopo, nunca no chamador (specs/cursos/design.md §3.2).
func TestAplicarEscopo_CarteiraEmAlvoDescendenteUsaCursoID(t *testing.T) {
	escopo := escopoDeCarteiraTeste(t, uuid.Must(uuid.NewV7()))
	clausula, _, err := AplicarEscopo(escopo, AlvoDesignacao, 1)
	if err != nil {
		t.Fatalf("AplicarEscopo: %v", err)
	}
	if !strings.Contains(clausula, "carteira_d.curso_id = designacao.curso_id") {
		t.Fatalf("esperava a coluna correlacionada designacao.curso_id: %s", clausula)
	}
}
