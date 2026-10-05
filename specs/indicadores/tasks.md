# Tasks: indicadores

**Data:** 28/09/2026 (revisão 2) · **Lei da construção:** `design.md` desta
pasta e `specs/_fundacao-metas.md`.
**Legenda:** 🗄️ depende do `dba` · 🐳 toca arquivos do `dev-docker-compose` ·
🔒 teste de mecanismo (obrigatório em qualquer fase, nunca adiável) ·
🆕 alterado pela revisão 2

> **Esta é a primeira feature da cadeia**, e por isso carrega o **Grupo 0** —
> a fundação comum às quatro. Nada do Grupo 1 em diante compila antes dele.
>
> **Pré-requisito fora desta lista:** as 14 correções de
> `autenticacao-usuarios/design.md` §16, em especial **T-101** (`AplicarEscopo`
> emite `instituicao_id IS NULL`) e **T-102** (guardas fixam assinatura). O
> Grupo 0 é **rebase sobre elas**, nunca substituição — ver `_fundacao-metas.md`
> §3.1.
>
> **Revisão 2:** o item de menu do catálogo passa a chamar-se **"Catálogo de
> metas"** — mudou **T-140**. A rota `/app/metas` **não** muda.

---

## Ordem de execução

```
Grupo 0  — fundação (sequencial, nada depois compila sem ele)
Grupo 1  — domínio e banco de indicador e meta   ┐ podem ir em paralelo
Grupo 2  — backend: use cases, adapters, rotas   ┘ depois do Grupo 1
Grupo 3  — frontend
Grupo 4  — E2E do fluxo crítico
⛔ PARE — o dono testa no ambiente de desenvolvimento
Grupo 5  — dba consolida migrations, depois revisão e QA
```

**Sem paralelismo entre features.** A cadeia
`indicadores → cursos → plano-acao → metas-coordenacao` é única: sem
indicador não há meta, sem meta não há item de plano. Dentro de cada feature há
paralelismo, e ele está marcado.

---

## Grupo 0 — Fundação comum (sequencial)

> **Status em 2026-09-29: concluído.** T-110 a T-118 implementados, testados
> (TDD, com comprovação negativa nos 🔒) e verificados: suíte Go completa
> (`go test ./... -count=1`) verde, suíte E2E Playwright (10 testes) verde,
> `go vet` limpo, `next build` das 8 rotas sem erro, `tsc --noEmit` e
> `biome check` limpos no frontend. T-117 fica **parcialmente satisfeito**
> por desenho: as portas de indicador/meta ainda não existem (chegam em
> T-125, Grupo 1) — os três guardas de mecanismo (`TestRepositoriosDeNegocio_
> TodoMetodoExigeEscopo`, `TestPortasEstreitas_ListaFechada`,
> `TestNenhumaPortaNovaSemEscopo`) continuam intactos e verdes hoje, prontos
> para T-125 estendê-los.

