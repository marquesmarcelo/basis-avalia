# Testes pendentes — plano-acao

Escritos: os testes de mecanismo (integridade e isolamento) listados em
`design.md` §10.1 — VOs (`Aprovacao`, `Quantidade`), `Plano.SituacaoEfetiva`/
`SemAprovacao`/ciclo de vida, cópia em lote (CP-02, CP-05, CP-07, CP-11,
CP-12), `PL-02` com corrida real (duas transações concorrentes contra o
índice único), isolamento e carteira do coordenador (SI-01, VI-01), o
gerador de `.docx` (zero marcador vazado, clonagem de linha, escape de
XML/quebra de linha), e o teste que fixa P-11 (coordenador gera documento
sem causar 409 no PI).

Este arquivo lista o que ficou fora, seguindo a mesma lista de `design.md`
§10.2, com uma ressalva importante abaixo.

## Ressalva: `entrega` ainda não existe

`specs/metas-coordenacao/` não foi implementada nesta sessão. Toda checagem
que depende da tabela `entrega` (`PLANO_COM_ENTREGA`, `ITEM_COM_ENTREGA`,
o campo `tem_entrega` na API) está **implementada no formato certo** mas
**stubada para sempre `false`** — mesmo padrão que `port/meta_repository.go`
já usou para `Planos: 0` antes de `item_plano` existir:

- `PlanoRepository.ExisteEntregaNoPlano` / `ExisteEntregaNoItem` — sempre
  `false`.
- `LinhaPlano.TemEntrega` / `ItemDoPlanoResponse.TemEntrega` — sempre
  `false` na resposta da API.
- O `SELECT ... FOR UPDATE` antes do `EXISTS` de entrega (design.md §5.1),
  que serializa a corrida entre despublicar/excluir e um registro de
  entrega concorrente, **não foi implementado** — não há nada para
  serializar ainda, e adicionar o lock agora sem o `EXISTS` real por trás
  seria teatro.

**Quando `metas-coordenacao` criar a tabela `entrega`:** os dois métodos
acima precisam de uma reescrita real (não é só tirar o stub — é a query
`EXISTS` de verdade, com `AplicarEscopo(AlvoEntrega, ...)`), e os
seguintes cenários passam a ser testáveis e devem ser adicionados:

| Cenário | O que valida |
|---|---|
| `PP-3` / `SI-10` (metade que falta) | Despublicar com 1 entrega registrada → 409 `PLANO_COM_ENTREGA` |
| `IT-10` | Remover item com 2 entregas → 409 `ITEM_COM_ENTREGA` |
| `SI-14` (metade que falta) | Excluir plano com entrega → 409 `PLANO_COM_ENTREGA` |
| corrida de despublicar × entrega concorrente | `SELECT ... FOR UPDATE` na linha do plano antes do `EXISTS`, com escritas concorrentes reais |
| `tem_entrega` no grid e no item | Habilita/desabilita `[⏸]` e `[✗]` corretamente na tela |

## Adiados por decisão de escopo/tempo (não bloqueiam a feature)

Unitários com o relógio injetado, cobertos apenas parcialmente por
integração:

- `PE-01` (cadastro simples de período) — coberto indiretamente pelo teste
  de integração de `PL-02`, que cria período; sem teste dedicado do
  happy-path isolado.
- `PL-01` (cadastro simples de plano) — idem, coberto indiretamente.
- `IT-08` (acrescentar item a plano vigente, com o aviso de que o
  coordenador verá na próxima visita — o aviso em si é comportamento de
  tela, `T-238`, não testado no frontend).

Filtros, ordenação e paginação de grid — sem teste automatizado (unitário
nem E2E) nesta rodada:

- `PE-02` (nome duplicado por instituição, índice único) — a constraint
  existe e é exercida manualmente; falta teste de integração dedicado.
- `PL-03` a `PL-06` (aprovação incompleta, órgão fora da lista, curso
  inativo, alinhamento livre) — cobertos pelas regras de domínio
  (`Aprovacao`, `OrgaoDeAprovacao`) com teste unitário do VO; falta o teste
  de integração ponta a ponta via HTTP.
- `IT-04`, `IT-06` (quantidade inválida, meta inativa) — cobertos pelo VO
  `Quantidade` e pela checagem de situação no use case; falta teste de
  integração via HTTP.
- `SI-12` (motivo obrigatório) — coberto no domínio (`Plano.Encerrar`);
  falta integração via HTTP.
- `CP-01` (cópia com múltiplos cursos, comparação de igualdade completa
  dos itens copiados) — coberto por unit test com mock; falta integração
  com banco real comparando os itens linha a linha.

Smoke de rota (401/403/404 básico por endpoint) — não escrito nesta
sessão para as ~20 rotas novas. Recomenda-se seguir o padrão de
`smoke_grupos6a9_test.go` / `smoke_indicadores_metas_test.go` já existentes
no projeto.

Frontend — sem teste automatizado (unitário de componente nem E2E) nesta
sessão além do `tsc --noEmit`, `next build` e `biome check` limpos:

- Estados de loading/erro/vazio dos grids de período e plano.
- Dirty state e `beforeunload` nos formulários (`Novo Plano`, item do
  plano) — implementado seguindo o padrão de `useDirtyState`, mas sem
  teste de componente.
- Atalhos de teclado (`Ctrl+N`, `Ctrl+S`, `Ctrl+F`) — herdados do shell
  global (`use-atalhos-de-teclado.ts`), não testados especificamente
  nestas telas novas.

## E2E recomendado — não escrito nesta sessão

`design.md` §10.3 recomenda `e2e/copia-em-lote.spec.ts` (o fluxo de maior
risco da feature) e um acréscimo para o coordenador gerando documento de
plano em rascunho. Nenhum dos dois foi escrito nesta sessão — ficam para a
fase de Release (CLAUDE.md, "Camada 2"), a menos que o arquiteto marque
"E2E antecipado" explicitamente.

## Simplificações registradas (não são gaps de teste, são decisões de escopo)

- **"Abrir esse plano" no erro `PLANO_DUPLICADO`** (`ux.md`, tela Novo
  Plano): o backend não devolve o `id` do plano conflitante no corpo do
  erro 409 — implementar essa afordância exigiria estender
  `domain.ErrPlanoDuplicado` para carregar o id e uma consulta adicional
  no caminho de erro. A tela mostra a mensagem de conflito sem o atalho.
- **`/app/meus-planos` como rota própria**: não foi criada. O coordenador
  usa as mesmas páginas `/app/planos` e `/app/planos/{id}`, que escolhem
  em runtime (via `useEu().permissoes`) se chamam `/api/v1/planos` ou
  `/api/v1/meus-planos` — é a leitura literal de `ux.md` ("o Coordenador
  chega à mesma rota").
