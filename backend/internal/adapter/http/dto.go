package http

// Envelope de erro e de listagem seguem o padrão de project.config.md e
// CLAUDE.md — este arquivo só define os corpos específicos desta feature.

type LoginRequest struct {
	InstituicaoID *string `json:"instituicao_id"`
	Email         string  `json:"email"`
	Senha         string  `json:"senha"`
}

type InstituicaoResumoResponse struct {
	ID    string `json:"id"`
	Nome  string `json:"nome"`
	Sigla string `json:"sigla"`
}

// ContextoDeSessaoResponse é o corpo de POST /auth/login e GET /auth/eu —
// o mesmo formato nos dois (design.md R2, R4). Perfis é o conjunto
// EFETIVO; PerfisDerivados é o subconjunto que não está gravado em
// usuario_perfil — mudança aditiva de specs/cursos, fundacao-metas.md
// §4.2 (nunca "coordena_hoje" na API, só o fato exposto: quantos cursos).
type ContextoDeSessaoResponse struct {
	ID                string                     `json:"id"`
	Nome              string                     `json:"nome"`
	Email             string                     `json:"email"`
	Perfis            []string                   `json:"perfis"`
	PerfisRotulos     []string                   `json:"perfis_rotulos"`
	PerfisDerivados   []string                   `json:"perfis_derivados"`
	CursosCoordenados int                        `json:"cursos_coordenados"`
	SenhaProvisoria   bool                       `json:"senha_provisoria"`
	Instituicao       *InstituicaoResumoResponse `json:"instituicao"`
	Permissoes        []string                   `json:"permissoes"`
}

type InstituicaoPublicaResponse struct {
	ID    string `json:"id"`
	Nome  string `json:"nome"`
	Sigla string `json:"sigla"`
}

type AlterarSenhaRequest struct {
	SenhaAtual string `json:"senha_atual"`
	SenhaNova  string `json:"senha_nova"`
}

// MetaPaginacao é o envelope de metadados de toda listagem paginada
// (design.md §6.2) — nunca array solto.
type MetaPaginacao struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func calcularTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	paginas := total / pageSize
	if total%pageSize != 0 {
		paginas++
	}
	return paginas
}

type InstituicaoRequest struct {
	Nome       string `json:"nome"`
	Sigla      string `json:"sigla"`
	CodigoEMec string `json:"codigo_emec"`
}

type InstituicaoAtualizarRequest struct {
	Nome       string `json:"nome"`
	Sigla      string `json:"sigla"`
	CodigoEMec string `json:"codigo_emec"`
	Versao     int    `json:"versao"`
}

type AlterarSituacaoRequest struct {
	Situacao string `json:"situacao"`
	Versao   int    `json:"versao"`
}

// UsuarioRequest é o corpo de POST /usuarios (Pesquisador Institucional
// criando usuário da própria instituição) — o único alcance que aceita
// "perfis" no corpo (design.md §4.5, §16 T-099).
type UsuarioRequest struct {
	Nome   string   `json:"nome"`
	Email  string   `json:"email"`
	Perfis []string `json:"perfis"`
	Senha  string   `json:"senha"`
}

// PesquisadorRequest é o corpo de POST /instituicoes/{id}/pesquisadores —
// sem "perfis": o alcance é fixo (pesquisador_institucional).
type PesquisadorRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// AdministradorRequest é o corpo de POST /administradores — sem "perfis":
// o alcance é fixo (administrador_sistema).
type AdministradorRequest struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

type UsuarioAtualizarRequest struct {
	Nome   string   `json:"nome"`
	Email  string   `json:"email"`
	Perfis []string `json:"perfis"`
	Versao int      `json:"versao"`
}

// PesquisadorAtualizarRequest é o corpo de PUT
// /instituicoes/{id}/pesquisadores/{usuario_id} — sem "perfis".
type PesquisadorAtualizarRequest struct {
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Versao int    `json:"versao"`
}

// AdministradorAtualizarRequest é o corpo de PUT
// /administradores/{usuario_id} — sem "perfis".
type AdministradorAtualizarRequest struct {
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Versao int    `json:"versao"`
}

