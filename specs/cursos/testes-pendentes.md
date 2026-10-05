# Testes pendentes — cursos

Formato de cada linha: **código do cenário · o que verificar · camada · por
que foi adiado.**

## Grupo 4 — Backend de curso e designação

- **V-5 (planner sob volume real)** · a consulta de sessão de fato usa
  `idx_designacao_coordenador` em vez de `Seq Scan` · integração
  (`EXPLAIN ANALYZE`) · mesmo achado já registrado para `V-3` em
  `specs/autenticacao-usuarios/testes-pendentes.md`: com a massa de um
  teste de integração (poucas linhas em `designacao`), o planner do
  Postgres escolhe corretamente `Seq Scan` por ser mais barato — um teste
  que afirma uso de índice nesse volume falha de forma intermitente
  (depende de quantas linhas outros testes da mesma suíte já inseriram) e
  passaria em produção. `TestAutenticacaoRepository_V5_
  IndiceDeCoordenadorExiste` verifica hoje, via `pg_indexes`, que a
  migration criou o índice correto — não que o planner o escolhe.
  Reavaliar quando houver volume representativo de designações em
  ambiente de teste.

## Grupo 6/7 — Frontend e E2E (T-191)

- **CU-12 (`plano_do_periodo` na linha de curso) — RESOLVIDO (T-115),
  ponta a ponta.** Decisão do arquiteto: compor no use case de consulta
  (`ListarCursosUseCase`), em 1 (cursos) + 1 (período aberto) + 1 (planos
  do período) consultas, injetando `port.PeriodoRepository` e
  `port.PlanoRepository` — nenhuma das duas mudou de contrato, e
  `CursoRepository` continua sem conhecer `plano`. Backend testado
  (`TestSmoke_Cursos_T115_PlanoDoPeriodoAbertoNaListagem`);
  `CursoResponse.plano_do_periodo` preenchido só em `GET /cursos`, nulo em
  respostas de curso único. Frontend: `Curso.plano_do_periodo` e
  `curso-table.tsx` renderizando o texto (ou "—"). Não coberto por E2E
  (o cenário exigiria período+plano vigentes no seed, fora do escopo do
  E2E existente) — se um E2E futuro cobrir o grid de cursos com plano,
  vale exercitar esta coluna também.
- **T-188 (aviso de perfil, texto completo)** · "encerrada por edição" vs.
  "encerrada por vencimento" com portaria e data · frontend
  (`use-sincronizacao-de-perfis.ts`) · o hook já distingue ganho/perda de
  `coordenador_curso` e nomeia o curso (e a portaria, no ganho, via uma
  consulta extra a `/cursos/{id}/designacoes?coordenador_id=`), mas não
  distingue as duas causas de encerramento — o frontend não tem como
  saber sem um sinal novo do backend (nenhum campo hoje diz "por que"
  encerrou). O redirecionamento de dentro de `/app/minhas-metas` ao
  perder o perfil **já está escrito** (`if
  (pathname.startsWith("/app/minhas-metas")) router.replace("/app")`) —
  a rota nasceu depois, em `metas-coordenacao`, então isto nunca foi
  exercitado em E2E aqui; vale um teste dedicado cobrindo os dois
  juntos quando alguém mexer nessa página.
- **Filtro "Coordenador" da tela de designações reaproveita o combobox de
  candidatos** · `ux.md` pede "só pessoas que já têm designação **neste**
  curso" · frontend (`designacao-filtro.tsx`) · não existe endpoint
  dedicado para "quem já foi designado neste curso especificamente" — o
  filtro usa `GET /designacoes/candidatos` (qualquer elegível na
  instituição), uma lista mais ampla que a pedida. Funcionalmente correto
  (filtrar por alguém sem designação no curso só devolve grid vazio, não
  erro), mas menos preciso que o texto do ux.md. Criar o endpoint dedicado
  se a diferença incomodar em uso real.

## T-122 — composição do Biome (número medido, não repetido de memória)

O número que circulava (~148) não estava registrado em lugar nenhum —
por isso não compunha. Esta seção registra a composição real, medida em
`frontend/`, para o próximo ciclo poder comparar em vez de repetir um
boato.

**Antes de qualquer mudança:** `npx @biomejs/biome ci .` → **168 erros**
(0 avisos, 1 info) no estado do frontend em 2026-09-29 (o número de hoje,
163, foi medido pela revisão em um commit ligeiramente anterior — a
diferença de 5 é o acréscimo normal do trabalho em andamento nos dois
`dev-fullstack`, não uma divergência de método).

**Decisão aplicada:** exclusão de `src/components/ui` (biblioteca
vendorizada, já decidida em ciclo anterior) via `files.includes` em
`biome.json` — sintaxe de pasta sem `/**` (Biome ≥ 2.2; com `/**` o
próprio Biome acusa `lint/suspicious/useBiomeIgnoreFolder`).

| Etapa | Erros | Delta |
|---|---|---|
| Antes (sem exclusão) | 168 | — |
| Depois de excluir `src/components/ui` | 160 | −8 |
| Depois de corrigir os 4 mecânicos (`format` ×3, `organizeImports` ×1, `biome check --write`, sem `--unsafe`) | **156** | −4 |

**Composição final (156), por regra — em código autoral a meta é zero;
cada classe abaixo tem decisão explícita:**

