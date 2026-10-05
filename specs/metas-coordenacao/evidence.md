# Evidence — revisão de qualidade e arquitetura (code-reviewer)

> **Nota de redistribuição.** Este arquivo cobre as **quatro** features novas
> (`indicadores`, `cursos`, `plano-acao`, `metas-coordenacao`), não só
> `metas-coordenacao`. Meu papel só permite escrever em `evidence.md` de uma
> feature; escrevi aqui por instrução explícita da tarefa. Pedir ao
> orquestrador para redistribuir as seções abaixo para
> `specs/indicadores/evidence.md`, `specs/cursos/evidence.md` e
> `specs/plano-acao/evidence.md` conforme o achado.

## Aviso de método — sem acesso a terminal nesta sessão

**Eu não tive a ferramenta Bash/terminal disponível nesta sessão** — meu
conjunto de ferramentas era `Read`, `Write`, `Grep`, `Glob` e
`SubagentHandback`, sem execução de comando. Isso significa que **não
executei**:

- `go test ./...` (34 pacotes) nem `go vet`/`golangci-lint`
- `npx knip`, `npx depcruise`, `npx biome ci .`, `npx vitest`
- Nenhum E2E Playwright
- Nenhum `docker compose` / ambiente de container

**Nenhuma saída de ferramenta abaixo é real** — não existe "saída real" para
colar, contrariando o que a tarefa pediu. Tudo que segue é **leitura de
código-fonte e busca estática (`grep`/contagem manual)**, com o cuidado de
não inferir resultado de execução onde eu não tenho como comprová-lo. Onde
consegui uma **contagem estática equivalente** ao que um guarda afirma (ex.:
contar interfaces exportadas em `internal/port` por leitura), registrei isso
como tal — verificação estática, não execução do teste. Peço que outro
agente com acesso a terminal rode de fato:

```
docker compose -f docker-compose.dev.yml run --rm backend go test ./... -cover
docker compose -f docker-compose.dev.yml run --rm backend go vet ./...
docker compose -f docker-compose.dev.yml run --rm frontend npx knip
docker compose -f docker-compose.dev.yml run --rm frontend npx depcruise --validate
docker compose -f docker-compose.dev.yml run --rm frontend npx biome ci .
docker compose -f docker-compose.dev.yml run --rm frontend npx vitest run
```

Também não encontrei `.golangci.yml`/`.yaml`/`.toml` em lugar nenhum do
repositório (busquei em toda a árvore) nem workflow de CI (`.github/workflows`
não existe, só `.github/dependabot.yml`) — então "golangci-lint no backend",
mencionado na tarefa como ferramenta configurada, **não tem configuração
localizável**; não posso confirmar se está de fato configurado em algum lugar
que meu grep não alcançou, mas não encontrei o arquivo.

---

## Achados de qualidade

### 🔴 Crítico — `usecase` importando `internal/adapter/metricas` (viola a regra hexagonal e §15.1)

CLAUDE.md, regra inegociável: *"`domain` e `usecase` nunca importam/dependem
de nada de `adapter`"*; e o design de `autenticacao-usuarios` §15.1, que
`fundacao-metas.md` §12.1 explicitamente estende às quatro features:
*"`domain` e `usecase` **nunca** importam `gin`, `sqlx`, `pgx`, `prometheus`,
`jwt`, `argon2`, cliente S3 ou cliente SMTP."*

Cinco arquivos de `usecase` (produção, não teste) importam
`github.com/basis-avalia/backend/internal/adapter/metricas`, que por sua vez
importa `github.com/prometheus/client_golang/prometheus` diretamente
(`backend/internal/adapter/metricas/catalogo_plataforma.go:4`):

- `backend/internal/usecase/servico/rele/rele.go:9` — chama
  `metricas.RegistrarNotificacoesPendentes` (linha 61) e
  `metricas.RegistrarNotificacaoEnviada` (linha 96)
- `backend/internal/usecase/command/indicador_plataforma/criar_indicador_plataforma.go:6` —
  chama `metricas.RegistrarEscritaCatalogoPlataforma("criar")` (linha 59)
