package autorizacao

import "github.com/basis-avalia/backend/internal/domain/valueobject"

// Permissao vive em valueobject (Perfil.Pode a usa) para evitar import
// cycle — este pacote reexporta via alias, que é o nome que o resto do
// sistema (rotas, use cases, métricas) usa (design.md §3.2).
type Permissao = valueobject.Permissao

const (
	InstituicaoListar      = valueobject.InstituicaoListar
	InstituicaoCriar       = valueobject.InstituicaoCriar
	InstituicaoEditar      = valueobject.InstituicaoEditar
	InstituicaoInativar    = valueobject.InstituicaoInativar
	PIGerenciar            = valueobject.PIGerenciar
	UsuarioListar          = valueobject.UsuarioListar
	UsuarioCriar           = valueobject.UsuarioCriar
	UsuarioEditar          = valueobject.UsuarioEditar
	UsuarioExcluir         = valueobject.UsuarioExcluir
	UsuarioRedefinirSenha  = valueobject.UsuarioRedefinirSenha
	AdministradorGerenciar = valueobject.AdministradorGerenciar

	IndicadorPlataformaGerenciar = valueobject.IndicadorPlataformaGerenciar
	IndicadorListar              = valueobject.IndicadorListar
	IndicadorGerenciar           = valueobject.IndicadorGerenciar
	MetaListar                   = valueobject.MetaListar
	MetaGerenciar                = valueobject.MetaGerenciar
	CursoListar                  = valueobject.CursoListar
	CursoGerenciar               = valueobject.CursoGerenciar
	CursoLerProprio              = valueobject.CursoLerProprio
	DesignacaoGerenciar          = valueobject.DesignacaoGerenciar
	PeriodoGerenciar             = valueobject.PeriodoGerenciar
	PlanoListar                  = valueobject.PlanoListar
	PlanoGerenciar               = valueobject.PlanoGerenciar
	PlanoLerProprio              = valueobject.PlanoLerProprio
	EntregaRegistrar             = valueobject.EntregaRegistrar
	EntregaListar                = valueobject.EntregaListar
	EntregaAvaliar               = valueobject.EntregaAvaliar
	RelatorioLer                 = valueobject.RelatorioLer
	RelatorioExportar            = valueobject.RelatorioExportar
	RelatorioLerProprio          = valueobject.RelatorioLerProprio

	Nenhuma       = valueobject.Nenhuma
	EscopoProprio = valueobject.EscopoProprio
)
