# Tasks: plano-acao

**Data:** 28/09/2026 (revisão 2 — QP-3 respondida) · **Lei da construção:**
`design.md` desta pasta e `specs/_fundacao-metas.md`.
**Legenda:** 🗄️ `dba` · 🐳 `dev-docker-compose` · 🔒 teste de mecanismo ·
🆕 acrescentado ou alterado pela revisão 2

> **Depende de `cursos` concluída** (curso, designação, carteira no `Escopo`) e
> de `indicadores` (meta). **MinIO entra no ambiente nesta feature** — o
> documento `.docx` chega antes dos anexos, e a infraestrutura vem com ele.
>
> **Revisão 2:** QP-3 respondida pelo dono. Mudaram **T-204**, **T-215**,
> **T-231**, **T-232**, **T-234** e **T-241**; entraram **T-231b**, **T-233b**,
> **T-241b** e **T-242b**; **saiu** a metade de `DO-05` que exigia 404 no
> rascunho. **`QP-6` de `design.md` §11.1 bloqueia o fechamento do critério de
> aceitação**, não a construção.

---

## Ordem de execução

```
Grupo 1  — infraestrutura: MinIO   🐳   (pode começar em paralelo ao Grupo 2)
Grupo 2  — domínio e banco
Grupo 3  — backend: período, plano, itens
Grupo 4  — cópia em lote
Grupo 5  — documento .docx
Grupo 6  — frontend
Grupo 7  — E2E
⛔ PARE — o dono testa
Grupo 8  — dba consolida, revisão, QA
```

Os Grupos 1 e 2 são independentes e podem ser tocados ao mesmo tempo. Os
Grupos 4 e 5 dependem do 3 e são independentes **entre si** — se houver dois
pares de mãos, é aqui que eles se separam.

---

## Grupo 1 — MinIO 🐳

| # | Tarefa | Verificação |
|---|---|---|
| **T-195** 🐳 | Serviço `minio` no `docker-compose.dev.yml` (9000 API, 9001 console), com volume nomeado. **Console não publicado no compose de produção** | `docker compose -f docker-compose.dev.yml up` sobe o MinIO; o compose de produção não expõe 9001 |
| **T-196** 🐳 | Serviço de um tiro que cria o *bucket* na subida, com `depends_on ... service_completed_successfully`, no padrão do `migrate` | Ambiente novo sobe com o bucket pronto, sem passo manual |
| **T-197** 🐳 | `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` no `.env.example` **sem valores** e no compose de desenvolvimento | `.env.example` versionado mostra as chaves, nunca os valores |
| **T-198** | `port.ArmazenamentoDeObjetos` e o adapter S3. **Nenhum método recebe `uuid.UUID`** | `TestNenhumaPortaNovaSemEscopo` continua verde **sem** ampliar a lista fechada das duas portas estreitas |
| **T-199** 🐳 | `/readyz` verifica o MinIO, **sem devolver mensagem de driver** | Com credencial inválida, o corpo traz `"indisponivel"` e **não** contém endereço, usuário, porta nem mensagem do SDK; o log do servidor contém |

---

## Grupo 2 — Domínio e banco

| # | Tarefa | Verificação |
|---|---|---|
| **T-200** | VOs `OrgaoDeAprovacao`, **`Aprovacao`** (par indivisível), **`Quantidade`** (≥ 1, tipo próprio), `SituacaoPlano` | `PL-03` cai de graça: não existe construtor que aceite data sem órgão |
| **T-201** | `Periodo` **reusando `Vigencia`** de `cursos` — sem coluna de situação | `PE-03`: as três situações derivam das datas |
| **T-202** | `Plano` com `SituacaoEfetiva(periodo, hoje)` e `SemAprovacao(...)` — **as únicas funções que decidem** | `SI-11`, `SI-07`: rascunho sem aprovação **não** recebe aviso |
| **T-203** | `ItemDoPlano` com `CursoID` e `InstituicaoID` denormalizados | FK compostas de F-05 presentes |
| **T-204** 🗄️🆕 | Migration `000006_plano_acao`: `periodo`, `plano`, `item_plano`, `documento`; o **índice único parcial `(curso_id, periodo_id)`**; os índices compostos de F-05; os `CHECK` de aprovação, encerramento e textos. Coluna chamada **`situacao_publicacao`**. **Sem `plano.documento_id`** (P-11) | `dba` valida V-1 a V-9 de `design.md` §4.2; confirma que **não existe** coluna `situacao` **nem `documento_id`** em `plano`, e que `idx_documento_plano (plano_id, gerado_em DESC)` está presente — é ele que sustenta a lateral que substituiu a coluna |
| **T-205** | `Alvo` ganha `AlvoPeriodo`, `AlvoPlano`, `AlvoItemPlano`, `AlvoDocumento` | `TestAlvos_ExcecaoDoCatalogoEmExatamenteUm` continua passando |
| **T-206** 🗄️ | Seed: os três períodos e os cinco planos da seção 7 da spec | Sem o plano em rascunho, `SI-01`/`SI-07`/`DO-01` não são observáveis; sem os dois sem aprovação, `SI-05` não é; sem o de Pedagogia (vago), `SI-08`/`DO-03` não são |

