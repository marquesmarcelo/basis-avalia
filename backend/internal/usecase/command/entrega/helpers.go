package entrega

import (
	entregadomain "github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

// entregaDeDetalhe reconstrói a entidade de domínio a partir da projeção
// de leitura — necessário porque as regras (PodeCorrigir, Corrigir,
// Avaliar, DesfazerAceitacao) vivem no domínio, e o repositório devolve a
// projeção já enriquecida com nomes para a tela, não a entidade crua.
func entregaDeDetalhe(d port.DetalheEntrega) *entregadomain.Entrega {
	situacao, _ := valueobject.NovaSituacaoEntrega(d.Situacao)
	rodadas, _ := valueobject.NovaRodadaDeRecusa(d.Rodadas)
	return &entregadomain.Entrega{
		ID: d.ID, ItemPlanoID: d.ItemPlanoID, CursoID: d.CursoID,
		EnviadaPor: d.EnviadaPorID, Observacao: d.Observacao,
		Situacao: situacao, Rodadas: rodadas, PrazoCorrecao: d.PrazoCorrecao,
		AvaliadaEm: d.AvaliadaEm, Motivo: d.Motivo,
		AvaliadorEraCoordenador: d.AvaliadorEraCoordenador,
		PendenciaVistaEm:        d.PendenciaVistaEm,
		CriadoEm:                d.CriadoEm, AtualizadoEm: d.AtualizadoEm, Versao: d.Versao,
	}
}
