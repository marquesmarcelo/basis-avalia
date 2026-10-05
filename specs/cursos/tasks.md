# Tasks: cursos

**Data:** 28/09/2026 · **Lei da construção:** `design.md` desta pasta e
`specs/_fundacao-metas.md` §4 (perfil derivado) e §3.4 (carteira).
**Legenda:** 🗄️ `dba` · 🐳 `dev-docker-compose` · 🔒 teste de mecanismo
(obrigatório em qualquer fase) · ⚠️ mexe em privilégio de pessoas reais

> **Depende de `indicadores` concluída** — o Grupo 0 da fundação está lá.
> **Esta é a feature que altera privilégio**: o perfil de Coordenador deixa de
> ser atribuído e passa a ser derivado, e uma migration remove o perfil do
> conjunto atribuído de todo mundo. Nada aqui é rotina.

> **Status em 2026-09-29:** Grupos 1 a 7 concluídos e verificados — `go
> build ./...` e `go test ./...` verdes, todos os testes de mecanismo e
> comprovações negativas exigidas (T-157, T-161/T-162 CP-14, T-166,
> T-171, T-176 DG-03, T-181 CP-15) passando, com sabotagem-e-reversão
> documentada em cada um. Migration `000003_cursos` (tabelas) e
> `000005_cursos_normalizacao` (privilégio) aplicadas e testadas
> `up → down → up` nos três bancos. OpenAPI regenerado
> (`backend/docs/swagger.yaml`). Frontend (`/app/cursos`,
> `/app/cursos/{id}/designacoes`) e E2E (`designacao-e-perfil.spec.ts`)
> implementados e verificados — 13/13 na suíte completa, duas execuções
> seguidas sem recriar o banco. `T-171` teve DUAS causas, as duas
> corrigidas (ver design.md §5.4 "Achado de revisão"): primeiro, o teste
> não replicava a transação real de produção (corrigido); com isso
> corrigido o teste ainda falhava ~2-4% em 100 execuções — bug real de
> produção, não só do teste: a FK `designacao.curso_id → curso.id`
> trava a linha-pai ao inserir, mas só garante que ela existe, nunca que
> `excluido_em` continua nulo (exclusão lógica não dispara FK).
> `CriarDesignacaoUseCase` ganhou `CursoRepository.TravarSeAtivo`
> (`FOR SHARE`) antes de inserir — 300 execuções consecutivas sem falha
> depois da correção. Outros dois achados de revisão corrigidos durante
> o Grupo 6/7: `designacaoParaResposta` nunca preenchia `situacao` na
> resposta HTTP (só a query/domínio estavam corretos) e o seed de
> `joao.ribeiro@ies.edu.br` no IVV ainda atribuía `coordenador_curso`
> direto, quebrado desde a migration de normalização — corrigido com
> curso+designação próprios do IVV.
>
> **T-115 concluída** (CU-12, decisão do arquiteto): `plano_do_periodo`
> composto em `ListarCursosUseCase`, injetando `port.PeriodoRepository` e
> `port.PlanoRepository` (nenhuma porta mudou de contrato,
> `CursoRepository` continua sem conhecer `plano` — nenhuma linha do
> `SELECT` de curso tocada). Testado
> (`TestSmoke_Cursos_T115_PlanoDoPeriodoAbertoNaListagem`) e com o
> frontend (`curso-table.tsx`) ligado à ponta a ponta. **T-116**
> (comentário de `port.PeriodoOpcao` citando `SemCarteira()` obsoleto) já
> estava corrigido pelo outro agente quando cheguei a ele — nada a fazer.
> Aguardando o dono para o Grupo 8.

---

## Ordem de execução

```
Grupo 1  — domínio, banco e a extensão do Alvo        (sequencial até T-144)
Grupo 2  — perfil derivado: sessão, conjunto efetivo, API   ⚠️
Grupo 3  — carteira dentro do Escopo                   🔒
Grupo 4  — backend de curso e designação
Grupo 5  — migration de normalização                   ⚠️
Grupo 6  — frontend
Grupo 7  — E2E
⛔ PARE — o dono testa
Grupo 8  — dba consolida, revisão, QA
```

