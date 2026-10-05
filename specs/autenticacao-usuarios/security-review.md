# Revisão de segurança — `autenticacao-usuarios` (T-077)

**Data:** 28/09/2026
**Revisor:** `security-reviewer`
**Referência:** OWASP Top 10:2025 · `CLAUDE.md` · `spec.md` (rev. 28/09) ·
`design.md` (revisão 4)
**Escopo revisado:** `backend/` (13.0 kLOC Go), `frontend/src/` (111 arquivos),
`docker-compose.yaml`, `docker-compose.dev.yml`, `Dockerfile` de ambos os
serviços, migrations `000002_autenticacao_usuarios`.

> **Autoridade.** Este documento descreve problemas e vetores de ataque.
> **Não decide a correção** — pela cadeia do `CLAUDE.md`, quem decide como
> corrigir é o `arquiteto`. Os achados 🔴 Crítico e 🟡 Alto abaixo **bloqueiam
> a entrega** até que o `arquiteto` decida o encaminhamento ou o dono do
> produto aceite o risco explicitamente.

---

## 1. Veredito

| Severidade | Qtd. | Bloqueia? |
|---|---|---|
| 🔴 Crítico | 2 | **sim** |
| 🟡 Alto | 2 | **sim** |
| 🟠 Médio | 4 | não |
| 🔵 Baixo / observação | 6 | não |

**O núcleo da feature está sólido.** A propriedade central do desenho — "esquecer
a verificação de autorização não compila" — **sustenta-se no código real**, e as
duas portas estreitas foram revisadas manualmente e estão corretas (§3). Todos os
quatro mitigantes de que dependem os riscos aceitos pelo dono **continuam de pé**
(§2). Nenhum dos achados abaixo está no caminho de autenticação, autorização ou
isolamento: os quatro que bloqueiam são **uma dependência vulnerável**,
**comentários no bundle** e **duas lacunas de configuração na borda HTTP**.

---

## 2. Riscos aceitos pelo dono do produto — **registro, não pendência**

Os três controles abaixo foram recusados pelo dono do produto com o alerta à
vista (`spec.md` §3.2, §3.3, §4.1). Registrados aqui como **risco aceito**,
conforme o encaminhamento da própria `spec.md` §4.1 e a regra do `CLAUDE.md`
("o dono do produto pode aceitar um risco explicitamente documentado").
**Nenhum controle recusado é reproposto neste documento.**

| # | Risco aceito | Referência |
|---|---|---|
| RA-1 | Sem limite de tentativas de login — sem bloqueio, sem 429, sem contador | spec 3.2, 4.1 |
| RA-2 | Sem registro de tentativa de login, de login bem-sucedido nem de logout na auditoria; sem métrica de falha de autenticação | spec 3.2, 4.1 |
| RA-3 | Sem política de senha — só "não pode ser vazia", mais o limite técnico de 1024 caracteres de fronteira de confiança (P12) | spec 3.3, 4.1 |
| RA-4 | Lista de instituições visível a qualquer visitante (rota pública do combo) | spec 3.17, 4.2 |

### 2.1 Verificação dos mitigantes de que esses riscos dependem

O risco aceito em 4.1 foi aceito **porque** quatro controles permaneceriam de
pé. Verifiquei os quatro no código real. **Nenhum regrediu.**

| Mitigante | Situação | Onde verifiquei |
|---|---|---|
| **Resposta de login genérica** | ✅ **De pé.** Todo caminho de falha devolve `domain.ErrCredenciaisInvalidas` → 401 `CREDENCIAIS_INVALIDAS`, "Instituição, e-mail ou senha inválidos.". Conta inexistente, conta excluída e instituição inativa são colapsadas no **próprio SQL** (`AND u.excluido_em IS NULL AND (u.instituicao_id IS NULL OR i.situacao = 'ativa')`), devolvendo `nil, nil` indistinguível | `usecase/command/sessao/autenticar.go:48-61`; `adapter/postgres/autenticacao_repository.go:70-86`; `adapter/http/middleware_erro.go:54` |
| **Tempo constante no login** | ✅ **De pé.** `credencial == nil` dispara `ConferirDescartavel`, que calcula argon2id contra um **hash de referência fixo gerado na construção do adapter**, com os mesmos parâmetros — nunca contra dado real, e passando pelo mesmo semáforo | `usecase/command/sessao/autenticar.go:51`; `adapter/argon2/hash_de_senha.go:74-80, 129-138` |
| **Hash lento (argon2id)** | ✅ **De pé, e acima do especificado.** m=144 MiB, t=3, p=4, salt 16 B, hash 32 B, formato PHC; comparação com `subtle.ConstantTimeCompare`; semáforo de 4 cálculos concorrentes que **enfileira, não recusa** (D-02). `bcrypt` corretamente rejeitado | `adapter/argon2/hash_de_senha.go:33-39, 66-77, 163` |
| **Token fora do alcance do JavaScript** | ✅ **De pé.** `SetSameSite(http.SameSiteStrictMode)` + `SetCookie(..., secure=true, httpOnly=true)`, sem ramificar por `APP_ENV`. **Zero ocorrências** de token/JWT em `localStorage` ou `sessionStorage` em todo o frontend (grep por `token`, `jwt`, `bearer`, `authorization` em `frontend/src/`: **nenhum resultado**) | `adapter/http/auth_handler.go:64-72`; varredura de `frontend/src/` |
| **Auditoria administrativa íntegra** | ✅ **De pé.** 13 ações auditadas, canal local **na mesma transação** da mutação; `detalhes` contém só nomes de campo, conjuntos de perfis e permissão exigida — **nunca senha, hash, token, cookie ou valor de campo pessoal**; syslog fire-and-forget pós-commit, cuja falha nunca derruba a operação | `domain/auditoria/evento.go`; `adapter/postgres/auditoria_repository.go`; `adapter/auditoria/composto.go` |

