# Evidence: autenticacao-usuarios

## Resultado da validação — T-079 (`qa-tester`), rodada 2

**Data:** 29/09/2026 · **Executor:** `qa-tester` · **Referência:** `spec.md` §7
(critérios de aceite) e §15.3 (critério de aceitação por aritmética de
conjuntos), `tasks.md` (T-079), `testes-pendentes.md`.

> Esta é a segunda rodada de T-079. A primeira (mesma data) apontou três
> lacunas: `U-02`/`U-03`/`U-09` sem teste apesar de obrigatórios (§15.1),
> os 12 cenários `SH-XX` do AppShell nem testados nem registrados como
> pendência, e rótulos trocados em `I-08`/`I-12`. As três foram endereçadas
> pelo `dev-fullstack` entre as duas rodadas. Refiz a aritmética do zero —
> não assumi que as correções bateram, conferi cada uma.

### 1. Suítes executadas

**Go — `go test ./... -count=1`** (contra `basisavalia_test`, banco real):

```
docker compose -f docker-compose.dev.yml exec -T \
  -e DATABASE_URL_TEST="postgres://basisavalia:basisavalia_dev_local@postgres:5432/basisavalia_test?sslmode=disable" \
  backend go test -count=1 ./...
```

**Verde.** Todos os pacotes `ok`, nenhuma falha — inclusive os três testes
novos, confirmados isolados também:

```
docker compose ... backend go test -count=1 -run \
  'TestUsuarioRepository_U02|TestUsuarioRepository_U03|TestUsuarioRepository_U09' \
  -v ./internal/adapter/postgres/...

--- PASS: TestUsuarioRepository_U02_EmailDuplicadoNaMesmaInstituicao (0.06s)
--- PASS: TestUsuarioRepository_U03_EmailDeExcluidoVoltaAFicarLivre (0.07s)
--- PASS: TestUsuarioRepository_U09_MesmoEmailEmDuasInstituicoes (0.08s)
PASS
```

Os três exercitam `Inserir`/`Atualizar` de verdade contra o índice único
parcial `(instituicao, email) WHERE excluido_em IS NULL` — não é smoke,
é o comportamento do banco real: `U-02` recusa duplicado na mesma
instituição sem criar linha; `U-03` libera o e-mail de um excluído
logicamente para novo cadastro sem tocar a linha antiga; `U-09` grava
duas contas independentes para o mesmo e-mail em instituições diferentes,
sem uma vazar para a outra.

**E2E — Playwright.** Consegui rodar nesta rodada (na anterior, bloqueado
por infraestrutura em provisionamento). Duas travas de permissão
apareceram e removi (mesma causa que o coordenador já tinha diagnosticado
— diretório `e2e/test-results` recriado com dono `root` pelo container):
corrigi a posse (`chown 1000:1000`, via container descartável, sem apagar
nada) em vez de repetir a remoção destrutiva.

**1ª tentativa (banco `basisavalia_e2e` com dado residual de execuções
anteriores, inclusive minhas próprias tentativas frustradas nesta
sessão):** 3 falhas em `auth/login-logout.spec.ts` e
`auth/mesmo-email-duas-instituicoes.spec.ts` — consistente com o
diagnóstico do coordenador: a suíte muda senha provisória e cria
usuários sem restaurar tudo, e o banco de E2E é persistente por desenho.

**Recriei o banco do zero** (`DROP DATABASE`/`CREATE DATABASE
basisavalia_e2e`, depois `migrate-e2e` + `seed-e2e`) e rodei de novo:

```
Running 12 tests using 1 worker

✓  auth/login-logout.spec.ts — 3/3
✓  auth/mesmo-email-duas-instituicoes.spec.ts — 3/3
✓  auth/primeiro-acesso.spec.ts — 1/1
✓  usuario/crud.spec.ts — 1/1
✓  usuario/isolamento.spec.ts — 1/1
✓  smoke/abre-sem-quebrar.spec.ts — 1/1
✓  metas/catalogo-comum.spec.ts:10 — 1/1
✘  metas/catalogo-comum.spec.ts:55 — falha na RESTAURAÇÃO do teste
   ("excluir admin temporário" devolveu não-ok), não na asserção do
   cenário em si

11 passed, 1 failed
```

