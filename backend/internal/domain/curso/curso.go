package curso

import (
	"time"

	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Curso — entidade de domínio (specs/cursos/design.md §3.2). Nenhuma
// coluna de coordenador aqui: quem responde hoje é resultado de consulta
// sobre Designacao, nunca campo persistido (achado de revisão se aparecer).
type Curso struct {
	ID            uuid.UUID
	InstituicaoID uuid.UUID
	Nome          valueobject.NomeCatalogo
	CodigoEMec    *valueobject.CodigoEMec
	Grau          valueobject.GrauDeCurso
	Modalidade    valueobject.ModalidadeDeCurso
	Situacao      valueobject.SituacaoCurso
	CriadoEm      time.Time
	AtualizadoEm  *time.Time
	ExcluidoEm    *time.Time
	Versao        int
}

func codigoEMecOpcional(bruto string) (*valueobject.CodigoEMec, error) {
	if bruto == "" {
		return nil, nil
	}
	codigo, err := valueobject.NovoCodigoEMec(bruto)
	if err != nil {
		return nil, err
	}
	if codigo.Nulo() {
		return nil, nil
	}
	return &codigo, nil
}

func NovoCurso(instituicaoID uuid.UUID, nomeBruto, codigoEMecBruto, grauBruto, modalidadeBruto string) (*Curso, error) {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return nil, err
	}
	codigoEMec, err := codigoEMecOpcional(codigoEMecBruto)
	if err != nil {
		return nil, err
	}
	grau, err := valueobject.NovoGrauDeCurso(grauBruto)
	if err != nil {
		return nil, err
	}
	modalidade, err := valueobject.NovaModalidadeDeCurso(modalidadeBruto)
	if err != nil {
		return nil, err
	}
	return &Curso{
		ID: uuid.Must(uuid.NewV7()), InstituicaoID: instituicaoID, Nome: nome, CodigoEMec: codigoEMec,
		Grau: grau, Modalidade: modalidade, Situacao: valueobject.CursoAtivo, CriadoEm: time.Now(), Versao: 1,
	}, nil
}

func (c *Curso) Atualizar(nomeBruto, codigoEMecBruto, grauBruto, modalidadeBruto string) error {
	nome, err := valueobject.NovoNomeCatalogo(nomeBruto)
	if err != nil {
		return err
	}
	codigoEMec, err := codigoEMecOpcional(codigoEMecBruto)
	if err != nil {
		return err
	}
	grau, err := valueobject.NovoGrauDeCurso(grauBruto)
	if err != nil {
		return err
	}
	modalidade, err := valueobject.NovaModalidadeDeCurso(modalidadeBruto)
	if err != nil {
		return err
	}
	c.Nome, c.CodigoEMec, c.Grau, c.Modalidade = nome, codigoEMec, grau, modalidade
	return nil
}

func (c *Curso) AlterarSituacao(nova valueobject.SituacaoCurso) {
	c.Situacao = nova
}