Um ponto merece registro explícito porque poderia parecer contradição com RA-2:
as métricas `hash_senha_duracao_segundos` e `hash_senha_em_espera` existem, mas
**não são um contador de falha de autenticação disfarçado** — não contam falha,
não têm rótulo de conta, e-mail ou origem, e não revelam se alguma tentativa
teve êxito. Coerente com `design.md` §8.2.

---

## 3. A01 — Controle de acesso: a propriedade do desenho **se sustenta**

Esta era a verificação central pedida. Resultado: **a propriedade é real**, não
apenas afirmada no documento.

### 3.1 `Escopo` obrigatório

- `autorizacao.Escopo` tem **todos os campos não exportados** e **nenhum
  construtor exportado** além do retorno de `Autorizar`. Fora do pacote,
  `Escopo{}` não compila e o valor zero é inválido (`Valido() == false`)
  — `domain/autorizacao/escopo.go`.
- `Autorizar` confere a permissão **antes** de produzir o `Escopo`; sem
  permissão devolve `ErrPermissaoNegada` e `Escopo{}` — `autorizar.go:75-78`.
- Os **três testes-guarda de reflexão passam** (`go test ./internal/port/...` →
  `ok`). Verifiquei que o guarda 1 inspeciona de fato `In(1)` de cada método das
  duas interfaces de negócio.
- `AplicarEscopo` é, de fato, **a única** função que monta o filtro de
  isolamento: grep por `AplicarEscopo` retorna 7 chamadas, **todas** em
  `usuario_repository.go`. `InstituicaoRepository` recusa qualquer escopo não
  `Plataforma()` em **todos** os 5 métodos.
- **Escalonamento de privilégio por payload está bloqueado em três camadas:**
  `ConjuntoInstitucional` recusa `administrador_sistema` (403
  `PERFIL_NAO_ATRIBUIVEL`); `conjuntoDoAlcance` ignora o payload nos alcances de
  PI e de administrador; e o banco **não consegue representar** o estado
  (`ck_usuario_perfil_estrutura` + FK composta).
- **A instituição do novo usuário nunca vem do payload** —
  `instituicaoDoAlcance` a tira do `Ator` ou do caminho
  (`criar_usuario.go:26-36`).

### 3.2 As duas portas estreitas — revisão humana (o ponto que dependia de mim)

**`AutenticacaoRepository` — 5 métodos. Aprovado.**

| Método | Identificador que recebe | Veredito |
|---|---|---|
| `BuscarCredencial` | `instituicaoID` + `email` do corpo do login | ✅ É o par de credenciais oferecido. Não vaza: o SQL já exclui conta excluída e instituição inativa, e a resposta é indistinguível de "não existe" |
| `CarregarContextoDeSessao` | `usuarioID` da claim `sub` de um **JWT já validado** | ✅ Roda antes de existir `Ator`; é o que o constrói. O identificador não é escolhido pelo cliente — vem de token assinado |
| `BuscarCredencialPropria` | `autorizacao.Proprio` | ✅ `Proprio` tem campos não exportados e **só nasce de `Ator.Proprio()`**. Impossível forjar de fora do pacote |
| `DefinirSenhaPropria` | `autorizacao.Proprio` | ✅ idem. `WHERE id=$3 AND excluido_em IS NULL` |
| `InvalidarSessoesProprias` | `autorizacao.Proprio` | ✅ idem |

Nenhum dos cinco aceita identificador escolhido pelo cliente via corpo, caminho
ou query string. A regra de admissão de `design.md` §4.4 **é respeitada**.

**`InstituicaoPublicaQuery.ListarParaCombo` — 1 método. Aprovado.**
Sem parâmetro algum. O `SELECT` projeta **exatamente** `id, nome, sigla` — nunca
`codigo_emec`, nunca contagem, nunca usuário. Filtra `situacao = 'ativa'`,
`excluido_em IS NULL` e exige `EXISTS` de detentor **ativo** do perfil de PI. O
DTO `InstituicaoPublicaResponse` tem só os três campos. Confere com spec 3.17 e
3.20, e o dano da rota está limitado ao que 4.2 aceita.