- `backend/internal/usecase/command/indicador_plataforma/atualizar_indicador_plataforma.go:6`
- `backend/internal/usecase/command/indicador_plataforma/alterar_situacao_indicador_plataforma.go:6`
- `backend/internal/usecase/command/indicador_plataforma/excluir_indicador_plataforma.go:6`

Todos os quatro arquivos de `indicador_plataforma` seguem o mesmo padrão do
exemplo lido (`criar_indicador_plataforma.go`): o `usecase` chama a função do
pacote `adapter/metricas` diretamente, depois do `uow.Executar`.

Não é caso isolado nem hipotético: é uma dependência de import real,
apontando para o cliente Prometheus, exatamente o que a lista de pacotes
proibidos nomeia. Compare com o único outro consumidor de
`adapter/metricas` fora dele mesmo —
`backend/internal/adapter/http/middleware_autorizacao.go` — que é
`adapter`→`adapter`, o caminho correto.

**Não encontrei nenhum guarda automatizado no lado Go** (teste de análise
estática equivalente a `TestGuardaA_PortaDeNegocioExigeEscopo`, ou regra de
lint tipo `depguard`) que impeça `usecase` de importar `adapter` — só existe
o guarda de portas (`Escopo`/`Proprio`), que é uma garantia diferente. Isso
significa que esta classe de violação **não tem mecanismo que a pegue**, nem
manual (não há `.golangci.yml` no repo) nem por teste — o mesmo padrão de
"garantia que ninguém revalida" que `specs/_status.md` acabou de registrar
como причина raiz recorrente (guardas fail-open). Não decido a solução (porta
`MetricasPublisher` em `/port`? evento de domínio consumido por um listener no
adapter?) — isso é do arquiteto.

### 🟡 Alto (achado a confirmar, não decidido como defeito) — `RemoverAnexo` não recebeu o mesmo tratamento de `Proprio` que T-114 deu a casos idênticos

`design.md` de `autenticacao-usuarios`, revisão 8, §4.4/§15.3: *"Se o método
alcança um recurso, `Proprio` não substitui `Escopo` — valem os dois"* e,
mais direto, o critério deixado pela própria revisão: *"antes de pedir uma
proveniência nova, olhe os pontos de chamada"* — porque, no ciclo T-114, três
das quatro dispensas propostas eram exatamente o padrão "identificador é o
próprio ator autenticado" e deveriam ter sido `autorizacao.Proprio` desde o
início.

`port.EntregaRepository.RemoverAnexo` (`backend/internal/port/entrega_repository.go:282`)
tem a assinatura:

```go
RemoverAnexo(ctx context.Context, escopo autorizacao.Escopo, anexoID uuid.UUID, autorID uuid.UUID) (chaveObjeto string, err error)
```

`autorID` é sempre `in.Ator.UsuarioID()` — o único ponto de chamada é
`backend/internal/usecase/command/anexo/remover.go:37`:
`uc.repo.RemoverAnexo(ctx, esc, in.AnexoID, in.Ator.UsuarioID())`. Nunca é o
id de um terceiro. E a implementação em
`backend/internal/adapter/postgres/entrega_repository.go:375-381` usa esse
valor exatamente como o `CoordenaCursoHoje` pré-T-114 usava o seu: uma
comparação (`if enviadaPor != autorID { return domain.ErrExclusaoDeEntregaAlheia }`)
que só nega, nunca concede — a mesma propriedade que a revisão 8 usou para
justificar que `cursoID` cru continua seguro em `CoordenaCursoHoje` mesmo sem
`Escopo`.

**O guarda A não pega isso** porque a checagem de tipo verificado só olha o
parâmetro na posição 1 (`escopo` já está lá) — não verifica cada
`uuid.UUID` adicional da assinatura. Isso não quer dizer que o guarda esteja
errado (ele nunca prometeu isso); quer dizer que este caso específico ficou
fora do alcance dos dois mecanismos (guarda automático e revisão manual de
T-114), pela mesma razão que os outros quatro ficaram até alguém "olhar os
pontos de chamada". Não decido se a correção é trocar `autorID uuid.UUID`
por `autorizacao.Proprio` (consistente com o padrão que T-114 estabeleceu) —
registro a inconsistência para o arquiteto avaliar contra o próprio critério
que a revisão 8 definiu.

