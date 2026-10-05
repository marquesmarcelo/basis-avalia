# Tasks: metas-coordenacao

**Data:** 28/09/2026 (revisão 2 — consequência de QP-3) · **Lei da
construção:** `design.md` desta pasta, `specs/_fundacao-metas.md` §7 (relê) e
§9 (infraestrutura).
**Legenda:** 🗄️ `dba` · 🐳 `dev-docker-compose` · 🔒 teste de mecanismo ·
🆕 alterado pela revisão 2

> **Última da cadeia.** Depende de `plano-acao` (item do plano, de onde
> a entrega pende), de `cursos` (carteira, designação) e de `indicadores`
> (meta e indicadores). **Mailpit entra no ambiente aqui**; o MinIO já entrou
> na feature anterior.
>
> **Revisão 2:** a resposta a QP-3 atravessou para cá. Mudaram **T-265** e
> **T-270**. **`QP-6`** (`plano-acao/design.md` §11.1) bloqueia o
> fechamento do critério de aceitação, não a construção.

---

## Ordem de execução

```
Grupo 1  — infraestrutura: Mailpit e o relê    🐳
Grupo 2  — domínio e banco
Grupo 3  — anexos: a fronteira de confiança          ← o mais arriscado
Grupo 4  — entrega e idempotência
Grupo 5  — avaliação, rodadas e desfazimento
Grupo 6  — notificação: reconciliação e restauração de prazo
Grupo 7  — relatório de desempenho e exportação       ← o mais fácil de errar
Grupo 8  — frontend
Grupo 9  — E2E
⛔ PARE — o dono testa
Grupo 10 — dba consolida, revisão, QA
```

**Grupo 3 antes do 4** porque a entrega não existe sem anexo válido, e a
validação de tipo e tamanho acontece **antes** de qualquer byte ir para o
armazenamento. Construir a entrega primeiro produziria um caminho que grava e
valida depois — exatamente o que `AN-03` proíbe.

**Grupos 6 e 7 são independentes entre si** e dependem do 5.

---

## Grupo 1 — Mailpit e o relê 🐳

