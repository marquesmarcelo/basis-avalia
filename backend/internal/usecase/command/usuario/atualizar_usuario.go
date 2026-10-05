package usuario

import (
	"context"

	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/auditoria"
	"github.com/basis-avalia/backend/internal/domain/autorizacao"
	"github.com/basis-avalia/backend/internal/domain/usuario"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/basis-avalia/backend/internal/port"
	"github.com/google/uuid"
)

type AtualizarUsuarioInput struct {
	Ator                 autorizacao.Ator
	Alcance              autorizacao.Alcance
	InstituicaoDoCaminho *uuid.UUID
	UsuarioID            uuid.UUID
	Nome                 string
	Email                valueobject.Email
	Perfis               []valueobject.Perfil // só lido em UsuariosDaPropriaInstituicao (R20, R26 não trocam perfil)
	Versao               int
}

// perfisRemovidos devolve os perfis que estavam em anterior e não estão
// mais em novo — é o que decide se a invariante de último detentor
// precisa ser verificada (design.md §5.8), nunca a lista completa.
func perfisRemovidos(anterior, novo valueobject.ConjuntoDePerfis) []valueobject.Perfil {
	var removidos []valueobject.Perfil
	for _, p := range anterior.Ordenado() {
		if !novo.Possui(p) {
			removidos = append(removidos, p)
		}
	}
	return removidos
}

// verificarInvarianteDeUltimoDetentor implementa design.md §5.8: trava
// antes de contar, para fechar a corrida (E-17, AS-06) — chamada sempre
// que pesquisador_institucional ou administrador_sistema sai do conjunto
// de alguém, seja por exclusão ou por edição do conjunto.
func verificarInvarianteDeUltimoDetentor(ctx context.Context, repo port.UsuarioRepository, esc autorizacao.Escopo, removidos []valueobject.Perfil) error {
	for _, perfil := range removidos {
		if perfil != valueobject.PesquisadorInstitucional && perfil != valueobject.AdministradorSistema {
			continue
		}
		if err := repo.TravarPopulacao(ctx, esc); err != nil {
			return err
		}
		total, err := repo.ContarDetentoresDoPerfil(ctx, esc, perfil)
		if err != nil {
			return err
		}
		if total <= 1 {
			if perfil == valueobject.PesquisadorInstitucional {
				return domain.ErrUltimoPesquisadorInstitucional
			}
			return domain.ErrUltimoAdministradorSistema
		}
	}
	return nil
}

// AtualizarUsuarioUseCase cobre E-01..E-04, E-09, E-10, E-14..E-17, T-03.
type AtualizarUsuarioUseCase struct {
	repo  port.UsuarioRepository
	audit port.AuditLogger
	uow   port.UnidadeDeTrabalho
}

func NovoAtualizarUsuarioUseCase(repo port.UsuarioRepository, audit port.AuditLogger, uow port.UnidadeDeTrabalho) *AtualizarUsuarioUseCase {
	return &AtualizarUsuarioUseCase{repo: repo, audit: audit, uow: uow}
}

func (uc *AtualizarUsuarioUseCase) Executar(ctx context.Context, in AtualizarUsuarioInput) (usuario.Usuario, error) {
	esc, err := autorizacao.Autorizar(in.Ator, in.Alcance, autorizacao.AcaoEditar, in.InstituicaoDoCaminho)
	if err != nil {
		return usuario.Usuario{}, err
	}

	// O conjunto final só é lido do payload no alcance que permite trocar
	// perfil (R20, R26 nunca trocam — design.md §6.2).
	trocaPerfis := in.Alcance == autorizacao.UsuariosDaPropriaInstituicao

	var resultado usuario.Usuario
	erro := uc.uow.Executar(ctx, func(ctx context.Context) error {
		alvo, err := uc.repo.BuscarPorID(ctx, esc, in.UsuarioID) // 404 por escopo (T-03)
		if err != nil {
			return err
		}

		campos := []string{"nome", "email"}
		perfisAnterior := alvo.Perfis
		perfisMudaram := false

		if trocaPerfis {
			perfisNovo, err := valueobject.ConjuntoInstitucional(in.Perfis)
			if err != nil {
				return err
			}
			if !perfisNovo.Igual(perfisAnterior) {
				if alvo.ID == in.Ator.UsuarioID() {
					return domain.ErrAlteracaoDosPropriosPerfisNegada // E-04
				}

				// E-09/E-17/AS-06: a trava roda ANTES da contagem, e só
				// para os perfis que efetivamente saíram do conjunto
				// (E-14, E-15: promover ou retirar outro perfil nunca
				// aciona a invariante).
				if err := verificarInvarianteDeUltimoDetentor(ctx, uc.repo, esc, perfisRemovidos(perfisAnterior, perfisNovo)); err != nil {
					return err
				}

				alvo.DefinirPerfis(perfisNovo)
				campos = append(campos, "perfis")
				perfisMudaram = true
			}
		}

		if err := alvo.RenomearEReenderecar(in.Nome, in.Email); err != nil {
			return err
		}

		if err := uc.repo.Atualizar(ctx, esc, alvo, in.Versao); err != nil {
			return err
		}
		if perfisMudaram {
			if err := uc.repo.SubstituirPerfis(ctx, esc, alvo.ID, alvo.Perfis); err != nil {
				return err
			}
		}

		acao := auditoria.AtualizarUsuario
		if in.Alcance == autorizacao.AdministradoresDaPlataforma {
			acao = auditoria.AtualizarAdministrador
		}
		atorID := in.Ator.UsuarioID()
		evento := auditoria.NovoEvento(acao, auditoria.ResultadoSucesso)
		evento.AtorID = &atorID
		evento.InstituicaoID = alvo.InstituicaoID
		evento.RecursoTipo = "Usuario"
		evento.RecursoID = &alvo.ID
		evento.Detalhes["campos_alterados"] = campos
		if perfisMudaram {
			// Os dois conjuntos completos, em ordem canônica, nunca a
			// diferença (E-02, design.md §8.1) — é o que distingue
			// correção legítima de retirada indevida de controle.
			evento.Detalhes["perfis_anterior"] = perfisComoTexto(perfisAnterior)
			evento.Detalhes["perfis_novo"] = perfisComoTexto(alvo.Perfis)
		}
		if err := uc.audit.Registrar(ctx, evento); err != nil {
			return err
		}
		resultado = *alvo
		return nil
	})
	if erro != nil {
		return usuario.Usuario{}, erro
	}
	return resultado, nil
}

func perfisComoTexto(c valueobject.ConjuntoDePerfis) []string {
	ordenado := c.Ordenado()
	texto := make([]string, len(ordenado))
	for i, p := range ordenado {
		texto[i] = string(p)
	}
	return texto
}
