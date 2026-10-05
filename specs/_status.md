# Status do Projeto

> Quadro de andamento. Toda sessão nova deve ler este arquivo antes de assumir
> que está começando do zero. Cada subagent atualiza a linha que tocou.

**Projeto:** basis-avalia — apoio ao acompanhamento das metas da coordenação,
com base nos instrumentos de avaliação do INEP.
Visão: `specs/00-visao-produto.md` · Stack: `project.config.md`

**Multi-institucional desde a v1.** **Um usuário tem um CONJUNTO de perfis**;
permissões são a **união**; **Administrador do Sistema não acumula**.

## Funcionalidades

| Funcionalidade | Fase | Atualização | Observação |
|---|---|---|---|
| autenticacao-usuarios | **qa** | 29/09/2026 | Correções dos blocos 1 e 2 em `design.md` §16 |
| indicadores | implementada | 29/09/2026 | — |
| cursos | implementada | 29/09/2026 | `coordenador_curso` derivado de designação com vigência |
| plano-acao | implementada | 29/09/2026 | Renomeada pelo dono a partir de "plano de metas" |
| metas-coordenacao | implementada | 29/09/2026 | Três pareceres do arquiteto abaixo |

---

## ⚠️ Decisão prioritária: os guardas de arquitetura eram fail-open

**Medição de 29/09: 23 portas declaradas, 14 verificadas, 9 escapando — e o
teste verde.** Os guardas 1 e 3 enumeravam **listas escritas à mão**, então
porta nova não acrescentada escapava por omissão. **Falha minha, e é a segunda
vez:** O-1 apontou a fraqueza, T-102 corrigiu a assinatura — metade — e deixou
a enumeração, que era a metade mais perigosa, porque assinatura errada quebra o
build de quem a muda e lista desatualizada não quebra nada.

**Decisão (D-26), agora restrição transversal:**

> **Verificação cuja cobertura depende de lista escrita à mão não é mecanismo —
> é documentação com sintaxe de teste.** Toda verificação estrutural **deriva a
> enumeração do código**, e o padrão para o que é novo é **reprovar**. Se a
> lista for inevitável, ela enumera **exceções**, nunca o que é verificado.

Os guardas passam a varrer o pacote por análise estática (**T-110 a T-112**),
com **comprovação negativa obrigatória**: acrescentar uma porta sem `Escopo`,
**sem tocar em lista nenhuma**, e ver o teste falhar. A ausência dessa
comprovação foi o que deixou T-102 passar pela metade.

**Estas tarefas têm prioridade sobre o bloco 1** — enquanto não entram, nenhuma
verificação estrutural do sistema vale o que afirma. **Um agente só** toca
`internal/port`, para não colidir com os três em paralelo.

Esta é a terceira vez que a mesma família aparece: o script que encadeava
etapas sem verificar código de saída, o comentário que afirmava exclusividade
que não tinha, e agora o guarda que afirmava cobertura que não tinha.
**Formato comum: afirmação verdadeira quando foi escrita, que ninguém
revalida, e cuja falsificação não produz sinal.**

---

## Pareceres sobre `metas-coordenacao`

### 1. A terceira porta estreita é aceita — o problema não é ela

`EntregaReleRepository` **fica**. O argumento de quem a criou está certo e é da
mesma natureza das outras duas: o relê roda em segundo plano, **não tem sessão,
não tem ator, não tem instituição** — varre todas por desenho, porque é o único
processo que precisa disso. `Escopo` modela "uma instituição ou a plataforma";
job de manutenção cross-tenant está fora do que ele modela. **Forçar um
`Escopo` fabricado seria pior:** ensinaria que é aceitável inventar escopo
quando não há ator, e o próximo a inventar teria ator.

**O que não foi aceitável foi ela entrar sem acionar a trava.** O implementador
documentou a decisão, o que foi correto; o mecanismo que deveria ter exigido a
minha decisão não funcionou. **A porta fica; o guarda muda.**

**A regra de admissão foi ampliada** (§4.4) — descrevia duas proveniências
quando existem três: credencial de autenticação; `Proprio` derivado de token; e
**identificador produzido pela consulta anterior do próprio processo, sem
requisição e sem ator**. Se vem do cliente, exige `Escopo`, sem exceção.

