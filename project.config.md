# Configuração do Projeto

> Preencha este arquivo uma única vez, no início do projeto. Todos os
> subagents leem este arquivo antes de agir — ele é a fonte única de
> verdade sobre qual tecnologia usar. Mudar de stack em um novo projeto
> significa editar apenas este arquivo, não os subagents.

**Projeto:** basis-avalia — acompanhamento das metas da coordenação, apoiado
nos instrumentos de avaliação do INEP.
**Escopo da v1:** autenticação e usuários, **cursos** e **relatório de metas da
coordenação** (ver `specs/00-visao-produto.md`).
**Multi-institucional:** uma instalação atende várias instituições, com
isolamento por coluna e banco único.

## Backend
- **Linguagem:** Go
- **Framework HTTP:** Gin
- **Biblioteca de acesso a dados:** sqlx
- **Skill correspondente:** `.claude/skills/backend-go/SKILL.md`
- **Subagent correspondente:** `dev-fullstack` (único agente de implementação,
  agnóstico de stack — carrega a skill correspondente antes de codificar)

## Frontend
- **Framework:** Next.js
- **Biblioteca de componentes:** shadcn/ui
  - Sistema **não é** portal público de governo federal — DSGOV não se aplica
- **Editor de texto rico:** **não se aplica ao escopo atual.** Nenhum campo das
  features especificadas justifica formatação rica. Se uma feature futura
  precisar, a decisão (biblioteca, formato de armazenamento e sanitização no
  backend com allowlist de tags) é tomada naquele momento, não antecipada
- **Skill correspondente:** `.claude/skills/frontend-nextjs-shadcn/SKILL.md`
- **Subagent correspondente:** `dev-fullstack`

## Serviços opcionais (adicionar ao docker-compose se o projeto precisar)

- **Armazenamento de arquivos:** **sim**
  - Serviço local (dev): compatível com AWS S3 SDK, porta 8000 (publicada
    em 9000 no host). **A imagem é `zenko/cloudserver`, não a MinIO** — a
    MinIO parou de distribuir imagem/binário para pull anônimo em 2025
    (Docker Hub removeu o repositório, quay.io exige login, dl.min.io
    devolve 410); CloudServer é AGPLv3 e S3-API-compatível, verificado
    nesta rodada com gravação/leitura reais. Sem console web. Ver
    comentário do serviço `minio` em `docker-compose.dev.yml`
  - Produção: MinIO self-hosted | AWS S3
  - **O que guarda:** o **documento `.docx` do plano de ação** (`plano-acao`)
    e os **arquivos de comprovante das entregas de meta** (`metas-coordenacao`)
  - Regra: **nenhum arquivo em disco local do container** — ver CLAUDE.md,
    "Design para múltiplos containers"
  - **Provisionado** — serviço `minio` no `docker-compose.dev.yml`, bucket
    criado na subida pelo serviço `minio-init` (`mc mb --ignore-existing`,
    mesmo padrão do `migrate`). Antecipado em `plano-acao`, antes dos
    anexos de `metas-coordenacao` — o documento chega primeiro
- **Envio de e-mail:** **sim**
  - Serviço local (dev): **Mailpit** (1025 SMTP, 8025 UI) — captura a mensagem
    e a exibe na interface dele, **sem entregar para fora**. É o que permite
    exercitar o fluxo de notificação em desenvolvimento sem risco de disparar
    e-mail real para pessoa real
  - Produção: SMTP corporativo da instituição | SendGrid | SES
  - **O que envia na v1:** notificação de **recusa de entrega de comprovante de
    meta** ao coordenador responsável (`metas-coordenacao`). Nenhum outro envio
    entra sem decisão registrada
  - A notificação **complementa, não substitui** o aviso dentro da aplicação: o
    badge de recusas não vistas continua existindo, porque e-mail se perde
  - Toda chamada ao servidor SMTP é adapter de saída atrás de uma interface em
    `/port` (ex: `EmailSender`), com circuit breaker — falha de e-mail **nunca**
    derruba a operação de negócio que o disparou
  - **Provisionado** — serviço `mailpit` no `docker-compose.dev.yml`
- **Mensageria / background jobs:** **não**
  - Sem RabbitMQ nesta versão. **Atenção do `arquiteto`:** o envio de e-mail
    junto com a gravação da recusa é um caso de **dual write** (CLAUDE.md). Sem
    broker, a decisão entre Outbox e reconciliação por job fica no `design.md`
    de `metas-coordenacao` — não presumir

> Quando marcar "sim", o `dev-docker-compose` adiciona o container correspondente
> ao `docker-compose.yaml` com a configuração pronta.

## Geração de PDF
- **Não se aplica ao escopo atual.** Nenhuma feature especificada gera
  documento. Se uma feature futura precisar, a decisão é tomada naquele
  momento — a regra que já vale é a arquitetural: seria adapter de saída em
  `/adapter` atrás de interface em `/port`, com circuit breaker.