### 🟡 Alto (a confirmar) — `DISTINCT` na consulta de "resumo" do relatório de desempenho, fora do alcance do guarda que prova a ausência dele

`design.md` de `metas-coordenacao`, M-10: *"O relatório não contém
`DISTINCT`, e um teste confere"*; §9.2: *"`DISTINCT` é proibido **nesta
consulta**, e um teste confere que a SQL gerada não o contém."*

O teste que existe, `TestMontarRelatorioBase_SemDistinct`
(`backend/internal/adapter/postgres/relatorio_repository_test.go:37-50`),
verifica só `selectCampos + from + where` devolvidos por
`montarRelatorioBase` — a consulta da **listagem** paginada.

Mas `RelatorioDesempenho` (`backend/internal/adapter/postgres/relatorio_repository.go:227-236`)
monta uma **segunda string SQL**, `consultaResumo`, concatenando os mesmos
`selectCampos + from + where` dentro de uma subconsulta e envolvendo com
`count(DISTINCT CASE WHEN coordenador_id IS NULL THEN curso_nome END)`
(linha 231) — e essa string **não passa pelo teste acima**, porque é
construída inline, dentro do próprio método `RelatorioDesempenho`, não
dentro de `montarRelatorioBase`.

Pelo que consigo ler estaticamente, este uso parece diferente do problema que
M-10 nomeia: ele deduplica **nomes de curso** (para contar quantos cursos
distintos estão sem coordenador), não deduplica linhas multiplicadas por
`aceitas`/indicadores — o risco específico que o design descreve. Pode ser
uma exceção legítima (o próprio enunciado da tarefa cita "`DISTINCT`
permitido numa consulta de períodos, onde não há agregado a mascarar" como
padrão de não-defeito neste projeto). **Mas**: (a) não está registrada em
lugar nenhum como exceção sancionada — nem em `design.md`, nem em comentário
no código apontando a diferença; e (b) o teste que a documentação cita como
"um teste confere" **não cobre esta string**. Dado que `specs/_status.md`
acabou de registrar, como decisão prioritária do projeto, que *"verificação
cuja cobertura depende de lista escrita à mão não é mecanismo"* — o mesmo
raciocínio se aplica aqui: a garantia "o relatório não contém DISTINCT" é
verdadeira para uma string e falsa para outra que o texto do design não
distingue, e ninguém validou a distinção. Não decido se o `consultaResumo`
precisa reescrever sem `DISTINCT` (ex.: `count(*) FILTER` sobre uma
subconsulta já deduplicada por curso) ou se basta estender o teste e
documentar a exceção — isso é do arquiteto.

### 🟢 Informativo/confirmado — guardas estruturais de `internal/port` (contagem estática, não execução)

Não executei `go test ./internal/port/...`. Fiz a verificação equivalente
por leitura:

- `backend/internal/port/guarda_estrutural_test.go` implementa exatamente o
  desenho descrito em `design.md` de `autenticacao-usuarios` §4.4/§15.4/§15.21:
  Guarda A varre `go/packages`+`go/types` sobre **todas** as interfaces
  exportadas do pacote, com `dispensaDeInterface` (lista de exceções, não de
  cobertura) e `dispensaDeMetodo` **vazio** (zero dispensas individuais).
- Contei manualmente as interfaces exportadas de `internal/port` via
  `grep -n "^type \w+ interface"`: **23** — bate com o número que
  `t.Logf` do próprio guarda afirmaria (`interfacesVistas`), e com o número
  citado em `specs/_status.md` ("23 portas declaradas").
- `dispensaDeInterface` tem exatamente três entradas:
  `AutenticacaoRepository`, `InstituicaoPublicaQuery`,
  `EntregaReleRepository` — as três portas estreitas, cada uma com
  proveniência estrutural documentada em §4.4 (credencial, `Proprio`
  derivado de token, identificador produzido pelo próprio processo sem
  ator). `dispensaDeMetodo = map[string]map[string]bool{}` — literalmente
  vazio no código-fonte, confirma "zero de método".