**T-160 a T-163 (Grupo 2) antes do Grupo 4**, porque todo alcance de carteira
depende do conjunto efetivo existir. Construir os use cases primeiro
produziria autorização que lê o conjunto errado e passa nos testes — a falha
que `CP-14` existe para pegar.

---

## Grupo 1 — Domínio e banco

| # | Tarefa | Verificação |
|---|---|---|
| **T-152** | VOs `GrauDeCurso`, `ModalidadeDeCurso`, `SituacaoCurso`, `Portaria` | `CU-04`: grau "mestrado" e modalidade "hibrida" → 400 `VALOR_INVALIDO` |
| **T-153** | VO **`Vigencia`** com `SituacaoEm(hoje DataLocal)`, escrito com `!fim.AntesDe(hoje)` — **sem `>` que alguém possa trocar por `>=`** | `DG-06` em unidade: 31/07 vigente, 01/08 encerrada, `fim` nulo vigente |
| **T-154** | `CodigoEMec` **reutilizado** de `autenticacao-usuarios`, não recriado | Nenhum VO novo com esse nome no repositório |
| **T-155** | Entidades `Curso` e `Designacao`. **Nenhuma coluna de coordenador em `Curso`** | Revisão: `Curso` não tem campo `CoordenadorID` |
| **T-156** 🗄️ | Migration `000005_cursos` parte 1: `btree_gist`, tabelas `curso` e `designacao`, `EXCLUDE USING gist` com `'[]'` e `COALESCE(data_fim,'infinity')`, índices únicos compostos de F-05, índices de listagem e o de `coordenador_id` | `dba` valida V-1 a V-4 e V-8 de `design.md` §4.4 |
| **T-157** 🔒 | `postgres.FragmentoDesignacaoVigente(alias, n)` — **a única função que escreve o predicado de vigência em SQL** | `TestVigencia_RotuloEPredicadoConcordam` passa nas três fronteiras de `DG-06`, **com o container em UTC** |
| **T-158** | `Alvo` ganha `AlvoCurso` e `AlvoDesignacao`, com `temColunaCurso` | `TestAlvos_ExcecaoDoCatalogoEmExatamenteUm` continua passando (os dois alvos novos **não** declaram a exceção) |
| **T-159** 🗄️ | Seed: os seis cursos e as sete designações da seção 7 da spec, **com as pessoas sem `coordenador_curso` no conjunto atribuído** | Sem Beatriz coordenando, `CP-08`/`CP-09`/`CP-11` não são observáveis; sem a designação futura, `DG-02`/`CP-13` não são; sem a que vence em 31/07, `CP-05`/`DG-06` não são |

---

## Grupo 2 — Perfil derivado ⚠️

| # | Tarefa | Verificação |
|---|---|---|
| **T-160** | `CarregarContextoDeSessao` ganha, **na mesma consulta**, o `EXISTS` de designação vigente e `cursos_coordenados`. **Sem junção com `curso`** | `CP-06`: designação vigente de curso **inativo** mantém o perfil. `dba` valida V-5: sem varredura sequencial, consulta de sessão em p95 < 20 ms |
| **T-161** 🔒 | `montarConjuntoEfetivo(atribuidos, coordenaHoje)` — **a única função do projeto que acrescenta `CoordenadorCurso` a um conjunto** | `TestConjuntoEfetivo_UnicaFonteDoPerfilDerivado`: `grep` por `CoordenadorCurso` em contexto de acréscimo encontra **um** lugar |
| **T-162** 🔒 | `CP-14`: um teste que faz a autorização ler só o conjunto atribuído faz **coordenador legítimo receber 403** | O teste falha quando a leitura é sabotada — comprovar e desfazer |
| **T-163** | `GET /auth/eu` passa a devolver `perfis` (efetivo), **`perfis_derivados`** e **`cursos_coordenados`** | `CP-01`: o conjunto atribuído do CRUD **não** contém Coordenador; o de `/auth/eu` contém. Mudança **aditiva** |
| **T-164** | `ConjuntoInstitucional` e `Perfil.PodeAtribuir` recusam `coordenador_curso` → 403 `PERFIL_NAO_ATRIBUIVEL` | Tentar atribuir pelo `PUT /usuarios/{id}` responde 403, nunca 400 e nunca silenciosamente |
| **T-165** | `CP-02`, `CP-03`, `CP-04` em unidade, com o relógio injetado | Coordenador **acrescenta**; encerrar a última o remove e **não altera nenhuma linha de `usuario_perfil`**; sair de uma entre três não remove |