---

## Grupo 3 — Backend de período, plano e itens

| # | Tarefa | Verificação |
|---|---|---|
| **T-207** | Portas e adapters — **todo método com `Escopo`** | Guarda estendido passa |
| **T-208** | Tradução do filtro de situação efetiva para o par (`situacao_publicacao`, predicado de data), **em uma função só** | `grep` por `'encerrado'` em `WHERE` não encontra nada: o valor não existe no banco |
| **T-209** | Use cases de período, com exclusão bloqueada por plano | `PE-02`, `PE-05`, `PE-06` |
| **T-210** 🔒 | `PL-02` em integração: o **segundo** plano do mesmo curso e período é recusado, inclusive em transações concorrentes | A garantia é o índice único, não a verificação prévia |
| **T-211** | Publicar, despublicar, encerrar e reabrir, com as pré-condições de §5.1 e `FOR UPDATE` antes do `EXISTS` de entrega | `SI-02`, `SI-04`, `SI-09`, `SI-10`, `SI-12`, `SI-13`, `SI-14`. **`SI-04`: publicar sem aprovação responde 200** e não existe erro de "plano não aprovado" |
| **T-212** | Use cases de item com a tabela de §5.2 | `IT-01`, `IT-02`, `IT-05`, `IT-07`, `IT-09`, `IT-10` |
| **T-213** | **Edição de item bloqueada com o plano encerrado por qualquer motivo** → 409 `PLANO_ENCERRADO_PARA_EDICAO`; editar os textos do plano continua permitido | Encerramento **antecipado** também bloqueia, e o backend recusa — não só a tela |
| **T-214** | Consulta de planos com `metas`, `total_exigido` (único `SUM`, de itens), `sem_aprovacao` e `tem_entrega` | `SI-05`: o grid mostra "sem aprovação" exatamente nos planos vigentes/encerrados sem os dados |
| **T-215** 🆕 | `GET /api/v1/meus-planos` com alcance `PlanosDaCarteira`, **em todas as situações, somente leitura** (P-10). O antigo `SomenteVigentes` **sai desta consulta** | Coordenador vê os planos em rascunho dos cursos dele, **sem nenhuma ação de escrita**; plano de curso fora da carteira → **404**; escrita → **403** (`VI-04`) |
| **T-216** | Handlers, rotas, DTOs, anotações OpenAPI para o consumidor | Rota sem permissão não compila |
| **T-217** | Smoke de cada rota | 401, 403, 404; Administrador do Sistema recebe **403** |

> **Atenção a T-215:** o `SomenteVigentes` **não some do projeto** — ele
> **muda de lugar**, para a consulta de obrigações (`minhas-metas`, tarefa
> T-270 de `metas-coordenacao`). Apagá-lo dos dois lugares faria plano em
> rascunho gerar obrigação e oferecer "Prestar contas" numa rota que responde
> 409. Ver `design.md` §5.4.

---

## Grupo 4 — Cópia em lote

