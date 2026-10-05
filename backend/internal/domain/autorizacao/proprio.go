package autorizacao

import "github.com/google/uuid"

// Proprio identifica o próprio ator numa operação sobre a identidade dele
// (trocar a própria senha, encerrar a própria sessão). Campo não exportado
// e nenhum construtor: só nasce de Ator.Proprio(), e Ator só nasce de
// NovoAtor, chamado pelo middleware de sessão a partir dos claims do token
// validado — é o análogo do Escopo para o eixo "sobre quem", quando a
// resposta é "sobre mim" (design.md §4.4).
type Proprio struct {
	usuarioID uuid.UUID
	valido    bool
}

func (p Proprio) UsuarioID() uuid.UUID { return p.usuarioID }
func (p Proprio) Valido() bool         { return p.valido }