### 3.3 Isolamento entre instituições

- Filtro centralizado em `AplicarEscopo`; nenhum outro arquivo do pacote
  `postgres` escreve `instituicao_id` em cláusula `WHERE` de `usuario`.
- **404 e não 403** confirmado: o escopo entra no `WHERE` da própria consulta, e
  `sql.ErrNoRows` vira `domain.ErrNaoEncontrado` → 404. Vale para `BuscarPorID`,
  `Atualizar`, `ExcluirLogicamente`, `DefinirSenha` e `SubstituirPerfis`.
- **Ordem de avaliação correta:** permissão (403) no middleware, **antes** do
  isolamento (404) no repositório. Coberto por
  `TestSmoke_Usuarios_T05_PermissaoAntesDeIsolamento_403NaoDao404`.
- Testes de isolamento **T-02, T-03, T-04, T-06** existem no nível HTTP e passam
  (`go test ./internal/adapter/...` → `ok`).
- **`instituicao_id` nulo falha fechado:** se um escopo institucional chegasse
  com instituição nula, o SQL vira `instituicao_id = NULL`, que nunca é
  verdadeiro → zero linhas → 404. Comportamento correto.
- **Exceção do catálogo comum:** varri o código à procura de qualquer caminho
  que relaxe o filtro de instituição. **Não existe.** Nada no código atual abre
  exceção ao isolamento. (Ver, porém, o achado M-3.)
- **Download de arquivo:** não se aplica — nenhuma rota de arquivo nesta feature
  (MinIO ainda não provisionado, `project.config.md`).

### 3.4 A07 — Autenticação

- **Cookie** `HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age` — verificado
  em `auth_handler.go:64-72`. `Secure` em todos os ambientes, sem ramificar.
- **O token identifica, não autoriza.** Claims: `sub`, `ins`, `emt`, `iat`,
  `exp`. **Perfil, nome e e-mail não entram no token.** O middleware relê do
  banco, a cada requisição: conjunto de perfis, `excluido_em`, vínculo
  institucional, situação da instituição, `senha_provisoria` e
  `sessoes_validas_a_partir_de`. `jwt.WithValidMethods([]string{"HS256"})`
  bloqueia `alg: none` e confusão de algoritmo.
- **Divergência de instituição** entre a claim `ins` e o vínculo atual → 401.
  `mesmaInstituicao` trata corretamente o caso "um nulo e o outro não".
- **Os três códigos de 401 não vazam para anônimo.** `CONTA_EXCLUIDA` e
  `INSTITUICAO_INATIVA` só são alcançáveis **depois** de
  `tokenDeSessao.Validar(cookie)` ter sucesso — ou seja, quem os recebe já
  apresentou um JWT assinado e não expirado. Todo o resto (cookie ausente,
  assinatura inválida, expirado, linha ausente, `emt` anterior à invalidação,
  divergência de instituição) devolve `SESSAO_EXPIRADA`. **No login a mensagem é
  sempre genérica.** A fronteira de spec 3.18 está corretamente implementada.
- **Porta de senha provisória é do servidor**, não da tela: o middleware recusa
  com 403 `SENHA_PROVISORIA` qualquer rota fora das quatro permitidas, usando
  `c.FullPath()` (padrão da rota, não string crua).
- Auto-exclusão, auto-redefinição de senha e alteração dos próprios perfis
  bloqueadas (`ErrAutoExclusaoNegada`, `ErrRedefinirPropriaSenhaNegada`,
  `ErrAlteracaoDosPropriosPerfisNegada`).

---

## 4. Achados

### 🔴 C-1 — A03 · Dependência vulnerável com correção disponível: `quic-go@v0.59.0`

**Onde:** `backend/go.mod:46` (`github.com/quic-go/quic-go v0.59.0 // indirect`).

**Saída real da varredura** (`docker compose -f docker-compose.dev.yml exec backend go run golang.org/x/vuln/cmd/govulncheck@latest ./...`):

```
Vulnerability #1: GO-2026-5676
    HTTP/3 QPACK Trailer Expansion Memory Exhaustion in
    github.com/quic-go/quic-go
  Module: github.com/quic-go/quic-go
    Found in: github.com/quic-go/quic-go@v0.59.0
    Fixed in: github.com/quic-go/quic-go@v0.59.1
    Example traces found:
      #1: cmd/api/main.go:176:22: api.main calls gin.Engine.Run, which
          eventually calls http3.ConfigureTLSConfig
Your code is affected by 1 vulnerability from 1 module.
This scan also found 0 vulnerabilities in packages you import and 6
vulnerabilities in modules you require, but your code doesn't appear to
call these vulnerabilities.
exit status 3
```