- `TestGuardaC_AdapterHTTPNaoReferenciaPortaDeManutencao` existe e varre
  `adapter/http` por AST em busca do identificador `EntregaReleRepository`.

**Isto confirma que a *forma* do guarda está correta e que a contagem bate
com o que os documentos afirmam — não confirma que `go test` realmente passa
verde.** Não posso substituir a execução por leitura; só posso dizer que não
encontrei, por leitura, nenhuma porta nova (`internal/port/*.go`) fora das 23
contadas, nem dispensa adicional não documentada.

Verifiquei também o caso citado no `_status.md` como corrigido (T-115):
`CursoRepository` (`backend/internal/port/curso_repository.go`) **não**
referencia `plano` em lugar nenhum (`grep -i plano` no arquivo não bateu). A
composição do campo "plano do período corrente" está no use case de
consulta `ListarCursosUseCase`
(`backend/internal/usecase/query/curso/listar_cursos.go`), injetando
`PeriodoRepository` e `PlanoRepository` e fazendo duas consultas adicionais
(não uma por linha) — conforme T-115 pedia.

### 🟡 Médio — inconsistência de nomenclatura/padrão, não vulnerabilidade: duplicação pequena de `offset := (page - 1) * pageSize`

`fundacao-metas.md` não fala disso diretamente, mas é o risco natural de
construção paralela que a tarefa pediu para observar. A expressão
`offset := (page - 1) * pageSize` aparece copiada, idêntica, em **11**
arquivos de `adapter/postgres` (`indicador_repository.go`,
`meta_repository.go`, `curso_repository.go`, `instituicao_repository.go`,
`usuario_repository.go`, `indicador_plataforma_repository.go`,
`entrega_fila_e_badges.go`, `periodo_repository.go`,
`designacao_repository.go`, `plano_repository.go`, `entrega_repository.go` —
mais uma variante encapsulada em `paginaEPageSize()`, só em
`entrega_repository.go:260`, que os outros dez não reusam). Severidade baixa
porque é uma linha (cabe na "escada de simplicidade" do CLAUDE.md, passo 6),
mas é exatamente o tipo de coisa que diverge na próxima correção (ex.: se o
cálculo de `offset` precisar mudar por causa de paginação por cursor no
futuro, são 11 lugares, não 1). Reporto como observação de duplicação, não
como defeito de segurança.

### 🔴 Crítico — comentários em arquivo autoral de frontend (regra de segurança do CLAUDE.md, não princípio 4)

CLAUDE.md, seção "Comentários que chegam ao usuário final — proibidos":
*"Frontend (Next.js, Angular): **zero comentários** no código que vai para o
bundle."* Achado explícito: *"Comentário proibido encontrado = achado
Crítico"*. Busquei `^\s*//` em `frontend/src/features/**`, `frontend/src/app/app/**`
e `frontend/src/components/**`; excluí do achado os comentários
`// eslint-disable-next-line ...` (débito já registrado e discutido no ciclo
anterior, tensão entre regras já levada ao arquiteto) e o componente
`ui/drawer.tsx` vendorizado (já registrado como não-achado no ciclo
anterior). Sobraram, **novos nestas quatro features**, comentários de prosa
explicando lógica de negócio:

- `frontend/src/features/plano/components/plano-aviso-aprovacao.tsx:9-11` —
  três linhas explicando quando o aviso aparece (`SI-05`/`SI-06`/`SI-07`)
- `frontend/src/features/plano/hooks/use-cursos-sugestoes.ts:8-10` — três
  linhas explicando por que reusa o endpoint de listagem de cursos
- `frontend/src/features/designacao/components/encerrar-designacao-modal.tsx:26-30` —
  cinco linhas explicando a armadilha de fuso horário ao somar um dia a uma
  data pura
