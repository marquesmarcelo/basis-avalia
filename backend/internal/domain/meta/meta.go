package meta

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Meta — entidade de domínio (specs/indicadores/design.md §3.3). Aponta
// para 1 a 5 indicadores; nunca tem quantidade EXIGIDA, período ou curso —
// isso entra em Item do Plano, noutra feature. QuantidadeSugerida (abaixo)
// não é exceção a essa regra: é sugestão herdada uma vez na criação do
// item, nunca a quantidade que vale para a apuração.
type Meta struct {
	ID            uuid.UUID
	InstituicaoID uuid.UUID
	Nome          valueobject.NomeCatalogo
	Descricao     string
	Indicadores   []uuid.UUID
	Situacao      valueobject.SituacaoCatalogo
	// QuantidadeSugerida — sugestão herdada pelo item do plano na
	// CRIAÇÃO, uma única vez; editar a sugestão depois não afeta item já
	// existente. Nunca lida pela apuração do relatório, que só conhece
	// item_plano.quantidade. Opcional: meta sem sugestão continua
	// exigindo que o PI informe a quantidade ao montar o plano.
	QuantidadeSugerida *valueobject.Quantidade
	CriadoEm           time.Time
	AtualizadoEm       *time.Time
	ExcluidoEm         *time.Time
	Versao             int
}

const maxIndicadoresPorMeta = 5

// DefinirIndicadores valida e substitui a lista inteira, e devolve a
// lista anterior (nunca a diferença) para a auditoria registrar as duas
// completas (MC-13). Não valida existência, situação ou instituição dos
// indicadores — isso depende do banco e é responsabilidade do use case.
func (m *Meta) DefinirIndicadores(novos []uuid.UUID) ([]uuid.UUID, error) {
	if len(novos) == 0 {
		return nil, domain.ErrIndicadorObrigatorio
	}
	if len(novos) > maxIndicadoresPorMeta {
		return nil, domain.ErrIndicadoresAcimaDoLimite
	}
	vistos := make(map[uuid.UUID]bool, len(novos))
	for _, id := range novos {
		if vistos[id] {
			return nil, domain.ErrIndicadorDuplicadoNaMeta
		}
		vistos[id] = true
	}
	anteriores := m.Indicadores
	m.Indicadores = novos
	return anteriores, nil
}

// NovaMeta cria a meta e já define os indicadores — "não existe meta sem
// indicador, em nenhuma rota" (MC-03) é garantido aqui, no construtor.
func NovaMeta(instituicaoID uuid.UUID, nomeBruto, descricao string, indicadores []uuid.UUID) (*Meta, error) {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return nil, err
	}
	m := &Meta{
		ID:            uuid.Must(uuid.NewV7()),
		InstituicaoID: instituicaoID,
		Nome:          nome,
		Descricao:     descricao,
		Situacao:      valueobject.CatalogoAtivo,
		CriadoEm:      time.Now(),
		Versao:        1,
	}
	if _, err := m.DefinirIndicadores(indicadores); err != nil {
		return nil, err
	}
	return m, nil
}

// Atualizar edita nome e descrição — os indicadores são trocados só por
// DefinirIndicadores, chamado à parte pelo use case.
func (m *Meta) Atualizar(nomeBruto, descricao string) error {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return err
	}
	m.Nome = nome
	m.Descricao = descricao
	return nil
}

func (m *Meta) AlterarSituacao(nova valueobject.SituacaoCatalogo) {
	m.Situacao = nova
}

// DefinirQuantidadeSugerida — nil limpa a sugestão (campo opcional).
// Método próprio, não parâmetro de NovaMeta/Atualizar: mantém as duas
// assinaturas estáveis para quem já as chama, e a sugestão é ortogonal
// ao nome/descrição/indicadores que elas já validam.
func (m *Meta) DefinirQuantidadeSugerida(bruta *int) error {
	if bruta == nil {
		m.QuantidadeSugerida = nil
		return nil
	}
	q, err := valueobject.NovaQuantidade(*bruta)
	if err != nil {
		return err
	}
	m.QuantidadeSugerida = &q
	return nil
}