type RedefinirSenhaRequest struct {
	SenhaNova string `json:"senha_nova"`
}

// UsuarioResponse é sempre serializada da entidade persistida, nunca
// ecoada do request (D-18) — é o que faz a coerção para {aluno} nunca ser
// silenciosa, e o conjunto sempre em ordem canônica.
type UsuarioResponse struct {
	ID              string   `json:"id"`
	Nome            string   `json:"nome"`
	Email           string   `json:"email"`
	Perfis          []string `json:"perfis"`
	PerfisRotulos   []string `json:"perfis_rotulos"`
	SenhaProvisoria bool     `json:"senha_provisoria"`
	CriadoEm        string   `json:"criado_em"`
	AtualizadoEm    *string  `json:"atualizado_em"`
	Versao          int      `json:"versao"`
}

type InstituicaoResponse struct {
	ID                  string  `json:"id"`
	Nome                string  `json:"nome"`
	Sigla               string  `json:"sigla"`
	CodigoEMec          *string `json:"codigo_emec"`
	Situacao            string  `json:"situacao"`
	PesquisadoresAtivos int     `json:"pesquisadores_ativos"`
	CriadoEm            string  `json:"criado_em"`
	AtualizadoEm        *string `json:"atualizado_em"`
	Versao              int     `json:"versao"`
}

// AlterarSituacaoCatalogoRequest é o corpo de PATCH .../situacao para as
// três entidades desta feature — indicador do INEP, indicador próprio e
// meta usam o mesmo par (situacao, versao).
type AlterarSituacaoCatalogoRequest struct {
	Situacao string `json:"situacao"`
	Versao   int    `json:"versao"`
}

// IndicadorPlataformaRequest é o corpo de POST/PUT
// /api/v1/plataforma/indicadores — referência do instrumento obrigatória.
// Sem campo de escopo: a rota já o impõe (ESCOPO_IMUTAVEL se enviado).
type IndicadorPlataformaRequest struct {
	Codigo                string `json:"codigo"`
	Nome                  string `json:"nome"`
	Descricao             string `json:"descricao"`
	ReferenciaInstrumento string `json:"referencia_instrumento"`
	Versao                int    `json:"versao"`
}

type IndicadorPlataformaResponse struct {
	ID                    string  `json:"id"`
	Escopo                string  `json:"escopo"`
	Codigo                string  `json:"codigo"`
	Nome                  string  `json:"nome"`
	Descricao             string  `json:"descricao"`
	ReferenciaInstrumento string  `json:"referencia_instrumento"`
	Situacao              string  `json:"situacao"`
	Metas                 int     `json:"metas"`
	CriadoEm              string  `json:"criado_em"`
	AtualizadoEm          *string `json:"atualizado_em"`
	Versao                int     `json:"versao"`
}

// IndicadorRequest é o corpo de POST/PUT /api/v1/indicadores — sem campo
// de referência do instrumento e sem campo de escopo (doutrina D-23
// herdada, IN-02).
type IndicadorRequest struct {
	Codigo    string `json:"codigo"`
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Versao    int    `json:"versao"`
}

type IndicadorResponse struct {
	ID                    string  `json:"id"`
	Escopo                string  `json:"escopo"`
	Codigo                string  `json:"codigo"`
	Nome                  string  `json:"nome"`
	Descricao             string  `json:"descricao"`
	ReferenciaInstrumento *string `json:"referencia_instrumento"`
	Situacao              string  `json:"situacao"`
	Metas                 int     `json:"metas"`
	CriadoEm              string  `json:"criado_em"`
	AtualizadoEm          *string `json:"atualizado_em"`
	Versao                int     `json:"versao"`
}

type IndicadorSugestaoResponse struct {
	ID                    string `json:"id"`
	Codigo                string `json:"codigo"`
	Nome                  string `json:"nome"`
	Escopo                string `json:"escopo"`
	ReferenciaInstrumento string `json:"referencia_instrumento"`
}

