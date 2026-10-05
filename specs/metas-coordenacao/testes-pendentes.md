# Testes pendentes — metas-coordenacao

**Critério de aceitação (spec.md §12.3):** `C = A ∪ B`, `A ∩ B = ∅`, onde `C` são
os identificadores de cenário da seção 7 do spec.md (`EN`, `AN`, `AV`, `VG`,
`NT`, `RD`, `VI`), `A` são os cobertos por teste automatizado agora, `B` é
esta lista. Ver `EN-03` — muda de código (404→409, M-14) e ainda não foi
reconciliado em `spec.md`/`QP-6`; o recálculo de `C` fica bloqueado até essa
reconciliação, conforme já registrado em `design.md` §15 e `tasks.md`.

## Cobertos agora (`A`)

Testes automatizados escritos e verdes nesta entrega:

- **EN-01, EN-02, EN-03 (409, M-14), EN-06 (mock e corrida real EN-06)** —
  `usecase/command/entrega/registrar_test.go`,
  `adapter/postgres/entrega_repository_test.go::TestEntregaRepository_EN06_...`
- **AN-02 (dois níveis: assinatura e ZIP)** —
  `domain/valueobject/tipo_de_anexo_test.go`
- **AV-01, AV-02, AV-06, AV-07 (mock), AV-15/AV-18 (marca de coincidência)** —
  `usecase/command/avaliacao/avaliar_test.go`
- **AV-11, AV-13, AV-14 (rodadas esgotadas via ciclo aceitar→desfazer)** —
  `domain/entrega/entrega_test.go`, `usecase/command/avaliacao/desfazer_aceitacao_test.go`
- **AV-03, AV-04, AV-05, AV-10 (prazo e rodadas, com relógio injetado)** —
  `domain/entrega/entrega_test.go`
- **RD-01 a RD-04 (não divide nem multiplica), RD-10 (filtro por indicador
  não duplica), ausência de `DISTINCT` na consulta de listagem/total, V-4
  (total bate com o filtrado)** — `adapter/postgres/relatorio_repository_test.go`
  — contra banco real
- **T-120 (o `DISTINCT` sancionado do resumo usa o identificador do curso,
  nunca o nome)** — defeito encontrado na revisão de qualidade:
  `count(DISTINCT ... curso_nome)` deduplicava por nome, e nome não é
  identidade — dois cursos homônimos sem coordenador contariam como um.
  Corrigido para `curso_id` (`relatorio_repository.go::resumoSQL`), com
  `item_plano.curso_id` acrescentado a `selectCampos`. Dois testes novos:
  `TestResumoSQL_DistinctSancionadoUsaIdentificadorDoCurso` (mecanismo puro,
  allowlist explícita da única ocorrência sancionada de `DISTINCT`
  admitida na base — a que converte grão de item para grão de curso, nunca
  a que esconderia multiplicação de junção) e
  `TestRelatorioDesempenho_T120_ResumoContaCursosVagosPorIdentificador`
  (contra banco real, dois cursos vagos distintos contam 2). O caso
  literal de dois cursos *homônimos* não é reproduzível em teste porque o
  schema já impede (`uq_curso_instituicao_nome`) — a correção vale como
  defesa em profundidade e por princípio geral, não só pelo caso
  observável hoje.
- **VI-06 (outra instituição → 404)** — `adapter/postgres/entrega_repository_test.go`
- **AN-04, AN-05, AN-06 (fronteira de isolamento do download de anexo — mesma
  instituição/carteira concede, outra instituição e curso fora da carteira
  negam com 404)** — `adapter/postgres/entrega_repository_test.go`. Promovidos
  de `B` para `A` por decisão do arquiteto: download de anexo de outra
  instituição é fronteira de isolamento, não espera a fase de Release.
- **M-08 (nunca ler além do limite combinado, mesmo com origem
  ilimitada)** — `adapter/http/entrega_multipart_test.go`, com um
  `io.Reader` sintético que nunca devolve EOF, provando que
  `lerConteudoLimitado` nunca lê mais que `limite+1` bytes independente do
  que a origem ofereça. Promovido de `B` para `A` pela mesma triagem: é
  fronteira de segurança (exaustão de recurso por payload não validado),
  não espera a Release.
