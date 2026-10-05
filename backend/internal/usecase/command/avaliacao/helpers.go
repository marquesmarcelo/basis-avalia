package avaliacao

import (
	entregadomain "github.com/basis-avalia/backend/internal/domain/entrega"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
)

// entregaDeDetalhe — mesma conversão de usecase/command/entrega, duplicada
// de propósito: os dois pacotes não importam um do outro só por isto.
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
