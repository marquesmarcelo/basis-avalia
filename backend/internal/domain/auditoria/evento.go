package auditoria

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type Acao string

const (
	AlterarSenhaPropria    Acao = "alterar_senha_propria"
	RedefinirSenhaUsuario  Acao = "redefinir_senha_usuario"
	CriarUsuario           Acao = "criar_usuario"
	AtualizarUsuario       Acao = "atualizar_usuario"
	ExcluirUsuario         Acao = "excluir_usuario"
	CriarAdministrador     Acao = "criar_administrador"
	AtualizarAdministrador Acao = "atualizar_administrador"
	ExcluirAdministrador   Acao = "excluir_administrador"
	CriarInstituicao       Acao = "criar_instituicao"
	AtualizarInstituicao   Acao = "atualizar_instituicao"
	InativarInstituicao    Acao = "inativar_instituicao"
	ReativarInstituicao    Acao = "reativar_instituicao"
	AcessoNegado           Acao = "acesso_negado"

	// specs/indicadores/spec.md §10 — catálogo do INEP (comum à
	// instalação) e catálogo de metas da instituição.
	CriarIndicadorInep       Acao = "criar_indicador_inep"
	AtualizarIndicadorInep   Acao = "atualizar_indicador_inep"
	InativarIndicadorInep    Acao = "inativar_indicador_inep"
	ReativarIndicadorInep    Acao = "reativar_indicador_inep"
	ExcluirIndicadorInep     Acao = "excluir_indicador_inep"
	CriarIndicador           Acao = "criar_indicador"
	AtualizarIndicador       Acao = "atualizar_indicador"
	AlterarSituacaoIndicador Acao = "alterar_situacao_indicador"
	ExcluirIndicador         Acao = "excluir_indicador"
	CriarMeta                Acao = "criar_meta"
	AtualizarMeta            Acao = "atualizar_meta"
	AlterarSituacaoMeta      Acao = "alterar_situacao_meta"
	ExcluirMeta              Acao = "excluir_meta"

	// specs/plano-acao/spec.md §10 — período, plano de ação
	// curso/coordenador, itens e documento.
	CriarPeriodo       Acao = "criar_periodo"
	AtualizarPeriodo   Acao = "atualizar_periodo"
	ExcluirPeriodo     Acao = "excluir_periodo"
	CriarPlano         Acao = "criar_plano"
	AtualizarPlano     Acao = "atualizar_plano"
	PublicarPlano      Acao = "publicar_plano"
	DespublicarPlano   Acao = "despublicar_plano"
	EncerrarPlano      Acao = "encerrar_plano"
	ReabrirPlano       Acao = "reabrir_plano"
	ExcluirPlano       Acao = "excluir_plano"
	CopiarPlano        Acao = "copiar_plano"
	CriarItemPlano     Acao = "criar_item_plano"
	AtualizarItemPlano Acao = "atualizar_item_plano"
	ExcluirItemPlano   Acao = "excluir_item_plano"
	GerarDocumento     Acao = "gerar_documento"
	BaixarDocumento    Acao = "baixar_documento"

	// specs/cursos/spec.md §10 — curso e designação de coordenação.
	CriarCurso          Acao = "criar_curso"
	AtualizarCurso      Acao = "atualizar_curso"
	InativarCurso       Acao = "inativar_curso"
	ReativarCurso       Acao = "reativar_curso"
	ExcluirCurso        Acao = "excluir_curso"
	CriarDesignacao     Acao = "criar_designacao"
	AtualizarDesignacao Acao = "atualizar_designacao"
	ExcluirDesignacao   Acao = "excluir_designacao"

	// specs/metas-coordenacao/spec.md §10 — entrega, anexo, avaliação e
	// relatório de desempenho.
	RegistrarEntrega            Acao = "registrar_entrega"
	CorrigirEntrega             Acao = "corrigir_entrega"
	ExcluirEntrega              Acao = "excluir_entrega"
	AvaliarEntrega              Acao = "avaliar_entrega"
	DesfazerAceitacao           Acao = "desfazer_aceitacao"
	RestaurarPrazoPorVacancia   Acao = "restaurar_prazo_por_vacancia"
	BaixarAnexo                 Acao = "baixar_anexo"
	ExportarRelatorioDesempenho Acao = "exportar_relatorio_desempenho"
)