- **T-113 (seletor de período de "Minhas metas" é projeção da carteira, não
  do catálogo da instituição — período sem item na carteira não aparece,
  plano em rascunho não aparece, sem duplicata)** —
  `adapter/postgres/entrega_periodos_test.go`
- **Mecanismo de circuito do EmailSender** — `adapter/smtp/email_sender_test.go`
- **T-256 (Alvo fechado)** — `adapter/postgres/alvo_test.go` (mecanismo herdado, continua verde)
- **T-123/T-125 (injeção em fronteira de saída — duas camadas)**:
  - **Domínio:** `motivo`/`observação` recusam caractere de controle
    (`\r`, `\n`, C0/DEL) via `valueobject.ProibirCaractereDeControle`,
    chamado em `NovaEntrega`, `Entrega.Corrigir`, `Entrega.DesfazerAceitacao`
    e `NovoResultadoDeAvaliacao` — testado em
    `domain/entrega/entrega_test.go` (3 casos novos) e
    `domain/valueobject/resultado_de_avaliacao_test.go` (3 casos novos,
    incluindo a tabela de `ProibirCaractereDeControle`).
  - **Banco (segunda camada, independente):** migration `000008` adiciona
    `CHECK (motivo !~ '[[:cntrl:]]')` e o mesmo para `observacao` — sobrevive
    a um caminho de escrita futuro que não passe pelo Value Object. `up` →
    `down` → `up` verificado em dev e teste.
  - **SMTP (T-123, adapter):** `montarMensagem` (`adapter/smtp/email_sender.go`)
    agora devolve `error` e RECUSA (nunca sanitiza) `\r`/`\n` em qualquer
    valor de cabeçalho — assunto, remetente, destinatário — antes de montar
    a mensagem; valor não-ASCII do assunto é codificado em RFC 2047
    (`mime.QEncoding`), o que também corrige o bug de exibição do
    travessão ("—") no assunto ("Entrega recusada — Curso"). `Enviar`
    valida ANTES de checar o circuit breaker. Quatro testes novos em
    `adapter/smtp/email_sender_test.go`.
  - **CSV (T-125, adapter):** aspas RFC 4180 não neutralizam injeção de
    fórmula (Excel/LibreOffice removem as aspas antes de avaliar) — nova
    função `neutralizarFormulaCSV` (`adapter/http/relatorio_handler.go`)
    prefixa `'` quando o campo começa com `=`, `+`, `-` ou `@`, aplicada só
    nas colunas textuais (curso, responsável, meta, indicadores, situação)
    — nunca nas numéricas (onde `-` é sinal legítimo). Três testes novos em
    `adapter/http/relatorio_handler_test.go`, incluindo a prova de que
    `Cumprimento: -10` (percentual negativo) permanece intacto.
  - Registrado como restrição permanente em `_fundacao-metas.md` §12.17-18
    (texto livre sem caractere de controle; neutralização de injeção
    desenhada antes de cada novo formato de saída).
