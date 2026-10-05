package itemplano

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// ItemDoPlano — entidade de domínio (specs/plano-acao/design.md §3.2): a
// meta do catálogo e a quantidade exigida NESTE plano. CursoID e
// InstituicaoID denormalizados com FK composta (fundacao-metas.md §3.5)
// — é o que permite o filtro de carteira ser um predicado de linha, sem
// junção.
type ItemDoPlano struct {
	ID            uuid.UUID
	PlanoID       uuid.UUID
	CursoID       uuid.UUID
	InstituicaoID uuid.UUID
	MetaID        uuid.UUID
	Quantidade    valueobject.Quantidade
	CriadoEm      time.Time
	AtualizadoEm  *time.Time
	ExcluidoEm    *time.Time
	Versao        int
}

func NovoItemDoPlano(planoID, cursoID, instituicaoID, metaID uuid.UUID, quantidadeBruta int) (*ItemDoPlano, error) {
	quantidade, err := valueobject.NovaQuantidade(quantidadeBruta)
	if err != nil {
		return nil, err
	}
	return &ItemDoPlano{
		ID: uuid.Must(uuid.NewV7()), PlanoID: planoID, CursoID: cursoID, InstituicaoID: instituicaoID,
		MetaID: metaID, Quantidade: quantidade, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

// AlterarQuantidade nunca divide nem multiplica — apenas substitui o
// valor (IT-01, IT-02: a quantidade mora só aqui).
func (i *ItemDoPlano) AlterarQuantidade(novaBruta int) error {
	nova, err := valueobject.NovaQuantidade(novaBruta)
	if err != nil {
		return err
	}
	i.Quantidade = nova
	return nil
}
