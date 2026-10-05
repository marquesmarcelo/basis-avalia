package port_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// Este arquivo substitui os antigos guardas 1 e 3 de portas_test.go
// (design.md de autenticacao-usuarios §4.4, revisão 7, T-110 a T-112).
//
// O defeito medido em 29/09: os guardas antigos enumeravam à mão QUAIS
// portas eram verificadas — sete repositórios de negócio no guarda 1,
// nove interfaces na lista `permitidas` do guarda 3. Uma porta nova que
// ninguém acrescentasse a essas listas nunca era verificada, e os testes
// continuavam verdes afirmando uma cobertura que não tinham (23 portas
// declaradas, 14 verificadas). Isso é o único ponto do desenho inteiro em
// que esquecer produz SILÊNCIO em vez de falha — toda outra garantia do
// sistema (Escopo obrigatório na assinatura, Escopo inconstruível fora de
// Autorizar, CHECK do banco, ramo de plataforma em AplicarEscopo) quebra
// alto quando alguém esquece.
//
// A correção inverte o sentido da lista: em vez de enumerar o que É
// verificado, o Guarda A varre TODAS as interfaces exportadas do pacote
// (via go/packages + go/types — a mesma técnica que
// TestEscopoInconstruivelForaDeAutorizar já usa, aqui com informação de
// tipo porque precisamos saber se um parâmetro É uuid.UUID ou entidade de
// domínio, não só a sintaxe) e só deixa de verificar o que estiver
// EXPLICITAMENTE dispensado — dispensaDeInterface e dispensaDeMetodo,
// abaixo. Porta nova não listada é verificada por padrão; quem quiser
// dispensa tem de escrevê-la aqui, com justificativa, no mesmo commit.

const (
	pacotePort         = "github.com/basis-avalia/backend/internal/port"
	caminhoUUID        = "github.com/google/uuid"
	caminhoAutorizacao = "github.com/basis-avalia/backend/internal/domain/autorizacao"
	prefixoDominio     = "github.com/basis-avalia/backend/internal/domain/"
)

// dispensaDeInterface — a lista fechada de portas estreitas (design.md
// §4.4, D-27): exatamente as três que Guarda B verifica assinatura por
// assinatura. Uma interface nova só entra aqui com a mesma seção do
// design.md alterada no mesmo commit.
var dispensaDeInterface = map[string]bool{
	"AutenticacaoRepository":  true,
	"InstituicaoPublicaQuery": true,
	"EntregaReleRepository":   true,
}

// dispensaDeMetodo — exceção MÉTODO A MÉTODO dentro de uma interface que
// continua sendo verificada nos demais métodos. Vazia hoje: as quatro
// dispensas propostas ao longo desta tarefa (CoordenaCursoHoje,
// InserirIdempotencia, BuscarEntregaPorChaveIdempotencia e o avaliadorID
// de BuscarPorID) eram, todas, o mesmo caso — um identificador que é o
// PRÓPRIO ator autenticado, nunca um terceiro — e autorizacao.Proprio já
// existia para isso. Nenhuma das quatro precisou de dispensa; precisou de
// leitura do ponto de chamada. Mantida declarada (em vez de removida)
// porque o mecanismo — método dispensável dentro de uma interface só
// parcialmente dispensada — é legítimo e pode ser necessário de novo; o
// que este ciclo mostrou é que a barra para usá-lo é mais alta do que
// pareceu da primeira vez: verificar a proveniência real antes de
// declarar "é só uma checagem de fato".
var dispensaDeMetodo = map[string]map[string]bool{}

// palavrasTripwireDeIdentidade — nomes de parâmetro que sugerem "id de uma
// PESSOA", usados pelo segundo guarda de T-118, abaixo. Uma palavra nova
// entra aqui quando um caso real aparecer — a lista cresce por incidente
// observado, não por antecipação.
var palavrasTripwireDeIdentidade = []string{
	"autor", "usuario", "ator", "avaliador", "coordenador", "criador", "responsavel",
}