**Com uma obrigação adicional, porque o risco desta proveniência é diferente:**
nas duas primeiras o identificador é impossível de forjar; nesta ele é apenas
*de fato* server-derived. `MarcarNotificacaoEnviada(ctx, entregaID)` chamado de
um handler com `c.Param("id")` seria escrita cross-tenant sem autorização. Hoje
isso é impedido por um comentário — afirmação, não mecanismo. **O guarda C
(T-112) transforma em verificação:** `adapter/http` não pode referenciar porta
de manutenção.

### 2. `GET /minhas-metas/periodos` — aprovo a rota, recuso o `SemCarteira()`

A rota resolve um problema real: o coordenador não tem a permissão de gerenciar
períodos e sem ela não popularia o seletor. **Mas a solução inverteu o
recorte.**

O nome da rota diz o que ela é: **os períodos em que eu tenho metas**. Isso não
é o catálogo da instituição — é uma projeção derivada da carteira do próprio
coordenador. Portanto a restrição de carteira deve ser **aplicada, não
removida**:

- a consulta é um `DISTINCT` de período sobre a mesma base que já produz
  "minhas metas", com a carteira no lugar;
- **nenhuma permissão nova é necessária** — e é exatamente por isso que a rota
  não precisa da permissão de gerenciar períodos: ela não lê nada além do que
  `MinhasMetas` já autoriza;
- o seletor deixa de oferecer período que produz tela vazia, que é o
  comportamento correto de um seletor.

Além do mais, `SemCarteira()` **no recurso de topo contraria a regra escrita no
comentário do próprio método** — ele existe para junção seguinte com alvo que
não tem dimensão de curso, depois que o recurso de topo já aplicou a carteira.
Usá-lo no topo afrouxa o recorte que o alcance acabou de aplicar. Virou
restrição §15.20. Tarefa **T-113**.

**Bifurcação que devolvo ao `analista-requisitos`:** se o produto quiser mesmo
que o coordenador veja **todos** os períodos da instituição — inclusive aqueles
em que não tem nada —, então é outra rota, com outro nome e decisão de produto
própria. Não é o que o nome atual descreve.

### 3. E2E: prevalece a `spec.md` — e o `tasks.md` era meu

Contradição real, e o erro é meu: a spec §12 difere E2E para a Release, e o
`tasks.md` pedia dois antes do fim. **Prevalece a spec**, porque ela reflete a
decisão de custo do dono, e porque dois agentes lendo documentos contraditórios
é pior que qualquer uma das duas opções.

**Os dois E2E saem do `tasks.md`** e entram na lista da Release e em
`testes-pendentes.md`. **Com uma ressalva que não é negociável:** se algum dos
fluxos adiados for **fronteira de segurança, isolamento ou integridade**, a
exceção da própria regra de custo se aplica — e aí ele vira **teste de
integração agora**, não E2E na Release. O que é fluxo de tela vai para a
Release; o que é fronteira não espera.

Vale o mesmo registro já feito para `autenticacao-usuarios`: os E2E que já
existem **excedem** o pedido, são bem-vindos, e precisam entrar em
`evidence.md` e na lista da Release **para não serem escritos duas vezes**.

---

## Decisões do arquiteto que atravessam as features de metas

| # | Decisão |
|---|---|
| 1 | **"Indicadores" fica no grupo Metas.** *Item de menu fica onde a tarefa do usuário acontece, não onde a natureza técnica do dado sugere.* O texto divergente é da spec, e corrigi-lo é do `analista-requisitos` |
| 2 | **Toda trava de interface tem contraparte no use case.** Trava só visual é sugestão, não regra. A tela pode ser mais restritiva; **nunca menos**. Família 409 |
| 3 | **Campos computados aprovados**, sob três condições: computado na resposta e **nunca persistido** (persistir é dual write); **um campo por fato**; nunca derivado no cliente quando depende de regra de domínio. **`data_fim` calculado em um lugar só** |
| 4 | **`Idempotency-Key` no envio de anexo: sim** — é o oposto do caso da autenticação, onde a unicidade no banco tornava a duplicação impossível. Anexo não tem unicidade natural. Chave por **tentativa de envio**, janela de dedupe declarada |
| 5 | **Designação é imutável em coordenador e data de início.** Alterá-la muda **retroativamente quem tinha privilégio quando**. A operação legítima é **encerrar e criar outra**, que preserva a história |

