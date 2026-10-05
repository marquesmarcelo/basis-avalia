package indicador

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Indicador — entidade de domínio (specs/indicadores/design.md §3.2). Uma
// única tabela, com Escopo determinando se InstituicaoID e
// ReferenciaInstrumento existem. Escopo é imutável após a criação: os dois
// construtores tornam impossível montar a combinação errada, e nenhum
// método muda o campo depois.
type Indicador struct {
	ID                    uuid.UUID
	Escopo                valueobject.EscopoIndicador
	InstituicaoID         *uuid.UUID
	Codigo                valueobject.CodigoIndicador
	Nome                  valueobject.NomeCatalogo
	Descricao             string
	ReferenciaInstrumento *valueobject.ReferenciaInstrumento
	Situacao              valueobject.SituacaoCatalogo
	CriadoEm              time.Time
	AtualizadoEm          *time.Time
	ExcluidoEm            *time.Time
	Versao                int
}

func validarCodigoNome(codigoBruto, nomeBruto string) (valueobject.CodigoIndicador, valueobject.NomeCatalogo, error) {
	codigo, err := valueobject.NovoCodigoIndicador(codigoBruto)
	if err != nil {
		return valueobject.CodigoIndicador{}, valueobject.NomeCatalogo{}, err
	}
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return valueobject.CodigoIndicador{}, valueobject.NomeCatalogo{}, err
	}
	return codigo, nome, nil
}

// NovoDaPlataforma cria um indicador do catálogo do INEP, comum à
// instalação — referência do instrumento obrigatória (IE-02).
func NovoDaPlataforma(codigoBruto, nomeBruto, descricao, referenciaBruta string) (*Indicador, error) {
	codigo, nome, err := validarCodigoNome(codigoBruto, nomeBruto)
	if err != nil {
		return nil, err
	}
	referencia, err := valueobject.NovaReferenciaInstrumento(referenciaBruta)
	if err != nil {
		return nil, err
	}
	if referencia.Vazia() {
		return nil, domain.ErrReferenciaInstrumentoInvalida
	}
	return &Indicador{
		ID:                    uuid.Must(uuid.NewV7()),
		Escopo:                valueobject.EscopoIndicadorPlataforma,
		Codigo:                codigo,
		Nome:                  nome,
		Descricao:             descricao,
		ReferenciaInstrumento: &referencia,
		Situacao:              valueobject.CatalogoAtivo,
		CriadoEm:              time.Now(),
		Versao:                1,
	}, nil
}

// NovoDaInstituicao cria um indicador próprio da instituição. Não recebe
// parâmetro de referência do instrumento (IN-02): a assinatura não aceita
// o dado, em vez de recusar em tempo de execução.
func NovoDaInstituicao(instituicaoID uuid.UUID, codigoBruto, nomeBruto, descricao string) (*Indicador, error) {
	codigo, nome, err := validarCodigoNome(codigoBruto, nomeBruto)
	if err != nil {
		return nil, err
	}
	return &Indicador{
		ID:            uuid.Must(uuid.NewV7()),
		Escopo:        valueobject.EscopoIndicadorInstituicao,
		InstituicaoID: &instituicaoID,
		Codigo:        codigo,
		Nome:          nome,
		Descricao:     descricao,
		Situacao:      valueobject.CatalogoAtivo,
		CriadoEm:      time.Now(),
		Versao:        1,
	}, nil
}

// AtualizarDaPlataforma edita código, nome, descrição e referência do
// instrumento de um indicador do catálogo comum. Vale para todas as
// instituições imediatamente (IE-05) — não é responsabilidade da entidade
// avisar quem usa, só mudar o próprio estado.
func (i *Indicador) AtualizarDaPlataforma(codigoBruto, nomeBruto, descricao, referenciaBruta string) error {
	referencia, err := valueobject.NovaReferenciaInstrumento(referenciaBruta)
	if err != nil {
		return err
	}
	if referencia.Vazia() {
		return domain.ErrReferenciaInstrumentoInvalida
	}
	codigo, nome, err := validarCodigoNome(codigoBruto, nomeBruto)
	if err != nil {
		return err
	}
	i.Codigo, i.Nome, i.Descricao, i.ReferenciaInstrumento = codigo, nome, descricao, &referencia
	return nil
}

// AtualizarDaInstituicao edita código, nome e descrição de um indicador
// próprio. Sem parâmetro de referência — mesma razão de NovoDaInstituicao.
func (i *Indicador) AtualizarDaInstituicao(codigoBruto, nomeBruto, descricao string) error {
	codigo, nome, err := validarCodigoNome(codigoBruto, nomeBruto)
	if err != nil {
		return err
	}
	i.Codigo, i.Nome, i.Descricao = codigo, nome, descricao
	return nil
}

func (i *Indicador) AlterarSituacao(nova valueobject.SituacaoCatalogo) {
	i.Situacao = nova
}
