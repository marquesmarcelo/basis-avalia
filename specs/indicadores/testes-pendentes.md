# Testes pendentes — indicadores

Cenários adiados conforme `design.md` §10.2 e o critério de custo do
`CLAUDE.md` ("escreve-se agora o que é fronteira de segurança, isolamento
ou integridade — o resto vai para este arquivo"). Cada linha: cenário, o
que verificar, prioridade.

## Prioridade alta

| Cenário | O que verificar | Motivo da prioridade |
|---|---|---|
| **MC-15** | A tela de meta, sem nenhum indicador disponível (nem INEP nem próprio), mostra o bloco instrutivo com o link para `/app/indicadores` — já implementado em `meta-form-modal.tsx`, falta teste de componente/E2E cobrindo o estado vazio de verdade | Sem ela, o PI fica diante de um autocomplete vazio sem saída — é a UX mais frágil da feature |
| **IV-03** | A contagem de uso (`metas_da_instituicao` para o PI, `metas` total para o Administrador) não vaza entre instituições — teste de integração com dado real de duas instituições e três/duas metas, batendo os números exatos | É onde um vazamento entre instituições seria mais discreto (um número errado, não um erro visível) |
| **MC-11 (item_plano)** | Bloqueio de exclusão de meta por uso em item de plano — **não implementável ainda**: a tabela `item_plano` só nasce em `specs/plano-acao`. Hoje `MetaRepository.Excluir` não bloqueia por uso nenhum (ver `port/meta_repository.go`). Quando `plano-acao` criar `item_plano`, este teste — e o bloqueio em si — precisam ser adicionados junto | Divergência registrada: `design.md`/`tasks.md` T-135 lista MC-11 como "escrito agora", mas a tabela que o cenário depende não existe nesta feature. Ver seção "Divergências" abaixo |

## Caminho feliz (adiado por já estar coberto indiretamente pelos testes de mecanismo)

| Cenário | O que verificar |
|---|---|
| IE-01 | Administrador cadastra indicador do INEP com sucesso (201, escopo plataforma, instituicao_id nulo) |
| IN-01 | PI cadastra indicador próprio com sucesso (201, escopo instituicao) |
| MC-01 | Meta com um indicador — sem quantidade, período, curso ou campo de INEP |
| MC-02 | Meta com dois indicadores — o caso do dono (parcialmente coberto pelo teste de domínio e pelo E2E) |
| MC-06 | Meta mistura indicador do INEP e indicador próprio na mesma lista |

## Filtros, ordenação e persistência de estado

| Cenário | O que verificar |
|---|---|
| IN-06 | Filtro "Origem" (Todos/Do INEP/Próprios) na tela `/app/indicadores` |
| IN-07 | A tela abre sem executar consulta — grid só aparece após "Pesquisar" |
| Persistência | `grid-state:indicadores-inep`, `grid-state:indicadores`, `grid-state:metas` restauram filtro/ordenação/paginação ao voltar à tela |

## Validações de formulário (client-side)

| Cenário | O que verificar |
|---|---|
| MC-07 (parte de validação) | Indicador inativo escolhido em meta nova → `FieldError` no cliente antes mesmo de chamar a API |
| MC-09 | Nome de meta repetido → 409 `NOME_META_DUPLICADO` mapeado para `FieldError` no campo Nome |

## Caminho conhecido (baixo risco, sem dado real ainda testado)

| Cenário | O que verificar |
|---|---|
| MC-12 | A mesma meta em dois planos com quantidades independentes — depende de `item_plano` (plano-acao) |
| MC-13 | Trocar indicadores da meta vale dali em diante para planos vigentes — a parte de UI (troca via `ComboboxEntidadeMultipla` já implementada; a auditoria com lista antes/depois já testada em `atualizar_meta_test.go`) |
| IV-07 | Coordenador lê indicadores/metas como contexto (entrega), mas recebe 403 ao chamar a listagem — precisa do fluxo de entrega de `metas-coordenacao` para ser testado de ponta a ponta |
| IV-08 | Professor/aluno recebem 403 em qualquer rota daqui — coberto parcialmente pelos smoke tests existentes (padrão idêntico ao de usuários), sem teste dedicado nesta feature |
| IV-09 | Permissão antes de isolamento — professor pedindo indicador de outra instituição recebe 403, não 404 — mesma cobertura do padrão geral de autorização, sem teste dedicado nesta feature |

## Microcópia

Textos exatos de toasts, mensagens de erro de campo e rótulos — verificação
manual/visual na fase de smoke do dono do produto, não automatizada agora.

---

## Divergências registradas durante a implementação

1. **MC-11 (bloqueio de exclusão de meta por uso em plano) não é
   implementável nesta feature.** `design.md` T-135 e a spec listam MC-11
   junto de IE-07/IN-05 como "bloqueio de exclusão com uso, escrito
   agora", mas o bloqueio depende da tabela `item_plano`, que só existe em
   `specs/plano-acao` (feature seguinte na cadeia). Implementei
   `MetaRepository.Excluir` **sem** nenhum bloqueio de uso por enquanto —
   a exclusão lógica sempre sucede quando o escopo permite. Quando
   `plano-acao` criar `item_plano` com FK para `meta`, o bloqueio
   real (na mesma mecânica de `FOR UPDATE` + contagem usada em
   `ExcluirSeSemUso` de indicador) precisa ser adicionado ali, junto do
   teste correspondente.
2. **Bug encontrado e corrigido: `todasAsPermissoes` em
   `obter_contexto_de_sessao.go` não incluía as permissões novas de
   `fundacao-metas.md` §6.1** (indicador, meta, curso, designação,
   período, plano, entrega, relatório) — o Grupo 0 (T-116) criou as
   constantes e a matriz de perfil, mas não atualizou a lista que monta o
   campo `permissoes` da resposta de `/auth/login` e `/auth/eu`. Sem essa
   lista, nenhum perfil jamais recebia essas permissões na sessão,
   independente do que `matrizPermissoes` dissesse — o menu "Metas" nunca
   apareceria para o PI. Corrigido acrescentando todas as permissões das
   quatro features de metas à lista, com os dois testes de
   `obter_contexto_de_sessao_test.go` atualizados para refletir a união
   completa esperada para o Pesquisador Institucional. Sem essa correção,
   `indicadores` não seria utilizável em nenhuma tela.
3. **T-140/E2E step 5 (IE-09 "tela de não encontrado, não a de sem
   permissão" para `/app/metas`) implementado como o mecanismo real do
   frontend permite.** O `useGuardaDePermissao` deste projeto trata toda
   falta de permissão de forma uniforme — redireciona para `/app` com um
   toast "Você não tem permissão para acessar esta área." — e não
   distingue 403 de 404 na camada de UI (a distinção é uma propriedade do
   backend, testada em `TestSmoke_IndicadoresPlataforma_IE09_
   IndicadorInstitucionalRecebe404`). O E2E `catalogo-comum.spec.ts`
   verifica o comportamento real e uniforme do guard para
   `/app/metas` (Administrador não alcança), e a distinção 403-vs-404
   fica coberta no nível correto (backend), não inventando uma tela nova
   só para esta feature.