**Os 11 specs de `autenticacao-usuarios` passaram 100%** — todos os seis
arquivos que tocam esta feature (`login-logout`, `mesmo-email-duas-
instituicoes`, `primeiro-acesso`, `crud`, `isolamento`, `abre-sem-
quebrar`). A única falha é em `metas/catalogo-comum.spec.ts`, que
**não é desta feature** — pertence a `metas-coordenacao`/`indicadores` —
e falha no próprio `afterEach` de restauração (exclusão de um
administrador temporário), não numa asserção de cenário. Não investiguei
mais fundo porque está fora do escopo de T-079 desta feature, mas
registro para quem for validar `metas-coordenacao`: a suíte E2E, hoje,
ainda tem pelo menos um ponto de fragilidade em banco não-virgem, fora de
`autenticacao-usuarios`.

**Confirma o que o coordenador reportou:** a suíte de `autenticacao-
usuarios` passa de ponta a ponta com banco virgem; com banco reaproveitado
de execuções anteriores, falha por resíduo de dado (senha deixou de ser
provisória, usuário de teste já existe, etc.) — a correção de fixtures
está em andamento pelo `dev-fullstack`, ainda não verifiquei essa correção
especificamente (não é o que me foi pedido nesta rodada).

---

### 2. Aritmética de conjuntos (spec §15.3) — refeita

`C` continua com **131** identificadores (o enunciado original desta
tarefa dizia 121; mantenho a correção já registrada na rodada anterior).

**O que mudou desde a rodada 1, conferido item a item:**

| Mudança | Efeito na conta |
|---|---|
| `U-02`, `U-03`, `U-09` ganharam teste de repositório dedicado, confirmado passando | Saem de `B`/`esquecido`, entram em `A` |
| Os 12 `SH-01, SH-03` a `SH-13` ganharam código explícito no Grupo 10 de `testes-pendentes.md` | Saem de `esquecido`, entram em `B` (continuam adiados, agora visíveis) |
| `I-08`: a linha antiga citava um teste (`TestInstituicaoRepository_I08_...`) que **não** verifica o `I-08` atual da spec — o `dev-fullstack` corrigiu o rótulo em vez de apagar a linha, e registrou que o `I-08` real está **parcialmente** coberto (metade "fica fora do combo" tem teste; metade "destacada no grid" não tem) | Antes eu tinha `I-08` errado em `A` (confiando no rótulo antigo). Corrijo: `I-08` não é obrigatório pela §15.1, e a metade que falta é a mais central do enunciado (o destaque visual) — sigo a conclusão do próprio `testes-pendentes.md` ("pendência real, não coberta") e movo `I-08` para `B` |
| `I-12`: rótulo também corrigido — a metade de backend (`409 CONFLITO_DE_VERSAO`) está coberta por `TestSmoke_Instituicoes_CicloCompletoDeStatusCodes`; a metade "mensagem na tela" não | `I-12` **é** obrigatório pela §15.1 ("Concorrência otimista"). A parte que a §15.1 protege — a escrita concorrente real, não a UX da mensagem — está coberta e passando. Mantenho `I-12` em `A` |

**`A` (88 cenários) — teste automatizado que passa:**

```
A-01 A-02 A-03 A-04 A-05 A-06 A-07
AS-02 AS-03 AS-06 AS-07 AS-08 AS-09 AS-10
E-01 E-03 E-04 E-05 E-08 E-09 E-10 E-11 E-12 E-14 E-15 E-16 E-17
G-02 G-03 G-04 G-07 G-08 G-09 G-10 G-17
I-01 I-02 I-04 I-05 I-09 I-11 I-12
L-01 L-02 L-03 L-04 L-05 L-08 L-11 L-12 L-13 L-14 L-15 L-16 L-17
S-02 S-03 S-09 S-10 S-11 S-12
SE-01 SE-02 SE-04 SE-05 SE-06 SE-07 SE-08
SH-02
T-01 T-02 T-03 T-04 T-05 T-06
U-01 U-02 U-03 U-04 U-06 U-07 U-09 U-10 U-11 U-12 U-13 U-14 U-15
```

**`B` (43 cenários) — pendência real em `testes-pendentes.md`:**

```
AS-01 AS-04 AS-05
E-02 E-06 E-07 E-13
G-01 G-05 G-06 G-11 G-12 G-13 G-14 G-15 G-16 G-18 G-19
I-03 I-06 I-07 I-08 I-10
L-09 L-10
S-01 S-04
SE-03 SE-09
SH-01 SH-03 SH-04 SH-05 SH-06 SH-07 SH-08 SH-09 SH-10 SH-11 SH-12 SH-13
U-05 U-08
```