| # | Tarefa | Verificação |
|---|---|---|
| **T-247** 🐳 | Serviço `mailpit` no `docker-compose.dev.yml` (1025 SMTP, 8025 UI) | Nenhuma mensagem sai para fora; a interface do Mailpit as exibe |
| **T-248** 🐳 | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_TLS` e **`RELE_HABILITADO`** no `.env.example` **sem valores** | `.env.example` versionado; `RELE_HABILITADO` documentado como o botão de desligar por configuração |
| **T-249** | `port.EmailSender` e adapter SMTP **com circuit breaker no adapter**, nunca no use case | Servidor fora do ar: o breaker abre e falha rápido; o use case não sabe que ele existe |
| **T-250** 🐳 | `/readyz` verifica o SMTP, **sem mensagem de driver** | Corpo com `"indisponivel"`; nenhum endereço, porta ou mensagem do cliente SMTP |
| **T-251** | Relê: goroutine com *ticker*, desligável por `RELE_HABILITADO`, com `SELECT ... FOR UPDATE SKIP LOCKED LIMIT n` | **`SKIP LOCKED` obrigatório**: teste com duas instâncias do relê sobre o mesmo banco envia **uma** mensagem por entrega |
| **T-252** | Métricas `notificacoes_pendentes` (gauge) e `notificacoes_enviadas_total{evento}` (counter) | Com o SMTP fora do ar, o gauge **sobe**; é o que transforma a exigência 2 da spec em alerta |

---

## Grupo 2 — Domínio e banco

| # | Tarefa | Verificação |
|---|---|---|
| **T-253** | VOs `SituacaoEntrega` (**três** valores), `RodadaDeRecusa`, `PrazoDeCorrecao`, `TipoDeAnexo`, `ResultadoDeAvaliacao` | `recusada_definitiva` **não** é valor de `SituacaoEntrega` |
| **T-254** | `Entrega.PodeCorrigir(agora)` devolvendo **o erro nomeado**, não booleano | Os três 409 saem do lugar certo: rodadas, prazo, aceita |
| **T-255** 🗄️ | Migration `000007_entregas`: `entrega`, `anexo`, `idempotencia`; as colunas de notificação; o **índice parcial de notificação pendente**; o parcial de recusadas por curso; as FK compostas de F-05 | `dba` valida V-1 a V-8 de `design.md` §4.1 |
| **T-256** | `Alvo` ganha `AlvoEntrega` e `AlvoAnexo` | `TestAlvos_ExcecaoDoCatalogoEmExatamenteUm` continua passando |
| **T-257** 🗄️ | Seed: entregas em todos os estados, **incluindo pelo menos uma avaliada por Beatriz em Biomedicina** e uma recusada com prazo expirado durante vacância | Sem a de Beatriz, `AV-15`, `RD-11` e a métrica de coincidência não são observáveis; sem a da vacância, `VG-06` não é |

---

## Grupo 3 — Anexos: a fronteira de confiança

| # | Tarefa | Verificação |
|---|---|---|
| **T-258** 🔒 | Detecção de tipo **pelo conteúdo, em dois níveis**: assinatura, e **inspeção do ZIP** quando o resultado é `application/zip` | `AN-02` com **dois** casos: `.exe` renomeado para `.pdf` (nível 1) **e um ZIP qualquer renomeado para `.docx`** (nível 2). O segundo é o que a implementação ingênua aceita |
| **T-259** 🔒 | Limites com `http.MaxBytesReader` na requisição e `io.LimitReader` de **10 MB + 1 byte** por parte; contagem e soma correntes. **Nunca `io.ReadAll`** | `AN-03`: arquivo de 12 MB → 400 e **nenhum byte no armazenamento**; 11º arquivo e soma acima de 50 MB → 400. `grep` por `io.ReadAll` no caminho de multipart não encontra nada |
| **T-260** | Hash SHA-256 por `io.TeeReader` no **mesmo passe** da transmissão; 512 bytes de detecção por `bufio.Peek`, sem consumir | Uma leitura, três usos: detecção, hash e gravação |
| **T-261** | Ordem de gravação: valida → grava objetos → transação com as linhas → remoção compensatória *best-effort* | O comentário do código diz que a remoção é *best-effort*, **nunca garantia**, e que se prefere objeto órfão a linha órfã |
| **T-262** 🔒 | `GET /anexos/{id}/conteudo`: busca a linha **com `Escopo`**, e só então usa `chave_objeto` | `AN-04` a `AN-07`: sem sessão → **401 e nenhum byte**; outra instituição → 404; curso fora da carteira → 404, nunca 403; **nenhuma resposta contém a URL do bucket**, em campo nenhum |
| **T-263** | Download auditado individualmente; **nome original, conteúdo e observação nunca em log** | `grep` no caminho de log não encontra `nome_original` |

---

## Grupo 4 — Entrega e idempotência

| # | Tarefa | Verificação |
|---|---|---|
| **T-264** | Portas e adapters de `Entrega` e `Anexo` — **todo método com `Escopo`** | Guarda estendido passa |
| **T-265** 🆕 | Use case de registrar, com as portas de `design.md` §5.1 **nesta ordem**. **Plano em rascunho responde 409 `PLANO_NAO_VIGENTE`, não 404** (M-14) | `EN-03` **reescrito**: o coordenador que acabou de ler o plano em rascunho e gerar o documento dele recebe **409**, não 404 — 404 deixaria de esconder e passaria a mentir. `EN-04`, `VG-01`, `VG-07` inalterados |
| **T-266** 🔒 | Idempotência: `INSERT` na tabela de chave **dentro da mesma transação**; violação de PK → relê a linha e devolve **200** com a entrega original | `EN-06` **com corrida real**: duas requisições simultâneas com a mesma chave produzem **uma** entrega. Sem lock, sem verificação prévia |
| **T-267** | Janela de deduplicação de **24 horas** e rotina de expurgo da tabela | `dba` valida V-7: a tabela não cresce sem limite |
| **T-268** | Correção: o ramo **passa por cima do período encerrado**, e usa `PodeCorrigir` | `AV-04`: corrigir em 03/08 com o período encerrado em 30/07 responde 200, **não** `PERIODO_ENCERRADO` |
| **T-269** | Exclusão: só o autor, só se não aceita, lógica | `EN-10`, `EN-11`: entrega alheia → **403 `EXCLUSAO_DE_ENTREGA_ALHEIA`**; aceita → 409 |
| **T-270** 🆕 | `GET /minhas-metas` com alcance de carteira, agrupado por curso, com o progresso e os indicadores de cada meta. **O `SomenteVigentes` mora aqui agora** (veio da consulta de planos) | `VI-01`: abre preenchida, **sem exigir "Pesquisar"** — e é a **única** rota com essa exceção. **Item de plano em rascunho NÃO aparece**: é a lista de obrigações, e rascunho não é uma. Apagar o filtro faria a tela oferecer "Prestar contas" numa rota que responde 409 |

---

## Grupo 5 — Avaliação, rodadas e desfazimento

| # | Tarefa | Verificação |
|---|---|---|
| **T-271** | Avaliar: aceitar e recusar, com motivo obrigatório na recusa e prazo de 7 dias pela `DataLocal` | `AV-01` a `AV-03`: recusa em 29/07 16h40 → prazo até **05/08 23h59min59s (-03:00)**, com o container em UTC |
| **T-272** | Limite de três rodadas; na terceira, **sem novo prazo** | `AV-10`: reenviar responde 409 `LIMITE_DE_RODADAS_ATINGIDO` |
| **T-273** 🔒 | `UPDATE ... WHERE id AND versao AND situacao = 'pendente_avaliacao'`; zero linhas → **relê a entrega** para escolher entre `CONFLITO_DE_VERSAO` e `ENTREGA_JA_AVALIADA` | `AV-06`, `AV-07` com **escritas concorrentes reais**: exatamente uma tem êxito, e a outra recebe o código correto — **relê, não adivinha** |
| **T-274** | Desfazer aceitação: **qualquer PI da instituição**, motivo obrigatório, volta a `recusada`, consome rodada, abre prazo | `AV-11`, `AV-13`, `AV-14`: com rodadas esgotadas → 409; sobre entrega não aceita → 409 `ENTREGA_NAO_ESTA_ACEITA`; o cumprimento **diminui na hora** |
| **T-275** 🔒 | `avaliador_era_coordenador` computado no ato e **gravado**; **nunca recalculado** | `AV-15` a `AV-18`, `CP-12` de `cursos`: continua verdadeiro depois que a pessoa deixa de coordenar |
| **T-276** 🔒 | **Não existe `AVALIACAO_DO_PROPRIO_CURSO` em lugar nenhum** | `grep` no repositório inteiro não encontra o identificador |
| **T-277** | `GET /avaliacoes` com **`coordenado_pelo_avaliador` por linha** | A tela avisa **antes** de abrir, não depois |
| **T-278** | `GET /metas/pendencias` com os dois badges conforme os **perfis efetivos** | `NT-04`: o do PI é o tamanho da fila; o do coordenador zera por item ao abrir a entrega |

---

## Grupo 6 — Notificação e restauração de prazo

| # | Tarefa | Verificação |
|---|---|---|
| **T-279** 🔒 | `notificacao_evento` e `notificacao_gerada_em` escritos **na transação de negócio**; o use case **não importa o adapter de SMTP** | Teste de contrato arquitetural: os pacotes de `usecase` desta feature não importam o pacote de e-mail |
| **T-280** | Relê tarefa 1: envia as pendentes, monta a mensagem **a partir da própria linha**, marca `notificacao_enviada_em` | `NT-01`: curso, meta, motivo, data-limite, rodada e o caminho. **Sem anexo, sem conteúdo de comprovante, sem dado de terceiro** |
| **T-281** 🔒 | `NT-03`: servidor fora do ar → **a recusa permanece**, o PI recebe 200 sem erro, o aviso interno aparece, as tentativas crescem e o último erro é registrado | A recusa nunca é desfeita por falha de envio |
| **T-282** | Backoff exponencial por tentativas, com teto; **linha pendente nunca apagada** | Nenhum `DELETE` no caminho do relê |
| **T-283** | Destinatário é o coordenador com designação **vigente na data do envio**, não quem era na recusa | `NT-06`: quem coordena e avalia notifica a si mesmo, e isso **não** é suprimido |
| **T-284** | Relê tarefa 2: restauração de prazo por vacância, com o predicado de §10 — **os dois lados**, e auditoria de `restaurar_prazo_por_vacancia` | `VG-06`: novo prazo até 17/08, **rodada não consumida**; e uma entrega cujo prazo expirou **antes** de o curso ficar vago **não** é restaurada |
| **T-285** | Aviso interno **derivado do estado da entrega**, sem passar pelo relê | `NT-02`: com o SMTP fora do ar, o badge e o destaque continuam funcionando |

---

## Grupo 7 — Relatório e exportação

| # | Tarefa | Verificação |
|---|---|---|
| **T-286** 🔒 | Consulta única com **contagem por junção lateral** e `exigido` lido da linha | `RD-01` a `RD-04`: três cursos leem 4, 2 e 2; uma entrega com 3 anexos conta **1 de 4**, nunca 3 de 4; pendentes e recusadas **não contam** |
| **T-287** 🔒 | Filtro por indicador e por origem com **`EXISTS`, nunca junção** | `RD-10`: a meta com 1.4 e 1.5 filtrada por origem aparece **uma vez**, e a linha não é deduplicada — **não há o que deduplicar** |
| **T-288** 🔒 | Teste de que a SQL gerada do relatório **não contém `DISTINCT`** | Introduzir `DISTINCT` faz o teste falhar — comprovar e desfazer. É o que impede o relatório plausível e falso |
| **T-289** 🗄️ | `total` da paginação contado **sem as junções laterais** | `dba` valida V-4: o `total` bate com o número de linhas do conjunto filtrado |
| **T-290** | Lista de indicadores por lateral agregada, com a junção a `indicador` **por `AplicarEscopo`**, nunca nua | Indicador de outra instituição não aparece nem por essa via |
| **T-291** | Marca de coincidência por `EXISTS` com `FragmentoDesignacaoVigente`, olhando **quem coordena hoje** | `AV-17`: a marca do relatório muda quando a coordenação muda; **a da auditoria não** |
| **T-292** | Situação por linha conforme o diagrama 2.2, com **"Sem responsável"** e o aviso agregado no `resumo` | `VG-02` a `VG-04`, `RD-06`; o `resumo` é calculado sobre o conjunto filtrado **inteiro**, não sobre a página |
| **T-293** | `RD-05`: plano em rascunho **não** entra no relatório | **Inalterado por QP-3** — ler o plano é uma coisa, apurá-lo é outra. Publicar faz as linhas aparecerem |
| **T-294** | Recorte do coordenador: só os cursos dele, **sem** filtro de responsável e **sem** exportação | `RD-14`: quem acumula PI e Coordenador vê o relatório completo, com as próprias linhas marcadas |
| **T-295** | Exportação CSV **em fluxo, com cursor**, UTF-8 **com BOM**, separador `;`, limite de 50.000 linhas, auditada com linhas e filtros, **403 para o coordenador** | `RD-12`, `RD-13`; `grep` não encontra a lista completa sendo montada em memória |

---

## Grupo 8 — Frontend

| # | Tarefa | Verificação |
|---|---|---|
| **T-296** | Menu: **Minhas metas, Avaliação de entregas e Desempenho dos cursos no grupo Metas**, com os dois badges | Quem acumula os dois perfis vê os dois badges, sem um esconder o outro |
| **T-297** | `/app/minhas-metas` — **a única exceção ao padrão de CRUD**, agrupada por curso, com `<h2>` reais | O `code-reviewer` trata como achado qualquer outra listagem que consulte na montagem |
| **T-298** | `AreaDeAnexos` novo em `shared/forms/` e `ProgressoDeEnvio` novo em `shared/ui/`, conforme `ux.md` | Botão "+ Adicionar arquivos" **sempre** funcional por teclado; arrastar-e-soltar é reforço, nunca único caminho |
| **T-299** | As duas granularidades de progresso: **agregada** no registro, **por arquivo** na correção | Reflete o contrato de API; não finge granularidade que o backend não oferece |
| **T-300** | Queda de conexão: o formulário volta a editável, nada se perde, e o novo clique **reusa a mesma `Idempotency-Key`**; a tela trata 200 e 201 igual | Sem mudança de contrato |
| **T-301** | Tela de avaliar, com o aviso de coincidência **`role="status"`, informativo e não bloqueante**, e o modal de desfazer com `role="alertdialog"` | O botão continua habilitado; a confirmação de desfazer diz que **o cumprimento vai diminuir** |
| **T-302** | Tela de desempenho, com as marcas em **texto**, o filtro próprio de coincidência e o aviso agregado no topo | As três camadas de `X12` e a quarta de `X14` na tela |
| **T-303** | Atualização dos badges: na troca de rota e após a ação local, de forma otimista. **Sem polling** | Gatilho registrado: só migrar para push sob necessidade medida |

---

## Grupo 9 — E2E

| # | Tarefa | Verificação |
|---|---|---|
| **T-304** | `e2e/ciclo-da-entrega.spec.ts` conforme `design.md` §12.3 | Os oito passos passam; **o passo 8 prova que o número do relatório diminui** |
| **T-305** | `e2e/anexo-autorizado.spec.ts` conforme `design.md` §12.3 | Os quatro passos passam; o passo 4 confere que nenhuma resposta contém o endereço do armazenamento |

---

## Grupo 10 — Depois do teste manual do dono

| # | Tarefa | Quem |
|---|---|---|
| **T-306** | `testes-pendentes.md` com os adiados de §12.2 | `dev-fullstack` |
| **T-307** 🗄️ | Consolidar as migrations da feature em uma | `dba` |
| **T-308** | Revisão de segurança + qualidade, em paralelo entre si | `security-reviewer`, `code-reviewer` |
| **T-309** 🆕 | `evidence.md` com as duas diferenças vazias, **com `C` recalculado depois de `QP-6` fechada** (`EN-03` muda de código) | `qa-tester` |

---

## Bloqueio registrado

**`QP-6` (`plano-acao/design.md` §11.1) não bloqueia a construção, mas
bloqueia o fechamento do critério de aceitação desta feature também**, porque
`EN-03` da `spec.md` daqui ainda diz 404 e o código correto passou a ser 409.

## Discordâncias

Nenhuma do `dev-fullstack`. As divergências entre documentos estão em
`design.md` §15 — em especial a recusa do `Idempotency-Key` por anexo
individual, com **o risco residual formalmente aceito e o gatilho nomeado**.
