package sessao

import (
	"context"
	"time"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AutenticarInput struct {
	InstituicaoID *uuid.UUID
	Email         valueobject.Email
	Senha         valueobject.SenhaEmTexto
}

type AutenticarOutput struct {
	Ator                   autorizacao.Ator
	InstanteDeAutenticacao time.Time
}

// AutenticarUseCase implementa L-01 a L-17 (design.md §2.1). A mensagem de
// erro nunca distingue causa (3.1) — todo caminho de falha devolve
// domain.ErrCredenciaisInvalidas.
//
// fusoDeExibicao existe para fixar Ator.DataDeReferencia (fundacao-metas.md
// §5.2) — o login não passa pelo middleware de sessão (a rota é pública),
// então é o único outro lugar que constrói um Ator do zero e precisa da
// mesma disciplina de "uma leitura do relógio por requisição", com o
// mesmo uc.relogio que já usa para InstanteDeAutenticacao.
type AutenticarUseCase struct {
	repo           port.AutenticacaoRepository
	hash           port.HashDeSenha
	relogio        port.Relogio
	fusoDeExibicao *time.Location
}

func NovoAutenticarUseCase(repo port.AutenticacaoRepository, hash port.HashDeSenha, relogio port.Relogio, fusoDeExibicao *time.Location) *AutenticarUseCase {
	return &AutenticarUseCase{repo: repo, hash: hash, relogio: relogio, fusoDeExibicao: fusoDeExibicao}
}

func (uc *AutenticarUseCase) Executar(ctx context.Context, in AutenticarInput) (AutenticarOutput, error) {
	credencial, err := uc.repo.BuscarCredencial(ctx, in.InstituicaoID, in.Email)
	if err != nil {
		return AutenticarOutput{}, err
	}

	// Sem conta, conta excluída ou instituição inativa: mesmo tratamento —
	// verificação de hash descartável para manter o tempo constante (L-08),
	// contra o hash de referência do adapter, nunca contra dado real.
	if credencial == nil {
		uc.hash.ConferirDescartavel(ctx, in.Senha)
		return AutenticarOutput{}, domain.ErrCredenciaisInvalidas
	}

	confere, err := uc.hash.Conferir(ctx, in.Senha, credencial.SenhaHash)
	if err != nil {
		return AutenticarOutput{}, err
	}
	if !confere {
		return AutenticarOutput{}, domain.ErrCredenciaisInvalidas
	}

	ator, err := autorizacao.NovoAtor(credencial.ID, credencial.Perfis, credencial.InstituicaoID)
	if err != nil {
		return AutenticarOutput{}, err
	}
	instante := uc.relogio.Agora()
	ator = ator.ComDataDeReferencia(valueobject.DataLocalDe(instante, uc.fusoDeExibicao))

	return AutenticarOutput{Ator: ator, InstanteDeAutenticacao: instante}, nil
}