```
|A| = 88     |B| = 43     |A| + |B| = 131 = |C|
A ∩ B = ∅
C \ (A ∪ B) = ∅
```

**Todos os cenários obrigatórios da §15.1 estão em `A`, nenhum em `B`** —
conferi a lista completa (isolamento T-01 a T-06; invariantes de corrida
E-09/E-17/AS-06; união de permissões A-07/U-14/U-15; matriz de
autorização A-01 a A-06/AS-03/AS-08/AS-10/U-10/T-05; concorrência
E-03/I-12; unicidade U-02/U-03/U-09/I-04/I-05; padrão de perfis
U-13/E-16; fronteira de divulgação L-08/SE-08) — os 33 itens, todos em
`A`.

**Observação de higiene, menor, sem efeito na conta:** `testes-
pendentes.md` ainda descreve `U-04 a U-08` como um intervalo "adiado",
mas `U-04`, `U-06` e `U-07` têm teste próprio, fora desse intervalo
(`TestEmail_U04_NormalizaTrimEMinusculas`, `TestEmail_FormatoInvalido` —
que verifica exatamente a entrada de `U-06`, `joao.ribeiro.ies.edu.br` —
e `TestConjuntoInstitucional_U07`/`TestCriarUsuario_U07`). Não afeta a
aritmética (classifiquei os três em `A`, com evidência direta de teste),
mas fica registrado para o `dev-fullstack` ajustar a redação do intervalo
quando mexer nesse arquivo de novo — é o mesmo tipo de rótulo impreciso
que já foi corrigido para `I-08`/`I-12`, só que sem consequência porque
não é obrigatório pela §15.1.

---

### 3. Veredito da aritmética

```
A ∪ B = C   e   A ∩ B = ∅   e   todo teste de A passa   e
todo cenário de 15.1 está em A, nunca em B
```

**As quatro condições da spec §15.3 estão satisfeitas. A feature está
aceita pelo critério mecânico.**

Isto não substitui a decisão do dono do produto sobre produção — é a
conta que a spec pede, e ela fecha.

---

### 4. Riscos aceitos (spec §4.1) — registro, não pendência

Sem mudança desde a rodada 1. Confirmo o que `security-review.md` (T-077,
§2) já registrou — não repito a revisão:

| # | Risco aceito | Referência |
|---|---|---|
| RA-1 | Sem limite de tentativas de login — sem bloqueio, sem 429, sem contador | spec 3.2, 4.1 |
| RA-2 | Sem registro de tentativa de login, de login bem-sucedido nem de logout; sem métrica de falha de autenticação | spec 3.2, 4.1 |
| RA-3 | Sem política de senha — só "não pode ser vazia" + limite técnico de 1024 caracteres (P12) | spec 3.3, 4.1 |
| RA-4 | Lista de instituições visível a qualquer visitante (rota pública do combo) | spec 3.17, 4.2 |

Os quatro mitigantes continuam de pé, conforme `security-review.md` §2.1
(verificado pelo `security-reviewer`, não repetido por mim): resposta de
login genérica e de tempo constante (confirmado também pela suíte Go —
`TestAutenticar_L02/L03`, `TestHashDeSenha_L08`, ambos em `A` e verdes);
hash lento (argon2id); cookie `HttpOnly`+`Secure`+`SameSite=Strict`; e
auditoria administrativa íntegra.

---

### 5. Validação manual do dono do produto — o que de fato houve, e o que fica sem cobrir

**O que de fato aconteceu:** o dono do produto aprovou as telas **em
bloco**, de forma agregada — não recebi, e não tenho como reconstruir,
uma lista de quais dos 43 cenários de `B` ele efetivamente exercitou na
tela. Não vou inventar essa confirmação. Registro a aprovação em bloco
como o que ela é: validação do fluxo principal e da aparência geral,
válida como Camada 3 da estratégia de testes do `CLAUDE.md` — mas **não**
uma confirmação cenário-a-cenário conforme a letra da §15.3(b).