- **T-124/T-127 (relê — data de referência e reivindicação atômica)**:
  - **T-124 (fuso):** `valueObjectDataLocalString` usava `.UTC()` para a
    data que decide quem recebe a notificação (via designação vigente) —
    inconsistente com o resto do mesmo arquivo, que já usava
    `fusoDeExibicao` para o prazo de correção poucas linhas abaixo. Corrigido:
    `rele.go` computa `hoje := valueobject.DataLocalDe(r.relogio.Agora(),
    r.fusoDeExibicao)` e passa para o repositório — a função `.UTC()` foi
    removida, não contornada.
  - **T-127 (reivindicação não era atômica):** `SELECT ... FOR UPDATE SKIP
    LOCKED` via `SelectContext` roda em autocommit — o lock é liberado ao
    fim do próprio `SELECT`, MUITO antes do envio de e-mail (latência
    imprevisível), então N réplicas podiam enviar a mesma notificação.
    Corrigido com reivindicação comprometida: `EntregaReleRepository.
    ListarPendentesDeNotificacao` foi renomeado para
    `ReivindicarNotificacoesPendentes` e reescrito em dois statements — (1)
    `UPDATE entrega SET notificacao_reivindicada_em = now() WHERE id IN
    (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING id`, atômico mesmo em
    autocommit porque é um único statement; (2) um `SELECT` comum (sem
    lock) que lê os dados das linhas já reivindicadas. O envio de e-mail
    roda inteiramente fora de qualquer lock ou transação, entre os dois
    passos.
  - **Uma coluna fecha duas pendências:** `notificacao_reivindicada_em`
    (migration `000008`) é ao mesmo tempo a marca de reivindicação (T-127)
    e a "última tentativa" que faltava para o backoff exponencial real
    (simplificação registrada anteriormente — "a janela entre execuções do
    relê já funciona como backoff natural" — agora fechada): uma linha só
    volta a ficar elegível depois de `2^tentativas` minutos (teto 10, ~17h)
    desde a última reivindicação.
  - **Semântica honesta, escrita no código:** at-least-once — se o processo
    cair entre o envio confirmado e `MarcarNotificacaoEnviada`, a
    reivindicação expira e a mesma entrega é reivindicada de novo,
    duplicando o e-mail. Deliberado (comentário em `rele.go`,
    `enviarNotificacoesPendentes`): preferimos duplicar a perder o alerta.
    Não corrigido com transação ao redor do envio — segurar uma transação
    através de uma chamada SMTP trocaria "risco raro de duplicata" por
    "conexão presa enquanto o servidor de e-mail está lento".
  - **Testes (container em UTC, como `TestVigencia_RotuloEPredicadoConcordam`)**
    — `adapter/postgres/entrega_rele_repository_test.go`:
    `TestEntregaRepository_T127_SegundaReivindicacaoNaMesmaJanelaNaoRepete`
    (duas reivindicações seguidas, a segunda não repete a mesma entrega),
    `TestEntregaRepository_T127_ReivindicacaoExpiradaVoltaAFicarElegivel`
    (reivindicação de 2 min atrás, tentativas=0, volta a ficar elegível
    depois da janela de 1 min expirar), e
    `TestEntregaRepository_T124_HojeDoFusoDecideDestinatario` (mesma
    fronteira DG-06 de `designacao_vigente_sql_test.go`: 23h58 de 31/07 em
    Brasília ainda reivindica, 00h02 de 01/08 em Brasília já não reivindica
    — a designação encerrou no dia local).
  - **Guarda B atualizado:** `internal/port/portas_test.go`
    (`TestPortasEstreitas_ListaFechada`) — a assinatura fechada de
    `EntregaReleRepository` foi ajustada para o método renomeado, com o
    novo parâmetro `valueobject.DataLocal`, no mesmo commit desta mudança
    de design (regra do próprio guarda).
  - **Achado registrado, não corrigido (fora de escopo desta rodada):**
    `ListarCandidatosARestauracaoDePrazo`/`RestaurarPrazoPorVacancia` (§10)
    têm a mesma fragilidade de SKIP LOCKED em autocommit, com consequência
    menor (auditoria duplicada, não e-mail duplicado, pois não há chamada
    externa entre as duas operações) — ver nota em `design.md` §10.

## Adiados (`B`) — o que falta e por quê

### Cenários da spec ainda sem teste automatizado dedicado

`EN-04, EN-05, EN-07 a EN-11 (integração real), AN-01, AN-03, AN-07
(401 sem sessão — coberto estruturalmente por middleware_sessao_test.go,
genérico a todas as rotas, não duplicado aqui), AV-08, AV-09, AV-12, AV-16,
AV-17, VG-01 a VG-07 (curso vago/inativo — restauração de prazo, VG-06, não
tem seed observável, ver abaixo), NT-01 a NT-06 (o relê de e-mail em
execução real, contra Mailpit), RD-05 a RD-09, RD-11 a RD-15, VI-01 a VI-05,
VI-07 a VI-10.`

Nenhum destes é fronteira de segurança/isolamento/integridade isolada (a
triagem pedida pelo arquiteto já promoveu os dois candidatos que eram —
download de anexo entre instituições/carteiras e o limite de upload, ver
acima); são principalmente fluxos completos de rota (HTTP fim-a-fim) e
comportamento do relê em execução real, cobertos em mecanismo pelos testes
herdados de `fundacao-metas.md` §3.7 e pelos testes de unidade do domínio.

A cobertura de **mecanismo** (isolamento genérico via `AplicarEscopo`,
concorrência otimista, deleção lógica) já está garantida pelos testes
herdados de `fundacao-metas.md` §3.7 — o que falta aqui é o teste
**específico da rota/fluxo** de metas-coordenacao repetindo esse mecanismo
com dados do domínio (ex: um teste de integração HTTP fim-a-fim por rota,
que a suíte de smoke ainda não cobre para esta feature).

### Simplificações de implementação registradas (não são bugs — são decisões deliberadas sob restrição de tempo, a rever com o arquiteto)

1. **Detecção de tipo de anexo usa buffer limitado em memória, não streaming
   puro.** O design (§6.3) descreve `bufio.Peek` + `io.TeeReader` num único
   passe streaming. A implementação real usa `io.LimitReader` até
   `LimiteBytesPorArquivo+1` seguido de leitura completa do buffer (≤10 MB,
   memory-safe) para permitir a inspeção do ZIP (DOCX/ODT exigem acesso ao
   diretório central, que fica no FIM do arquivo — impossível com
   `Peek`/streaming puro). O invariante real de M-08 (nunca ler dado
   ilimitado antes de validar o tamanho) é preservado; o que muda é que a
   gravação no S3 não é feita a partir do MESMO reader de rede, e sim do
   buffer já validado.
2. **Backoff exponencial do relê não é literal.** O schema aprovado
   (`entrega`) não tem uma coluna de "última tentativa" — só
   `notificacao_gerada_em` (fixo, não muda em retentativas) e
   `notificacao_tentativas`. A implementação usa o intervalo do próprio
   `ticker` (30 s) como backoff natural — uma falha só é retentada na volta
   seguinte, nunca imediatamente. Backoff exponencial de verdade exigiria
   uma coluna nova (`notificacao_ultima_tentativa_em`), fora do schema já
   migrado — decisão para o `dba`/`arquiteto` avaliar antes da consolidação.
3. **`InclusEntregasAnteriores` (RD-07, "inclui entregas de gestão
   anterior") é heurística, não testada.** Calculada como
   `EXISTS(entrega.enviada_por <> coordenador atual)` — plausível, mas sem
   teste de integração que prove os casos de borda (ex: coordenador atual
   nunca enviou nada).
4. **VG-06 (prazo restaurado após vacância) não tem dado de seed
   observável.** A massa de desenvolvimento existente (`cursos`) não tem
   nenhum curso com histórico "designação → vacância → nova designação": o
   único curso vago (Pedagogia) é **permanentemente** vago, de propósito
   (CU-05 de `cursos`), e criar uma reatribuição para ele quebraria esse
   cenário. A lógica de `ListarCandidatosARestauracaoDePrazo` e
   `RestaurarPrazoPorVacancia` (`adapter/postgres/entrega_rele_repository.go`)
   está implementada e o SQL foi revisado, mas **não foi exercitada contra
   dado real** — nem por teste automatizado nem por seed. Registrado para o
   `arquiteto`: precisa de um curso de desenvolvimento dedicado com essa
   história de designações (fora do escopo de `cursos`, que já está
   concluída) para se tornar observável.
5. **Progresso de upload é indeterminado, não percentual real.** `ux.md`
   pede `onUploadProgress` (XHR) para o envio agregado da entrega nova. A
   implementação usa `fetch` + `FormData`, que não expõe progresso de
   upload no navegador — o componente mostra um indicador indeterminado
   ("Enviando...") em vez do percentual. Migrar para `XMLHttpRequest` (ou
   `fetch` com `ReadableStream` no corpo, quando o suporte de navegador for
   maduro) é a correção completa.
6. ~~Seletor de período — divergência a confirmar~~ **Resolvido.** O
   arquiteto decidiu (T-113): a rota `GET /minhas-metas/periodos` fica, mas
   a implementação inicial (que usava `Escopo.SemCarteira()` sobre
   `AlvoPeriodo`, tratando período como catálogo da instituição) estava
   errada — "os períodos em que eu tenho metas" é projeção da **carteira**
   do coordenador, não do catálogo. Reimplementado com `AplicarEscopo`
   sobre `AlvoItemPlano` (que suporta carteira), na mesma base que produz
   `MinhasMetas` — plano vigente, item dentro do recorte — com `SELECT
   DISTINCT` de período (deduplicação de linha, não de agregado — não é o
   padrão M-10 proibido no relatório). Nenhuma permissão nova é necessária:
   a rota não lê nada além do que `MinhasMetas` já autoriza. Testado em
   `adapter/postgres/entrega_periodos_test.go` (T-113, acima).

   **Nota para quem mexer em `internal/port/entrega_repository.go` depois:**
   o comentário de `port.PeriodoOpcao` ainda cita `escopo.SemCarteira()` —
   ficou desatualizado porque essa correção é só na implementação
   (`adapter/postgres`), e fui instruído a não tocar no pacote `internal/port`
   nesta rodada (outro agente está corrigindo o guarda de portas estreitas
   lá). Precisa de um ajuste de uma linha quando esse trabalho terminar.

### Frontend — microcópia, estados finos e a11y

Conforme spec.md §12.2 (explicitamente adiado): formatação exata de
mensagens de erro por cenário (a tabela completa de `ux.md`), skeleton com
estrutura fiel por tela, agrupamento visual do card "aceitação desfeita"
(AV-12) com destaque âmbar, atalhos de teclado desta feature, textos exatos
de estado vazio/erro em cada tela, paginação/ordenação/persistência de
filtro nas telas de Avaliação e Desempenho (o mecanismo existe via
`useEstadoDeGrid`, mas não foi testado tela a tela), componente
`avaliar-acoes.tsx`/`desfazer-aceitacao-dialog.tsx` separados (a
implementação atual está inline na página, funcional mas sem a decomposição
de componentes que `ux.md` sugere).

## E2E (Grupo 9 de tasks.md) — decidido: fica para a Release

`design.md` §12.3 recomenda dois arquivos (`e2e/ciclo-da-entrega.spec.ts`,
`e2e/anexo-autorizado.spec.ts`) e `tasks.md` os lista no Grupo 9, antes do
"PARE — o dono testa". `spec.md` §12, por outro lado, registra "E2E
antecipado: não" e defere a suíte para a fase de Release (`CLAUDE.md`,
"Camada 2"). **Decisão do arquiteto: prevalece `spec.md`** — a spec reflete
a decisão de custo do dono, e os dois E2E saem do `tasks.md` e vão para a
lista da Release.

**Ressalva aplicada:** a exceção da própria regra de custo vale para fluxo
que seja fronteira de segurança/isolamento/integridade — nesse caso vira
teste de **integração agora**, não E2E na Release. Fiz essa triagem (ver
seção "Cobertos agora" acima): download de anexo entre instituições/carteiras
(AN-04/05/06) e o limite de upload (M-08) foram promovidos e já estão
cobertos por teste de integração real. O resto do fluxo (registrar → avaliar
→ recusar → corrigir → aceitar → desfazer, e o relê contra Mailpit) está
implementado ponta a ponta e testável manualmente agora, mas o E2E
automatizado fica para a Release.

## Encabeçam a prioridade (spec.md §12.2)

Nesta ordem, se/quando o dono aprovar tempo para fechar `B`:
`VG-06` (precisa de seed novo primeiro), `AV-12`, `RD-07`, `RD-08`, `NT-01`,
`NT-02`, `RD-12`, `RD-13`.