// dispensaDeTripwire — igual a dispensaDeMetodo, mas só para o guarda de
// nome (abaixo). Separada porque as duas listas respondem perguntas
// diferentes: dispensaDeMetodo é "este método não precisa de Escopo/Proprio
// na posição 1"; esta é "este parâmetro, fora da posição 1, tem um nome
// que soa como identidade mas não é".
//
// UsuarioRepository.SubstituirPerfis(ctx, escopo, usuarioID, perfis):
// usuarioID não é o ator — é o alvo da edição (alvo.ID em
// AtualizarUsuarioUseCase.Executar, sempre um usuário diferente de
// in.Ator: há checagem explícita alvo.ID == in.Ator.UsuarioID() logo
// acima, que barra a auto-edição de perfil com E-04). É identificador de
// recurso de negócio de verdade, já recortado por escopo — o mesmo papel
// que anexoID tem em RemoverAnexo. O nome só soa como identidade porque
// "usuário" é, ao mesmo tempo, o nome da pessoa e o nome do recurso.
var dispensaDeTripwire = map[string]map[string]bool{
	"UsuarioRepository": {"SubstituirPerfis": true},
}

func carregarPacotePort(t *testing.T) *packages.Package {
	t.Helper()
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedName}
	pkgs, err := packages.Load(cfg, pacotePort)
	if err != nil {
		t.Fatalf("packages.Load(%s): %v", pacotePort, err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("esperava carregar exatamente 1 pacote, obtive %d", len(pkgs))
	}
	pkg := pkgs[0]
	for _, e := range pkg.Errors {
		t.Fatalf("erro ao carregar %s: %v", pacotePort, e)
	}
	return pkg
}

// ehEntidadeDeDominio — pacote de domínio que representa entidade sob
// controle de um repositório (Meta, Plano, Curso, Usuario...), nunca:
//   - domain/autorizacao (Escopo, Ator, Alcance — mecanismo, não identificador);
//   - domain/valueobject (Email, DataLocal, SenhaHash, Perfil... — usados
//     livremente pelas portas estreitas, não carregam identidade sozinhos);
//   - domain/auditoria (Evento é o PAYLOAD de um log — AuditLogger é
//     infraestrutura por design.md §4.2, e Evento carrega seus próprios
//     ids internamente sem que o MÉTODO Registrar esteja "operando sobre
//     um recurso" — é a mesma razão pela qual DadosDoDocumento e
//     ItemDoDocumento, ambos de port puro, também não entram aqui).
func ehEntidadeDeDominio(caminhoPacote string) bool {
	if !strings.HasPrefix(caminhoPacote, prefixoDominio) {
		return false
	}
	resto := strings.TrimPrefix(caminhoPacote, prefixoDominio)
	for _, excluido := range []string{"autorizacao", "valueobject", "auditoria"} {
		if resto == excluido || strings.HasPrefix(resto, excluido+"/") {
			return false
		}
	}
	return true
}

func semPonteiro(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}

// ehIdentificadorDeNegocio — design.md §4.4, Guarda A: uuid.UUID (ou
// ponteiro) ou entidade de domínio. É o gatilho que exige Escopo.
func ehIdentificadorDeNegocio(t types.Type) bool {
	named, ok := semPonteiro(t).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}
	if pkg.Path() == caminhoUUID && obj.Name() == "UUID" {
		return true
	}
	return ehEntidadeDeDominio(pkg.Path())
}

// ehTipoDeIdentidadeVerificada — Escopo OU Proprio: os dois únicos tipos
// deste sistema que só nascem depois de uma verificação (Autorizar ou
// Ator.Proprio(), ambos inconstruíveis fora dali). Escopo verifica "este
// recurso, desta instituição/carteira"; Proprio verifica "este ator,
// sobre si mesmo" — a mesma exigência de proveniência do design.md §4.4,
// só que na outra dimensão. Um método com identificador de negócio pode
// satisfazer o Guarda A com qualquer um dos dois, nunca com um uuid.UUID
// cru.
func ehTipoDeIdentidadeVerificada(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != caminhoAutorizacao {
		return false
	}
	return obj.Name() == "Escopo" || obj.Name() == "Proprio"
}