---

## Riscos aceitos pelo dono do produto

| Risco | Registro |
|---|---|
| Sem política de senha e sem limite/registro de tentativas de login | spec 4.1 |
| Lista de instituições visível a qualquer visitante | spec 4.2 |
| Testes reduzidos ao mínimo nas fases iniciais | spec 4.3 |

**Regra de custo de testes:** escreve-se só fronteira de segurança, isolamento
ou integridade. **Exceção obrigatória em qualquer fase:** testes-guarda de
arquitetura e de invariante com corrida — **são mecanismo, não cobertura**, e
não entram no orçamento de teste, do mesmo modo que um `CHECK` de banco não
entra.

## Pendências que atravessam o projeto

| Item | Situação |
|---|---|
| **Guardas de arquitetura (T-110 a T-112)** | **Prioridade máxima.** Um agente só toca `internal/port` |
| Correções de segurança e qualidade | Bloco 1 de `design.md` §16. **T-096 (bump) em PR próprio**, nunca com feature |
| Automação de dependências | **Não existe.** T-109 — sem ela, o próximo `quic-go` só aparece na próxima revisão manual |
| Manifesto Kubernetes | Não existe. A porta administrativa já é a decisão que ele herda |
| Restrição de implantação | Frontend e API sob o **mesmo domínio registrável**, senão o cookie nunca é enviado |
| Dual write de e-mail | Pertence a `metas-coordenacao`; decisão entre Outbox e reconciliação no design dela |
| Proposta ao dono | Acrescentar `base-uri 'self'` e `form-action 'self'` à CSP padrão do `CLAUDE.md` — a lacuna é do padrão, não de uma implementação |

## Conversão de modais em páginas (concluída — 29/09/2026)

Decisão do dono: **todo cadastro do sistema deixa de ser modal e passa a
ser página com rota própria.** Padrão único em
[`specs/_padrao-formularios.md`](_padrao-formularios.md) — aprovado.

As **oito** conversões estão feitas: Instituição, Usuário, Administrador,
Pesquisador Institucional, Curso, Designação, Indicador, Indicador do
INEP, Meta e Período. Modais antigos removidos após conferir que nada os
referenciava.

**Continuam em janela, por decisão:** confirmação de exclusão, redefinir
senha, encerrar designação e **item do plano** (não é cadastro autônomo,
é ação repetida dentro da montagem de um plano — §1.2 do padrão).

**Peças compartilhadas criadas nesta rodada:**
- `useGuardaDeSaida` (`components/shared/hooks/`) — pergunta ao sair com
  formulário sujo. **Não é** o guarda de senha provisória, que bloqueia
  sem perguntar.
- `SelectComRotulo` (`components/shared/forms/`) — o `Select` do Base UI
  mostra o **valor** quando não recebe `items`; 16 telas exibiam id ou
  número cru. Nenhum `Select` cru restou no projeto.

## Armadilhas de ambiente descobertas (leia antes de diagnosticar bug)

| Sintoma | Causa real |
|---|---|
| Rota responde `404 page not found` do Gin, mas o log mostra a rota registrada | `air` falhava ao religar a porta (`address already in use`) e deixava **binário antigo** servindo. Corrigido com `send_interrupt` + `kill_delay` em `backend/.air.toml` |
| Suíte E2E passa mas não reflete o código | **`backend-e2e` roda imagem sem hot-reload** — precisa de rebuild, senão a evidência é de um binário velho |
| `go test` diz `ok` mas **137 testes de integração (33 arquivos) não rodaram** | Rodar pelo serviço `backend` do compose, que **não** define `DATABASE_URL_TEST`: `testhelpers.BancoDeTeste` faz `t.Skip`, e `go test` retorna `ok`. Use o serviço **`test`**, que define o banco isolado. Verificar sempre com `-v`: `SKIP` em massa é o sintoma |
| Suíte Go intermitente sem causa aparente | Rodar a suíte **enquanto um agente salva arquivos** faz o compilador ler fonte parcial. Não é flakiness do teste |

## Como usar
- Ao iniciar sessão nova: ler esta tabela antes de decidir o que fazer.
- Ao terminar sua parte, o subagent responsável atualiza a linha aqui.
- Quando o dono declarar uma versão, o arquiteto registra em "Versões lançadas".