- `frontend/src/components/layout/nav-config.ts:121-130` — dez linhas
  explicando `montarNav` e a semântica "qualquer uma" de `fundacao-metas.md`
  §8 (este arquivo é `layout/`, compartilhado, mas foi tocado pela mudança
  de `nav-config` desta rodada — F-13/§8 da fundação)

Todos os quatro são exatamente o tipo de comentário que o princípio 4 do
CLAUDE.md **elogia** para código de backend ("explica o porquê, não o quê")
— o que confirma que não é desleixo, é a regra de segurança específica de
frontend sendo esquecida, não a regra geral de qualidade de comentário. A
correção, seguindo o próprio CLAUDE.md, é mover o conteúdo para `spec.md`
(ou `design.md`), não apagar o conhecimento — mas a linha tem que sair do
arquivo `.tsx`/`.ts`.

Não achei comentário de prosa (fora de `eslint-disable`) em nenhum outro
arquivo novo das quatro features além destes quatro — a maioria do código
frontend novo está limpo.

### 🟢 Baixo/informativo — débito de lint (`~148`) não localizado como número corrente

Não encontrei, em nenhum dos quatro `testes-pendentes.md` novos, em nenhum
`design.md` das quatro features, nem em `specs/_status.md`, um número
atualizado de avisos tolerados de Biome/ESLint pós-inclusão das quatro
features. O que encontrei foi o registro do ciclo anterior
(`specs/autenticacao-usuarios/testes-pendentes.md:469-509`): ESLint
`react-hooks/set-state-in-effect` 18 erros/2 avisos, Biome 49 problemas —
soma 69, não 148. Não sei se "~148" citado na tarefa é o total após somar o
débito das quatro features novas (que eu não consegui medir sem `biome ci .`)
ou uma referência a outra medição que não encontrei nos arquivos do projeto.
`specs/plano-acao/testes-pendentes.md:83` afirma "`biome check` limpos" para
a sessão de `plano-acao` — não tenho como confirmar essa afirmação sem
executar a ferramenta, e ela não está acompanhada de número (quantos
problemas havia antes/depois). **Recomendo que o `qa-tester` ou outro agente
com terminal rode `npx biome ci .` e registre o número atual**, exatamente
pelo motivo que a própria tarefa deu: "um número que sumiu do relatório" é o
padrão de falha que este projeto já identificou duas vezes.

### 🟢 Confirmado — não são defeitos (decisões verificadas contra o design antes de reportar)

- **`DISTINCT` na consulta de "minhas-metas"/carteira**: não encontrei
  `DISTINCT` fora do já discutido acima (resumo do relatório); as junções de
  indicador em `entrega_minhas_metas.go`, `plano_repository.go`,
  `entrega_repository.go` usam `EXISTS`/lateral, nunca `DISTINCT`, conforme
  M-11.
- **`GET /minhas-metas` sem botão "Pesquisar"**: confirmado em
  `frontend/src/app/app/minhas-metas/page.tsx` — a tela chama `carregar()`
  automaticamente ao montar (dentro do `useEffect` que espera
  `periodos`/`localStorage`). Isto é o comportamento **documentado** como
  exceção única: `port/entrega_repository.go:196-199` chama
  explicitamente "a única tela sem 'Pesquisar'". Não é falha do padrão CRUD
  do CLAUDE.md — é a exceção nomeada.
- **Catálogo de indicadores compartilhado entre instituições**: mecanismo
  `Alvo.admiteCatalogoComum`, verdadeiro só em `AlvoIndicador`
  (`backend/internal/adapter/postgres/alvo.go`, coerente com
  `fundacao-metas.md` §3.2/§3.3); há teste dedicado
  (`TestAlvos_ExcecaoDoCatalogoEmExatamenteUm`, não executado por mim, mas
  presente no código de `alvo_test.go`).
- **`Escopo.SemCarteira()` só usado em sub-junções, nunca no recurso de
  topo**: os quatro usos que encontrei (`entrega_repository.go:117`,
  `plano_repository.go:308`, `entrega_minhas_metas.go:19`) são todos junções
  para `AlvoIndicador` dentro de uma consulta de recurso de topo diferente
  (`meta`, `plano`, `entrega`), nunca o `Alvo` do próprio recurso da rota —
  consistente com a restrição §15.20 e com o parecer 2 do arquiteto em
  `_status.md` sobre `GET /minhas-metas/periodos` (T-113), cujo comentário
  em `port/entrega_repository.go:196-203` já reflete a correção (não cita
  mais `SemCarteira()` no recurso de topo) — T-116 (corrigir o comentário)
  parece já aplicado.
