# Tasks: autenticacao-usuarios

**Data desta versão:** 28/09/2026 (revisão 5)
**Lei da construção:** `design.md` (revisão 6) — a lista executável completa,
com verificação por item, está em **`design.md` §16**.

> **Estado:** feature implementada, suíte verde, migrations consolidadas —
> e, a pedido do dono do produto em 28/09/2026, **unificadas num único
> arquivo** (`000001_autenticacao_usuarios.{up,down}.sql`): o no-op
> `000001_init` foi absorvido, porque o motivo dele (pasta vazia derruba o
> `migrate`) deixou de existir a partir do momento em que há migration real.
> `up`→`down`→`up` limpo, schema idêntico por `pg_dump --schema-only`, e o
> ponteiro de `schema_migrations` de `dev`, `test` e `e2e` realinhado por
> `migrate force` (sem executar SQL) — nenhum dado de desenvolvimento
> perdido.
> Revisões **T-077 (segurança)** e **T-078 (qualidade)** concluídas.
> **14 correções decididas pelo arquiteto**, abaixo. Nenhuma altera
> comportamento de negócio observável.
>
> **Atualização de 28/09/2026: as 14 correções (T-096 a T-109) estão
> implementadas e verificadas** — ver a coluna "Status" nas duas tabelas
> abaixo, cada uma com a evidência concreta (a maioria com comprovação
> negativa, não só teste passando). `go build`/`go vet`/`govulncheck`
> limpos; suíte Go completa (unitária + integração contra o banco real)
> verde; `tsc`/`eslint`/`biome`/`knip`/`dependency-cruiser`/`vitest`/
> `next build` limpos; suíte E2E revalidada. Falta só **T-079
> (`qa-tester`)**, que só começa depois desta confirmação.

---

## Ordem de execução

```
1. T-096  (bump de dependência)     ← PR PRÓPRIO, sem nenhuma feature junto
2. T-101 e T-102                    ← mecanismo: fail-open latente e guardas
3. T-098, T-099, T-100, T-103       ← borda HTTP (podem ir em paralelo)
4. T-097, T-104 a T-108             ← demais, em paralelo
5. T-109                            ← automação de dependências
6. T-079 qa-tester                  ← depois de tudo verde
```

**Por que T-096 isolada:** CLAUDE.md — *nunca atualizar dependência junto com
feature no mesmo PR*. Quando algo quebra, ninguém sabe qual dos dois foi.

**Por que T-101 e T-102 logo depois:** são **mecanismo**, não cobertura. T-101
fecha um fail-open latente na função que o desenho declara ser o único ponto de
isolamento; T-102 faz os guardas entregarem a garantia que o documento
descrevia. Enquanto não entram, a propriedade central do desenho está apoiada
em coincidência e em revisão humana.

---

## Bloqueantes da entrega (decididos)

| # | O quê | Achado | Decisão em | Status |
|---|---|---|---|---|
| **T-096** | Bump de `quic-go` para **v0.59.1**, em PR próprio | 🔴 C-1 | §14, §16 | ✅ `go.mod`/`go.sum` isolados dos demais; `govulncheck ./...` limpo (0 vulnerabilidades no código); suíte Go completa verde |
| **T-097** | Remover os **2 comentários que descrevem mecanismo de defesa** realocando o conteúdo para `design.md`, e os **3 curtos** dando nome ao que explicavam. **Manter** diretivas de lint e `components/ui/` | 🔴 C-2 | §11.6 | ✅ Removidos de `use-guarda-senha-provisoria.ts` e `api-client.ts`, conteúdo em `design.md` §11.8. Os 3 curtos viraram `decodificarValorArmazenado`, `lerUltimaInstituicaoSalva`, `lerJsonSeguro`. Débito de Biome estável em 48 (nada novo) |
| **T-098** | `middleware_seguranca.go` na borda da API + `Vary: Origin` no CORS | 🟡 A-1, O-2 | §4.5 | ✅ `curl -D-` confirma os 5 cabeçalhos + `Vary: Origin`; HSTS só com `APP_ENV=production` (testado nos dois ambientes) |
| **T-099** | Swagger **só fora de produção**; anotações reescritas para o consumidor; **DTO de requisição por alcance** | 🟡 A-2 | §4.5 | ✅ `/swagger/*` → 404 em produção (testado); `doc.json` sem `design.md`/`CLAUDE.md`/`T-0`/`allowlist`/`ignorado`; `perfis` fora de `PesquisadorRequest`/`AdministradorRequest`; `rotas.Publica` com 1 ocorrência |

## Não bloqueantes, decididos e a executar

