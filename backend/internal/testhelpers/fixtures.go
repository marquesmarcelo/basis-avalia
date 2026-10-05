package testhelpers

import (
	"strings"
	"testing"
	"time"

	"github.com/basis-avalia/backend/internal/adapter/relogio"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// HashDeTeste é um hash PHC argon2id sintético — nunca é conferido de
// verdade nos testes que só precisam de um usuário existir no banco.
const HashDeTeste = "$argon2id$v=19$m=65536,t=3,p=4$c2FsdGVzYWx0$aGFzaGZha2Vmb3J0ZXN0"

type OpcoesInstituicao struct {
	Nome       string
	Sigla      string
	CodigoEMec string
	Situacao   string
}

// CriarInstituicao cria uma instituição no banco e registra a própria
// limpeza via t.Cleanup — quem chama não precisa lembrar de nada.
func CriarInstituicao(t *testing.T, db *sqlx.DB, opcoes OpcoesInstituicao) uuid.UUID {
	t.Helper()

	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	nome := opcoes.Nome
	if nome == "" {
		nome = "Instituição de Teste " + sufixo
	}
	sigla := opcoes.Sigla
	if sigla == "" {
		sigla = "T" + strings.ToUpper(sufixo[:5])
	}
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "ativa"
	}
	var codigoEMec any
	if opcoes.CodigoEMec != "" {
		codigoEMec = opcoes.CodigoEMec
	}

	_, err := db.Exec(
		`INSERT INTO instituicao (id, nome, sigla, codigo_emec, situacao) VALUES ($1,$2,$3,$4,$5)`,
		id, nome, sigla, codigoEMec, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarInstituicao: %v", err)
	}

	t.Cleanup(func() {
		// Apaga primeiro qualquer auditoria que aponte para esta
		// instituição — testes de smoke exercitam use cases reais que
		// escrevem auditoria de verdade, e a FK bloquearia esta limpeza.
		db.Exec(`DELETE FROM auditoria WHERE instituicao_id = $1`, id)
		db.Exec(`DELETE FROM instituicao WHERE id = $1`, id)
	})
	return id
}

type OpcoesUsuario struct {
	InstituicaoID   *uuid.UUID
	Nome            string
	Email           string
	SenhaHash       string
	Perfil          string   // um só perfil — atalho equivalente a Perfis: []string{Perfil}
	Perfis          []string // conjunto completo; tem prioridade sobre Perfil quando preenchido
	SenhaProvisoria *bool
}