| # | Tarefa | Verificação |
|---|---|---|
| **T-110** ✅ | `valueobject.DataLocal` — data pura no fuso de exibição, com `DataLocalDe`, `AntesDe`, `MaisDias`, `FimDoDia`, `String` | Testes de fronteira: 31/07 23h58 e 01/08 00h02 em `America/Sao_Paulo` com o **processo em UTC** produzem dias diferentes e corretos — `data_local_test.go`, 7 testes verdes |
| **T-111** ✅ | `Ator` e `Escopo` ganham `dataDeReferencia`; o middleware de sessão a fixa **uma vez** por requisição a partir de `Relogio` e `APP_TIMEZONE` | `TestMiddlewareSessao_T111_UmaSoLeituraDeRelogioPorRequisicao` — `relogioContador` prova exatamente 1 chamada. `Ator.ComDataDeReferencia` é *wither* imutável (não mudou a assinatura de `NovoAtor`, evitando reescrever os ~14 call sites pré-existentes que não usam data) |
| **T-112** ✅ | `Escopo` ganha `restritoACarteiraDe`; `Autorizar` é a única a preenchê-lo. Campos continuam não exportados | `TestEscopoInconstruivelForaDeAutorizar` (novo, `escopo_inconstruivel_test.go`) prova por reflection (zero campo exportado) **e** por AST (nenhuma função exportada além de `Autorizar` devolve `Escopo`) — mais forte que o texto do critério pedia |
| **T-113** 🔒 ✅ | `postgres.Alvo` — lista fechada com `alias`, `temColunaCurso`, `admiteCatalogoComum`, `admiteExigePerfil`. Por enquanto só `AlvoUsuario` e `AlvoIndicador` | `TestAlvos_ExcecaoDoCatalogoEmExatamenteUm` passa; comprovação negativa feita e desfeita (acrescentar `admiteCatalogoComum: true` a `AlvoUsuario` quebrou o teste com a mensagem exata esperada) |
| **T-114** 🔒 ✅ | `AplicarEscopo` passa a receber `Alvo`. Migra `usuario_repository` para `AlvoUsuario` **sem alterar comportamento**. Emite o ramo de plataforma para todos os alvos, o ramo do catálogo só no alvo que o declara, entre parênteses, com `instituicao_id IS NULL` **e** `escopo = 'plataforma'` | Suíte de `autenticacao-usuarios` **verde sem alteração** (7 call sites em `usuario_repository.go` migrados, saída SQL idêntica pois `AlvoUsuario.alias == "usuario"`); `TestAplicarEscopo_PlataformaExigeInstituicaoNula` continua passando; `TestAplicarEscopo_SemExcecaoForaDoIndicador` e `TestAplicarEscopo_ExcecaoDoCatalogoTemIsNullEParenteses` (novos) passam |
| **T-115** 🔒 ✅ | `AplicarEscopo` devolve `ErrEscopoInvalido` quando o `Escopo` é incompatível com o `Alvo` (carteira em alvo sem curso; `exigePerfil` fora de `usuario`) | `TestAplicarEscopo_EscopoIncompativelComAlvoFalha` passa (metade `exigePerfil`); comprovação negativa feita e desfeita. **Metade "carteira em alvo sem curso" registrada como pendente de dado real** — nenhum alcance ainda produz `Escopo` com `restritoACarteiraDe` preenchido fora de `Autorizar` (por desenho, `Escopo` é inconstruível de fora); o ramo de validação já existe em `escopo_sql.go` e ganha o segundo teste quando `CursosDaCarteira` for exercitado de verdade em `specs/cursos` |
| **T-116** ✅ | Permissões e alcances novos de `_fundacao-metas.md` §6, com a matriz atualizada | `autorizar_fundacao_test.go`: `TestAlcancesNovos_TemLinhaEmPermissaoExigida` (nenhum alcance sem linha), `TestAutorizar_AlcancesDeInstituicao_EscopoDevolvidoCampoACampo` (8 alcances), `TestAutorizar_IndicadoresDaPlataforma_EscopoDePlataformaSemPerfil`, `TestAutorizar_AlcancesDeCarteira_RestritoACarteiraDeSempreLigado` (4 alcances, `restritoACarteiraDe == &ator.usuarioID` sempre) — todos verdes |
| **T-117** 🔒 ⏳ | `TestRepositoriosDeNegocio_TodoMetodoExigeEscopo` e `TestNenhumaPortaNovaSemEscopo` estendidos aos repositórios e portas novos | **Não estendível ainda**: as portas `IndicadorRepository`/`MetaRepository` só nascem em T-125 (Grupo 1) — os testes usam lista hardcoded de `reflect.TypeOf` (não reflection automática sobre o pacote), então "estender" é literalmente impossível sem os tipos existirem. Confirmado: os três guardas continuam passando, inalterados, e `TestPortasEstreitas_ListaFechada` continua com exatamente as duas portas estreitas. Fecha de verdade quando T-125 rodar |
| **T-118** ✅ | `nav-config` do frontend: item declara `permissoes: Permissao[]` com semântica **qualquer uma**, e a montagem deduplica por `path` | `nav-config.test.ts` (4 testes, Vitest): item aparece com qualquer uma das permissões; duas permissões do mesmo item não duplicam o `path`; item some sem nenhuma permissão; grupo some sem nenhum item. `montarNav` extraída como função pura testável, reaproveita `possuiAlgumaPermissao` já existente em `lib/permissoes.ts`. `use-guarda-de-permissao.ts` também migrado para a semântica de array |

**Verificação final do Grupo 0** (2026-09-29): `go build ./... && go vet ./...`
limpos; `go test ./... -count=1` — todos os pacotes `ok`, zero falhas;
`npx tsc --noEmit`, `npx biome check` e `npx vitest run` limpos no frontend;
`npx next build` gera as 8 rotas sem erro; suíte E2E Playwright completa
(10 testes, todos os arquivos) verde contra `backend-e2e`/`frontend-e2e`.

**Por que T-113 a T-115 antes de qualquer feature:** são **mecanismo, não
cobertura**. Enquanto não entram, a propriedade central do desenho — "só existe
um lugar que monta filtro de isolamento" — vale para uma tabela e é uma
promessa para as outras nove.

---

## Grupo 1 — Domínio e banco (paralelizável entre as duas colunas)

