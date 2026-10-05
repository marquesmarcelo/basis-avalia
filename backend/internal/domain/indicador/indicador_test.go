package indicador

import (
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// TestNovoDaPlataforma_RecusaSemReferencia prova IE-02: referência é
// obrigatória no catálogo do INEP.
func TestNovoDaPlataforma_RecusaSemReferencia(t *testing.T) {
	_, err := NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "")
	if err != domain.ErrReferenciaInstrumentoInvalida {
		t.Fatalf("esperava ErrReferenciaInstrumentoInvalida, obtido %v", err)
	}
}

func TestNovoDaPlataforma_CriaComEscopoPlataformaEInstituicaoNula(t *testing.T) {
	ind, err := NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "",
		"Instrumento de Avaliação de Cursos de Graduação 2017 - Dimensão 1, indicador 1.4")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ind.Escopo != valueobject.EscopoIndicadorPlataforma {
		t.Fatalf("esperava escopo plataforma, obtido %v", ind.Escopo)
	}
	if ind.InstituicaoID != nil {
		t.Fatal("esperava instituicao_id nulo no escopo plataforma")
	}
	if ind.ReferenciaInstrumento == nil || ind.ReferenciaInstrumento.Vazia() {
		t.Fatal("esperava referência do instrumento preenchida")
	}
}

// TestNovoDaInstituicao_NaoTemParametroDeReferencia é IN-02: a garantia é
// de assinatura, não de validação em tempo de execução — não há como
// nem tentar passar uma referência a este construtor.
func TestNovoDaInstituicao_CriaComEscopoInstituicaoSemReferencia(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ind, err := NovoDaInstituicao(instituicaoID, "GEST-01", "Reuniões com representação discente", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ind.Escopo != valueobject.EscopoIndicadorInstituicao {
		t.Fatalf("esperava escopo instituicao, obtido %v", ind.Escopo)
	}
	if ind.InstituicaoID == nil || *ind.InstituicaoID != instituicaoID {
		t.Fatal("esperava instituicao_id preenchido com a instituição da sessão")
	}
	if ind.ReferenciaInstrumento != nil {
		t.Fatal("indicador institucional não pode ter referência do instrumento")
	}
}

func TestNovoDaPlataforma_RecusaCodigoVazio(t *testing.T) {
	_, err := NovoDaPlataforma("", "Nome", "", "referência")
	if err == nil {
		t.Fatal("esperava erro para código vazio")
	}
}

// TestAtualizarDaPlataforma_RecusaSemReferencia garante que a edição
// preserva a obrigatoriedade — não só a criação.
func TestAtualizarDaPlataforma_RecusaSemReferencia(t *testing.T) {
	ind, err := NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência original")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	err = ind.AtualizarDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "")
	if err != domain.ErrReferenciaInstrumentoInvalida {
		t.Fatalf("esperava ErrReferenciaInstrumentoInvalida, obtido %v", err)
	}
}

func TestAtualizarDaPlataforma_CorrigeReferencia(t *testing.T) {
	ind, err := NovoDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência original")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := ind.AtualizarDaPlataforma("1.4", "Núcleo Docente Estruturante", "", "referência corrigida"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if ind.ReferenciaInstrumento.String() != "referência corrigida" {
		t.Fatalf("esperava referência corrigida, obtido %q", ind.ReferenciaInstrumento.String())
	}
}

func TestIndicador_AlterarSituacao(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	ind, err := NovoDaInstituicao(instituicaoID, "GEST-01", "Nome", "")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	ind.AlterarSituacao(valueobject.CatalogoInativo)
	if ind.Situacao != valueobject.CatalogoInativo {
		t.Fatal("esperava situação inativa")
	}
}