| # | Tarefa | Verificação |
|---|---|---|
| **T-218** | `GET /planos/{id}/destinos-copia` — **uma** consulta, com `ja_tem_plano` e `vago` por curso; cursos inativos **não** oferecidos | `dba` valida V-5: uma ida ao banco, não uma por curso. `CP-09`, `CP-10` |
| **T-219** | Use case de cópia: **uma transação por plano**, `INSERT` multi-linha dos itens, auditoria na mesma transação | `CP-12`: N registros de auditoria, **nunca um do lote inteiro**. `dba` valida V-6 |
| **T-220** 🔒 | **Violação do índice único tratada como resultado esperado**: o curso vira "pulado", nunca substituído; os demais permanecem | `CP-05` **com corrida real**: criar o plano entre a montagem da tela e a confirmação; `CP-07`: falha no sétimo não desfaz os seis |
| **T-221** 🔒 | `CP-02` e `CP-03`: a cópia **não** leva entregas, anexos, avaliações, prazos, **os dados de aprovação** nem a situação; origem e cópia são independentes | É o defeito silencioso da feature: copiar a aprovação atribuiria a nove planos uma aprovação institucional que não aconteceu |
| **T-222** | Limites do lote: nenhum curso → 400 `CURSOS_OBRIGATORIOS`; acima de 100 → 400 `LOTE_ACIMA_DO_LIMITE` | `CP-11` |
| **T-223** | Resposta **200** com `{ criados, pulados: [{curso, motivo}] }` | `CP-06`: nunca apenas "sucesso"; nunca 207 |

---

## Grupo 5 — Documento `.docx`

| # | Tarefa | Verificação |
|---|---|---|
| **T-224** | Modelo `.docx` versionado no repositório, com os marcadores **em `run` único e sem formatação** e a linha-modelo da tabela entre `{{#itens}}` e `{{/itens}}` | Abrir o modelo no Word e conferir que nenhum marcador está partido entre `runs` |
| **T-225** | `port.GeradorDeDocumento` e adapter sobre `archive/zip` + `encoding/xml`. **Sem biblioteca, sem container** | Nenhuma dependência nova no `go.mod` |
| **T-226** 🔒 | **Falha alta quando um marcador declarado não é encontrado** | Teste renderiza o modelo versionado e afirma **zero `{{` no resultado**; remover um marcador do mapa faz o teste falhar com erro, não com documento torto |
| **T-227** 🔒 | `xml.EscapeText` em toda substituição; quebra de linha vira `<w:br/>` | Texto com `&`, `<` e `\n` produz `.docx` que abre, com os parágrafos separados |
| **T-228** | Clonagem de `<w:tr>` por item, com nome da meta, **todos os indicadores** (código, nome, origem) e a quantidade | `DO-02` |
| **T-229** | Marcas e aprovação: `RASCUNHO`, `ENCERRADO — período encerrado em DD/MM/AAAA`, nada em vigente; `Não aprovado` com a observação **só quando vigente** | `DO-01`, `SI-05`, `SI-07`. **Vale igual para o documento gerado pelo coordenador** |
| **T-230** | Coordenador responsável na data de emissão, com a portaria — ou **"Sem responsável"** | `DO-03`; vem de `FragmentoDesignacaoVigente`, como tudo o mais |
| **T-231** 🆕 | Gravação: **objeto primeiro, linha depois**, dentro da transação; remoção compensatória *best-effort*. **Gerar documento é `INSERT` em `documento` e nada mais — não escreve em `plano`** (P-11) | `DO-06`: duas gerações, dois objetos, o anterior **não** sobrescrito |
| **T-231b** 🔒🆕 | Teste do 409 espúrio que P-11 eliminou | PI abre o plano na `versao` 1 · coordenador gera o documento · PI salva → **200, nunca 409**. Reintroduzir `plano.documento_id` com `UPDATE` que toca `versao` **faz o teste falhar** |
| **T-232** 🆕 | `POST /planos/{id}/documentos` e `GET /documentos/{id}/conteudo` **sem ramificação por situação do plano**: o único recorte é o `Escopo` — instituição para o PI, instituição **e carteira** para o coordenador | `DO-04` **ampliado**: coordenador gera o documento de um plano **em rascunho** do curso dele → **200**, arquivo marcado RASCUNHO; coordenador pedindo documento de curso **fora da carteira** → **404**; sem sessão → 401; **nenhuma resposta contém a URL do bucket**, em campo nenhum |
| **T-233** 🆕 | Geração e download auditados, **com quem foi** | `gerado_por` distingue PI de coordenador — depois de QP-3 deixa de ser formalidade |
| **T-233b** 🆕 | `ultimo_documento` no **detalhe** do plano, por `LEFT JOIN LATERAL` de um registro. **Não aparece na listagem** | `dba` valida V-9: sem varredura sequencial em `documento`; `null` quando nunca houve geração |