> **Status: concluído.** Migration aplicada como `000002_indicadores`
> (único incremento sobre a `000001_autenticacao_usuarios` unificada que
> já existia no repositório — não há `000002`/`000003` anteriores a
> renumerar). `migrate up` → `down` → `up` verificado nos três bancos
> (dev, test, e2e).

| # | Tarefa | Verificação |
|---|---|---|
| **T-120** ✅ | VOs `EscopoIndicador`, `CodigoIndicador`, `ReferenciaInstrumento`, `SituacaoCatalogo`, `NomeCatalogo` | Construtores recusam valores inválidos com os códigos de `design.md` §6 — testes em `internal/domain/valueobject/*_test.go`, verdes |
| **T-121** ✅ | Entidade `Indicador` com **dois construtores** (`NovoDaPlataforma`, `NovoDaInstituicao`) e as invariantes de §3.2 | `NovoDaInstituicao` **não tem parâmetro** de referência do instrumento — verificável na assinatura; `internal/domain/indicador/indicador_test.go` verde |
| **T-122** ✅ | Entidade `Meta` com `DefinirIndicadores` devolvendo a lista anterior | Lista vazia, acima de 5 e com repetido devolvem os três erros nomeados; a anterior volta completa, nunca a diferença — `internal/domain/meta/meta_test.go` verde |
| **T-123** 🗄️ ✅ | Migration `000002_indicadores`: tabelas `indicador`, `meta`, `meta_indicador`, os quatro `CHECK`, o índice único com `NULLS NOT DISTINCT`, os índices de listagem e o inverso da associação. Com `down` funcional | `migrate up` → `down` → `up` passa nos três bancos; V-1/V-2/V-4 verificados por teste real (`indicador_coerencia_test.go`, `meta_repository_test.go`); V-3/V-5/V-6 (ausência de Seq Scan, latência) **não verificados por dba nesta rodada — pendente de revisão formal do dba antes do merge** |
| **T-124** 🗄️ ✅ | Seed de desenvolvimento: catálogo do INEP da seção 7 da spec (**sem instituição**), indicadores próprios da FSA e do IVV, metas da FSA — **incluindo a de dois indicadores, a de escopos misturados, o `1.4` próprio homônimo e o indicador inativo em uso** | Rodado duas vezes seguidas contra `basisavalia_dev` — segunda execução não duplicou nada (idempotente); 4 indicadores de plataforma (1 inativo), 4 indicadores próprios, 5 metas |

---

## Grupo 2 — Backend (depois do Grupo 1)

> **Status: concluído.** `go build ./... && go vet ./...` limpos;
> `go test ./... -count=1` — todos os pacotes `ok`. T-117 fecha aqui: os
> três guardas de mecanismo (`TestRepositoriosDeNegocio_TodoMetodoExigeEscopo`,
> `TestPortasEstreitas_ListaFechada`, `TestNenhumaPortaNovaSemEscopo`) foram
> estendidos aos três repositórios novos e continuam verdes.

