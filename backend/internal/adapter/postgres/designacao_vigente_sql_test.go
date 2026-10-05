package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/testhelpers"
	"github.com/google/uuid"
)

// TestVigencia_RotuloEPredicadoConcordam prende o rótulo do domínio
// (Vigencia.SituacaoEm) e o predicado SQL (FragmentoDesignacaoVigente) à
// mesma resposta, nas fronteiras de DG-06 (specs/_fundacao-metas.md §5.3):
// 31/07/2026 23h58 -03:00 (ainda 31/07 em Brasília) e 01/08/2026 00h02
// -03:00 (já 01/08 em Brasília), mais o caso de data_fim nula. O teste roda
// com o container em UTC de propósito — se o adapter algum dia usar
// CURRENT_DATE/now() em vez do parâmetro recebido, a divergência entre a
// data UTC do servidor e a data de Brasília aparece exatamente aqui.
func TestVigencia_RotuloEPredicadoConcordam(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	fusoBrasilia, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("carregar fuso: %v", err)
	}

	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	coordenadorID := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID})
	// Dois cursos distintos: as duas designações abaixo têm intervalos que
	// se tocariam se fossem do mesmo curso, e o EXCLUDE (DG-03) recusaria a
	// segunda — o teste não é sobre esse mecanismo.
	cursoComFim := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	cursoSemFim := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})

	fimEmJulho := "2026-07-31"
	designacaoComFim := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoComFim, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID,
		DataInicio: "2026-01-01", DataFim: &fimEmJulho,
	})
	designacaoSemFim := testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{
		CursoID: cursoSemFim, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID,
		Portaria: "2/2026", DataInicio: "2025-06-01",
	})

	casos := []struct {
		nome         string
		instante     time.Time
		designacaoID uuid.UUID
		esperado     valueobject.SituacaoDesignacao
	}{
		{
			nome:         "23h58 de 31/07 (Brasília) — ainda vigente",
			instante:     time.Date(2026, 7, 31, 23, 58, 0, 0, fusoBrasilia),
			designacaoID: designacaoComFim,
			esperado:     valueobject.Vigente,
		},
		{
			nome:         "00h02 de 01/08 (Brasília) — já encerrada",
			instante:     time.Date(2026, 8, 1, 0, 2, 0, 0, fusoBrasilia),
			designacaoID: designacaoComFim,
			esperado:     valueobject.Encerrada,
		},
		{
			nome:         "data_fim nula — vigente em qualquer instante futuro",
			instante:     time.Date(2030, 1, 1, 12, 0, 0, 0, fusoBrasilia),
			designacaoID: designacaoSemFim,
			esperado:     valueobject.Vigente,
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			hoje := valueobject.DataLocalDe(c.instante, fusoBrasilia)

			// Lado do domínio: reconstrói a Vigencia a partir das colunas
			// gravadas e usa o MESMO SituacaoEm que o grid de designações usa.
			var inicioTexto string
			var fimTexto *string
			if err := db.QueryRow(`SELECT data_inicio::text, data_fim::text FROM designacao WHERE id = $1`, c.designacaoID).Scan(&inicioTexto, &fimTexto); err != nil {
				t.Fatalf("ler designação: %v", err)
			}
			inicio, err := valueobject.DataLocalTexto(inicioTexto)
			if err != nil {
				t.Fatalf("parse início: %v", err)
			}
			var fim *valueobject.DataLocal
			if fimTexto != nil {
				f, err := valueobject.DataLocalTexto(*fimTexto)
				if err != nil {
					t.Fatalf("parse fim: %v", err)
				}
				fim = &f
			}
			vigencia, err := valueobject.NovaVigencia(inicio, fim)
			if err != nil {
				t.Fatalf("NovaVigencia: %v", err)
			}
			situacaoDoDominio := vigencia.SituacaoEm(hoje)

			// Lado do SQL: o mesmo predicado que AplicarEscopo e o EXISTS da
			// sessão usam.
			predicado := FragmentoDesignacaoVigente("designacao", 2)
			consulta := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM designacao WHERE id = $1 AND %s)`, predicado)
			var vigenteNoSQL bool
			if err := db.QueryRow(consulta, c.designacaoID, hoje.String()).Scan(&vigenteNoSQL); err != nil {
				t.Fatalf("consulta SQL: %v", err)
			}
			situacaoDoSQL := valueobject.Encerrada
			if vigenteNoSQL {
				situacaoDoSQL = valueobject.Vigente
			}
			// A fronteira "futura" não é exercitada pelo predicado SQL desta
			// função (ele só distingue vigente de não-vigente); os casos
			// deste teste nunca caem em "futura", então a comparação é
			// direta entre os dois "vigente/não-vigente".

			if situacaoDoDominio != c.esperado {
				t.Fatalf("domínio: esperava %v, obtido %v", c.esperado, situacaoDoDominio)
			}
			if situacaoDoSQL != c.esperado {
				t.Fatalf("SQL: esperava %v, obtido %v", c.esperado, situacaoDoSQL)
			}
		})
	}
}