// CriarUsuario cria um usuário e o conjunto de perfis dele no banco, e
// registra a própria limpeza via t.Cleanup — executado antes da limpeza da
// instituição (LIFO), o que respeita a FK sem exigir nada de quem chama.
func CriarUsuario(t *testing.T, db *sqlx.DB, opcoes OpcoesUsuario) uuid.UUID {
	t.Helper()

	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	nome := opcoes.Nome
	if nome == "" {
		nome = "Usuário de Teste " + sufixo
	}
	email := opcoes.Email
	if email == "" {
		email = "usuario-" + sufixo + "@teste.local"
	}
	hash := opcoes.SenhaHash
	if hash == "" {
		hash = HashDeTeste
	}
	perfis := opcoes.Perfis
	if len(perfis) == 0 {
		perfil := opcoes.Perfil
		if perfil == "" {
			perfil = "professor"
		}
		perfis = []string{perfil}
	}
	provisoria := true
	if opcoes.SenhaProvisoria != nil {
		provisoria = *opcoes.SenhaProvisoria
	}

	// sessoes_validas_a_partir_de vem do relógio da aplicação (time.Now()),
	// nunca do DEFAULT now() da coluna — comparar um valor cravado pelo
	// Postgres (relógio do container do banco) contra o emt de um login
	// gerado pela aplicação (relógio do container Go) tem folga zero na
	// comparação de middleware_sessao.go; sob carga (suíte inteira em
	// paralelo) o desvio entre os dois relógios já supera o tempo real
	// decorrido entre criar o usuário e logar, derrubando a sessão recém
	// criada com SESSAO_EXPIRADA de forma intermitente. Investigado e
	// comprovado via log temporário no ponto de comparação (systematic-
	// debugging): delta real capturado de +3.3ms num login que falhou.
	//
	// SEGUNDA CAUSA DA MESMA FAMÍLIA (encontrada depois, sob carga real de
	// `go test ./...`): mesmo com o relógio certo, um `time.Now()` cru tem
	// precisão de nanossegundo; o `emt` do login passa por
	// relogio.Relogio.Agora(), que TRUNCA para microssegundo (a precisão
	// do TIMESTAMPTZ). Sem a mesma truncagem aqui, o Postgres ARREDONDA o
	// resto de nanossegundo ao gravar — se arredondar para cima, o valor
	// gravado fica MAIOR que o `agora` original, e um login capturado
	// poucos microssegundos depois (truncado, nunca arredondado para
	// cima) pode cair ANTES do valor gravado mesmo sendo cronologicamente
	// posterior. Sob carga, a janela entre criar o usuário e logar encolhe
	// o suficiente para essa margem de arredondamento decidir a
	// comparação — reproduzido via `go test ./... -count=1` repetido
	// (nunca em execução isolada do teste, só sob a suíte inteira).
	//
	// TERCEIRA CAUSA, DISTINTA (encontrada isolando só TestSmoke_R5, sem
	// nenhuma interferência de outro pacote): o relógio de parede do
	// container (WSL2/Docker Desktop) retrocede sob carga de CPU — provado
	// com uma amostragem pura de time.Now() sob a mesma carga que argon2
	// gera: 1 retrocesso de ~550ms capturado em 83 milhões de amostras/20s.
	// A correção ficou em relogio.Relogio: Agora() garante monotonicidade
	// (nunca retrocede em relação à própria última chamada) — mas só
	// protege quem passa por ela. Esta fixture usava time.Now() cru, fora
	// de qualquer Relogio, então continuava exposta ao mesmo retrocesso
	// entre "criar o usuário" e o login do teste (T02, T05, AS03, AS10,
	// G07 ainda falhavam depois da primeira correção). A fixture agora usa
	// relogio.Novo() — que devolve sempre a MESMA instância no processo —
	// para entrar na mesma sequência monotônica que o login do teste usa.
	agora := relogio.Novo().Agora()
	_, err := db.Exec(
		`INSERT INTO usuario (id, instituicao_id, nome, email, senha_hash, senha_provisoria, sessoes_validas_a_partir_de)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, opcoes.InstituicaoID, nome, email, hash, provisoria, agora,
	)
	if err != nil {
		t.Fatalf("fixture CriarUsuario: %v", err)
	}

	for _, perfil := range perfis {
		var instituicaoDoVinculo any
		if perfil != "administrador_sistema" {
			instituicaoDoVinculo = opcoes.InstituicaoID
		}
		if _, err := db.Exec(
			`INSERT INTO usuario_perfil (usuario_id, perfil, instituicao_id) VALUES ($1,$2,$3)`,
			id, perfil, instituicaoDoVinculo,
		); err != nil {
			t.Fatalf("fixture CriarUsuario: vínculo de perfil %q: %v", perfil, err)
		}
	}

	t.Cleanup(func() {
		// Mesma razão do cleanup de CriarInstituicao: apaga auditoria que
		// referencie este usuário antes de tentar remover a linha dele.
		// usuario_perfil sai sozinho via ON DELETE CASCADE.
		db.Exec(`DELETE FROM auditoria WHERE ator_id = $1`, id)
		db.Exec(`DELETE FROM usuario WHERE id = $1`, id)
	})
	return id
}

// CriarUsuarioExcluido cria um usuário já excluído logicamente, com o
// hash anulado (3.8) — para cenários como L-04, U-03 e G-17.
func CriarUsuarioExcluido(t *testing.T, db *sqlx.DB, opcoes OpcoesUsuario) uuid.UUID {
	t.Helper()

	id := CriarUsuario(t, db, opcoes)
	_, err := db.Exec(`UPDATE usuario SET excluido_em = now(), senha_hash = NULL WHERE id = $1`, id)
	if err != nil {
		t.Fatalf("fixture CriarUsuarioExcluido: %v", err)
	}
	return id
}

// sufixoAleatorio usa os últimos caracteres da UUIDv7 (bits aleatórios),
// nunca os primeiros (que carregam o timestamp e mal mudam entre chamadas
// consecutivas dentro do mesmo teste) — evita colisão de sigla/e-mail
// quando o teste cria vários registros em sequência rápida.
func sufixoAleatorio(id uuid.UUID) string {
	s := id.String()
	return s[len(s)-8:]
}

type OpcoesIndicadorPlataforma struct {
	Codigo                string
	Nome                  string
	ReferenciaInstrumento string
	Situacao              string
}

// CriarIndicadorPlataforma cria um indicador do catálogo do INEP (escopo
// plataforma, instituicao_id nulo) e registra a própria limpeza.
func CriarIndicadorPlataforma(t *testing.T, db *sqlx.DB, opcoes OpcoesIndicadorPlataforma) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	codigo := opcoes.Codigo
	if codigo == "" {
		codigo = "IND-" + sufixo
	}
	nome := opcoes.Nome
	if nome == "" {
		nome = "Indicador do INEP " + sufixo
	}
	referencia := opcoes.ReferenciaInstrumento
	if referencia == "" {
		referencia = "Instrumento de teste " + sufixo
	}
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "ativo"
	}
	_, err := db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, referencia_instrumento, situacao)
		 VALUES ($1,'plataforma',NULL,$2,$3,$4,$5)`,
		id, codigo, nome, referencia, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarIndicadorPlataforma: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_indicador WHERE indicador_id = $1`, id)
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id)
	})
	return id
}

type OpcoesIndicadorInstituicao struct {
	InstituicaoID uuid.UUID
	Codigo        string
	Nome          string
	Situacao      string
}

// CriarIndicadorInstituicao cria um indicador próprio de uma instituição.
func CriarIndicadorInstituicao(t *testing.T, db *sqlx.DB, opcoes OpcoesIndicadorInstituicao) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	codigo := opcoes.Codigo
	if codigo == "" {
		codigo = "IND-" + sufixo
	}
	nome := opcoes.Nome
	if nome == "" {
		nome = "Indicador próprio " + sufixo
	}
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "ativo"
	}
	_, err := db.Exec(
		`INSERT INTO indicador (id, escopo, instituicao_id, codigo, nome, referencia_instrumento, situacao)
		 VALUES ($1,'instituicao',$2,$3,$4,NULL,$5)`,
		id, opcoes.InstituicaoID, codigo, nome, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarIndicadorInstituicao: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_indicador WHERE indicador_id = $1`, id)
		db.Exec(`DELETE FROM indicador WHERE id = $1`, id)
	})
	return id
}

