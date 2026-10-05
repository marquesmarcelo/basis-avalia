package autorizacao

import (
	"github.com/basis-avalia/backend/internal/domain"
	"github.com/basis-avalia/backend/internal/domain/valueobject"
	"github.com/google/uuid"
)

type Acao int

const (
	AcaoListar Acao = iota
	AcaoBuscar
	AcaoCriar
	AcaoEditar
	AcaoExcluir
	AcaoRedefinirSenha
	AcaoInativar
	AcaoAvaliar  // entrega.avaliar (fundacao-metas.md §6.1)
	AcaoExportar // relatorio.exportar
)

func permissaoExigida(alcance Alcance, acao Acao) (Permissao, error) {
	switch alcance {
	case UsuariosDaPropriaInstituicao:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return UsuarioListar, nil
		case AcaoCriar:
			return UsuarioCriar, nil
		case AcaoEditar:
			return UsuarioEditar, nil
		case AcaoExcluir:
			return UsuarioExcluir, nil
		case AcaoRedefinirSenha:
			return UsuarioRedefinirSenha, nil
		}
	case PesquisadoresDeUmaInstituicao:
		return PIGerenciar, nil
	case AdministradoresDaPlataforma:
		return AdministradorGerenciar, nil
	case InstituicoesDaPlataforma:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return InstituicaoListar, nil
		case AcaoCriar:
			return InstituicaoCriar, nil
		case AcaoEditar:
			return InstituicaoEditar, nil
		case AcaoInativar:
			return InstituicaoInativar, nil
		}
	case IndicadoresDaPlataforma:
		return IndicadorPlataformaGerenciar, nil
	case CatalogoDeIndicadores:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return IndicadorListar, nil
		default:
			return IndicadorGerenciar, nil
		}
	case MetasDaInstituicao:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return MetaListar, nil
		default:
			return MetaGerenciar, nil
		}
	case CursosDaInstituicao:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return CursoListar, nil
		default:
			return CursoGerenciar, nil
		}
	case DesignacoesDaInstituicao:
		return DesignacaoGerenciar, nil
	case PeriodosDaInstituicao:
		return PeriodoGerenciar, nil
	case PlanosDaInstituicao:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return PlanoListar, nil
		default:
			return PlanoGerenciar, nil
		}
	case EntregasDaInstituicao:
		switch acao {
		case AcaoListar, AcaoBuscar:
			return EntregaListar, nil
		default:
			return EntregaAvaliar, nil
		}
	case DesempenhoDaInstituicao:
		switch acao {
		case AcaoExportar:
			return RelatorioExportar, nil
		default:
			return RelatorioLer, nil
		}
	case CursosDaCarteira:
		return CursoLerProprio, nil
	case PlanosDaCarteira:
		return PlanoLerProprio, nil
	case EntregasDaCarteira:
		return EntregaRegistrar, nil
	case DesempenhoDaCarteira:
		return RelatorioLerProprio, nil
	}
	return Nenhuma, domain.ErrEscopoInvalido
}

// Autorizar é a única porta de entrada para um Escopo. Mapeia
// (alcance, ação) para a permissão exigida, confere contra a UNIÃO das
// permissões do conjunto de perfis do ator (A-07) e devolve o recorte
// correspondente (design.md §3.2).
//
// instituicaoDoCaminho é obrigatório em PesquisadoresDeUmaInstituicao e
// ignorado nos demais alcances.
func Autorizar(ator Ator, alcance Alcance, acao Acao, instituicaoDoCaminho *uuid.UUID) (Escopo, error) {
	permissao, err := permissaoExigida(alcance, acao)
	if err != nil {
		return Escopo{}, err
	}
	if !ator.perfis.Pode(permissao) {
		return Escopo{}, domain.ErrPermissaoNegada
	}

	switch alcance {
	case UsuariosDaPropriaInstituicao:
		return Escopo{
			valido:           true,
			instituicaoID:    ator.instituicaoID,
			dataDeReferencia: ator.dataDeReferencia,
		}, nil
	case PesquisadoresDeUmaInstituicao:
		if instituicaoDoCaminho == nil {
			return Escopo{}, domain.ErrEscopoInvalido
		}
		exige := valueobject.PesquisadorInstitucional
		return Escopo{
			valido:           true,
			instituicaoID:    instituicaoDoCaminho,
			exigePerfil:      &exige,
			dataDeReferencia: ator.dataDeReferencia,
		}, nil
	case AdministradoresDaPlataforma:
		exige := valueobject.AdministradorSistema
		return Escopo{
			valido:           true,
			plataforma:       true,
			exigePerfil:      &exige,
			dataDeReferencia: ator.dataDeReferencia,
		}, nil
	case InstituicoesDaPlataforma:
		return Escopo{valido: true, plataforma: true, dataDeReferencia: ator.dataDeReferencia}, nil
	case IndicadoresDaPlataforma:
		return Escopo{valido: true, plataforma: true, dataDeReferencia: ator.dataDeReferencia}, nil
	case CatalogoDeIndicadores, MetasDaInstituicao, CursosDaInstituicao, DesignacoesDaInstituicao,
		PeriodosDaInstituicao, PlanosDaInstituicao, EntregasDaInstituicao, DesempenhoDaInstituicao:
		return Escopo{
			valido:           true,
			instituicaoID:    ator.instituicaoID,
			dataDeReferencia: ator.dataDeReferencia,
		}, nil
	case CursosDaCarteira, PlanosDaCarteira, EntregasDaCarteira, DesempenhoDaCarteira:
		// Os quatro alcances de carteira ligam a restrição SEMPRE, nunca
		// condicionalmente — a rota é que escolhe o alcance
		// (fundacao-metas.md §6.2, doutrina D-08 herdada).
		usuarioID := ator.usuarioID
		return Escopo{
			valido:              true,
			instituicaoID:       ator.instituicaoID,
			restritoACarteiraDe: &usuarioID,
			dataDeReferencia:    ator.dataDeReferencia,
		}, nil
	}
	return Escopo{}, domain.ErrEscopoInvalido
}
