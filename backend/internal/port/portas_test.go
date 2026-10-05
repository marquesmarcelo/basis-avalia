package port_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

// Os antigos guardas 1 e 3 ("TestRepositoriosDeNegocio_TodoMetodoExigeEscopo"
// e "TestNenhumaPortaNovaSemEscopo") saíram deste arquivo — cada um
// enumerava à mão QUAIS interfaces eram verificadas, e uma porta nova que
// ninguém acrescentasse à lista nunca era verificada, com o teste
// permanecendo verde (design.md de autenticacao-usuarios §4.4, revisão 7,
// T-110). Os dois foram substituídos por
// TestGuardaA_PortaDeNegocioExigeEscopo, em guarda_estrutural_test.go, que
// varre TODA interface exportada do pacote por análise estática
// (go/packages + go/types) e só deixa de verificar o que estiver
// explicitamente dispensado.

// TestPortasEstreitas_ListaFechada — guarda B (design.md §4.4, §16 T-102,
// T-111, achado O-1): as três portas sem Escopo têm exatamente os métodos
// declarados no design, **e a lista exata de tipos de parâmetro de cada
// um**. Antes deste guarda, o teste só comparava nomes — trocar o que
// BuscarCredencial recebe, mantendo o nome, não quebrava nada. Acrescentar
// método, remover, ou trocar um tipo de parâmetro quebra este teste e exige
// decisão do arquiteto com alteração de design.md §4.4 no mesmo commit.
func TestPortasEstreitas_ListaFechada(t *testing.T) {
	tipoCtx := reflect.TypeOf((*context.Context)(nil)).Elem()
	tipoUUID := reflect.TypeOf(uuid.UUID{})
	tipoUUIDPtr := reflect.TypeOf(&uuid.UUID{})
	tipoEmail := reflect.TypeOf(valueobject.Email{})
	tipoProprio := reflect.TypeOf(autorizacao.Proprio{})
	tipoSenhaHash := reflect.TypeOf(valueobject.SenhaHash{})
	tipoTime := reflect.TypeOf(time.Time{})
	tipoDataLocal := reflect.TypeOf(valueobject.DataLocal{})
	tipoInt := reflect.TypeOf(0)
	tipoString := reflect.TypeOf("")

	casos := []struct {
		nome      string
		it        reflect.Type
		esperados map[string][]reflect.Type
	}{
		{
			nome: "AutenticacaoRepository",
			it:   reflect.TypeOf((*port.AutenticacaoRepository)(nil)).Elem(),
			esperados: map[string][]reflect.Type{
				"BuscarCredencial":         {tipoCtx, tipoUUIDPtr, tipoEmail},
				"CarregarContextoDeSessao": {tipoCtx, tipoUUID, tipoDataLocal},
				"BuscarCredencialPropria":  {tipoCtx, tipoProprio},
				"DefinirSenhaPropria":      {tipoCtx, tipoProprio, tipoSenhaHash, tipoTime},
				"InvalidarSessoesProprias": {tipoCtx, tipoProprio, tipoTime},
			},
		},
		{
			nome: "InstituicaoPublicaQuery",
			it:   reflect.TypeOf((*port.InstituicaoPublicaQuery)(nil)).Elem(),
			esperados: map[string][]reflect.Type{
				"ListarParaCombo": {tipoCtx},
			},
		},
		{
			// Terceira porta estreita (T-111, design.md §4.4, D-27): o
			// relê em segundo plano não tem sessão, ator nem instituição
			// — varre todas por desenho. Guarda C (em
			// guarda_estrutural_test.go) garante que ela nunca é
			// alcançável por adapter/http.
			nome: "EntregaReleRepository",
			it:   reflect.TypeOf((*port.EntregaReleRepository)(nil)).Elem(),
			esperados: map[string][]reflect.Type{
				// T-124/T-127 (design.md §8.4, revisão pós code-review):
				// reivindicação atômica (UPDATE...RETURNING) substitui o
				// antigo SELECT...FOR UPDATE SKIP LOCKED isolado, que não
				// protegia nada em autocommit; `hoje` (DataLocal, fuso de
				// exibição) decide quem recebe — nunca UTC (DG-06).
				"ReivindicarNotificacoesPendentes":    {tipoCtx, tipoDataLocal, tipoInt},
				"ContarNotificacoesPendentes":         {tipoCtx},
				"MarcarNotificacaoEnviada":            {tipoCtx, tipoUUID},
				"RegistrarFalhaDeNotificacao":         {tipoCtx, tipoUUID, tipoString},
				"ListarCandidatosARestauracaoDePrazo": {tipoCtx, tipoInt},
				"RestaurarPrazoPorVacancia":           {tipoCtx, tipoUUID, tipoTime},
				"ExpurgarIdempotenciaAntesDe":         {tipoCtx, tipoTime},
			},
		},
	}

	for _, c := range casos {
		encontrados := make(map[string]bool, c.it.NumMethod())
		for i := 0; i < c.it.NumMethod(); i++ {
			m := c.it.Method(i)
			encontrados[m.Name] = true

			esperado, existe := c.esperados[m.Name]
			if !existe {
				t.Errorf("%s: método novo %q não está na lista fechada de design.md §4.4 — exige decisão do arquiteto", c.nome, m.Name)
				continue
			}
			if m.Type.NumIn() != len(esperado) {
				t.Errorf("%s.%s: esperava %d parâmetros, tem %d — ver design.md §4.4", c.nome, m.Name, len(esperado), m.Type.NumIn())
				continue
			}
			for p, tipoEsperado := range esperado {
				if m.Type.In(p) != tipoEsperado {
					t.Errorf("%s.%s: parâmetro %d é %s, esperava %s — assinatura fixada por design.md §4.4",
						c.nome, m.Name, p, m.Type.In(p), tipoEsperado)
				}
			}
		}
		for nome := range c.esperados {
			if !encontrados[nome] {
				t.Errorf("%s: método esperado %q não existe mais — ver design.md §4.4", c.nome, nome)
			}
		}
	}
}