// IndicadorEmbutidoResponse é o indicador embutido na resposta de meta —
// design.md §6.
type IndicadorEmbutidoResponse struct {
	ID                    string `json:"id"`
	Codigo                string `json:"codigo"`
	Nome                  string `json:"nome"`
	Escopo                string `json:"escopo"`
	ReferenciaInstrumento string `json:"referencia_instrumento"`
	Situacao              string `json:"situacao"`
}

// MetaRequest é o corpo de POST/PUT /api/v1/metas — sem nenhum campo de
// INEP (3.4): a informação de instrumento é derivada dos indicadores.
type MetaRequest struct {
	Nome               string   `json:"nome"`
	Descricao          string   `json:"descricao"`
	Indicadores        []string `json:"indicadores"`
	QuantidadeSugerida *int     `json:"quantidade_sugerida"`
	Versao             int      `json:"versao"`
}

type MetaResponse struct {
	ID                 string                      `json:"id"`
	Nome               string                      `json:"nome"`
	Descricao          string                      `json:"descricao"`
	Situacao           string                      `json:"situacao"`
	Planos             int                         `json:"planos"`
	Indicadores        []IndicadorEmbutidoResponse `json:"indicadores"`
	QuantidadeSugerida *int                        `json:"quantidade_sugerida"`
	CriadoEm           string                      `json:"criado_em"`
	AtualizadoEm       *string                     `json:"atualizado_em"`
	Versao             int                         `json:"versao"`
}

type MetaSugestaoResponse struct {
	ID                 string                      `json:"id"`
	Nome               string                      `json:"nome"`
	Indicadores        []IndicadorEmbutidoResponse `json:"indicadores"`
	QuantidadeSugerida *int                        `json:"quantidade_sugerida"`
}

// --- specs/plano-acao — período, plano de ação curso/coordenador, itens,
// documento (design.md §6). ---

type PeriodoRequest struct {
	Nome       string `json:"nome"`
	DataInicio string `json:"data_inicio"`
	DataFim    string `json:"data_fim"`
	Versao     int    `json:"versao"`
}

type PeriodoResponse struct {
	ID           string  `json:"id"`
	Nome         string  `json:"nome"`
	DataInicio   string  `json:"data_inicio"`
	DataFim      string  `json:"data_fim"`
	Situacao     string  `json:"situacao"`
	Planos       int     `json:"planos"`
	CriadoEm     string  `json:"criado_em"`
	AtualizadoEm *string `json:"atualizado_em"`
	Versao       int     `json:"versao"`
}

type PlanoRequest struct {
	CursoID             string `json:"curso_id"`
	PeriodoID           string `json:"periodo_id"`
	Descricao           string `json:"descricao"`
	ObjetivoGeral       string `json:"objetivo_geral"`
	ResultadosEsperados string `json:"resultados_esperados"`
	AlinhamentoPDI      string `json:"alinhamento_pdi"`
	AlinhamentoPPC      string `json:"alinhamento_ppc"`
	AprovacaoData       string `json:"aprovacao_data"`
	AprovacaoOrgao      string `json:"aprovacao_orgao"`
	Versao              int    `json:"versao"`
}

type CursoEmbutidoResponse struct {
	ID         string `json:"id"`
	Nome       string `json:"nome"`
	Grau       string `json:"grau"`
	Modalidade string `json:"modalidade"`
	CodigoEMec string `json:"codigo_emec,omitempty"`
	Vago       bool   `json:"vago"`
}

type PeriodoEmbutidoResponse struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

type AprovacaoResponse struct {
	Data  string `json:"data"`
	Orgao string `json:"orgao"`
}

type UltimoDocumentoResponse struct {
	ID       string `json:"id"`
	GeradoEm string `json:"gerado_em"`
}

