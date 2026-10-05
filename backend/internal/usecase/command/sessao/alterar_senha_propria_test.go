package sessao

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type usuarioRepoSenhaMock struct {
	hashAtual          valueobject.SenhaHash
	chamouDefinirSenha bool
	hashDefinido       valueobject.SenhaHash
	instanteDefinido   time.Time
}

func (m *usuarioRepoSenhaMock) BuscarCredencialPropria(ctx context.Context, p autorizacao.Proprio) (*port.CredencialUsuario, error) {
	return &port.CredencialUsuario{ID: p.UsuarioID(), SenhaHash: m.hashAtual}, nil
}
func (m *usuarioRepoSenhaMock) DefinirSenhaPropria(ctx context.Context, p autorizacao.Proprio, hash valueobject.SenhaHash, instante time.Time) error {
	m.chamouDefinirSenha = true
	m.hashDefinido = hash
	m.instanteDefinido = instante
	return nil
}

func atorDeTeste(t *testing.T, usuarioID uuid.UUID) autorizacao.Ator {
	t.Helper()
	ator, err := autorizacao.NovoAtor(usuarioID, valueobject.ConjuntoDeAdministrador(), nil)
	if err != nil {
		t.Fatalf("ator de teste: %v", err)
	}
	return ator
}

type hashSenhaMock struct {
	conferirResultado bool
}

func (m *hashSenhaMock) Gerar(ctx context.Context, senha valueobject.SenhaEmTexto) (valueobject.SenhaHash, error) {
	return valueobject.NovaSenhaHash("$argon2id$v=19$m=1,t=1,p=1$c2FsdA$" + strings.ReplaceAll(senha.Revelar(), " ", "Xw"))
}
func (m *hashSenhaMock) Conferir(ctx context.Context, senha valueobject.SenhaEmTexto, hash valueobject.SenhaHash) (bool, error) {
	return m.conferirResultado, nil
}
func (m *hashSenhaMock) ConferirDescartavel(ctx context.Context, senha valueobject.SenhaEmTexto) {}

type uowFake struct{}