**A distinção que consigo fazer, sem inventar dado que não tenho:** dos 43
cenários de `B`, alguns descrevem o que **qualquer** uso normal das telas
principais (login, listagem de usuários, listagem de instituições,
administradores) necessariamente atravessa — é difícil usar a tela sem
ver essas coisas acontecerem. Outros exigem uma ação deliberada, um dado
específico ou duas sessões simultâneas que uma navegação comum não
produz sozinha. Separo os dois grupos para que o coordenador leve ao dono
só a pergunta que falta, não a lista inteira:

**Provavelmente cobertos pela aprovação em bloco (uso normal das telas
principais)** — AS-04 (menu do administrador só com "Sistema" — visível
de cara ao logar como Rafael); AS-05 metade de cabeçalho/rodapé (mesma
observação); SH-01, SH-04, SH-05, SH-06, SH-07, SH-08, SH-09, SH-12,
SH-13 (regiões do shell, menu expansível/recolhido, barra de progresso,
toasts, layout, rodapé/cabeçalho — tudo visível em qualquer navegação
básica); G-01 (grid não aparece antes de pesquisar); G-19 metade
"aparece" (mas não o truncamento "+N", ver abaixo).

**Exigem ação específica, dado específico ou duas sessões — a aprovação
em bloco muito provavelmente NÃO os cobriu, e eu não afirmo que cobriu:**

| Cenário | Por que não é plausível ter sido validado só "olhando as telas" |
|---|---|
| `SE-03`, `SE-09` | Exigem duas sessões simultâneas (um usuário logado enquanto outro muda os perfis dele) |
| `G-19` (truncamento "+N") | Exige um usuário com 3+ perfis — o seed não tem nenhum assim por padrão |
| `U-05`, `U-08` (duplo clique) | Ação deliberada de teste, não fluxo de uso |
| `E-06`, `E-07` (texto exato do diálogo de confirmação; loading isolado só na linha certa) | Detalhe fino de comportamento, não algo que se nota sem procurar |
| `E-02`, `E-13` (auditoria de campo alterado; dirty state do grupo de checkboxes) | Não é visível na tela sem inspecionar o banco (E-02) ou tentar fechar com dado sujo de propósito (E-13) |
| `L-10` (toast específico ao tentar URL sem sessão) | Exige digitar a URL deliberadamente sem estar logado |
| `S-01` (navegar para outra tela durante senha provisória) | Ação deliberada de burlar o fluxo |
| `S-04` (mensagem exata de confirmação divergente) | Exige digitar duas senhas diferentes de propósito |
| `SH-10` (responsividade em 360px/768px) | Exige redimensionar a janela ou usar DevTools — não é o uso padrão em desktop |
| `SH-11` (navegação só por teclado) | Exige testar sem mouse, deliberadamente |
| `SH-03` (elipse do breadcrumb abaixo de 360px) | Mesma razão de `SH-10` |
| `AS-01` (fluxo encadeado completo: instituição → modal de PI → combo público) | É um fluxo de várias etapas em sequência; plausível que só partes tenham sido vistas |
| `I-03`, `I-06`, `I-07`, `I-10` | Mensagens exatas, ordenação específica, formato de campo — precisam de tentativa deliberada de errar |
| `I-08` (metade "destacada no grid, com texto de apoio") | Exige uma instituição sem PI no ambiente que o dono usou — não é o caminho comum |

Não decido se isso é suficiente ou não para produção — é do dono. Só
registro que, se a aceitação por §15.3(b) depender de confirmação por
cenário, a lista acima é a que falta perguntar a ele.

---

### 6. O que ficou adiado e onde está registrado

Todo o conteúdo de `B` (43 cenários) está detalhado, com motivo e camada,
em `specs/autenticacao-usuarios/testes-pendentes.md`. Migra para suíte
E2E na fase de Release, conforme spec §13 e §15.2.

**Os testes E2E que já existem excedem o que a spec pedia** (E2E
antecipado era "não" — §15) e devem entrar na lista de fluxos já prontos
da fase de Release, para não serem reescritos: `login-logout.spec.ts`
(L-01, L-17, SE-01, parte de L-09), `mesmo-email-duas-instituicoes.spec.ts`
(L-12, L-13), `primeiro-acesso.spec.ts` (S-01 parcial, S-02),
`crud.spec.ts` (U-01, E-01, parte de E-06/E-07), `isolamento.spec.ts`
(isolamento de tela, complementa T-01 a T-06), `abre-sem-quebrar.spec.ts`
(smoke geral do shell).