type OpcoesMeta struct {
	InstituicaoID uuid.UUID
	Nome          string
	Situacao      string
	Indicadores   []uuid.UUID
}

// CriarMeta cria uma meta com o vínculo de indicadores e registra a
// própria limpeza — quem chama precisa garantir que os indicadores já
// existem (fixture própria ou passada por parâmetro).
func CriarMeta(t *testing.T, db *sqlx.DB, opcoes OpcoesMeta) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	nome := opcoes.Nome
	if nome == "" {
		nome = "Meta de teste " + sufixo
	}
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "ativo"
	}
	_, err := db.Exec(
		`INSERT INTO meta (id, instituicao_id, nome, situacao) VALUES ($1,$2,$3,$4)`,
		id, opcoes.InstituicaoID, nome, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarMeta: %v", err)
	}
	for _, indicadorID := range opcoes.Indicadores {
		if _, err := db.Exec(`INSERT INTO meta_indicador (meta_id, indicador_id) VALUES ($1,$2)`, id, indicadorID); err != nil {
			t.Fatalf("fixture CriarMeta: vínculo com indicador %s: %v", indicadorID, err)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM meta_indicador WHERE meta_id = $1`, id)
		db.Exec(`DELETE FROM meta WHERE id = $1`, id)
	})
	return id
}

type OpcoesCurso struct {
	InstituicaoID uuid.UUID
	Nome          string
	Grau          string
	Modalidade    string
	Situacao      string
}

// CriarCurso cria um curso e registra a própria limpeza.
func CriarCurso(t *testing.T, db *sqlx.DB, opcoes OpcoesCurso) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	nome := opcoes.Nome
	if nome == "" {
		nome = "Curso de Teste " + sufixo
	}
	grau := opcoes.Grau
	if grau == "" {
		grau = "bacharelado"
	}
	modalidade := opcoes.Modalidade
	if modalidade == "" {
		modalidade = "presencial"
	}
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "ativo"
	}
	_, err := db.Exec(
		`INSERT INTO curso (id, instituicao_id, nome, grau, modalidade, situacao) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, opcoes.InstituicaoID, nome, grau, modalidade, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarCurso: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM designacao WHERE curso_id = $1`, id)
		db.Exec(`DELETE FROM curso WHERE id = $1`, id)
	})
	return id
}

type OpcoesDesignacao struct {
	CursoID        uuid.UUID
	InstituicaoID  uuid.UUID
	CoordenadorID  uuid.UUID
	Portaria       string
	DataInicio     string // "AAAA-MM-DD"
	DataFim        *string
	Autodesignacao bool
}

// CriarDesignacao cria uma designação e registra a própria limpeza.
func CriarDesignacao(t *testing.T, db *sqlx.DB, opcoes OpcoesDesignacao) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	portaria := opcoes.Portaria
	if portaria == "" {
		portaria = "1/2026"
	}
	dataInicio := opcoes.DataInicio
	if dataInicio == "" {
		dataInicio = "2026-01-01"
	}
	_, err := db.Exec(
		`INSERT INTO designacao (id, curso_id, instituicao_id, coordenador_id, portaria, data_inicio, data_fim, autodesignacao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, opcoes.CursoID, opcoes.InstituicaoID, opcoes.CoordenadorID, portaria, dataInicio, opcoes.DataFim, opcoes.Autodesignacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarDesignacao: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM designacao WHERE id = $1`, id)
	})
	return id
}

type OpcoesPeriodo struct {
	InstituicaoID uuid.UUID
	Nome          string
	DataInicio    string // "AAAA-MM-DD"
	DataFim       string
}

// CriarPeriodo cria um período (specs/plano-acao) e registra a própria
// limpeza.
func CriarPeriodo(t *testing.T, db *sqlx.DB, opcoes OpcoesPeriodo) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	sufixo := sufixoAleatorio(id)
	nome := opcoes.Nome
	if nome == "" {
		nome = "Período de Teste " + sufixo
	}
	dataInicio := opcoes.DataInicio
	if dataInicio == "" {
		dataInicio = "2026-01-01"
	}
	dataFim := opcoes.DataFim
	if dataFim == "" {
		dataFim = "2026-07-30"
	}
	_, err := db.Exec(
		`INSERT INTO periodo (id, instituicao_id, nome, data_inicio, data_fim) VALUES ($1,$2,$3,$4,$5)`,
		id, opcoes.InstituicaoID, nome, dataInicio, dataFim,
	)
	if err != nil {
		t.Fatalf("fixture CriarPeriodo: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM plano WHERE periodo_id = $1`, id)
		db.Exec(`DELETE FROM periodo WHERE id = $1`, id)
	})
	return id
}

