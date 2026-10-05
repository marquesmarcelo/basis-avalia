package designacao

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Designacao — entidade de domínio (specs/cursos/design.md §3.2): a
// portaria com vigência de que o perfil de Coordenador de Curso é
// derivado. Nunca gravada como perfil — ver fundacao-metas.md §4.
type Designacao struct {
	ID             uuid.UUID
	CursoID        uuid.UUID
	InstituicaoID  uuid.UUID
	CoordenadorID  uuid.UUID
	Portaria       valueobject.Portaria
	Vigencia       valueobject.Vigencia
	Autodesignacao bool
	CriadoEm       time.Time
	AtualizadoEm   *time.Time
	ExcluidoEm     *time.Time
	Versao         int
}

// NovaDesignacao grava autodesignacao no ato de criação, nunca recalculada
// depois (C-06, specs/cursos/design.md §5.2) — descreve o instante, como a
// marca de avaliação.
func NovaDesignacao(
	cursoID, instituicaoID, coordenadorID uuid.UUID,
	portariaBruta string, inicio valueobject.DataLocal, fim *valueobject.DataLocal,
	autodesignacao bool,
) (*Designacao, error) {
	portaria, err := valueobject.NovaPortaria(portariaBruta)
	if err != nil {
		return nil, err
	}
	vigencia, err := valueobject.NovaVigencia(inicio, fim)
	if err != nil {
		return nil, err
	}
	return &Designacao{
		ID: uuid.Must(uuid.NewV7()), CursoID: cursoID, InstituicaoID: instituicaoID, CoordenadorID: coordenadorID,
		Portaria: portaria, Vigencia: vigencia, Autodesignacao: autodesignacao, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

func (d *Designacao) SituacaoEm(hoje valueobject.DataLocal) valueobject.SituacaoDesignacao {
	return d.Vigencia.SituacaoEm(hoje)
}

// AtualizarCompleta reescreve todos os campos — só chamada pelo use case
// quando a situação ATUAL é Futura (design.md §5.3): designação futura não
// concedeu privilégio ainda, e a spec já permite excluí-la inteira.
func (d *Designacao) AtualizarCompleta(coordenadorID uuid.UUID, portariaBruta string, inicio valueobject.DataLocal, fim *valueobject.DataLocal) error {
	portaria, err := valueobject.NovaPortaria(portariaBruta)
	if err != nil {
		return err
	}
	vigencia, err := valueobject.NovaVigencia(inicio, fim)
	if err != nil {
		return err
	}
	d.CoordenadorID = coordenadorID
	d.Portaria = portaria
	d.Vigencia = vigencia
	return nil
}

// AtualizarFimEPortaria é o único caminho de edição quando a situação
// ATUAL é Vigente ou Encerrada (design.md §5.3): "encerrar antes do
// previsto" e "prorrogar" são o mesmo mecanismo — editar data_fim. Tentar
// mudar coordenador ou início nessa situação é 409 DESIGNACAO_COM_EFEITO,
// decidido pelo use case antes de chamar este método.
func (d *Designacao) AtualizarFimEPortaria(portariaBruta string, fim *valueobject.DataLocal) error {
	portaria, err := valueobject.NovaPortaria(portariaBruta)
	if err != nil {
		return err
	}
	vigencia, err := valueobject.NovaVigencia(d.Vigencia.Inicio(), fim)
	if err != nil {
		return err
	}
	d.Portaria = portaria
	d.Vigencia = vigencia
	return nil
}

// PodeExcluir — só designação Futura é histórico ainda não escrito
// (DG-08, design.md §3.2).
func (d *Designacao) PodeExcluir(hoje valueobject.DataLocal) bool {
	return d.SituacaoEm(hoje) == valueobject.Futura
}

// ErrSeNaoForFutura devolve ErrDesignacaoComEfeito quando a situação atual
// não é Futura — usado pelo use case antes de aceitar mudança de
// coordenador ou de início, e antes de excluir.
func ErrSeNaoForFutura(situacaoAtual valueobject.SituacaoDesignacao) error {
	if situacaoAtual != valueobject.Futura {
		return domain.ErrDesignacaoComEfeito
	}
	return nil
}