---

## Grupo 6 — Frontend

| # | Tarefa | Verificação |
|---|---|---|
| **T-234** 🆕 | Menu: **Períodos e Planos no grupo Metas**, junto de **Indicadores** e **Catálogo de metas** | O item do catálogo chama-se **"Catálogo de metas"**, não "Metas" — "Metas → Metas" confundia |
| **T-235** | Telas de período (grid + modal), com o cuidado da data de fim inclusiva no texto | `PE-03` visível |
| **T-236** | Grid de planos com a coluna Aprovação em **texto** e o filtro "Sem aprovação" | `SI-05`: o filtro lista exatamente os planos pendentes |
| **T-237** | Página do plano: aviso persistente de aprovação pendente com `role="status"`, **nunca `role="alert"`** | É informativo, não erro. Some sozinho quando os dois campos são preenchidos (`SI-06`) |
| **T-238** | Seção "Metas do plano" com os indicadores em lista e a nota de que a quantidade não é multiplicada por eles | `IT-03`; `[✗]` some quando o item tem entrega |
| **T-239** | Página de cópia em lote com **`Checkbox` reais**, os já-com-plano desabilitados com o motivo em `aria-describedby`, e o contador anunciado | `CP-05` na tela; a seleção em massa respeita os desabilitados |
| **T-240** | Modal de resultado da cópia, com a lista e o motivo de cada pulado, anunciado com papel de alerta | `CP-06` |
| **T-241** 🆕 | Botão de documento com `LoadingButton`, `aria-busy`, e a **nota contextual ao lado** conforme a situação — **também na visão do coordenador** | Nenhum modal; é contexto, não confirmação. A nota do rascunho **deixa de dizer que o documento é material de trabalho do PI** |
| **T-241b** 🆕 | Na lista do coordenador, o plano em rascunho aparece **com a situação em texto e sem ações de escrita**, com a explicação de por que não há o que fazer ali | Plano sem ações e sem explicação parece quebrado — é a terceira consequência de `design.md` §5.4 |

---

## Grupo 7 — E2E

| # | Tarefa | Verificação |
|---|---|---|
| **T-242** | `e2e/copia-em-lote.spec.ts` conforme `design.md` §10.3 | Os cinco passos passam; **o passo 4 falha se alguém copiar a aprovação ou a situação** |
| **T-242b** 🆕 | Acréscimo por QP-3: coordenador abre um plano **em rascunho** do curso dele, gera o documento e recebe o arquivo marcado RASCUNHO; o mesmo coordenador pedindo o documento de curso que não coordena recebe não encontrado | Pode entrar no arquivo de `cursos` ou em arquivo próprio |

---

## Grupo 8 — Depois do teste manual do dono

| # | Tarefa | Quem |
|---|---|---|
| **T-243** 🆕 | `testes-pendentes.md` com os adiados de §10.2, **incluindo `DO-05` reescrito** | `dev-fullstack` |
| **T-244** 🗄️ | Consolidar as migrations da feature em uma | `dba` |
| **T-245** 🆕 | Revisão de segurança + qualidade, em paralelo entre si. **Ponto de atenção para o `security-reviewer`:** `QP-6` — se o dono optar pela leitura estreita, o recorte de leitura do plano passa a ser aparência, não proteção | `security-reviewer`, `code-reviewer` |
| **T-246** 🆕 | `evidence.md` com as duas diferenças vazias, **com `C` recalculado depois de `QP-6` fechada** | `qa-tester` |

---

## Bloqueio registrado

**`QP-6` (`design.md` §11.1) não bloqueia a construção, mas bloqueia o
fechamento do critério de aceitação.** A resposta a QP-3 foi aplicada em §3.5
da `spec.md` e contradiz sete passagens que não foram alteradas — incluindo
`DO-05`, que diz literalmente o oposto da decisão, e `EN-03` de
`metas-coordenacao`. Enquanto o `analista-requisitos` não reconciliar, o
conjunto `C` de cenários está inconsistente e `A ∪ B = C` não é verificável.

## Discordâncias

Nenhuma do `dev-fullstack`.
