package plano

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/periodo"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Plano — entidade de domínio (specs/plano-acao/design.md §3.2). Nasce em
// rascunho; a coluna armazenada (SituacaoPublicacao) nunca guarda
// "encerrado" — essa terceira situação é sempre derivada por
// SituacaoEfetiva, a única função que decide.
type Plano struct {
	ID                  uuid.UUID
	InstituicaoID       uuid.UUID
	CursoID             uuid.UUID
	PeriodoID           uuid.UUID
	Descricao           string
	ObjetivoGeral       string
	ResultadosEsperados string
	AlinhamentoPDI      string
	AlinhamentoPPC      string
	Aprovacao           *valueobject.Aprovacao
	SituacaoPublicacao  valueobject.SituacaoPlano // rascunho | vigente, nunca encerrado
	EncerradoEm         *time.Time
	EncerramentoMotivo  string
	CriadoEm            time.Time
	AtualizadoEm        *time.Time
	ExcluidoEm          *time.Time
	Versao              int
}

func validarTextosObrigatorios(descricao, objetivoGeral, resultadosEsperados string) error {
	if trimVazio(descricao) {
		return &domain.ErrValidacao{Campo: "descricao", Mensagem: "A descrição é obrigatória."}
	}
	if trimVazio(objetivoGeral) {
		return &domain.ErrValidacao{Campo: "objetivo_geral", Mensagem: "O objetivo geral é obrigatório."}
	}
	if trimVazio(resultadosEsperados) {
		return &domain.ErrValidacao{Campo: "resultados_esperados", Mensagem: "Os resultados esperados são obrigatórios."}
	}
	return nil
}

func trimVazio(s string) bool {
	for _, r := range s {
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}

type DadosDoPlano struct {
	Descricao, ObjetivoGeral, ResultadosEsperados string
	AlinhamentoPDI, AlinhamentoPPC                string
	AprovacaoData, AprovacaoOrgao                 string
}

func NovoPlano(instituicaoID, cursoID, periodoID uuid.UUID, d DadosDoPlano) (*Plano, error) {
	if err := validarTextosObrigatorios(d.Descricao, d.ObjetivoGeral, d.ResultadosEsperados); err != nil {
		return nil, err
	}
	aprovacao, err := aprovacaoOpcional(d.AprovacaoData, d.AprovacaoOrgao)
	if err != nil {
		return nil, err
	}
	return &Plano{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: instituicaoID, CursoID: cursoID, PeriodoID: periodoID,
		Descricao: d.Descricao, ObjetivoGeral: d.ObjetivoGeral, ResultadosEsperados: d.ResultadosEsperados,
		AlinhamentoPDI: d.AlinhamentoPDI, AlinhamentoPPC: d.AlinhamentoPPC, Aprovacao: aprovacao,
		SituacaoPublicacao: valueobject.PlanoRascunho, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

func aprovacaoOpcional(dataBruta, orgaoBruto string) (*valueobject.Aprovacao, error) {
	aprovacao, preenchida, err := valueobject.NovaAprovacao(dataBruta, orgaoBruto)
	if err != nil {
		return nil, err
	}
	if !preenchida {
		return nil, nil
	}
	return &aprovacao, nil
}

// AtualizarDados edita os campos de texto e a aprovação — nunca curso,
// período ou situação, que têm seu próprio caminho de comando.
func (p *Plano) AtualizarDados(d DadosDoPlano) error {
	if err := validarTextosObrigatorios(d.Descricao, d.ObjetivoGeral, d.ResultadosEsperados); err != nil {
		return err
	}
	aprovacao, err := aprovacaoOpcional(d.AprovacaoData, d.AprovacaoOrgao)
	if err != nil {
		return err
	}
	p.Descricao, p.ObjetivoGeral, p.ResultadosEsperados = d.Descricao, d.ObjetivoGeral, d.ResultadosEsperados
	p.AlinhamentoPDI, p.AlinhamentoPPC = d.AlinhamentoPDI, d.AlinhamentoPPC
	p.Aprovacao = aprovacao
	return nil
}

// SituacaoEfetiva é a ÚNICA função que decide a situação exposta ao
// cliente (design.md §3.2, P-02, P-03): encerrado quando o plano foi
// encerrado antecipadamente OU o período já terminou na data de
// referência — nunca gravado como coluna.
func (p *Plano) SituacaoEfetiva(per periodo.Periodo, hoje valueobject.DataLocal) valueobject.SituacaoPlano {
	if p.SituacaoPublicacao == valueobject.PlanoRascunho {
		return valueobject.PlanoRascunho
	}
	if p.EncerradoEm != nil {
		return valueobject.PlanoEncerrado
	}
	if per.Vigencia.SituacaoEm(hoje) == valueobject.Encerrada {
		return valueobject.PlanoEncerrado
	}
	return valueobject.PlanoVigente
}

// SemAprovacao é verdadeiro só quando a situação efetiva é vigente ou
// encerrada e não há aprovação registrada (SI-05, SI-07) — plano em
// rascunho sem aprovação é o estado normal de quem está montando, e não
// recebe aviso nenhum.
func (p *Plano) SemAprovacao(per periodo.Periodo, hoje valueobject.DataLocal) bool {
	situacao := p.SituacaoEfetiva(per, hoje)
	if situacao == valueobject.PlanoRascunho {
		return false
	}
	return p.Aprovacao == nil
}

// Publicar exige ao menos um item (verificado pelo use case, que conhece
// a contagem) e período não encerrado.
func (p *Plano) Publicar(temItem bool, periodoEncerrado bool) error {
	if !temItem {
		return domain.ErrPlanoSemItem
	}
	if periodoEncerrado {
		return domain.ErrPeriodoEncerrado
	}
	p.SituacaoPublicacao = valueobject.PlanoVigente
	return nil
}

// Despublicar só é permitido sem nenhuma entrega — a verificação de
// entrega é do use case (depende de repositório).
func (p *Plano) Despublicar() {
	p.SituacaoPublicacao = valueobject.PlanoRascunho
}

func (p *Plano) Encerrar(motivo string, agora time.Time) error {
	if trimVazio(motivo) {
		return domain.ErrMotivoObrigatorio
	}
	p.EncerradoEm = &agora
	p.EncerramentoMotivo = motivo
	return nil
}

// Reabrir limpa o encerramento antecipado — se o período já encerrou, a
// situação efetiva permanece "encerrado" mesmo assim (SI-13): a situação
// é derivada de duas fontes, e limpar uma não anula a outra.
func (p *Plano) Reabrir() {
	p.EncerradoEm = nil
	p.EncerramentoMotivo = ""
}