## Geo (preencher se o projeto tiver dados geográficos)
- **Habilitar geo:** **não**
- PostGIS, GeoServer e SRID não se aplicam a este projeto

## Banco de dados
- **Produção:** PostgreSQL
- **Desenvolvimento local:** **PostgreSQL** (mesmo banco de produção — SQLite
  está descartado por decisão explícita: o projeto usa **índice único composto e
  parcial** no isolamento multi-institucional, índice único com `NULLS NOT
  DISTINCT` para o administrador sem vínculo, e trilha de auditoria com `INET` —
  recursos que precisam ser validados contra o banco real desde a primeira linha
  de código)
- **Ferramenta de migration:** golang-migrate (arquivos pareados `up`/`down`)
- **Motor de busca:** busca nativa do Postgres (`tsvector` + índice GIN);
  Elasticsearch só sob gatilho — ver Roadmap em CLAUDE.md

## Testes E2E
- **Ferramenta:** Playwright
- **Skill correspondente:** `.claude/skills/e2e-playwright/SKILL.md`
- **Decisão revista em 2026-09-28:** E2E antecipado para os fluxos críticos
  de `autenticacao-usuarios` (login/logout, primeiro acesso, mesmo e-mail em
  duas instituições, CRUD de usuário, isolamento entre instituições) —
  decisão do dono do produto, motivada por dois bugs reais que só um teste
  de navegador de verdade pegaria (integração com biblioteca de terceiros no
  combo de instituição; navegação pós-logout). Suíte em `e2e/`, roda contra
  `DATABASE_URL_E2E` — banco **próprio**, separado de `DATABASE_URL_TEST`
  (descoberto nesta rodada: o seed do E2E cria dado de referência estável
  que não é limpo entre execuções e contamina invariantes globais dos
  testes de integração Go, ex: "último administrador do sistema") — via o
  serviço `playwright` do `docker-compose.dev.yml` (`RUN_TESTS=true`). Fora
  desses fluxos, a decisão de custo original
  continua: construção de testes no **mínimo necessário** (fronteiras de
  segurança, isolamento e integridade de dado); o restante é anotado em
  `specs/<feature>/testes-pendentes.md` e vira suíte na fase de Release.

## Cache
- **Tecnologia:** Redis (padrão do projeto — sessão e cache-aside de consultas)
- **Uso hoje:** nenhum. Nenhuma funcionalidade especificada até agora depende
  dele. Registrado para o `arquiteto` não inventar uso sem caso concreto.

## Infraestrutura
- **Containerização:** Docker + Docker Compose (`docker-compose.dev.yml` para
  desenvolvimento com hot-reload via `air`; `docker-compose.yaml` para produção)
- **Orquestração:** a definir quando houver deploy — não decidido nesta rodada

## Inteligência Artificial / LLM
- **Provedor:** **nenhum nesta versão**
- Nenhuma capacidade de IA é implementada sem aprovação registrada em
  `oportunidades-ia.md`. O `ai-consultor` pode ser acionado depois para avaliar
  oportunidades sobre o escopo atual.

## Padrão de comunicação (universal, independente de stack)
- Formato de payload: **JSON**
- Convenção de erro: `{"error": {"code": "...", "message": "..."}}`
- Versionamento de API: prefixo `/api/v1`

## Observabilidade (conforme CLAUDE.md)
- `/metrics` (Prometheus, `prometheus/client_golang` + middleware global)
- `/healthz` (liveness) e `/readyz` (readiness — verifica Postgres, Redis,
  MinIO e o servidor SMTP)

## Variáveis de ambiente padrão (todo projeto)
- `APP_ENV` — `development` | `production`; controla bypass de autenticação,
  CSP, CORS, source maps e stack trace. Ausência = `production`
- `CORS_ALLOWED_ORIGINS` — lista de origens permitidas, por ambiente
- `RUN_TESTS` — `true` roda migrations + seed de teste + testes; padrão `false`
- `DATABASE_URL` / `DATABASE_URL_TEST` / `DATABASE_URL_E2E` (banco próprio
  da suíte Playwright — nunca o mesmo de `DATABASE_URL_TEST`)
- `REDIS_URL`
- `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY` (MinIO/S3)
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`,
  `SMTP_TLS` — envio de e-mail. Em dev apontam para o Mailpit
- `RELE_HABILITADO` — liga/desliga o relê de notificações em segundo plano
  (`metas-coordenacao`) por configuração, sem deploy; padrão `true`
- `SEED_ADMIN_EMAIL`, `SEED_ADMIN_SENHA` — Administrador do Sistema inicial;
  chaves no `.env.example` **sem valores**
- `APP_TIMEZONE` — padrão `America/Sao_Paulo`
- `SYSLOG_SERVER`, `SYSLOG_PROTOCOL`, `SYSLOG_APP_NAME` (auditoria canal 2)

---

> Quando este arquivo mudar de projeto para projeto (ex: trocar Go por
> Node.js, ou Next.js por outro framework), nenhum subagent precisa ser
> editado — apenas este arquivo e a skill correspondente à nova tecnologia.