type PlanoResponse struct {
	ID                  string                   `json:"id"`
	Curso               CursoEmbutidoResponse    `json:"curso"`
	Periodo             PeriodoEmbutidoResponse  `json:"periodo"`
	Descricao           string                   `json:"descricao"`
	ObjetivoGeral       string                   `json:"objetivo_geral"`
	ResultadosEsperados string                   `json:"resultados_esperados"`
	AlinhamentoPDI      string                   `json:"alinhamento_pdi"`
	AlinhamentoPPC      string                   `json:"alinhamento_ppc"`
	Situacao            string                   `json:"situacao"`
	Metas               int                      `json:"metas"`
	TotalExigido        int                      `json:"total_exigido"`
	Aprovacao           *AprovacaoResponse       `json:"aprovacao"`
	SemAprovacao        bool                     `json:"sem_aprovacao"`
	TemEntrega          bool                     `json:"tem_entrega"`
	EncerramentoMotivo  string                   `json:"encerramento_motivo,omitempty"`
	UltimoDocumento     *UltimoDocumentoResponse `json:"ultimo_documento,omitempty"`
	CoordenadorNome     *string                  `json:"coordenador_nome"`
	Itens               []ItemPlanoResponse      `json:"itens,omitempty"`
	CriadoEm            string                   `json:"criado_em"`
	AtualizadoEm        *string                  `json:"atualizado_em"`
	Versao              int                      `json:"versao"`
}

type ItemPlanoResponse struct {
	ID          string                      `json:"id"`
	MetaID      string                      `json:"meta_id"`
	MetaNome    string                      `json:"meta_nome"`
	Indicadores []IndicadorEmbutidoResponse `json:"indicadores"`
	Quantidade  int                         `json:"quantidade"`
	TemEntrega  bool                        `json:"tem_entrega"`
	Versao      int                         `json:"versao"`
}

type ItemPlanoRequest struct {
	MetaID     string `json:"meta_id"`
	Quantidade int    `json:"quantidade"`
	Versao     int    `json:"versao"`
}

type EncerrarPlanoRequest struct {
	Motivo string `json:"motivo"`
	Versao int    `json:"versao"`
}

type VersaoRequest struct {
	Versao int `json:"versao"`
}

type CopiaEmLoteRequest struct {
	PeriodoDestinoID string   `json:"periodo_destino_id"`
	Cursos           []string `json:"cursos"`
}

type CursoPuladoResponse struct {
	Curso  CursoEmbutidoResponse `json:"curso"`
	Motivo string                `json:"motivo"`
}

type CopiaEmLoteResponse struct {
	Criados int                   `json:"criados"`
	Pulados []CursoPuladoResponse `json:"pulados"`
}

type DestinoDeCopiaResponse struct {
	CursoID         string  `json:"curso_id"`
	CursoNome       string  `json:"curso_nome"`
	CoordenadorNome *string `json:"coordenador_nome"`
	Vago            bool    `json:"vago"`
	JaTemPlano      bool    `json:"ja_tem_plano"`
}

// --- specs/cursos — curso e designação de coordenação (design.md §6). ---

// CursoRequest é o corpo de POST/PUT /api/v1/cursos — sem campo de
// coordenador (design.md §3.2): a entidade nasce e sempre existe sem ele.
type CursoRequest struct {
	Nome       string `json:"nome"`
	CodigoEMec string `json:"codigo_emec"`
	Grau       string `json:"grau"`
	Modalidade string `json:"modalidade"`
	Situacao   string `json:"situacao"`
}

type CursoAtualizarRequest struct {
	Nome       string `json:"nome"`
	CodigoEMec string `json:"codigo_emec"`
	Grau       string `json:"grau"`
	Modalidade string `json:"modalidade"`
	Versao     int    `json:"versao"`
}

// CoordenadorDoCursoResponse é o campo computado "coordenador" — projeção
// de designação vigente, nunca persistido (design.md §5.5). nil quando o
// curso está vago (CU-05).
type CoordenadorDoCursoResponse struct {
	ID                             string  `json:"id"`
	Nome                           string  `json:"nome"`
	DataFim                        *string `json:"data_fim"`
	TambemPesquisadorInstitucional bool    `json:"tambem_pesquisador_institucional"`
}