| Regra | Qtde | Decisão |
|---|---|---|
| `lint/correctness/useExhaustiveDependencies` | 149 | **Tolerado, com motivo registrado.** Todo `useEffect` de busca-ao-montar do projeto usa `// eslint-disable-next-line react-hooks/exhaustive-deps` — mas esse comentário suprime a regra do plugin `eslint-plugin-react-hooks`, não a regra equivalente do Biome (`lint/correctness/useExhaustiveDependencies`), que é um linter diferente. É convenção **já estabelecida e pervasiva** em todo o projeto (confirmado: até `features/meta/components/meta-filtro.tsx`, código anterior a qualquer feature deste ciclo, tem a mesma ocorrência) — corrigir exigiria decidir, projeto inteiro, entre (a) trocar todos os `useEffect` para incluir a função na dependência com `useCallback` estável, ou (b) usar `// biome-ignore lint/correctness/useExhaustiveDependencies: <motivo>` no lugar do comentário de ESLint. Decisão de convenção de projeto, não de uma feature — registro aqui para o arquiteto decidir em qual ciclo. `--write --unsafe` **não é opção**: adicionar a dependência sugerida (a função de busca) sem estabilizá-la primeiro com `useCallback` cria loop de re-render em vários desses componentes. |
| `lint/a11y/noSvgWithoutTitle` | 5 | **Tolerado.** `public/*.svg` são os quatro/cinco SVG de boilerplate do `create-next-app` (`file.svg`, `globe.svg`, `next.svg`, `vercel.svg`, `window.svg`) — nunca importados por nenhum componente do projeto (confirmado por `grep -rn` em `src/`, zero ocorrências dos cinco nomes). Não são código autoral; candidatos a exclusão de `public/**` do `biome.json`, mas essa é uma decisão de escopo do ciclo anterior (só `src/components/ui` foi decidida) — registrado para o arquiteto incluir ou não na próxima. |
| `lint/a11y/useAriaPropsSupportedByRole` | 1 | **Pendente, não corrigido agora.** `src/components/layout/sidebar-nav.tsx:41` — `aria-label` num elemento cujo role não o suporta. Fora do escopo de `cursos` (é `components/layout`, compartilhado); corrigir exige entender a estrutura de foco do menu lateral para não regredir acessibilidade sem revisão. Registrado para quem tem posse de `layout/`. |
| `lint/a11y/useSemanticElements` | 1 | **Pendente, não corrigido agora.** `src/components/shared/ui/progresso-de-envio.tsx:10` — Biome sugere `<fieldset>` no lugar do role atual. Fora do escopo de `cursos`; mesmo motivo do item acima. |

**Resultado líquido desta rodada: 168 → 156** (−12: −8 de exclusão de
vendorizado, −4 de correção mecânica). As duas pendências a11y (2) e a
convenção de `useExhaustiveDependencies` (149) ficam registradas como
decisão explícita, não como número esquecido.

## T-126 — substitui T-121, com verificação no build (a regra que crescia sem checagem)

A revisão de segurança achou o quinto arquivo que T-121 não pegou:
`lib/formato.ts` (4 comentários, em `formatarDataPura`) — total real **21
comentários em 5 arquivos** (os quatro de T-121 mais este). Conteúdo
realocado, não apagado:

| Arquivo | Destino |
|---|---|
| `lib/formato.ts` (4) | `specs/_fundacao-metas.md`, nova §5.4 "O par frontend de `DataLocal`: `formatarDataPura`" — a armadilha de fuso horário (`new Date` interpretando data pura como UTC e exibindo o dia anterior ao reformatar em `America/Sao_Paulo`) é a mesma do lado do servidor (§5.1), só que sem `DataLocal` para se apoiar no cliente. |

(Os outros quatro arquivos já tinham sido corrigidos em T-121, antes desta
mensagem chegar — nenhuma mudança adicional neles.)

**Verificação nova, para a regra parar de crescer sem checagem** (a frase
do arquiteto: "regra sem verificação cresce" — esta cresceu de 4 para 21
comentários em dois ciclos, contada à mão três vezes):

- `examples/quality/validar-comentarios-frontend.mjs` — varre `.ts`/`.tsx`
  de `frontend/src`, reprova qualquer comentário fora de duas exceções:
  diretiva de lint/tipo (`eslint-disable`, `biome-ignore`, `@ts-expect-error`,
  `@ts-ignore`) e `components/ui/**` (vendorizado). Arquivos `.test.ts`/
  `.spec.ts` ficam fora do escopo — nunca entram no bundle do navegador.
- Script novo em `frontend/package.json`: `npm run verificar-comentarios`.
- Registrado na tabela de fases de CI do `CLAUDE.md` (`comentarios-fe`).

**Comprovação negativa** (as duas pontas, script direto e via `npm run`):

```
$ node examples/quality/validar-comentarios-frontend.mjs $(find frontend/src -regex '.*\.tsx\?$' ...)
253 arquivo(s) verificado(s), zero comentários.        # antes de inserir — passa

# insere "// comentário de propósito..." em features/curso/types.ts:1
$ npm run verificar-comentarios
[FALHA] src/features/curso/types.ts
   1: // comentário de propósito só para a comprovação negativa de T-126 (via npm run) — removido a seguir
1 comentário(s) em arquivo autoral de frontend — ...
EXIT=1                                                  # falha corretamente

# desfeito o comentário
$ npm run verificar-comentarios
253 arquivo(s) verificado(s), zero comentários.
EXIT=0                                                  # passa de novo
```

Estado final: os 5 arquivos com zero comentários (só a diretiva de lint,
onde já existia); `tsc --noEmit`, `next build` e `depcruise` limpos; Biome
inalterado (156 — remoção de comentário não muda contagem de lint).