**Vetor de ataque:** expansão descontrolada de trailers QPACK em conexão HTTP/3
leva a exaustão de memória do processo — negação de serviço sem autenticação.

**Por que é Crítico:** pela regra desta revisão e do `CLAUDE.md`
("vulnerabilidade com correção disponível e severidade alta ou crítica"), há
correção publicada (v0.59.1) e `govulncheck` classifica o código como afetado.
A fase `security` do pipeline **falha o build** neste estado.

**Contexto atenuante que o `arquiteto` deve pesar:** a aplicação **não serve
HTTP/3** — `router.Run(":"+porta)` é `http.ListenAndServe` puro, HTTP/1.1. O
`quic-go` entra como dependência **indireta** do `gin v1.12.0`, e os traços
mostrados pelo `govulncheck` são conservadores (o traço #3, `rand.Read` →
`http3.countingByteReader.Read`, é claramente artefato de análise estática). A
exploração real exigiria que alguém habilitasse HTTP/3 no servidor.

**Não proponho a correção.** Registro o dado: existe versão corrigida a uma
linha de distância e o gate de CI está vermelho. O `arquiteto` decide entre
bump direto, bump do `gin`, ou exceção documentada; o `dev-fullstack` executa
depois da decisão. **Nunca junto com feature no mesmo PR** (`CLAUDE.md`).

**Frontend e raiz estão limpos:** `npm audit --audit-level=high` em
`frontend/` → `found 0 vulnerabilities` (901 dependências); na raiz (Playwright,
husky, commitlint) → `found 0 vulnerabilities`.

---

### 🔴 C-2 — Comentários em código que chega ao navegador

**Regra violada:** `CLAUDE.md`, "Comentários que chegam ao usuário final —
proibidos" → achado **Crítico**, não Baixo nem Médio. Também contraria
`design.md` §11.5 e §15.10 ("Zero comentários em `.ts`/`.tsx`/`.css`").

**Levantamento:** varredura própria com remoção de literais de string antes da
detecção. **35 comentários em 13 arquivos** de `frontend/src/`, todos em módulos
que entram no bundle do navegador.

**Os dois com conteúdo que efetivamente descreve mecanismo de segurança
— são estes que carregam risco real:**

`frontend/src/features/auth/hooks/use-guarda-senha-provisoria.ts:5-8`
```
// useGuardaSenhaProvisoria cancela a navegação em vez de navegar e voltar
// (design.md §2.3): clique em qualquer âncora é interceptado em fase de
// captura antes do React Router agir, e o botão Voltar do navegador é
// neutralizado reempilhando o estado atual.
```

`frontend/src/lib/api-client.ts:40-43`
```
// O login (e a própria troca de senha provisória) chamam a API antes
// de haver sessão válida — um 401 ali é resultado esperado da tela
// (credenciais erradas), não perda de sessão. O interceptor global só
// se aplica às chamadas feitas de dentro do shell autenticado.
```

**Vetor de ataque:** ambos descrevem, em prosa, **como um controle do cliente
funciona e onde ele não se aplica** — o primeiro detalha a técnica exata de
bloqueio da navegação com senha provisória (fase de captura, `pushState`), o
segundo identifica quais chamadas escapam do interceptor global de 401. Para
quem examina o bundle no DevTools, isso é um mapa de onde tentar contornar o
fluxo obrigatório de troca de senha (spec 3.4), que é o controle de que depende
a atribuibilidade de todo o relatório de metas. Ambos ainda **citam documento
interno de arquitetura** (`design.md §2.3`).

**Os demais 33** (18 diretivas `// eslint-disable-next-line`, 12 marcadores de
seção em `components/ui/drawer.tsx` vindos do shadcn, e três notas curtas em
`use-local-storage.ts:16`, `login-form.tsx:35`) violam a mesma regra, mas não
revelam nada de arquitetura. Separo-os porque a decisão sobre eles é de outra
natureza — inclusive porque remover as diretivas de lint muda o comportamento do
`biome ci`.

**Contexto que o `arquiteto` deve pesar:** no build de produção (`next build`,
`output: standalone`) o SWC minifica e **remove comentários**, e
`productionBrowserSourceMaps` não está habilitado — então, **num deploy de
produção, esses comentários não chegam ao navegador**. Eles chegam hoje, sim, em
`next dev` (`docker-compose.dev.yml`), onde os módulos são servidos sem
minificação. A regra do `CLAUDE.md` é escrita no nível do código-fonte e é
categórica, e por isso classifico como Crítico — mas registro honestamente que a
exploração concreta depende de o ambiente estar em modo de desenvolvimento.

---

### 🟡 A-1 — A02 · A API não emite **nenhum** cabeçalho de segurança, em nenhum ambiente