---

## Grupo 3 — Carteira dentro do `Escopo` 🔒

| # | Tarefa | Verificação |
|---|---|---|
| **T-166** 🔒 | `AplicarEscopo` emite o `EXISTS` de carteira usando `FragmentoDesignacaoVigente`, para alvos com `temColunaCurso` | `TestAplicarEscopo_CarteiraUsaOFragmentoUnico`: a cláusula é byte-a-byte a saída da função |
| **T-167** 🔒 | `CursosDaCarteira` liga `restritoACarteiraDe` **sempre**; o teste que percorre os alcances confere | Nenhum alcance de carteira devolve `Escopo` sem a restrição |
| **T-168** | `GET /api/v1/meus-cursos` — cursos **ativos** com designação **vigente** do ator | `CV-01`: curso de outro coordenador responde **404**, nunca 403 |

---

## Grupo 4 — Backend de curso e designação

| # | Tarefa | Verificação |
|---|---|---|
| **T-169** | Portas e adapters de `Curso` e `Designacao` — **todo método com `Escopo`** | T-117 estendido passa |
| **T-170** | Use cases de curso: criar, atualizar, alterar situação, excluir | `CU-01` a `CU-03`, `CU-06` a `CU-08` |
| **T-171** | Exclusão de curso bloqueada por vínculo, com `FOR UPDATE` na linha antes do `EXISTS` | `CU-07`: 409 `CURSO_COM_VINCULO`; a janela entre verificar e gravar está fechada |
| **T-172** | Inativar curso **não toca** em designação, plano, entrega nem perfil | `CU-06`: nenhuma linha dessas tabelas é alterada; a designação continua vigente |
| **T-173** | Use case de criar designação, com `autodesignacao` computada no ato e **gravada** | `CP-09`: o registro traz `autodesignacao` verdadeiro; não há recusa por ser a própria pessoa |
| **T-174** | Use case de atualizar designação com a tabela de §5.3: `futura` editável; `vigente`/`encerrada` recusam `coordenador_id` e `data_inicio` → **409 `DESIGNACAO_COM_EFEITO`** | Os três estados testados; **409, não 400** |
| **T-175** | Excluir designação só quando `futura` | `DG-08`: vigente e encerrada → 409 |
| **T-176** 🔒 | `DG-03` em integração, com **escritas concorrentes reais** | Duas transações simultâneas com vigências que se tocam: exatamente uma tem êxito, e a garantia vem do `EXCLUDE`, não da aplicação |
| **T-177** | `GET /api/v1/designacoes/candidatos` — "possui algum perfil além de aluno", na instituição da sessão | `CP-07`: traz Professor, PI e quem já coordena; **não** traz só-aluno nem o Administrador; identificador de outra instituição → **404** |
| **T-178** | `GET /api/v1/cursos` com `coordenador` por lateral: `nome`, `data_fim`, `tambem_pesquisador_institucional`; e `outras_designacoes_vigentes` na designação | `dba` valida V-6: **uma** consulta, não uma por linha |
| **T-179** | Handlers, rotas, DTOs e anotações OpenAPI escritas para o consumidor | Rota sem permissão declarada não compila |
| **T-180** | Smoke de cada rota | 401, 403, 404 — com o Administrador recebendo **403**, nunca 404 (`CV-04`) |

---

## Grupo 5 — Migration de normalização ⚠️

| # | Tarefa | Verificação |
|---|---|---|
| **T-181** 🗄️⚠️ | Migration `000005_cursos` parte 2, **nesta ordem**: auditoria de normalização (uma linha por usuário afetado) → `DELETE FROM usuario_perfil WHERE perfil = 'coordenador_curso'` → `CHECK` estreitado | `CP-15`: com linhas presentes, o perfil sai do conjunto atribuído de todos; quem tem designação vigente **continua coordenador pela derivação**; quem não tem mantém os demais perfis; a auditoria tem uma linha por pessoa, **nunca um rebaixamento por pessoa** |
| **T-182** 🗄️ | `down` que recria o `CHECK` largo e as tabelas, **com a limitação declarada no arquivo**: as linhas apagadas não voltam | `migrate down` roda; o comentário do arquivo diz que a reversão não restaura linhas, e que reverter a imagem **não exige** reverter o banco |