**Ressalva registrada nesta rodada:** a suíte E2E, hoje, só passa 100%
contra banco `basisavalia_e2e` virgem — execuções repetidas sem recriar o
banco acumulam dado residual (senha deixa de ser provisória, usuários de
teste já existem) e falham. O `dev-fullstack` está corrigindo os
fixtures para restaurarem o estado ao final de cada teste; não verifiquei
essa correção nesta rodada porque ainda está em andamento. **Isto não
afeta o veredito da seção 3** (a suíte passou com banco limpo, que é a
evidência que uso), mas é informação que o dono precisa ter antes de
decidir sobre produção: hoje, rodar a suíte E2E duas vezes seguidas sem
resetar o banco produz falso-negativo.

---

## Achados de qualidade (T-078 · `code-reviewer`)

> **Nota de processo:** esta execução do `code-reviewer` não teve acesso a
> ferramenta de shell/container (`Bash`) — apenas `Read`/`Grep`/`Glob`/`Write`.
> Todos os achados abaixo vêm de **leitura estática do código-fonte**, não de
> execução real de `go test ./...`, `npx knip`, `npx depcruise` nem
> `node examples/quality/validar-mermaid.mjs`. Onde a verificação exigiria
> rodar algo, isso está marcado explicitamente como "não executado — só lido".
> Recomendo que alguém com acesso a shell rode essas quatro checagens antes de
> fechar T-078, especialmente `go test ./internal/port/...` (os três
> testes-guarda) e `npx knip`/`npx depcruise` no frontend.
>
> Também: o pedido original desta tarefa foi para escrever em
> `specs/autenticacao-usuarios/code-review.md`. Meu papel fixo só permite
> escrever na seção "Achados de qualidade" de `evidence.md` — nenhum outro
> arquivo do projeto. Registrei aqui; quem orquestra pode copiar o conteúdo
> para `code-review.md` se for esse o arquivo que o fluxo T-077/T-078/T-079
> espera.

### Resumo por severidade

| Severidade | Qtde |
|---|---|
| 🔴 Crítico | 2 (grupos de achados) |
| 🟡 Alto/Moderado | 1 |
| 🟢 Baixo/Menor | 3 |

---

### 🔴 Crítico — comentários no código do frontend (bundle)

CLAUDE.md, seção "Comentários que chegam ao usuário final — proibidos":
*"Frontend (Next.js, Angular): zero comentários no código que vai para o
bundle... Comentário proibido encontrado = achado Crítico (não Baixo nem
Médio)."* `design.md` §15, item 10, repete a mesma regra como restrição
inegociável ("Zero comentários em `.ts`/`.tsx`/`.css`").

**a) Comentários explicativos genuínos (4 ocorrências, 4 arquivos)** —
não são diretiva de ferramenta, são prosa explicando decisão/comportamento:

- `frontend/src/features/auth/components/login-form.tsx:35` —
  `// valor corrompido — ignora`
- `frontend/src/lib/api-client.ts:40-43` — bloco de 4 linhas explicando por
  que o interceptor de 401 não se aplica ao login
- `frontend/src/features/auth/hooks/use-guarda-senha-provisoria.ts:5-8` —
  bloco de 4 linhas de cabeçalho explicando a estratégia de cancelar
  navegação (inclusive cita `design.md §2.3` dentro do comentário — exatamente
  o tipo de contexto que o princípio 6 do CLAUDE.md manda manter em
  `spec.md`, não no código)
- `frontend/src/components/shared/hooks/use-local-storage.ts:16` —
  `// valor corrompido — mantém o inicial`

Nenhum desses vaza segredo ou rota interna sensível, mas a regra do
CLAUDE.md não tem limiar de sensibilidade — qualquer comentário em código
que vai para o bundle é Crítico por definição.

**b) Comentários `// eslint-disable-next-line react-hooks/exhaustive-deps`
(9 ocorrências, 7 arquivos)** — mesma regra, natureza diferente: são
diretiva de ferramenta, não prosa. Ocorrem em:
`frontend/src/features/usuario/components/usuario-form-modal.tsx:120`,
`frontend/src/features/instituicao/components/instituicao-form-modal.tsx:79`,
`frontend/src/features/auth/hooks/use-sincronizacao-de-perfis.ts:42,57`,
`frontend/src/components/shared/hooks/use-local-storage.ts:20,38`,
`frontend/src/app/app/usuarios/page.tsx:63,70`,
`frontend/src/app/app/instituicoes/page.tsx:69,81`,
`frontend/src/app/app/administradores/page.tsx:69,76`,
`frontend/src/app/app/instituicoes/[id]/pesquisadores/page.tsx:42`.