type OpcoesPlano struct {
	InstituicaoID       uuid.UUID
	CursoID             uuid.UUID
	PeriodoID           uuid.UUID
	Descricao           string
	ObjetivoGeral       string
	ResultadosEsperados string
	SituacaoPublicacao  string // "rascunho" | "vigente"
}

// CriarPlano cria um plano de ação curso/coordenador (specs/plano-acao) e
// registra a própria limpeza.
func CriarPlano(t *testing.T, db *sqlx.DB, opcoes OpcoesPlano) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	descricao := opcoes.Descricao
	if descricao == "" {
		descricao = "Descrição de teste"
	}
	objetivoGeral := opcoes.ObjetivoGeral
	if objetivoGeral == "" {
		objetivoGeral = "Objetivo de teste"
	}
	resultadosEsperados := opcoes.ResultadosEsperados
	if resultadosEsperados == "" {
		resultadosEsperados = "Resultados de teste"
	}
	situacao := opcoes.SituacaoPublicacao
	if situacao == "" {
		situacao = "rascunho"
	}
	_, err := db.Exec(
		`INSERT INTO plano (id, instituicao_id, curso_id, periodo_id, descricao, objetivo_geral, resultados_esperados, situacao_publicacao)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, opcoes.InstituicaoID, opcoes.CursoID, opcoes.PeriodoID, descricao, objetivoGeral, resultadosEsperados, situacao,
	)
	if err != nil {
		t.Fatalf("fixture CriarPlano: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM documento WHERE plano_id = $1`, id)
		db.Exec(`DELETE FROM item_plano WHERE plano_id = $1`, id)
		db.Exec(`DELETE FROM plano WHERE id = $1`, id)
	})
	return id
}

type OpcoesItemPlano struct {
	PlanoID       uuid.UUID
	CursoID       uuid.UUID
	InstituicaoID uuid.UUID
	MetaID        uuid.UUID
	Quantidade    int
}

// CriarItemPlano cria um item de plano e registra a própria limpeza.
func CriarItemPlano(t *testing.T, db *sqlx.DB, opcoes OpcoesItemPlano) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	quantidade := opcoes.Quantidade
	if quantidade == 0 {
		quantidade = 1
	}
	_, err := db.Exec(
		`INSERT INTO item_plano (id, plano_id, curso_id, instituicao_id, meta_id, quantidade) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, opcoes.PlanoID, opcoes.CursoID, opcoes.InstituicaoID, opcoes.MetaID, quantidade,
	)
	if err != nil {
		t.Fatalf("fixture CriarItemPlano: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM item_plano WHERE id = $1`, id)
	})
	return id
}

