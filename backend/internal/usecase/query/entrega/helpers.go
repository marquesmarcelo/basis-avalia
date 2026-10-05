package entrega

import (
	"errors"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
)

// autorizarEntregaLeitura tenta EntregasDaInstituicao (PI) e, só se
// negado, EntregasDaCarteira (Coordenador) — mesmo padrão de
// documento/baixar_documento.go: a rota serve os dois perfis, e "quem
// acumula os dois enxerga pelos dois recortes" (spec.md 3.9, VI-05) é a
// união que Perfil.Pode() já garante — aqui só decidimos qual Escopo usar
// quando a rota não escolhe de antemão.
func autorizarEntregaLeitura(ator autorizacao.Ator) (autorizacao.Escopo, error) {
	esc, err := autorizacao.Autorizar(ator, autorizacao.EntregasDaInstituicao, autorizacao.AcaoListar, nil)
	if err == nil {
		return esc, nil
	}
	if !errors.Is(err, domain.ErrPermissaoNegada) {
		return autorizacao.Escopo{}, err
	}
	return autorizacao.Autorizar(ator, autorizacao.EntregasDaCarteira, autorizacao.AcaoListar, nil)
}