**A ordem de T-181 é parte da decisão**, não detalhe: auditar antes de apagar
(depois não há o que registrar) e o `CHECK` depois do `DELETE` (a restrição não
é aceita com linhas que a violam).

**Achado do Grupo 1, registrado para não ser esquecido no Grupo 5:** o seed
de desenvolvimento (`cmd/seed/dev.go`) mantém João Ribeiro com
`coordenador_curso` **atribuído diretamente** no IVV — de propósito, e
diferente de Ana Lima (FSA), que já perdeu a atribuição direta neste Grupo 1
porque a seção 7 da spec cria uma designação real para ela. O IVV não tem
nenhum curso nem designação no seed desta feature (só a FSA está na seção 7),
então não existe caminho de derivação para repor o que João perderia — e
`e2e/auth/mesmo-email-duas-instituicoes.spec.ts` afirma explicitamente que ele
vê "Coordenador de Curso" no menu ao entrar no IVV. Remover a atribuição dele
agora quebraria esse E2E sem nenhuma alternativa. Quando T-181 rodar, o
`DELETE FROM usuario_perfil WHERE perfil = 'coordenador_curso'` **também**
remove a linha dele — nesse momento, decidir entre (a) dar ao IVV um curso e
uma designação vigente para ele no seed, ou (b) atualizar o E2E para não
depender mais desse perfil nesta conta. Os dois exigem tocar o seed e o teste
no mesmo commit que aplica T-181, nunca antes.

---

## Grupo 6 — Frontend

| # | Tarefa | Verificação |
|---|---|---|
| **T-183** | Menu: **Cursos no grupo Administração** | Inalterado em relação ao `ux.md` e à visão |
| **T-184** | Tela `/app/cursos` — coluna Coordenador com "Vago ⚠", "até DD/MM" e "também é PI ⓘ", **todos em texto** | Excluir **escondido** quando há vínculo |
| **T-185** | Tela `/app/cursos/{id}/designacoes`, com "ⓘ autodesignação" em texto e `[✗]` só em futura | Campos desabilitados conforme §5.3 — **e o backend recusa** (T-174) |
| **T-186** | Modal de nova designação com o aviso informativo de acúmulo de papéis | `role="status"`, **não** `role="alert"` — é revelação, não repreensão, e o botão continua habilitado |
| **T-187** | Modal de encerrar designação com o texto afirmativo quando `outras_designacoes_vigentes == 0` | "Ana deixa de ter o perfil..." afirmativo, não hedged |
| **T-188** | Avisos de mudança de perfil (3.9): comparação de `perfis_derivados` na troca de rota e no retorno de foco; os dois textos | Ganho: `role="status"`. Perda dentro de `Minhas metas`: leva à tela inicial **com a mensagem** — nunca tela em branco, nunca 403 seco |
| **T-189** | `ComboboxEntidade` com `badge` **reutilizado** de `indicadores` | Nenhuma segunda implementação |

---

## Grupo 7 — E2E

| # | Tarefa | Verificação |
|---|---|---|
| **T-190** | `e2e/designacao-e-perfil.spec.ts` conforme `design.md` §10.3 — o relógio não se move, **o dado se move** | Os cinco passos passam; o passo 3 prova o perfil derivado e o aviso funcionando juntos |

---

## Grupo 8 — Depois do teste manual do dono

| # | Tarefa | Quem |
|---|---|---|
| **T-191** | `testes-pendentes.md` com os adiados de §10.2, **`CU-12` incluído** com a nota de que depende de `plano-acao` | `dev-fullstack` |
| **T-192** 🗄️ | Consolidar as migrations da feature em uma | `dba` |
| **T-193** | Revisão de segurança + qualidade, em paralelo entre si | `security-reviewer`, `code-reviewer` |
| **T-194** | `evidence.md` com `C \ (A ∪ B)` e `A ∩ B` vazias | `qa-tester` |

---

## Discordâncias registradas

Nenhuma discordância do `dev-fullstack` até aqui. As divergências entre
documentos estão em `design.md` §12 — em especial a fronteira exata da
imutabilidade da designação (§5.3), que reconcilia a decisão 5 do dono com o
`ux.md`.