type OpcoesEntrega struct {
	ItemPlanoID    uuid.UUID
	CursoID        uuid.UUID
	InstituicaoID  uuid.UUID
	EnviadaPor     uuid.UUID
	Situacao       string // "pendente_avaliacao" | "aceita" | "recusada"
	Rodadas        int
	PrazoCorrecao  *string // instante ISO completo, ex: "2026-08-05T23:59:59-03:00"
	AvaliadaPor    *uuid.UUID
	Motivo         string
	AvaliadorEraCoordenador bool
	PendenciaVistaEm *string
}

// CriarEntrega cria uma entrega (specs/metas-coordenacao) e registra a
// própria limpeza — executada antes da limpeza do item_plano (LIFO).
func CriarEntrega(t *testing.T, db *sqlx.DB, opcoes OpcoesEntrega) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	situacao := opcoes.Situacao
	if situacao == "" {
		situacao = "pendente_avaliacao"
	}
	var avaliadaEm any
	if opcoes.AvaliadaPor != nil {
		avaliadaEm = time.Now()
	}
	_, err := db.Exec(
		`INSERT INTO entrega (id, item_plano_id, curso_id, instituicao_id, enviada_por, situacao,
		                      rodadas_de_recusa, prazo_correcao_ate, avaliada_por, avaliada_em, motivo,
		                      avaliador_era_coordenador, pendencia_vista_em)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, opcoes.ItemPlanoID, opcoes.CursoID, opcoes.InstituicaoID, opcoes.EnviadaPor, situacao,
		opcoes.Rodadas, opcoes.PrazoCorrecao, opcoes.AvaliadaPor, avaliadaEm, opcoes.Motivo,
		opcoes.AvaliadorEraCoordenador, opcoes.PendenciaVistaEm,
	)
	if err != nil {
		t.Fatalf("fixture CriarEntrega: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM anexo WHERE entrega_id = $1`, id)
		db.Exec(`DELETE FROM entrega WHERE id = $1`, id)
	})
	return id
}

type OpcoesAnexo struct {
	EntregaID     uuid.UUID
	CursoID       uuid.UUID
	InstituicaoID uuid.UUID
	NomeOriginal  string
	Tipo          string
	TamanhoBytes  int64
	ChaveObjeto   string
	HashSHA256    string
}

// CriarAnexo cria um anexo e registra a própria limpeza.
func CriarAnexo(t *testing.T, db *sqlx.DB, opcoes OpcoesAnexo) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	nomeOriginal := opcoes.NomeOriginal
	if nomeOriginal == "" {
		nomeOriginal = "comprovante.pdf"
	}
	tipo := opcoes.Tipo
	if tipo == "" {
		tipo = "pdf"
	}
	tamanho := opcoes.TamanhoBytes
	if tamanho == 0 {
		tamanho = 1024
	}
	chave := opcoes.ChaveObjeto
	if chave == "" {
		chave = "teste/" + id.String()
	}
	hash := opcoes.HashSHA256
	if hash == "" {
		hash = strings.Repeat("0", 64)
	}
	_, err := db.Exec(
		`INSERT INTO anexo (id, entrega_id, curso_id, instituicao_id, nome_original, tipo, tamanho_bytes, chave_objeto, hash_sha256)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, opcoes.EntregaID, opcoes.CursoID, opcoes.InstituicaoID, nomeOriginal, tipo, tamanho, chave, hash,
	)
	if err != nil {
		t.Fatalf("fixture CriarAnexo: %v", err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM anexo WHERE id = $1`, id)
	})
	return id
}

// contarUsuariosDaInstituicao é usado pelo teste do próprio helper
// (T-012): prova que o banco volta ao estado anterior depois do
// t.Cleanup. Sempre restrito a uma instituição — nunca a tabela inteira,
// porque `go test ./...` roda pacotes em paralelo contra o mesmo banco de
// teste compartilhado.
func contarUsuariosDaInstituicao(t *testing.T, db *sqlx.DB, instituicaoID uuid.UUID) int {
	t.Helper()
	var n int
	if err := db.Get(&n, `SELECT count(*) FROM usuario WHERE instituicao_id = $1`, instituicaoID); err != nil {
		t.Fatalf("contarUsuariosDaInstituicao: %v", err)
	}
	return n
}