// acoesQueVaoParaSyslogSempre implementa a coluna "Syslog" da §11 da spec
// para o caso incondicional. AtualizarUsuario e AtualizarAdministrador são
// as exceções condicionais ("✅ quando os perfis mudaram") e são decididas
// em adapter/auditoria/composto.go, olhando Detalhes["perfis_anterior"].
var acoesQueVaoParaSyslogSempre = map[Acao]bool{
	AlterarSenhaPropria:   true,
	RedefinirSenhaUsuario: true,
	ExcluirUsuario:        true,
	ExcluirAdministrador:  true,
	CriarInstituicao:      true,
	InativarInstituicao:   true,
	ReativarInstituicao:   true,
	AcessoNegado:          true,

	// Toda operação sobre o catálogo comum vai para o syslog, sem exceção
	// (specs/indicadores/spec.md §10, design.md §8): é a única escrita do
	// sistema que afeta todas as instituições de uma vez.
	CriarIndicadorInep:     true,
	AtualizarIndicadorInep: true,
	InativarIndicadorInep:  true,
	ReativarIndicadorInep:  true,
	ExcluirIndicadorInep:   true,

	// Indicador e meta próprios: só a exclusão vai ao syslog (spec.md §10).
	ExcluirIndicador: true,
	ExcluirMeta:      true,

	// plano-acao: só exclusão de período e de plano — mesma doutrina
	// acima (specs/plano-acao/spec.md §10).
	ExcluirPeriodo: true,
	ExcluirPlano:   true,

	// specs/cursos/spec.md §10: toda ação sobre designação (o fato
	// auditado que sustenta o perfil derivado) vai ao syslog, sempre.
	CriarDesignacao:     true,
	AtualizarDesignacao: true,
	ExcluirDesignacao:   true,

	// specs/metas-coordenacao/spec.md §10 — excluir, avaliar e desfazer
	// vão sempre ao syslog; baixar_anexo só quando negado (condicional,
	// decidido em adapter/auditoria/composto.go, mesmo padrão de
	// AtualizarUsuario).
	ExcluirEntrega:               true,
	AvaliarEntrega:               true,
	DesfazerAceitacao:            true,
	ExportarRelatorioDesempenho:  true,
}

func (a Acao) VaiParaSyslog() bool {
	return acoesQueVaoParaSyslogSempre[a]
}

type Resultado string

const (
	ResultadoSucesso Resultado = "sucesso"
	ResultadoNegado  Resultado = "negado"
	ResultadoFalha   Resultado = "falha"
	ResultadoErro    Resultado = "erro"
)

// Evento — o que entra em Detalhes é limitado por design (design.md §8.1):
// nomes de campos alterados, perfis_anterior/perfis_novo — os dois
// conjuntos completos, nunca a diferença — e permissao_exigida. Nunca
// senha, hash, token ou valor de cookie.
type Evento struct {
	ID            uuid.UUID
	Acao          Acao
	Resultado     Resultado
	AtorID        *uuid.UUID
	InstituicaoID *uuid.UUID
	RecursoTipo   string
	RecursoID     *uuid.UUID
	Detalhes      map[string]any
	IPOrigem      netip.Addr
	ExecutadoEm   time.Time
}

func NovoEvento(acao Acao, resultado Resultado) Evento {
	return Evento{
		ID:          uuid.Must(uuid.NewV7()),
		Acao:        acao,
		Resultado:   resultado,
		Detalhes:    map[string]any{},
		ExecutadoEm: time.Now(),
	}
}