type CursoResponse struct {
	ID           string                      `json:"id"`
	Nome         string                      `json:"nome"`
	CodigoEMec   *string                     `json:"codigo_emec"`
	Grau         string                      `json:"grau"`
	Modalidade   string                      `json:"modalidade"`
	Situacao     string                      `json:"situacao"`
	Coordenador  *CoordenadorDoCursoResponse `json:"coordenador"`
	TemVinculo   bool                        `json:"tem_vinculo"`
	// PlanoDoPeriodo — só preenchido em GET /cursos (T-115); nil nas
	// respostas de curso único (design.md §5.5: campo de grid, não de
	// recurso individual).
	PlanoDoPeriodo *string `json:"plano_do_periodo"`
	CriadoEm       string  `json:"criado_em"`
	AtualizadoEm   *string `json:"atualizado_em"`
	Versao         int     `json:"versao"`
}

// DesignacaoRequest é o corpo de POST /api/v1/cursos/{id}/designacoes.
type DesignacaoRequest struct {
	CoordenadorID string  `json:"coordenador_id"`
	Portaria      string  `json:"portaria"`
	DataInicio    string  `json:"data_inicio"`
	DataFim       *string `json:"data_fim"`
}

// DesignacaoAtualizarRequest é o corpo de PUT /api/v1/designacoes/{id} —
// os mesmos quatro campos de criação, mais versão (design.md §5.3: o
// backend decide o que aceita, conforme a situação atual no banco).
type DesignacaoAtualizarRequest struct {
	CoordenadorID string  `json:"coordenador_id"`
	Portaria      string  `json:"portaria"`
	DataInicio    string  `json:"data_inicio"`
	DataFim       *string `json:"data_fim"`
	Versao        int     `json:"versao"`
}

type CoordenadorDaDesignacaoResponse struct {
	ID   string `json:"id"`
	Nome string `json:"nome"`
}

type DesignacaoResponse struct {
	ID                        string                          `json:"id"`
	Coordenador               CoordenadorDaDesignacaoResponse `json:"coordenador"`
	Portaria                  string                          `json:"portaria"`
	DataInicio                string                          `json:"data_inicio"`
	DataFim                   *string                         `json:"data_fim"`
	Situacao                  string                          `json:"situacao"`
	Autodesignacao            bool                            `json:"autodesignacao"`
	OutrasDesignacoesVigentes int                             `json:"outras_designacoes_vigentes"`
	Versao                    int                             `json:"versao"`
}

// CandidatoResponse é a linha do autocomplete de coordenador (CP-07) —
// com os perfis de cada um, para a tela avisar acúmulo de papel.
type CandidatoResponse struct {
	ID     string   `json:"id"`
	Nome   string   `json:"nome"`
	Email  string   `json:"email"`
	Perfis []string `json:"perfis"`
}

// --- specs/metas-coordenacao — entrega, anexo, avaliação e relatório
// de desempenho (design.md §11). ---

type EntregaCorrigirRequest struct {
	Observacao string `json:"observacao"`
	Versao     int    `json:"versao"`
}

type AvaliarRequest struct {
	Resultado string `json:"resultado"`
	Motivo    string `json:"motivo"`
	Versao    int    `json:"versao"`
}

type DesfazerAceitacaoRequest struct {
	Motivo string `json:"motivo"`
	Versao int    `json:"versao"`
}

type AnexoResponse struct {
	ID           string `json:"id"`
	NomeOriginal string `json:"nome_original"`
	Tipo         string `json:"tipo"`
	TamanhoBytes int64  `json:"tamanho_bytes"`
	HashSHA256   string `json:"hash_sha256"`
	CriadoEm     string `json:"criado_em"`
}

