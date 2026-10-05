package port

// ItemDoDocumento — uma linha da tabela de metas do .docx (specs/plano-acao/
// spec.md 3.5, item 6).
type ItemDoDocumento struct {
	MetaNome    string
	Indicadores []string // "1.4 · Do INEP", já formatado (código + origem)
	Quantidade  int
}

// DadosDoDocumento é tudo que o gerador precisa para substituir os
// marcadores do modelo — nenhum dado é buscado pelo adapter, ele só
// renderiza o que o use case já reuniu (specs/plano-acao/design.md §7.5).
type DadosDoDocumento struct {
	InstituicaoNome, InstituicaoSigla                      string
	CursoNome, CursoGrau, CursoModalidade, CursoCodigoEMec string
	CoordenadorNome, CoordenadorPortaria                   string // vazios = "Sem responsável"
	PeriodoNome, PeriodoInicio, PeriodoFim                 string

	Descricao, ObjetivoGeral, ResultadosEsperados string
	AlinhamentoPDI, AlinhamentoPPC                string

	Itens []ItemDoDocumento

	// AprovacaoData/AprovacaoOrgao vazios = "Não aprovado". Pendente=true
	// só quando o plano está vigente sem aprovação (SI-05/SI-07): "Aprovação
	// ainda não registrada" nunca aparece no rascunho.
	AprovacaoData, AprovacaoOrgao string
	AprovacaoPendente             bool

	// MarcaSituacao — "RASCUNHO", "ENCERRADO — período encerrado em
	// DD/MM/AAAA", ou vazio quando vigente (PP-7).
	MarcaSituacao string

	GeradoEmTexto string // data e hora de emissão, já formatada
}

// GeradorDeDocumento — modelo .docx com marcadores, sem container de
// conversão (P-07). Falha alta (P-08, domain.ErrMarcadorNaoEncontrado)
// quando um marcador declarado pelo modelo não é encontrado nos dados —
// nunca produz documento com "{{" vazado.
type GeradorDeDocumento interface {
	Gerar(dados DadosDoDocumento) ([]byte, error)
}
