package usuario

import (
	"strings"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

// Usuario — entidade de domínio (design.md §3.3).
type Usuario struct {
	ID                      uuid.UUID
	InstituicaoID           *uuid.UUID
	Nome                    string
	Email                   valueobject.Email
	SenhaHash               *valueobject.SenhaHash
	Perfis                  valueobject.ConjuntoDePerfis
	SenhaProvisoria         bool
	SessoesValidasAPartirDe time.Time
	ProvedorIdentidade      valueobject.ProvedorIdentidade
	IdentificadorExterno    *string
	CriadoEm                time.Time
	AtualizadoEm            *time.Time
	ExcluidoEm              *time.Time
	Versao                  int
}

func validarNome(bruto string) (string, error) {
	nome := strings.TrimSpace(bruto)
	if nome == "" {
		return "", &domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."}
	}
	if len(nome) > 200 {
		return "", &domain.ErrValidacao{Campo: "nome", Mensagem: "O nome é obrigatório."}
	}
	return nome, nil
}

// NovoUsuario gera o UUIDv7 no domínio (nunca no banco) e aplica as
// invariantes de nascimento: senha provisória, provedor local, sessões
// válidas a partir da criação (U-01).
func NovoUsuario(nomeBruto string, email valueobject.Email, senhaHash valueobject.SenhaHash, perfis valueobject.ConjuntoDePerfis, instituicaoID *uuid.UUID) (*Usuario, error) {
	nome, err := validarNome(nomeBruto)
	if err != nil {
		return nil, err
	}
	agora := time.Now()
	return &Usuario{
		ID:                      uuid.Must(uuid.NewV7()),
		InstituicaoID:           instituicaoID,
		Nome:                    nome,
		Email:                   email,
		SenhaHash:               &senhaHash,
		Perfis:                  perfis,
		SenhaProvisoria:         true,
		SessoesValidasAPartirDe: agora,
		ProvedorIdentidade:      valueobject.CredencialLocal,
		IdentificadorExterno:    nil,
		CriadoEm:                agora,
		Versao:                  1,
	}, nil
}

func (u *Usuario) RenomearEReenderecar(nomeBruto string, email valueobject.Email) error {
	nome, err := validarNome(nomeBruto)
	if err != nil {
		return err
	}
	u.Nome = nome
	u.Email = email
	return nil
}

// DefinirPerfis substitui o conjunto inteiro e devolve o anterior, para a
// auditoria registrar os dois completos (E-02, spec §11). Nunca aplica
// diferença: o chamador informa o conjunto final (design.md §3.3).
func (u *Usuario) DefinirPerfis(novo valueobject.ConjuntoDePerfis) (anterior valueobject.ConjuntoDePerfis) {
	anterior = u.Perfis
	u.Perfis = novo
	return anterior
}

// DefinirSenhaPropria é usado por alterar_senha_propria: zera a senha
// provisória e invalida as sessões anteriores no mesmo instante (§7.1).
func (u *Usuario) DefinirSenhaPropria(hash valueobject.SenhaHash, instante time.Time) {
	u.SenhaHash = &hash
	u.SenhaProvisoria = false
	u.SessoesValidasAPartirDe = instante
}

// ReceberSenhaProvisoria é usado por redefinir_senha_usuario: a conta
// volta ao estado de primeiro acesso e a sessão em andamento cai (E-11).
func (u *Usuario) ReceberSenhaProvisoria(hash valueobject.SenhaHash, instante time.Time) {
	u.SenhaHash = &hash
	u.SenhaProvisoria = true
	u.SessoesValidasAPartirDe = instante
}

// ExcluirLogicamente marca ExcluidoEm e anula o hash — a credencial não
// tem mais finalidade (3.8). Nome, e-mail e perfis são preservados.
func (u *Usuario) ExcluirLogicamente(instante time.Time) {
	u.ExcluidoEm = &instante
	u.SenhaHash = nil
}

func (u *Usuario) InvalidarSessoes(instante time.Time) {
	u.SessoesValidasAPartirDe = instante
}
