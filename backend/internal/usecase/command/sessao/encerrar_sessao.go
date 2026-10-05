package sessao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/port"
)

// autenticacaoDeLogout — interface estreita: só o método que opera sobre a
// própria sessão, via Proprio (design.md §4.4).
type autenticacaoDeLogout interface {
	InvalidarSessoesProprias(ctx context.Context, p autorizacao.Proprio, instante time.Time) error
}

// EncerrarSessaoUseCase — logout (SE-01). Idempotente por natureza: gravar
// sessoes_validas_a_partir_de de novo não muda o resultado observável.
// Não gera auditoria (3.2): sem registro de entrada, o de saída não
// produz informação utilizável.
type EncerrarSessaoUseCase struct {
	autenticacao autenticacaoDeLogout
	relogio      port.Relogio
}

func NovoEncerrarSessaoUseCase(autenticacao autenticacaoDeLogout, relogio port.Relogio) *EncerrarSessaoUseCase {
	return &EncerrarSessaoUseCase{autenticacao: autenticacao, relogio: relogio}
}

func (uc *EncerrarSessaoUseCase) Executar(ctx context.Context, ator autorizacao.Ator) error {
	return uc.autenticacao.InvalidarSessoesProprias(ctx, ator.Proprio(), uc.relogio.Agora())
}