| # | Tarefa | Verificação |
|---|---|---|
| **T-125** ✅ | Portas `IndicadorRepository`, `IndicadorPlataformaRepository`, `MetaRepository` — **todo método com `Escopo`** | T-117 passa com elas — `internal/port/portas_test.go` estendido e verde |
| **T-126** ✅ | Adapters de persistência, com `AlvoIndicador` e `AlvoMeta`. **Nenhum `WHERE instituicao_id` fora de `AplicarEscopo`** | `grep -rn "instituicao_id" internal/adapter/postgres/*.go` só encontra `escopo_sql.go`, os `INSERT`/lateral de contagem e a exceção sancionada já nomeada (`ContarDetentoresDoPerfil`) |
| **T-127** 🔒 ✅ | Teste de isolamento com a exceção: PI da FSA enxerga os do INEP e os dela; **nenhuma linha do IVV**; a única sem instituição é de escopo plataforma | `IV-04` — `TestIndicadorRepository_IV04_CatalogoComumVisivelATodasSemVazarInstituicao`; comprovação negativa feita e desfeita (desativar `admiteCatalogoComum` quebra o teste com a mensagem exata) |
| **T-128** 🔒 ✅ | Teste de que a exceção **não** existe em meta | `IV-05` — `TestAplicarEscopo_SemExcecaoParaMeta` (unitário) + `TestMetaRepository_IsolamentoInstitucional` (banco real) |
| **T-129** ✅ | Use cases de comando do catálogo comum (criar, atualizar, alterar situação, excluir), com auditoria na mesma transação e **syslog sempre** | `IE-01` a `IE-03`, `IE-05` a `IE-08`, `IE-10` — testes de integração em `indicador_repository_test.go`; `acoesQueVaoParaSyslogSempre` cobre as cinco ações do catálogo comum sem exceção |
| **T-130** ✅ | Use cases de comando de indicador próprio e de meta | `IN-*`, `MC-*` de comando — `internal/usecase/command/{indicador,meta}` com testes unitários (mocks) e de integração |
| **T-131** ✅ | **403 para o PI escrevendo em indicador do INEP; 404 para o Administrador pedindo indicador institucional ou meta** | `IE-04` (403) — `TestAtualizarIndicador_IE04_*`, `TestAlterarSituacaoIndicador_IE04_*`, `TestExcluirIndicador_IE04_*`; `IE-09` (404) — `TestSmoke_IndicadoresPlataforma_IE09_IndicadorInstitucionalRecebe404` |
| **T-132** ✅ | Consultas de listagem com a contagem de uso por lateral, nas **duas** variantes (total agregado × da instituição) | `IV-03` parcialmente coberto (contagem funcional provada em `TestIndicadorPlataformaRepository_IE07_ExcluirBloqueadoComUso`, total=2 de duas instituições); o cenário completo dos três atores (Maria/Renata/Rafael) fica em `testes-pendentes.md` |
| **T-133** 🔒 ✅ | Teste de que a resposta e os erros da família `/plataforma/` **não contêm identificador nem nome de instituição em campo nenhum**, inclusive na mensagem de 409 `INDICADOR_COM_META` | `IE-07`, PI-5 — `TestSmoke_IndicadoresPlataforma_CicloCompleto` verifica corpo da resposta de criação e da mensagem de 409 |
| **T-134** ✅ | Consulta de metas com indicadores agregados por lateral e filtro por `EXISTS`; **sem `DISTINCT`** | `MC-14` — `TestMetaRepository_MC14_FiltrarPorOrigemNaoDuplicaAMeta` (funcional) + `TestMetaRepository_MC14_ConsultaNuncaContemDISTINCT` (estático, código-fonte) |
| **T-135** ✅ | Bloqueio de exclusão com uso, dentro da transação, com `FOR UPDATE` na linha | `IE-07`, `IN-05` — implementados e testados contra banco real. **`MC-11` (bloqueio por uso em `item_plano`) não é implementável nesta feature** — tabela inexistente ainda; registrado em `testes-pendentes.md` como divergência, não contornado em silêncio |
| **T-136** ✅ | Handlers, rotas e DTOs **por família**, com a permissão na assinatura do registro; anotações OpenAPI escritas para quem consome | Rotas registradas em `cmd/api/main.go` com permissão obrigatória na assinatura; `docs/swagger.json` gerado via `swag init` contém as 17 rotas novas, descrições sem detalhe interno |
| **T-137** ✅ | Smoke dos endpoints das duas famílias | 401 sem sessão, 403 sem permissão, 404 fora do alcance — `internal/adapter/http/smoke_indicadores_metas_test.go`, 7 testes verdes |

---

## Grupo 3 — Frontend (depois do Grupo 2)

> **Status: concluído.** `npx tsc --noEmit`, `npx vitest run` e
> `npx next build` limpos (11 rotas, incluindo as 3 novas). `npx biome
> check` sem achados novos — os únicos avisos remanescentes
> (`useExhaustiveDependencies` nos `useEffect` de pesquisa) reproduzem
> byte a byte o padrão já aceito em `instituicoes/page.tsx` e
> `usuarios/page.tsx` (mesma contagem de ocorrências por arquivo),
> portanto não são regressão desta feature.

