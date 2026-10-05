package designacao

import (
	"context"
	"testing"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/curso"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

func cursoDeTeste(t *testing.T, instituicaoID uuid.UUID) curso.Curso {
	t.Helper()
	c, err := curso.NovoCurso(instituicaoID, "Engenharia de Software", "", "bacharelado", "presencial")
	if err != nil {
		t.Fatalf("setup curso: %v", err)
	}
	return *c
}

func usuarioDeTeste(t *testing.T, instituicaoID uuid.UUID, id uuid.UUID, perfis ...valueobject.Perfil) *usuario.Usuario {
	t.Helper()
	conjunto, err := valueobject.NovoConjunto(perfis...)
	if err != nil {
		t.Fatalf("setup conjunto: %v", err)
	}
	return &usuario.Usuario{ID: id, InstituicaoID: &instituicaoID, Nome: "Candidato de Teste", Perfis: conjunto}
}

// TestCriarDesignacao_CP09_AutodesignacaoGravadaQuandoAtorEhOCoordenador
// prova CP-09: quando o autor da designação é a pessoa designada,
// autodesignacao nasce verdadeiro, sem recusa por ser a própria pessoa.
func TestCriarDesignacao_CP09_AutodesignacaoGravadaQuandoAtorEhOCoordenador(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	cursoRepo := &cursoRepoMock{itemParaBuscar: port.ItemCurso{Curso: cursoDeTeste(t, instituicaoID)}}
	usuarioRepo := &usuarioRepoMock{itemParaBuscar: usuarioDeTeste(t, instituicaoID, ator.UsuarioID(), valueobject.Professor, valueobject.PesquisadorInstitucional)}
	repo := &designacaoRepoMock{}
	audit := &auditMock{}
	uc := NovoCriarDesignacaoUseCase(repo, cursoRepo, usuarioRepo, audit, uowFake{})

	out, err := uc.Executar(context.Background(), CriarDesignacaoInput{
		Ator: ator, CursoID: uuid.Must(uuid.NewV7()), CoordenadorID: ator.UsuarioID(),
		Portaria: "70/2026", DataInicio: "2026-03-01",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !out.Autodesignacao {
		t.Fatal("CP-09: esperava autodesignacao verdadeiro quando o autor designa a si mesmo")
	}
	if !repo.inserirChamado {
		t.Fatal("esperava Inserir chamado")
	}
	if len(audit.eventos) != 1 || audit.eventos[0].Detalhes["autodesignacao"] != true {
		t.Fatalf("esperava auditoria com autodesignacao=true, obtido %+v", audit.eventos)
	}
}

// TestCriarDesignacao_DesignacaoPorOutrem_AutodesignacaoFalso — a PI
// designa outra pessoa: autodesignacao nasce falso.
func TestCriarDesignacao_DesignacaoPorOutrem_AutodesignacaoFalso(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)
	coordenadorID := uuid.Must(uuid.NewV7())

	cursoRepo := &cursoRepoMock{itemParaBuscar: port.ItemCurso{Curso: cursoDeTeste(t, instituicaoID)}}
	usuarioRepo := &usuarioRepoMock{itemParaBuscar: usuarioDeTeste(t, instituicaoID, coordenadorID, valueobject.Professor)}
	repo := &designacaoRepoMock{}
	uc := NovoCriarDesignacaoUseCase(repo, cursoRepo, usuarioRepo, &auditMock{}, uowFake{})

	out, err := uc.Executar(context.Background(), CriarDesignacaoInput{
		Ator: ator, CursoID: uuid.Must(uuid.NewV7()), CoordenadorID: coordenadorID,
		Portaria: "47/2026", DataInicio: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if out.Autodesignacao {
		t.Fatal("esperava autodesignacao falso quando o autor designa outra pessoa")
	}
}

// TestCriarDesignacao_CP07_CandidatoApenasAlunoDevolveCoordenadorInvalido
// prova a validação de 400 do design.md §5.1: quem só tem o perfil de
// aluno não é candidato válido.
func TestCriarDesignacao_CP07_CandidatoApenasAlunoDevolveCoordenadorInvalido(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)
	alunoID := uuid.Must(uuid.NewV7())

	cursoRepo := &cursoRepoMock{itemParaBuscar: port.ItemCurso{Curso: cursoDeTeste(t, instituicaoID)}}
	usuarioRepo := &usuarioRepoMock{itemParaBuscar: usuarioDeTeste(t, instituicaoID, alunoID, valueobject.Aluno)}
	repo := &designacaoRepoMock{}
	uc := NovoCriarDesignacaoUseCase(repo, cursoRepo, usuarioRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarDesignacaoInput{
		Ator: ator, CursoID: uuid.Must(uuid.NewV7()), CoordenadorID: alunoID,
		Portaria: "1/2026", DataInicio: "2026-01-01",
	})
	if err != domain.ErrCoordenadorInvalido {
		t.Fatalf("esperava ErrCoordenadorInvalido, obtido %v", err)
	}
	if repo.inserirChamado {
		t.Fatal("não deveria ter chamado Inserir")
	}
}

// TestCriarDesignacao_CP07_CandidatoDeOutraInstituicaoDevolveNaoEncontrado
// prova o 404 (nunca 403 nem 400) de CP-07.
func TestCriarDesignacao_CP07_CandidatoDeOutraInstituicaoDevolveNaoEncontrado(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	cursoRepo := &cursoRepoMock{itemParaBuscar: port.ItemCurso{Curso: cursoDeTeste(t, instituicaoID)}}
	usuarioRepo := &usuarioRepoMock{erroBuscar: domain.ErrNaoEncontrado}
	repo := &designacaoRepoMock{}
	uc := NovoCriarDesignacaoUseCase(repo, cursoRepo, usuarioRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarDesignacaoInput{
		Ator: ator, CursoID: uuid.Must(uuid.NewV7()), CoordenadorID: uuid.Must(uuid.NewV7()),
		Portaria: "1/2026", DataInicio: "2026-01-01",
	})
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obtido %v", err)
	}
}

// TestCriarDesignacao_T171_CursoExcluidoDevolveNaoEncontrado prova o
// achado de revisão de T-171: TravarSeAtivo (FOR SHARE, port.CursoRepository)
// é o que impede criar uma designação para um curso já excluído — o FK
// designacao→curso não bastava, porque a exclusão é lógica e a FK nunca
// dispara sobre excluido_em. Aqui só confirma que o use case propaga o
// erro de TravarSeAtivo sem tentar inserir mesmo assim; a corrida real
// (curso_repository_concorrencia_test.go) prova a trava no banco.
func TestCriarDesignacao_T171_CursoExcluidoDevolveNaoEncontrado(t *testing.T) {
	instituicaoID := uuid.Must(uuid.NewV7())
	hoje := dataTeste(t, "2026-03-15")
	ator := atorPIComData(t, instituicaoID, hoje)

	cursoRepo := &cursoRepoMock{erroBuscar: domain.ErrNaoEncontrado}
	usuarioRepo := &usuarioRepoMock{itemParaBuscar: usuarioDeTeste(t, instituicaoID, ator.UsuarioID(), valueobject.Professor)}
	repo := &designacaoRepoMock{}
	uc := NovoCriarDesignacaoUseCase(repo, cursoRepo, usuarioRepo, &auditMock{}, uowFake{})

	_, err := uc.Executar(context.Background(), CriarDesignacaoInput{
		Ator: ator, CursoID: uuid.Must(uuid.NewV7()), CoordenadorID: ator.UsuarioID(),
		Portaria: "1/2026", DataInicio: "2026-01-01",
	})
	if err != domain.ErrNaoEncontrado {
		t.Fatalf("esperava ErrNaoEncontrado, obtido %v", err)
	}
	if repo.inserirChamado {
		t.Fatal("não deveria ter chamado Inserir quando o curso está excluído")
	}
}