// ehTipoUUIDCru — uuid.UUID ou *uuid.UUID, sem passar por
// ehIdentificadorDeNegocio (que também aceita entidade de domínio): o
// tripwire de nome (T-118, abaixo) só faz sentido para o tipo que não
// carrega NENHUMA informação própria além de dezesseis bytes — é
// exatamente por isso que dois parâmetros do mesmo tipo (anexoID,
// autorID) só se distinguem pelo nome.
func ehTipoUUIDCru(t types.Type) bool {
	named, ok := semPonteiro(t).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == caminhoUUID && obj.Name() == "UUID"
}

func nomeSugereIdentidadeDePessoa(nomeParametro string) bool {
	nomeMinusculo := strings.ToLower(nomeParametro)
	for _, palavra := range palavrasTripwireDeIdentidade {
		if strings.Contains(nomeMinusculo, palavra) {
			return true
		}
	}
	return false
}

// TestGuardaA_PortaDeNegocioExigeEscopo é o guarda fail-closed: varre TODA
// interface exportada de internal/port — nenhuma lista de "o que
// verificar", só a lista de dispensa acima (design.md §4.4, T-110).
func TestGuardaA_PortaDeNegocioExigeEscopo(t *testing.T) {
	pkg := carregarPacotePort(t)
	scope := pkg.Types.Scope()

	interfacesVistas := 0
	for _, nome := range scope.Names() {
		obj := scope.Lookup(nome)
		tn, ok := obj.(*types.TypeName)
		if !ok || !tn.Exported() {
			continue
		}
		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}
		interfacesVistas++

		if dispensaDeInterface[nome] {
			continue
		}

		for i := 0; i < iface.NumMethods(); i++ {
			metodo := iface.Method(i)
			sig, ok := metodo.Type().(*types.Signature)
			if !ok {
				continue
			}
			params := sig.Params()

			precisaDeEscopo := false
			for j := 0; j < params.Len(); j++ {
				if ehIdentificadorDeNegocio(params.At(j).Type()) {
					precisaDeEscopo = true
					break
				}
			}
			if !precisaDeEscopo {
				continue
			}
			if dispensaDeMetodo[nome][metodo.Name()] {
				continue
			}

			if params.Len() < 2 || !ehTipoDeIdentidadeVerificada(params.At(1).Type()) {
				t.Errorf(
					"port.%s.%s recebe identificador de negócio sem autorizacao.Escopo (ou autorizacao.Proprio) como segundo parâmetro (logo após ctx) — design.md §4.4, Guarda A. "+
						"Se for uma porta estreita de verdade, registre em design.md §4.4 e na dispensa deste arquivo, no mesmo commit.",
					nome, metodo.Name())
			}
		}
	}

	totalMetodosDispensados := 0
	for _, metodos := range dispensaDeMetodo {
		totalMetodosDispensados += len(metodos)
	}
	t.Logf("Guarda A: %d interfaces exportadas classificadas em internal/port (%d dispensadas por inteiro, %d método(s) com dispensa individual)",
		interfacesVistas, len(dispensaDeInterface), totalMetodosDispensados)
}