| # | Tarefa | Verificação |
|---|---|---|
| **T-140** 🆕 ✅ | Menu: **Indicadores** e **Catálogo de metas** no grupo **Metas** (decisão 1 do dono); "Indicadores do INEP" no grupo Sistema | Trilhas `Início → Metas → Indicadores` e `Início → Metas → Catálogo de metas`. **As rotas `/app/indicadores`, `/app/metas` e `/app/indicadores-inep` não mudam** — verificado por E2E (breadcrumb e navegação reais) |
| **T-141** ✅ | `ComboboxEntidade` ganha `badge`/`badgeVariant` **opcionais** — **uma única implementação**, reutilizada por `cursos` | Nenhum uso existente muda de comportamento (props opcionais, `tsc`/`vitest` verdes) |
| **T-142** ✅ | `ComboboxEntidadeMultipla` novo em `shared/forms/`, conforme `ux.md` — **sem** `ComboboxChips` | Cada item selecionado mostra a referência do instrumento em segunda linha; contador anunciado (`aria-live`); limite de 5 com texto |
| **T-143** ✅ | Tela `/app/indicadores-inep` — grid no padrão obrigatório, modal, contagem total com a nota de que instituição não é informação desta tela | Grid **só após "Pesquisar"**; estado persistido por tela (`grid-state:indicadores-inep`) — verificado por E2E |
| **T-144** ✅ | Tela `/app/indicadores` — coluna Origem, linhas do INEP **travadas com explicação em texto na própria linha** | Botão de excluir **escondido** quando `metas > 0`; linha travada verificada por E2E (`Somente leitura`, zero botões na linha) |
| **T-145** 🆕 ✅ | Tela `/app/metas` (**rótulo "Catálogo de metas"**) — coluna de indicadores em lista, modal com o seletor múltiplo, estado vazio de `MC-15` | `MC-15` implementado em `meta-form-modal.tsx` (bloco instrutivo + link para `/app/indicadores`); teste E2E dedicado fica em `testes-pendentes.md` |
| **T-146** ✅ | **Zero comentários** em arquivo autoral de frontend, alcance de §11.6 herdado | Verificado por grep dirigido: única ocorrência de comentário fora do padrão pré-existente (`// eslint-disable-next-line`, idêntico em todo o projeto) |

---

## Grupo 4 — E2E do fluxo crítico

> **Status: concluído.** `docker compose -f docker-compose.dev.yml run
> --rm -e RUN_TESTS=true playwright` — suíte completa **12/12** (10
> pré-existentes + 2 novos), incluindo o E2E desta feature.

| # | Tarefa | Verificação |
|---|---|---|
| **T-147** ✅ | `e2e/metas/catalogo-comum.spec.ts` conforme `design.md` §10.3 | Os passos 1 a 4 implementados e verdes. **Passo 5 adaptado**: `/app/metas` para o Administrador usa o mecanismo real do `useGuardaDePermissao` (redireciona para `/app` com toast), que não distingue 403 de 404 na UI — a distinção 403-vs-404 de IE-09 é testada no nível correto, o backend (`TestSmoke_IndicadoresPlataforma_IE09_IndicadorInstitucionalRecebe404`). Divergência registrada em `testes-pendentes.md` |

---

## Grupo 5 — Depois do teste manual do dono

| # | Tarefa | Quem |
|---|---|---|
| **T-148** ✅ | `testes-pendentes.md` com os cenários adiados de §10.2, um por linha, com prioridade | `dev-fullstack` — `specs/indicadores/testes-pendentes.md` criado, com as três divergências encontradas durante a implementação registradas na última seção |
| **T-149** 🗄️ | Consolidar as migrations da feature em **uma** antes da revisão | `dba` |
| **T-150** | Revisão de segurança + revisão de qualidade, **em paralelo entre si** | `security-reviewer`, `code-reviewer` |
| **T-151** | `evidence.md` com `C \ (A ∪ B)` e `A ∩ B` — **ambas vazias** | `qa-tester` |

---

## Discordâncias registradas

Nenhuma sobre decisão de design. Três achados durante a implementação,
todos registrados em detalhe em `specs/indicadores/testes-pendentes.md`
(seção "Divergências registradas") em vez de contornados em silêncio:

1. **Bug pré-existente corrigido**: `todasAsPermissoes` em
   `internal/usecase/query/sessao/obter_contexto_de_sessao.go` (criado no
   Grupo 0) não incluía nenhuma das permissões novas de
   `fundacao-metas.md` §6.1 — sem a correção, nenhum perfil jamais
   receberia `indicador.listar`/`meta.listar`/etc. na sessão, e o menu
   "Metas" nunca apareceria para ninguém. Corrigido, com os dois testes
   de `obter_contexto_de_sessao_test.go` atualizados para a união
   completa esperada do Pesquisador Institucional.
2. **MC-11 (bloqueio de exclusão de meta por uso em plano) não é
   implementável nesta feature** — depende de `item_plano`, que só existe
   em `specs/plano-acao`. `MetaRepository.Excluir` não bloqueia
   por uso nenhum por ora; o bloqueio real entra junto da criação de
   `item_plano`.
3. **T-147/passo 5 (E2E)**: adaptado ao mecanismo real de
   `useGuardaDePermissao` (redireciona com toast, não distingue 403 de
   404 na UI) — a distinção de IE-09 já está coberta no nível do backend.

As divergências entre documentos preexistentes (grupo de menu, rótulo
"Catálogo de metas") seguem em `design.md` §12, sem alteração — a
correção de texto cabe ao `analista-requisitos`.