- **404 antes de exclusividade/vinculo, dual write só em
  `metas-coordenacao`**: `usecase/servico/rele/rele.go` implementa
  reconciliação sobre o estado da entrega (varre `entrega`/`anexo`
  diretamente via `EntregaReleRepository`), não uma tabela `outbox` — bate
  com a decisão registrada em `fundacao-metas.md` §10 ("Dual write: existe
  em uma só feature... Solução: reconciliação sobre o estado da entrega, não
  Outbox").
- **`FOR UPDATE ... SKIP LOCKED`** confirmado em
  `backend/internal/adapter/postgres/entrega_rele_repository.go:36,115` —
  concorrência entre réplicas resolvida no banco, não em memória, conforme
  §7 da fundação.
- **Fixtures registram a própria limpeza**: as nove fixtures novas
  (`CriarIndicadorPlataforma`, `CriarIndicadorInstituicao`, `CriarMeta`,
  `CriarCurso`, `CriarDesignacao`, `CriarPeriodo`, `CriarPlano`,
  `CriarItemPlano`, `CriarEntrega`, `CriarAnexo`) em
  `backend/internal/testhelpers/fixtures.go` chamam `t.Cleanup` cada uma
  dentro de si mesma — nenhuma delegou a limpeza para quem chama.
- **Campos base + deleção lógica + `versao`**: conferido por amostragem em
  `migrations/000004_plano_acao.up.sql` (`periodo`, `plano`, `item_plano`
  têm os quatro campos base + `versao INT DEFAULT 1`; `documento` não tem
  `versao`/`atualizado_em`, o que bate com a exceção documentada em
  `fundacao-metas.md` §10 — "imutável"). Não conferi as seis migrations
  inteiras linha a linha; é amostragem, não cobertura total.
- **`LoadingButton`**: usado em 42 arquivos das quatro features (132
  ocorrências), incluindo listagens, modais de formulário e diálogos de
  ação destrutiva — não encontrei `<Button disabled>` cru substituindo o
  padrão nas telas novas que abri.
- **`shared/` do frontend não importa de `features/`**: busquei
  `from ["']@/features` e `from ["']\.\./\.\./features` dentro de
  `frontend/src/components/shared` — zero ocorrências.
- **`domain` não importa `adapter`**: zero ocorrências de
  `internal/adapter` dentro de `internal/domain` (confirmado por grep na
  árvore inteira do pacote).

---

## O que não deu para verificar nesta sessão (peço encaminhamento)

1. **Nenhum teste foi executado.** Toda a suíte (34 pacotes Go, 13 E2E)
   precisa rodar de verdade por outro agente com terminal antes de qualquer
   decisão de merge — eu só confirmei *forma* de código, nunca
   comportamento em runtime.
2. `npx depcruise`, `npx knip`, `npx biome ci .`, `npx vitest` — nenhum
   executado; as observações acima sobre duplicação e comentários vieram de
   `grep`, que é mais lento e mais sujeito a falso-negativo que as
   ferramentas reais.
3. Não revisei as seis migrations inteiras nem todas as ~40 rotas HTTP
   linha a linha — dado o volume (343 arquivos Go, 255+ arquivos TS/TSX só
   fora de `node_modules`), priorizei os pontos que a tarefa marcou como
   prioridade (guardas, hexagonal, dual write, DISTINCT, 404/403, comentário
   de frontend) e amostrei o resto.
4. Não verifiquei os cenários "rascunho visível ao coordenador" e "404 no
   lugar de 403" ponto a ponto em cada rota — confiei na leitura de
   `fundacao-metas.md` (que documenta os dois como decisão) e no que já
   apareceu nos arquivos que abri para outros fins, sem uma varredura
   dedicada.
