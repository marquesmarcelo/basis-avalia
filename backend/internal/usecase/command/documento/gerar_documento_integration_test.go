package documento

import (
	"context"
	"io"
	"testing"

	"github.com/basis-avalia/backend/internal/adapter/auditoria"
	"github.com/basis-avalia/backend/internal/adapter/postgres"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/plano"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/basis-avalia/backend/internal/testhelpers"
	planocmd "github.com/basis-avalia/backend/internal/usecase/command/plano"
	"github.com/google/uuid"
)

type armazenamentoFake struct{}

func (armazenamentoFake) Gravar(ctx context.Context, chave string, conteudo io.Reader, tipo string, tamanho int64) error {
	_, err := io.ReadAll(conteudo)
	return err
}
func (armazenamentoFake) Ler(ctx context.Context, chave string) (io.ReadCloser, string, error) {
	return nil, "", nil
}
func (armazenamentoFake) Remover(ctx context.Context, chave string) error { return nil }

type geradorFake struct{}

func (geradorFake) Gerar(dados port.DadosDoDocumento) ([]byte, error) {
	return []byte("docx-fake"), nil
}

type uowIntegracaoFake struct{}

func (uowIntegracaoFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func atorComPerfil(t *testing.T, usuarioID, instituicaoID uuid.UUID, perfil valueobject.Perfil, hoje string) autorizacao.Ator {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(perfil)
	if err != nil {
		t.Fatalf("conjunto: %v", err)
	}
	ator, err := autorizacao.NovoAtor(usuarioID, conjunto, &instituicaoID)
	if err != nil {
		t.Fatalf("NovoAtor: %v", err)
	}
	data, err := valueobject.DataLocalTexto(hoje)
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	return ator.ComDataDeReferencia(data)
}

// TestGerarDocumento_P11_CoordenadorGeraDocumentoNuncaCausa409NoPI prova
// design.md §7.4: gerar documento é INSERT em documento e nada mais —
// nunca escreve em plano. PI abre o plano na versão 1, o coordenador gera
// o documento do outro lado, e o PI salva depois: 200, nunca
// CONFLITO_DE_VERSAO.
func TestGerarDocumento_P11_CoordenadorGeraDocumentoNuncaCausa409NoPI(t *testing.T) {
	db := testhelpers.BancoDeTeste(t)
	instituicaoID := testhelpers.CriarInstituicao(t, db, testhelpers.OpcoesInstituicao{})
	pi := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "pesquisador_institucional"})
	coordenador := testhelpers.CriarUsuario(t, db, testhelpers.OpcoesUsuario{InstituicaoID: &instituicaoID, Perfil: "professor"})
	cursoID := testhelpers.CriarCurso(t, db, testhelpers.OpcoesCurso{InstituicaoID: instituicaoID})
	testhelpers.CriarDesignacao(t, db, testhelpers.OpcoesDesignacao{CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenador, DataInicio: "2026-01-01"})
	periodoID := testhelpers.CriarPeriodo(t, db, testhelpers.OpcoesPeriodo{InstituicaoID: instituicaoID})
	planoID := testhelpers.CriarPlano(t, db, testhelpers.OpcoesPlano{InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID})

	planoRepo := postgres.NovoPlanoRepository(db)
	documentoRepo := postgres.NovoDocumentoRepository(db)
	sys, _ := auditoria.NovoSyslog("", "tcp", "teste")
	auditRepo := postgres.NovoAuditoriaRepository(db)
	audit := auditoria.NovoComposto(auditRepo, sys)
	uow := uowIntegracaoFake{}

	gerarUC := NovoGerarDocumentoUseCase(planoRepo, documentoRepo, armazenamentoFake{}, geradorFake{}, audit, uow)
	atualizarPlanoUC := planocmd.NovoAtualizarPlanoUseCase(planoRepo, audit, uow)

	atorCoordenador := atorComPerfil(t, coordenador, instituicaoID, valueobject.CoordenadorCurso, "2026-03-15")
	_, err := gerarUC.Executar(context.Background(), GerarDocumentoInput{
		Ator: atorCoordenador, Alcance: autorizacao.PlanosDaCarteira, PlanoID: planoID,
	})
	if err != nil {
		t.Fatalf("coordenador deveria conseguir gerar o documento do próprio curso: %v", err)
	}

	atorPI := atorComPerfil(t, pi, instituicaoID, valueobject.PesquisadorInstitucional, "2026-03-15")
	_, err = atualizarPlanoUC.Executar(context.Background(), planocmd.AtualizarPlanoInput{
		Ator: atorPI, PlanoID: planoID, Versao: 1,
		Dados: plano.DadosDoPlano{Descricao: "Nova descrição", ObjetivoGeral: "Novo objetivo", ResultadosEsperados: "Novos resultados"},
	})
	if err != nil {
		t.Fatalf("PI salvando na versão 1 depois do coordenador gerar o documento deveria responder 200, obtido: %v", err)
	}
}