| # | O quê | Achado | Status |
|---|---|---|---|
| **T-100** | Porta administrativa para `/metrics`, `/healthz`, `/readyz` 🐳 | 🟠 M-1 | ✅ Porta pública devolve 404 para as três; porta administrativa (9090) devolve 200 de dentro da rede; não publicada no host em nenhum compose |
| **T-101** | `AplicarEscopo` emite `instituicao_id IS NULL` no ramo de plataforma, com teste que fixa a cláusula | 🟠 M-3 | ✅ `TestAplicarEscopo_PlataformaExigeInstituicaoNula` — comprovação negativa feita: falha sem a cláusula, passa com ela |
| **T-102** | Guardas de §4.4 passam a fixar **assinatura**, não só nome; detectam `*uuid.UUID` | 🔵 O-1 | ✅ Comprovação negativa feita: trocar o tipo de `BuscarCredencial` quebra `TestPortasEstreitas_ListaFechada`; desfeito e confirmado verde |
| **T-103** | `/readyz` sem mensagem de driver | 🟠 M-2 | ✅ `TestReadiness_CredencialInvalida_CorpoNaoVazaDetalheDoDriver` — corpo sem usuário/senha/IP/porta/SQLSTATE; log do servidor contém tudo |
| **T-104** | `USER` não-root no `Dockerfile` do backend 🐳 | 🟠 M-4 | ✅ `adduser -u 1000 appuser` + `USER appuser`; `COPY --chown` |
| **T-105** | Descrição do produto em `layout.tsx` + varredura de resíduos de escopo antigo | 🔵 elevado | ✅ Descrição corrigida; varredura encontrou 1 resíduo real (`page.tsx`, corrigido) e 1 falso positivo verificado (`"ata"` no seed é o dado literal de S-01, não resíduo) |
| **T-106** | Comentário de `escopo_sql.go` nomeia a exceção sancionada — **corrigir, não apagar** | 🔵 | ✅ `ContarDetentoresDoPerfil` nomeada explicitamente, com referência a §5.8 |
| **T-107** | CSP do frontend ganha `base-uri` e `form-action` | 🔵 O-5 | ✅ Presentes nos dois ramos (dev/produção) de `proxy.ts` |
| **T-108** | Escapar `%`, `_` e `\` no termo de busca | 🔵 O-4 | ✅ `escaparCuringasLike` + `TestUsuarioRepository_O4_CuringasDaBuscaSaoEscapados` — comprovação negativa feita |
| **T-109** | Renovate ou Dependabot no repositório 🐳 | C-1 (recorrência) | ✅ `.github/dependabot.yml` — gomod/npm (frontend, e2e, raiz)/docker, patch+minor agrupados semanalmente, major em PR separado. Abertura do primeiro PR depende do repositório estar no GitHub — fora do que dá para verificar localmente |

**Critério de verificação de cada uma: `design.md` §16.** Cada linha traz a
verificação concreta — e três delas pedem a **comprovação negativa** (remover a
cláusula, trocar o tipo, usar credencial inválida) para provar que o teste
testa alguma coisa.

## Sem achado — registrado para não voltar

| Item | Decisão |
|---|---|
| Senhas fictícias do seed em arquivo versionado (O-3) | **Ficam.** Pessoas inventadas, portão `APP_ENV=development` fechado, administrador de produção só por variável. Trocar dificultaria o ambiente sem proteger nada — §9 |
| `auditoria` append-only só por convenção (O-6) | **Fica.** `REVOKE` no mesmo papel da aplicação é reversível por quem já o tem. A garantia real é o **canal 2 (syslog)**, fora do banco. Gatilho: separação de papéis de banco — §8 |

---

## Fase de revisão

- [x] **T-077 · `security-reviewer`** — 2 Críticos, 2 Altos, 4 Médios, 6
  observações. **Nenhum em autenticação, autorização ou isolamento**; os quatro
  mitigantes dos riscos aceitos **não regrediram**. Relatório em
  `security-review.md`.
- [x] **T-078 · `code-reviewer`** — achados em `evidence.md`, seção "Achados de
  qualidade". Garantias estruturais verificadas item a item e de pé.
- [ ] **T-079 · `qa-tester` → `evidence.md`** — **depois** das correções acima.
  Aplica o critério de conjuntos da spec 15.3: para cada cenário da seção 7,
  **ou** teste automatizado que passa, **ou** linha em `testes-pendentes.md` —
  nunca os dois, nunca nenhum. Registra os conjuntos `esquecidos` e
  `duplicados`, **ambos vazios para aceitar**.
  **Registrar também:** os testes E2E existentes **excedem** o que a entrega
  pedia; precisam entrar em `evidence.md` e na lista da fase de Release **para
  não serem escritos duas vezes**.

## Fechamento — confirmar o que ficou pronto

- [ ] **T-070** CSP e cabeçalhos do frontend 🐳 · **T-071** `/version` e
  `VERSION` 🐳 · **T-072** anotações OpenAPI (**absorve T-099**) ·
  **T-074** Biome, commitlint, knip, dependency-cruiser 🐳 ·
  **T-075** `analista-requisitos` atualiza a `spec.md`

> ⛔ **PARE** — aprovação para produção é do dono do produto. Só depois dela
> entra a fase de Release: E2E dos fluxos críticos (aproveitando o que já
> existe), `release.sql`, snapshot do `openapi.yaml`, `CHANGELOG.md`, `VERSION`.