O débito de lint subjacente (`useExhaustiveDependencies` em 9 arquivos)
está registrado e justificado em `testes-pendentes.md` ("Grupo 11 —
Fechamento"), com gatilho de resolução (`react-query`) já aprovado em
`design.md` §11.5 — **isso eu não repito como achado**. O que não vejo
registrado em nenhum dos dois documentos é a tensão específica com a regra
"zero comentários": o ESLint não tem hoje outra forma de suprimir uma linha
específica sem um comentário inline, e a regra do CLAUDE.md não abre
exceção para diretiva de ferramenta. É uma divergência real entre duas
regras do próprio CLAUDE.md (qualidade de lint vs. zero-comentário), não
decidida em nenhum dos dois documentos — cabe ao arquiteto decidir (ex.:
suprimir a regra por configuração do ESLint para esses hooks específicos
em vez de comentário por linha, ou registrar exceção explícita).

**Divergência que NÃO reporto como achado:** os comentários de agrupamento
dentro de `frontend/src/components/ui/drawer.tsx:110-132` (`// Base.`,
`// Nested.`, `// Bleed.` etc.). É componente `shadcn`/`base-ui` vendorizado
(gerado pela CLI, não escrito à mão pelo time) — mencionado aqui só para
registro, severidade 🟢 Baixo/informativo: se a regra "zero comentários" se
aplica também a código vendorizado copiado pela CLI do shadcn, isso é
decisão do arquiteto (a alternativa seria patchear todo componente shadcn
que a CLI gerar com comentário, o que é atrito recorrente a cada
`shadcn add`).

---

### 🟡 Moderado — botão "Editar" da linha do grid sem `LoadingButton`

CLAUDE.md, "UI assíncrona por padrão": *"Botão 'Editar' também dispara
loading: mesmo que a edição abra um modal local (sem chamada à API), o
clique deve mostrar feedback imediato enquanto o modal carrega os dados."*

- `frontend/src/features/usuario/components/usuario-table.tsx:110-117` —
  botão "Editar" é `Button` comum, sem `loading`/`idEditando` por linha
  (o componente só rastreia `idExcluindo` para o botão Excluir).
- `frontend/src/features/instituicao/components/instituicao-table.tsx` —
  mesma forma (confirmado por `ux.md:674`, que já especifica o botão Editar
  do grid de instituições como `Button variant="ghost" size="icon"`, sem
  menção a `LoadingButton`).

Não encontrei em `ux.md` nem em `design.md` uma decisão registrada que
dispense o botão Editar dessa regra (diferente de outras divergências desta
feature, que estão explicitamente registradas — ex.: mensagem genérica de
login, 404 em vez de 403, ausência de limite de tentativas). Como a edição
aqui é só abrir um modal com dado já em memória (sem fetch adicional), o
ganho prático de um spinner é pequeno, mas a regra do CLAUDE.md não
condiciona a exigência a "haver fetch" — condiciona a "iniciar uma operação
que o usuário percebe como processamento". Reporto como achado, não decido
a solução (poderia ser adicionar `idAbrindo`/loading local de ~100ms, ou o
arquiteto registrar exceção explícita em `ux.md` por já não haver chamada
de rede).

---

### 🟢 Baixo — descrição de metadados desatualizada (resíduo do modelo anterior)

`frontend/src/app/layout.tsx:20` —
`description: "Cadastrador e assinador de atas apoiado nos instrumentos de avaliação do INEP."`

`specs/00-visao-produto.md` (versão atual, aprovada) descreve o produto como
apoio ao **acompanhamento das metas da coordenação** — não há mais conceito
de "atas" nem de "assinatura de atas" na visão vigente (a feature de atas
foi removida, conforme contexto desta rodada de revisão). Este texto de
`<meta name="description">` é visível a qualquer visitante (view-source,
aba do navegador, mecanismos de busca) e descreve um produto que não é mais
este. Não é falha de arquitetura nem de segurança — é resíduo textual de um
pivô de produto anterior, correção de baixo custo.

---

### 🟢 Baixo — comentário do `escopo_sql.go` reivindica exclusividade que não é exata

`backend/internal/adapter/postgres/escopo_sql.go:14-16` afirma: *"Nenhum
outro arquivo deste pacote escreve `instituicao_id` numa cláusula WHERE —
só aqui, nos INSERT e nas duas portas estreitas (AutenticacaoRepository,
InstituicaoPublicaQuery)."*

Isso não é exato: `ContarDetentoresDoPerfil`
(`backend/internal/adapter/postgres/usuario_repository.go:393-413`) também
monta `WHERE up.instituicao_id = $1 ...` / `WHERE up.instituicao_id IS NULL
...` manualmente, fora de `AplicarEscopo`, fora de um INSERT e fora das duas
portas estreitas citadas. **Isto não é violação do item (a) do checklist**
("nenhuma query montando o filtro de isolamento à mão — só `aplicarEscopo`")
porque essa consulta específica está desenhada e sancionada literalmente em
`design.md` §5.8 (contagem de invariante de último detentor, SQL exibido
ali linha a linha) — não é filtro de isolamento de leitura de recurso, é
contagem de invariante de concorrência, categoria diferente que o próprio
design separa. O achado é só que o comentário-fonte de `escopo_sql.go`
overclaima exclusividade e pode confundir quem procurar "o único lugar que
escreve instituicao_id em WHERE" no futuro. Sugiro (não decido) atualizar o
comentário para citar `ContarDetentoresDoPerfil` como terceira exceção
documentada.

---

## O que foi verificado e está conforme (sem achado)

Verificação estática, ponto a ponto do checklist de T-078 (`tasks.md`
linhas 74-87) e das garantias estruturais do design (`design.md` §15):

- **(a)** Filtro de isolamento centralizado em `AplicarEscopo`
  (`escopo_sql.go`) — todas as consultas de `usuario_repository.go`,
  `instituicao_repository.go` que retornam recurso ao cliente usam a
  função; única exceção documentada tratada acima (Baixo).
- **(b)** Os três testes-guarda existem em
  `backend/internal/port/portas_test.go`
  (`TestRepositoriosDeNegocio_TodoMetodoExigeEscopo`,
  `TestPortasEstreitas_ListaFechada`, `TestNenhumaPortaNovaSemEscopo`) e
  fazem exatamente o que `design.md` §4.4 descreve — **confirmado passando
  na execução real de `go test ./...` feita pelo `qa-tester` em T-079**
  (seção acima), inclusive os três testes-guarda deste pacote.
- **(c)** `rotas.Publica` tem exatamente 1 ocorrência
  (`cmd/api/main.go:132`); `AutenticadaSemPermissao` tem exatamente 3
  (`cmd/api/main.go:133-135`) — dentro do limite de 4.
  `POST /auth/login` é registrada à parte, com comentário explicando por
  que fica fora das duas contagens (`main.go:137-142`) — consistente com
  `design.md`.
- **(d)** Nenhuma linha do backend ou do frontend acrescenta `aluno` a um
  conjunto não vazio. `ConjuntoInstitucional` (`conjunto_de_perfis.go`) só
  atribui `{Aluno}` quando a entrada está vazia; `usuario-form-modal.tsx`
  só seta `"aluno"` via `handlePerfisChange` quando `novoValor.length === 0`.
- **(e)** Toda regra de conjunto de perfis está contida em
  `ConjuntoDePerfis`/`ConjuntoInstitucional`/`ConjuntoDeAdministrador`
  (`domain/valueobject/conjunto_de_perfis.go`). Os use cases
  (`criar_usuario.go`, `atualizar_usuario.go`) só chamam esses construtores
  e nunca revalidam/coagem por conta própria.
- **(f)** `atualizar_usuario.go:145-151` registra `perfis_anterior` e
  `perfis_novo` completos, em ordem canônica, nunca a diferença — confirma
  design §8.1/E-02.
- **(g)** `docker-compose.dev.yml`, serviços `test` (linhas 83-93) e
  `playwright` (linhas 312-322), ambos começam com `set -e` antes de
  encadear migration/build/teste.
- `CHECK (perfil = 'administrador_sistema') = (instituicao_id IS NULL)` en
  `usuario_perfil` presente na migration consolidada
  (`000002_autenticacao_usuarios.up.sql:112-114`), igual ao design.
- `ConjuntoInstitucional`/`NovoConjunto` recusam a combinação de
  administrador com outro perfil (dupla camada, domínio + banco).
- Hexagonal: nenhum import de `internal/adapter` em `internal/domain` ou
  `internal/usecase` (fora de arquivos `_test.go`, que legitimamente usam
  adapters reais para setup de teste de integração).
- CQRS-leve: `usecase/command/{sessao,usuario,instituicao}` e
  `usecase/query/{sessao,usuario,instituicao}` — uma subpasta por entidade,
  como exigido.
- Campos base, deleção lógica, UUIDv7 gerado em `usuario.NovoUsuario`
  (nunca no banco), concorrência otimista (`versao` + `UPDATE ... AND
  versao = $n` + 409 `CONFLITO_DE_VERSAO`, tratado no frontend com tela
  própria de conflito) — conforme.
- Paginação: envelope `{data, meta}`, allowlist de `sort` por handler
  (`ordenacaoUsuarioAllowlist`), valor fora da allowlist responde 400
  `PARAMETRO_INVALIDO` sem clamping (`paginacao.go`), `perfil` de fato
  ausente da allowlist de ordenação de usuários.
- Fixtures de teste (`testhelpers/fixtures.go`) registram a própria
  limpeza via `t.Cleanup`, incluindo a ordem de FK (perfil → usuário →
  auditoria) — conforme a regra inegociável do CLAUDE.md.
- `shared/ui` e `shared/forms`/`shared/hooks` não importam de `features/`
  (confirmado por grep, zero ocorrências); `components/layout/` importa de
  `features/auth` para tipos/hooks de sessão, o que é aceitável pois
  `layout/` é uma categoria à parte de `shared/` (documentado em `ux.md`
  §"Componentes shadcn/ui", não sujeita à mesma regra de isolamento).
  `.dependency-cruiser.js` tem a regra
  `shared-nao-importa-features` configurada corretamente — **não executei
  `npx depcruise`** para confirmar dinamicamente.
- `LoadingButton` presente em `shared/ui/loading-button.tsx`, usado em
  filtros, formulários e no botão Excluir por linha com rastreio de
  `idExcluindo` por item (não boolean global) — conforme, exceto o gap do
  botão Editar já reportado acima.
- Ferramentas de qualidade JS/TS presentes no repo: `biome.json`,
  `knip.config.ts`, `.dependency-cruiser.js`, `commitlint.config.js` (raiz),
  `.husky/` — presença confirmada por leitura de arquivo; **não executei
  `npx knip`** para confirmar ausência real de código morto.
- Débito de lint do Grupo 11 (`react-hooks/set-state-in-effect`,
  `useExhaustiveDependencies`, `lint/a11y/*`): decisão está registrada em
  `testes-pendentes.md` com números específicos, causa raiz e gatilho de
  resolução (`react-query`) — não é número que "sumiu" com a troca de
  ferramenta; ESLint e Biome contam achados separados e ambos apontam para
  a mesma causa. Único ponto não coberto por essa decisão é a tensão com
  "zero comentários" já relatada acima.
- Erros de domínio traduzidos centralmente em `middleware_erro.go`, 500
  sempre genérico, nunca stack trace/SQL/tabela — conforme design §6.3.
- Dual write: auditoria local sempre dentro da transação do comando
  (`uc.uow.Executar`), syslog agendado via `postgres.AgendarAposCommit`
  (fire-and-forget pós-commit, nunca durante a transação) — sem dual write,
  conforme D-09 e a saída de escape já documentada no CLAUDE.md para canal
  não-autoritativo.
- Ordem de avaliação 403-antes-de-404 confirmada:
  `middleware_autorizacao.go` responde 403 `PERMISSAO_NEGADA` antes de
  qualquer consulta ao repositório; o 404 `NAO_ENCONTRADO` só surge depois,
  dentro do use case, quando `AplicarEscopo` não encontra a linha.

## Divergências que PARECEM defeito e não são (confirmei contra o design antes de não reportar)

- Mensagem de login genérica, ausência de limite de tentativas, 404 em vez
  de 403 para recurso de outra instituição, menu oculto em vez de
  desabilitado para grupos inteiros, teste de desempenho que verifica
  propriedade (ausência de `Seq Scan`) em vez de nomear o índice — todos
  conferidos linha a linha contra `spec.md`/`design.md` e implementados
  exatamente como as decisões registradas descrevem. Nenhum é achado.