// TestGuardaA_TripwireDeIdentidadeForaDaPosicaoUm (T-118) — segunda metade
// do Guarda A, e MAIS FRACA que a primeira: leia isto antes de confiar
// nela.
//
// A posição 1 (imediatamente após o ctx) é verificada POR TIPO, em
// TestGuardaA_PortaDeNegocioExigeEscopo: um uuid.UUID cru ali NUNCA passa,
// não importa como se chama. Tipo não se disfarça.
//
// As DEMAIS posições não têm essa garantia, e não podem ter com a técnica
// atual: o compilador não distingue um `anexoID uuid.UUID` (identificador
// de RECURSO, legítimo — o próprio Escopo, na posição 1, já o recorta) de
// um `autorID uuid.UUID` (identidade de PESSOA, que deveria ser Proprio) —
// os dois são o mesmo tipo Go. O caso real que motivou este teste
// (EntregaRepository.RemoverAnexo, T-118) era exatamente isso: `anexoID`
// na posição 2 e `autorID` na posição 3, ambos uuid.UUID, só o segundo
// errado.
//
// Por isso este guarda é um TRIPWIRE POR NOME, não uma prova: reprova um
// parâmetro fora da posição 1 se, ao MESMO TEMPO, (a) o tipo é uuid.UUID
// cru e (b) o NOME contém uma palavra de palavrasTripwireDeIdentidade
// (acima). Um parâmetro de identidade batizado para escapar — `x
// uuid.UUID`, `id2`, `quemPediu` — passa por este guarda sem ser pego.
// Isto é aceito conscientemente: o mecanismo forte de verdade (um tipo
// Go distinto para "identidade de pessoa", em vez de uuid.UUID cru em
// TODO lugar) é refatoração larga para fechar uma fresta estreita, e não
// foi feita agora.
//
// Gatilho registrado: se este tripwire deixar passar um caso real que só
// seja pego em revisão humana, a resposta é criar esse tipo distinto — e
// aí a verificação volta a ser por tipo em todas as posições, como já é
// na posição 1.
func TestGuardaA_TripwireDeIdentidadeForaDaPosicaoUm(t *testing.T) {
	pkg := carregarPacotePort(t)
	scope := pkg.Types.Scope()

	parametrosInspecionados := 0
	for _, nome := range scope.Names() {
		obj := scope.Lookup(nome)
		tn, ok := obj.(*types.TypeName)
		if !ok || !tn.Exported() {
			continue
		}
		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		for i := 0; i < iface.NumMethods(); i++ {
			metodo := iface.Method(i)
			sig, ok := metodo.Type().(*types.Signature)
			if !ok {
				continue
			}
			params := sig.Params()
			if dispensaDeTripwire[nome][metodo.Name()] {
				continue
			}
			for j := 2; j < params.Len(); j++ {
				param := params.At(j)
				parametrosInspecionados++
				if !ehTipoUUIDCru(param.Type()) {
					continue
				}
				if !nomeSugereIdentidadeDePessoa(param.Name()) {
					continue
				}
				t.Errorf(
					"port.%s.%s: parâmetro %q (posição %d) tem nome de identidade de pessoa mas é uuid.UUID cru — "+
						"use autorizacao.Proprio (T-118, mesmo caso de EntregaRepository.RemoverAnexo). "+
						"Se for mesmo um identificador de recurso, registre a dispensa em dispensaDeTripwire, no mesmo commit.",
					nome, metodo.Name(), param.Name(), j)
			}
		}
	}

	t.Logf("Guarda A (tripwire de nome): %d parâmetro(s) fora da posição 1 inspecionados em internal/port — "+
		"tripwire por nome, não prova de tipo (ver comentário desta função)", parametrosInspecionados)
}

// TestGuardaC_AdapterHTTPNaoReferenciaPortaDeManutencao prova, por análise
// estática (mesma técnica de TestEscopoInconstruivelForaDeAutorizar), que
// nenhum arquivo de internal/adapter/http referencia
// port.EntregaReleRepository — a porta de manutenção do relê nunca pode
// ser alcançável por uma rota HTTP (design.md §4.4, D-27, T-112).
func TestGuardaC_AdapterHTTPNaoReferenciaPortaDeManutencao(t *testing.T) {
	const nomePortaDeManutencao = "EntregaReleRepository"
	const diretorioHTTP = "../adapter/http"

	fset := token.NewFileSet()
	pacotes, err := parser.ParseDir(fset, diretorioHTTP, func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("ParseDir(%s): %v", diretorioHTTP, err)
	}

	for _, pacote := range pacotes {
		for nomeArquivo, arquivo := range pacote.Files {
			ast.Inspect(arquivo, func(n ast.Node) bool {
				// *ast.Ident cobre os dois casos que importam: o Sel de
				// um SelectorExpr (port.EntregaReleRepository) É um
				// *ast.Ident e é visitado como nó próprio por ast.Inspect
				// — checar SelectorExpr também contaria a mesma
				// referência duas vezes.
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == nomePortaDeManutencao {
					t.Errorf("%s referencia %s — porta de manutenção não pode ser alcançada por adapter/http (design.md §4.4, Guarda C)",
						nomeArquivo, nomePortaDeManutencao)
				}
				return true
			})
		}
	}
}