func (u uowFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type auditLoggerMock struct {
	eventos []auditoria.Evento
}

func (m *auditLoggerMock) Registrar(ctx context.Context, e auditoria.Evento) error {
	m.eventos = append(m.eventos, e)
	return nil
}

func hashDeTesteValido(t *testing.T) valueobject.SenhaHash {
	t.Helper()
	h, err := valueobject.NovaSenhaHash("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA")
	if err != nil {
		t.Fatalf("hash de teste: %v", err)
	}
	return h
}

func TestAlterarSenhaPropria_S02_TrocaComSucesso(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoSenhaMock{hashAtual: hashDeTesteValido(t)}
	hashAdapter := &hashSenhaMock{conferirResultado: true}
	audit := &auditLoggerMock{}
	agora := time.Now()
	uc := NovoAlterarSenhaPropriaUseCase(repo, hashAdapter, audit, relogioMock{agora: agora}, uowFake{})

	senhaAtual, _ := valueobject.NovaSenhaEmTexto("senha-atual-123")
	senhaNova, _ := valueobject.NovaSenhaEmTexto("colegiado")

	out, err := uc.Executar(context.Background(), AlterarSenhaPropriaInput{
		Ator: atorDeTeste(t, usuarioID), SenhaAtual: senhaAtual, SenhaNova: senhaNova,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !out.Instante.Equal(agora) {
		t.Fatalf("instante devolvido incorreto")
	}
	if !repo.chamouDefinirSenha {
		t.Fatal("esperava que DefinirSenhaPropria fosse chamado")
	}
	if !repo.instanteDefinido.Equal(agora) {
		t.Fatal("instante gravado deveria ser o mesmo devolvido ao handler")
	}
	if len(audit.eventos) != 1 || audit.eventos[0].Resultado != auditoria.ResultadoSucesso {
		t.Fatalf("esperava 1 evento de sucesso, obtido %+v", audit.eventos)
	}
}

func TestAlterarSenhaPropria_S03_SenhaAtualIncorreta(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoSenhaMock{hashAtual: hashDeTesteValido(t)}
	hashAdapter := &hashSenhaMock{conferirResultado: false}
	audit := &auditLoggerMock{}
	uc := NovoAlterarSenhaPropriaUseCase(repo, hashAdapter, audit, relogioMock{agora: time.Now()}, uowFake{})

	senhaAtual, _ := valueobject.NovaSenhaEmTexto("errada")
	senhaNova, _ := valueobject.NovaSenhaEmTexto("nova-senha")

	_, err := uc.Executar(context.Background(), AlterarSenhaPropriaInput{
		Ator: atorDeTeste(t, usuarioID), SenhaAtual: senhaAtual, SenhaNova: senhaNova,
	})
	if !errors.Is(err, domain.ErrSenhaAtualIncorreta) {
		t.Fatalf("esperava ErrSenhaAtualIncorreta, obtido %v", err)
	}
	if repo.chamouDefinirSenha {
		t.Fatal("nada deveria ser alterado quando a senha atual está incorreta")
	}
	if len(audit.eventos) != 1 || audit.eventos[0].Resultado != auditoria.ResultadoFalha {
		t.Fatalf("esperava 1 evento de falha, obtido %+v", audit.eventos)
	}
	for _, e := range audit.eventos {
		for _, v := range e.Detalhes {
			if s, ok := v.(string); ok && (strings.Contains(s, "errada") || strings.Contains(s, "nova-senha")) {
				t.Fatal("auditoria não pode conter qualquer parte da senha")
			}
		}
	}
}

func TestAlterarSenhaPropria_S09_Senha100CaracteresComAcentoEEspaco(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoSenhaMock{hashAtual: hashDeTesteValido(t)}
	hashAdapter := &hashSenhaMock{conferirResultado: true}
	audit := &auditLoggerMock{}
	uc := NovoAlterarSenhaPropriaUseCase(repo, hashAdapter, audit, relogioMock{agora: time.Now()}, uowFake{})

	senhaAtual, _ := valueobject.NovaSenhaEmTexto("atual")
	original := strings.Repeat("á é í ó ú ç ão ", 6) + "final-com-acento"
	senhaNova, err := valueobject.NovaSenhaEmTexto(original)
	if err != nil {
		t.Fatalf("senha de 100 caracteres deveria ser aceita: %v", err)
	}

	_, err = uc.Executar(context.Background(), AlterarSenhaPropriaInput{
		Ator: atorDeTeste(t, usuarioID), SenhaAtual: senhaAtual, SenhaNova: senhaNova,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}

func TestAlterarSenhaPropria_S11_SenhaCurtaEhAceita(t *testing.T) {
	usuarioID := uuid.Must(uuid.NewV7())
	repo := &usuarioRepoSenhaMock{hashAtual: hashDeTesteValido(t)}
	hashAdapter := &hashSenhaMock{conferirResultado: true}
	audit := &auditLoggerMock{}
	uc := NovoAlterarSenhaPropriaUseCase(repo, hashAdapter, audit, relogioMock{agora: time.Now()}, uowFake{})

	senhaAtual, _ := valueobject.NovaSenhaEmTexto("atual")
	senhaNova, err := valueobject.NovaSenhaEmTexto("ata")
	if err != nil {
		t.Fatalf("senha curta deveria ser aceita: %v", err)
	}

	if _, err := uc.Executar(context.Background(), AlterarSenhaPropriaInput{
		Ator: atorDeTeste(t, usuarioID), SenhaAtual: senhaAtual, SenhaNova: senhaNova,
	}); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
}

// S-10 (senha vazia) e S-12 (acima de 1024 runas) são garantidos pelo tipo
// valueobject.SenhaEmTexto — não é possível montar um Input com uma senha
// nessas condições, porque NovaSenhaEmTexto já as recusa (ver T-001).
