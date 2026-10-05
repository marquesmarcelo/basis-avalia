package anexo

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type RemoverAnexoInput struct {
	Ator    autorizacao.Ator
	AnexoID uuid.UUID
}

// RemoverAnexoUseCase — DELETE /anexos/{id}: exclusão lógica no banco; o
// objeto só sai pela rotina de retenção (design.md §3.6). Só o autor da
// entrega remove — o repositório confere isso contra entrega.enviada_por.
type RemoverAnexoUseCase struct {
	repo port.EntregaRepository
}

func NovoRemoverAnexoUseCase(repo port.EntregaRepository) *RemoverAnexoUseCase {
	return &RemoverAnexoUseCase{repo: repo}
}

func (uc *RemoverAnexoUseCase) Executar(ctx context.Context, in RemoverAnexoInput) error {
	esc, err := autorizacao.Autorizar(in.Ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoEditar, nil)
	if err != nil {
		return err
	}
	// A chave do objeto NUNCA é usada aqui para chamar o armazenamento —
	// a remoção do objeto físico é responsabilidade da rotina de retenção
	// (design.md §3.6), não deste comando. O retorno existe só para o
	// caso futuro em que a remoção imediata for decidida; hoje é ignorado
	// de propósito.
	_, err = uc.repo.RemoverAnexo(ctx, esc, in.AnexoID, in.Ator.Proprio())
	return err
}