type EntregaResponse struct {
	ID                      string          `json:"id"`
	ItemPlanoID             string          `json:"item_plano_id"`
	CursoID                 string          `json:"curso_id"`
	CursoNome               string          `json:"curso_nome"`
	MetaNome                string          `json:"meta_nome,omitempty"`
	Quantidade              int             `json:"quantidade,omitempty"`
	Indicadores             []IndicadorEmbutidoResponse `json:"indicadores,omitempty"`
	Situacao                string          `json:"situacao"`
	Rodadas                 int             `json:"rodadas"`
	PrazoCorrecao           *string         `json:"prazo_correcao_ate"`
	EnviadaPorID            string          `json:"enviada_por_id"`
	EnviadaPorNome          string          `json:"enviada_por_nome"`
	CorrigidaPorNome        *string         `json:"corrigida_por_nome"`
	Observacao              string          `json:"observacao"`
	Motivo                  string          `json:"motivo"`
	AvaliadaPorNome         *string         `json:"avaliada_por_nome"`
	AvaliadaEm              *string         `json:"avaliada_em"`
	AvaliadorEraCoordenador bool            `json:"avaliador_era_coordenador"`
	PendenciaVistaEm        *string         `json:"pendencia_vista_em"`
	CoordenadoPeloAvaliador bool            `json:"coordenado_pelo_avaliador"`
	Anexos                  []AnexoResponse `json:"anexos"`
	CriadoEm                string          `json:"criado_em"`
	Versao                  int             `json:"versao"`
}

type ItemMinhasMetasResponse struct {
	ItemPlanoID       string                       `json:"item_plano_id"`
	MetaNome          string                       `json:"meta_nome"`
	Indicadores       []IndicadorEmbutidoResponse `json:"indicadores"`
	Quantidade        int                          `json:"quantidade"`
	Aceitas           int                          `json:"aceitas"`
	Pendentes         int                          `json:"pendentes"`
	EmCorrecao        int                          `json:"em_correcao"`
}

type GrupoMinhasMetasResponse struct {
	CursoID   string                     `json:"curso_id"`
	CursoNome string                     `json:"curso_nome"`
	Itens     []ItemMinhasMetasResponse `json:"itens"`
}

type PeriodoOpcaoResponse struct {
	ID      string `json:"id"`
	Nome    string `json:"nome"`
	DataFim string `json:"data_fim"`
	Aberto  bool   `json:"aberto"`
}

type PendenciasResponse struct {
	PendentesDeAvaliacao int `json:"pendentes_de_avaliacao"`
	PendenciasNaoVistas  int `json:"pendencias_nao_vistas"`
}

type RelatorioLinhaResponse struct {
	ItemPlanoID                     string                       `json:"item_plano_id"`
	CursoNome                       string                       `json:"curso_nome"`
	CursoVago                       bool                         `json:"curso_vago"`
	ResponsavelNome                 *string                      `json:"responsavel_nome"`
	ResponsavelDesde                *string                      `json:"responsavel_desde"`
	VagoDesde                       *string                      `json:"vago_desde"`
	MetaNome                        string                       `json:"meta_nome"`
	Indicadores                     []IndicadorEmbutidoResponse `json:"indicadores"`
	Exigido                         int                          `json:"exigido"`
	Aceitas                         int                          `json:"aceitas"`
	Pendentes                       int                          `json:"pendentes"`
	EmCorrecao                      int                          `json:"em_correcao"`
	Cumprimento                     float64                      `json:"cumprimento"`
	Situacao                        string                       `json:"situacao"`
	InclusEntregasAnteriores        bool                         `json:"inclui_entregas_anteriores"`
	AvaliacaoPeloProprioCoordenador bool                         `json:"avaliacao_pelo_proprio_coordenador"`
}

type ResumoRelatorioResponse struct {
	CursosSemCoordenador     int `json:"cursos_sem_coordenador"`
	MetasNaoCumpridasDeVagos int `json:"metas_nao_cumpridas_de_vagos"`
}

type ItemDesempenhoPorCursoResponse struct {
	CursoID         string  `json:"curso_id"`
	CursoNome       string  `json:"curso_nome"`
	ResponsavelNome string  `json:"responsavel_nome"`
	ExigidoTotal    int     `json:"exigido_total"`
	AceitasTotal    int     `json:"aceitas_total"`
	Percentual      float64 `json:"percentual"`
}

type MetaDesempenhoPorCursoResponse struct {
	TotalCursosComCoordenador int `json:"total_cursos_com_coordenador"`
	TotalCursosSemCoordenador int `json:"total_cursos_sem_coordenador"`
}