**Onde:** `backend/cmd/api/main.go:119-123` — a cadeia de middlewares é
`Recovery → Metricas → CORS → Erro`. **Não existe middleware de segurança.**

**Divergência do desenho aprovado:** `design.md` §4.3 lista
`adapter/http/middleware_seguranca` entre os arquivos da feature. **O arquivo
não existe** (`find backend -name "middleware_seguranca*"` → vazio; grep por
`X-Content-Type-Options|X-Frame-Options|Strict-Transport-Security|Content-Security-Policy|Referrer-Policy|Permissions-Policy`
em todo o `backend/` → **nenhuma ocorrência**).

**Verificação empírica:**
```
$ curl -s -D - -o /dev/null http://localhost:3001/api/v1/publico/instituicoes
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 28 Sep 2026 21:21:58 GMT
Content-Length: 181
```
Nenhum dos cinco cabeçalhos exigidos pelo `CLAUDE.md` em `production`.

**Por que o `proxy.ts` não cobre isto:** por D-12, o **navegador chama a API
direto** (`NEXT_PUBLIC_API_URL`, `credentials: "include"`) — a origem
`:3001` nunca passa pelo middleware do Next. Os cabeçalhos do `proxy.ts`
protegem apenas os documentos servidos pelo `:3000`.

**Vetores concretos na origem da API:**
1. **Sem `X-Content-Type-Options: nosniff`** — respostas JSON cujo conteúdo é
   parcialmente controlado pelo usuário (nome, e-mail, mensagem de erro com
   `campo`) ficam sujeitas a MIME sniffing por navegadores legados quando
   carregadas fora de contexto `fetch`.
2. **Sem `X-Frame-Options` / `frame-ancestors`** — a API **serve HTML**:
   `/swagger/index.html` responde 200 (ver A-2). Essa página é enquadrável em
   `<iframe>` de terceiro.
3. **Sem `Strict-Transport-Security`** na origem que transporta o cookie de
   sessão. O cookie é `Secure`, o que impede o envio em texto claro, mas nada
   força o navegador a nem tentar HTTP nessa origem.
4. **Sem `Referrer-Policy`** — URLs da API contêm identificadores de usuário e
   de instituição no caminho (`/instituicoes/{id}/pesquisadores/{usuario_id}`).

Pela régua desta revisão (`APP_ENV=production` sem cabeçalhos de segurança →
🟡 Alto) e por ser divergência de um item nominalmente presente no `design.md`,
**bloqueia a entrega**.

---

### 🟡 A-2 — A02 · Swagger UI e especificação OpenAPI públicos, sem autenticação e sem ramificação por ambiente, revelando implementação interna

**Onde:** `backend/cmd/api/main.go:130` —
```go
router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
```
Registrada **fora** de `grupoAPI`, sem middleware de sessão e **sem nenhuma
condicional por `APP_ENV`**. O mesmo código roda em produção.

**Verificação empírica:**
```
$ curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3001/swagger/index.html
200
$ curl -s -o /dev/null -w "%{http_code} bytes=%{size_download}\n" http://localhost:3001/swagger/doc.json
200 bytes=65273
```

**Vetor de ataque:** qualquer visitante anônimo obtém o mapa completo das 29
rotas, com métodos, parâmetros, esquemas de corpo e **a tabela de códigos de
erro por rota** — incluindo `PERMISSAO_NEGADA`, `SENHA_PROVISORIA`,
`ULTIMO_PESQUISADOR_INSTITUCIONAL` e os três códigos de 401. Combinado com o
risco aceito RA-4 (lista pública de instituições) e RA-1 (sem limite de
tentativas), isso entrega ao atacante o alvo **e** a superfície completa, sem
precisar descobrir nada.

**Agravante — a especificação revela implementação interna**, o que o
`CLAUDE.md` classifica separadamente ("Descrição de OpenAPI que revela
implementação interna"). Extraído do `doc.json` servido agora:

| Rota | Descrição publicada |
|---|---|
| `GET /api/v1/usuarios` | "A mesma operação serve três alcances (**design.md §4.3**): usuários da própria instituição (PI), pesquisadores de uma instituição (Administrador) e administradores da plataforma (Administrador)." |
| `POST /api/v1/administradores` | "\"perfis\" só é lido na rota de usuários (PI) — nas demais o conjunto é fixo pelo alcance e **o campo é ignorado** (**design.md §6.2**)." |
| `PUT /api/v1/administradores/{usuario_id}` | "\"perfis\" é sempre o conjunto final, nunca uma diferença (**design.md §6.2, T-056**). Concorrência otimista via \"versao\" — ver **CLAUDE.md**." |
| `GET .../pesquisadores` (resposta 400) | "PARAMETRO_INVALIDO (**allowlist de ordenação**, perfil de filtro inválido)" |

Nomes de documentos internos (`design.md`, `CLAUDE.md`), números de seção,
identificador de tarefa (`T-056`) e a existência de uma allowlist de ordenação
são detalhe de implementação publicado a anônimos. A linha sobre o campo
`perfis` ser "ignorado" nas rotas de administrador chega a descrever o
comportamento de uma **fronteira de privilégio**.

---

### 🟠 M-1 — A02/A09 · `/metrics` exposto sem autenticação

**Onde:** `backend/cmd/api/main.go:129`; `docker-compose.yaml:49-50` publica
`3001:3001` no host em produção.

```
$ curl -s -o /dev/null -w "HTTP %{http_code} bytes=%{size_download}\n" http://localhost:3001/metrics
HTTP 200  bytes=16477
```

**Vetor:** um anônimo lê `autorizacao_decisoes_total{permissao,resultado}`
(quantas tentativas de acesso negado, por permissão),
`auditoria_eventos_total{canal,acao,resultado}` (volume de exclusões,
redefinições de senha, inativações de instituição), `hash_senha_em_espera` e
`hash_senha_duracao_segundos` (que, na prática, dá a **contagem de tentativas de
login**), além de `go_*` e `process_*` com versão exata do runtime e uso de
memória — insumo para dimensionar o vetor de exaustão que D-02 existe para
conter. Nada disso é dado pessoal, mas é inteligência operacional gratuita.

Em Kubernetes o padrão é o scrape ficar na rede interna; **não há manifesto K8s
neste repositório ainda**, e o `docker-compose.yaml` de produção publica a porta
inteira no host. Decisão de exposição é do `arquiteto`.

---

### 🟠 M-2 — A10 · `/readyz` devolve a mensagem crua do driver a um anônimo

**Onde:** `backend/internal/adapter/http/health_handler.go:40-46`
```go
if err := h.db.PingContext(ctx); err != nil {
    falhas["postgres"] = err.Error()
}
...
c.JSON(http.StatusServiceUnavailable, gin.H{"status": "indisponivel", "falhas": falhas})
```

**Reproduzido** (teste descartável no container, com credencial inválida, para
capturar exatamente a string que iria para o corpo da resposta):

```
failed to connect to `user=basisavalia database=basisavalia_dev`:
172.18.0.3:5432 (postgres): failed SASL auth: FATAL: password
authentication failed for user "basisavalia" (SQLSTATE 28P01)
```

**Vetor:** quando o Postgres ou o Redis ficam indisponíveis — momento em que o
atacante mais quer informação, e que ele pode simplesmente esperar ou induzir
com carga — a rota **pública e sem autenticação** `/readyz` entrega **usuário do
banco, nome do banco, endereço IP interno, porta, nome de host do serviço e o
SQLSTATE**. Combinado com RA-1 e RA-3, o nome do usuário do banco é exatamente
o tipo de dado que alimenta a próxima tentativa.

Contraria `CLAUDE.md` ("Respostas de erro da API: nunca incluir stack trace,
nome de tabela, query SQL ou caminho de arquivo — apenas código de erro e
mensagem genérica") e `design.md` §15 item 11. Note que o `MiddlewareErro` faz a
coisa certa para todo o resto (`ERRO_INTERNO` genérico + detalhe só no log do
servidor) — `/readyz` é a exceção que escapou porque não passa por ele.

---

### 🟠 M-3 — A01 · `AplicarEscopo` omite `instituicao_id IS NULL` no ramo de plataforma (fail-open latente)

**Onde:** `backend/internal/adapter/postgres/escopo_sql.go:31-36`

```go
if !escopo.Plataforma() {
    condicoes = append(condicoes, fmt.Sprintf("usuario.instituicao_id = $%d", n))
    ...
}
// ← não há ramo "else" que acrescente `usuario.instituicao_id IS NULL`
```

`design.md` §4.3 especifica explicitamente o contrário:

```sql
-- quando o escopo é de plataforma
AND u.instituicao_id IS NULL
```

**Não é explorável hoje.** O único alcance de plataforma que alcança o
`UsuarioRepository` é `AdministradoresDaPlataforma`, que sempre traz
`exigePerfil = administrador_sistema`; como
`ck_usuario_perfil_estrutura` garante no banco que esse perfil implica
`instituicao_id IS NULL`, o `EXISTS` acaba produzindo o mesmo recorte.
`InstituicoesDaPlataforma` nunca chega aqui (o `InstituicaoRepository` tem
consultas próprias). Confirmei as 7 chamadas de `AplicarEscopo`, todas em
`usuario_repository.go`.

**Vetor futuro, concreto:** no dia em que existir um alcance de plataforma **sem
`exigePerfil`** — e o candidato óbvio já se desenha, uma tela de suporte que
liste usuários de toda a plataforma — `AplicarEscopo` devolverá apenas
`usuario.excluido_em IS NULL`, ou seja, **todos os usuários de todas as
instituições**, sem nenhum outro guarda no caminho. E cairá exatamente na função
que o próprio comentário declara ser "a ÚNICA função do sistema que monta o
fragmento WHERE de isolamento", que é o último lugar onde alguém iria procurar.
A propriedade "esquecer o isolamento não compila" não cobre este caso: compila,
passa nos três guardas de reflexão, e vaza.

---

### 🟠 M-4 — A02 · Container do backend roda como `root`

**Onde:** `backend/Dockerfile` — o estágio `runner` não tem diretiva `USER`. O
`frontend/Dockerfile:20` faz o certo (`USER node`), o que torna a inconsistência
provavelmente um esquecimento e não uma decisão.

**Vetor:** qualquer execução de código no processo Go (via dependência
comprometida — ver C-1 — ou falha futura de memória) começa com uid 0 dentro do
container, ampliando o alcance de uma escapada de container e permitindo
escrita em todo o sistema de arquivos da imagem, inclusive nas migrations
copiadas em `/app/migrations`.

---

### 🔵 Observações (não bloqueiam)

**O-1 — Os guardas de §4.4 congelam os *nomes*, não as *assinaturas*, das portas
estreitas.** `design.md` §4.4 afirma que os testes "verificam a **assinatura**
dos métodos e o conjunto de métodos das portas estreitas". Na prática
(`internal/port/portas_test.go`), `TestPortasEstreitas_ListaFechada` compara
**apenas nomes**, e `TestNenhumaPortaNovaSemEscopo` **pula** as quatro portas
principais via o mapa `permitidas` — além de só detectar `uuid.UUID` cru,
deixando passar `*uuid.UUID`. Consequência: trocar
`BuscarCredencial(ctx, *uuid.UUID, Email)` por um método que receba um
identificador escolhido pelo cliente, **mantendo o nome**, não quebra nenhum
teste. A revisão humana que o desenho reserva a este ponto (§3.2 acima)
continua sendo, de fato, a única barreira — e o documento a descreve como se
fosse redundante com um mecanismo automático que não existe nessa força.

**O-2 — CORS sem `Vary: Origin`.** `cmd/api/main.go:181-196` ecoa a origem
recebida em `Access-Control-Allow-Origin` sem emitir `Vary: Origin`. Com um
cache compartilhado à frente da API, uma resposta com o cabeçalho de uma origem
pode ser servida a outra. A allowlist em si está correta (comparação exata,
nunca `*`, e `permitida != ""` impede que `CORS_ALLOWED_ORIGINS` vazia libere
origem vazia). Impacto real limitado por `SameSite=Strict`.

**O-3 — Senhas fictícias de seed em arquivo versionado.**
`backend/cmd/seed/dev.go:50-61` traz dez senhas em texto claro
(`plataforma-2026`, `reuniao-nde-2026`, …). Leitura literal do `CLAUDE.md`
("nenhuma senha em texto plano em arquivo versionado") é violada, mas o portão
está correto: a massa só é criada com `APP_ENV=development`
(`cmd/seed/main.go:31`), o padrão na ausência da variável é `production`, e as
pessoas são fictícias (spec §8 — sem dado real, LGPD atendida). O
administrador inicial de produção vem só de `SEED_ADMIN_EMAIL`/`SEED_ADMIN_SENHA`,
sem valor padrão, e o processo aborta se faltarem. Registro por completude.

**O-4 — `LIKE` sem escapar `%` e `_` na busca.**
`usuario_repository.go:140` e `instituicao_repository.go:113` montam
`"%"+filtro.Busca+"%"` como **parâmetro vinculado** — não é injeção de SQL. Mas
um `%` digitado pelo usuário vira coringa, o que é comportamento inesperado e,
com `unaccent(lower(...))` sobre a coluna, um padrão como `%_%_%_%` pode
encarecer a varredura. Custo baixo dado o volume de P7.

**O-5 — CSP de produção sem `base-uri` e `form-action`.** `frontend/src/proxy.ts`
implementa **exatamente** a política prescrita no `CLAUDE.md`, e acrescenta
corretamente a origem da API em `connect-src`. Como `base-uri` e `form-action`
não herdam de `default-src`, uma injeção de HTML poderia usar `<base>` ou
redirecionar um `<form>`. Isso é lacuna do padrão do projeto, não desta
implementação — registro para o `arquiteto` decidir se o padrão muda.

**O-6 — Tabela `auditoria` é append-only apenas por convenção.** Não há
`REVOKE UPDATE, DELETE` nem trigger impeditiva; a integridade depende de nenhum
caminho de código escrever ali (verifiquei: só há `INSERT`). Como a trilha
administrativa é justamente o mitigante que sustenta o risco aceito RA-2,
vale a pena que a decisão de reforçá-la ou não seja explícita.

---

## 5. Itens verificados **sem** achado

| Categoria | Verificação | Resultado |
|---|---|---|
| **A01** | `Escopo` inconstruível fora de `Autorizar`; 3 guardas de reflexão passando; `AplicarEscopo` único; `InstituicaoRepository` exige `Plataforma()` | ✅ |
| **A01** | Duas portas estreitas revisadas manualmente, método a método | ✅ (§3.2) |
| **A01** | Isolamento 404-não-403; testes T-02..T-06 passando; nenhuma exceção ao isolamento no código atual | ✅ |
| **A01** | Escalonamento de privilégio por payload bloqueado em 3 camadas (VO, alcance, `CHECK` do banco) | ✅ |
| **A01** | Instituição e conjunto de perfis do novo usuário vêm do ator/alcance, nunca do payload | ✅ |
| **A01** | SSRF: nenhuma chamada de saída a URL controlada pelo cliente nesta feature | ✅ |
| **A01** | IDOR por enumeração: UUIDv7 gerado no domínio; nenhum `SERIAL`/`BIGSERIAL` no schema; nenhum id sequencial em rota ou resposta | ✅ |
| **A02** | `APP_ENV=development` aparece **somente** em `docker-compose.dev.yml` (4 serviços); `docker-compose.yaml` de produção não o define; padrão no código Go é `production` em `main.go` e em `cmd/seed/main.go` | ✅ |
| **A02** | CORS por allowlist exata, nunca `*` | ✅ |
| **A02** | Nenhum segredo em arquivo versionado. `.env.example` tem só chaves vazias; `.gitignore` cobre `.env`/`.env.*`; `git ls-files` não traz nenhum `.env` real | ✅ |
| **A02** | Source maps de produção: `productionBrowserSourceMaps` não habilitado em `next.config.ts` (padrão `false`) | ✅ |
| **A03** | `npm audit --audit-level=high`: 0 vulnerabilidades em `frontend/` e na raiz | ✅ |
| **A04** | argon2id (não bcrypt), PHC, `ConstantTimeCompare`, salt de `crypto/rand`; JWT HS256 com `exp` (8 h) e `WithValidMethods`; `JWT_SECRET` obrigatório sem padrão | ✅ |
| **A04** | `SenhaEmTexto.String()` → `"[senha omitida]"`; `SenhaHash.String()` → `"[hash omitido]"`; nenhum `log` com senha, hash, token ou cookie | ✅ |
| **A05** | Toda query com parâmetros vinculados. `sort` por allowlist **em duas camadas** (`ParsePaginacao` com 400 explícito, sem clamp; e mapa no adapter); `order` só `asc`/`desc`; `page_size` máx. 100 com 400 fora da faixa | ✅ |
| **A05** | XSS: zero `dangerouslySetInnerHTML`, `innerHTML`, `eval`, `new Function`, `document.write` em `frontend/src/` | ✅ |
| **A05** | Open redirect: todos os `window.location` usam destino literal | ✅ |
| **A06** | Idempotência: dispensada com justificativa (as unicidades tornam a duplicação impossível — spec 3.14). Concorrência otimista com `versao` → 409 | ✅ |
| **A08** | Migrations pareadas `up`/`down`; `NULLS NOT DISTINCT`, FK composta e `ck_usuario_perfil_estrutura` presentes conforme o desenho; sem `DELETE` físico de entidade | ✅ |
| **A09** | 13 ações administrativas auditadas na mesma transação da mutação; `acesso_negado` registrado com IP de origem; syslog para as ações críticas | ✅ |
| **A10** | `MiddlewareErro` nunca devolve stack trace, SQL, nome de tabela ou caminho; detalhe só no log do servidor; `gin.Recovery()` ativo | ✅ (exceto `/readyz` — M-2) |
| **LGPD** | Nenhum identificador forte coletado (sem CPF, RG, telefone — spec "Não-objetivos"). `detalhes` da auditoria guarda **nomes de campo** e conjuntos de perfis, nunca valores. Seed não-produtivo só com dado fictício. Nenhum envio a API externa. Retenção declarada em P9 | ✅ |
| **Testes** | `go test ./internal/port/... ./internal/domain/... ./internal/adapter/...` → todos `ok` | ✅ |

---

## 6. Encaminhamento

1. **C-1, C-2, A-1 e A-2 bloqueiam a entrega.** Precisam de decisão do
   `arquiteto` sobre como corrigir, ou de aceitação explícita de risco pelo dono
   do produto, registrada como os riscos da seção 2.
2. **M-1 a M-4 e O-1 a O-6** não bloqueiam. Recomendo que o `arquiteto` avalie
   **M-3** junto com a próxima feature que introduza alcance de plataforma, e
   **O-1** ao revisar a redação de `design.md` §4.4 — hoje o documento descreve
   uma garantia automática mais forte do que a que os testes entregam.
3. **Não decidi nenhuma solução.** Onde escrevi "contexto atenuante", é insumo
   para a decisão, não a decisão.
4. Este documento não toca `specs/_status.md` (`code-reviewer` trabalhando em
   paralelo em T-078).
