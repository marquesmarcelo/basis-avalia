-- ============================================================
-- Unificação de 000001_init (no-op) e 000002_autenticacao_usuarios
-- em uma única migration — o motivo do no-op (pasta vazia derruba
-- o serviço migrate) deixou de valer a partir do momento em que
-- esta migration real existe. Nenhuma linha de DDL foi alterada
-- em relação ao que estava em 000002_autenticacao_usuarios.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TABLE instituicao (
    id            UUID        PRIMARY KEY,
    nome          TEXT        NOT NULL,
    sigla         TEXT        NOT NULL,
    codigo_emec   TEXT,
    situacao      TEXT        NOT NULL DEFAULT 'ativa',
    criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em TIMESTAMPTZ,
    excluido_em   TIMESTAMPTZ,
    versao        INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_instituicao_situacao CHECK (situacao IN ('ativa','inativa')),
    CONSTRAINT ck_instituicao_nome     CHECK (btrim(nome) <> '' AND length(nome) <= 200),
    CONSTRAINT ck_instituicao_sigla    CHECK (sigla = upper(btrim(sigla))
                                              AND sigla <> ''
                                              AND length(sigla) <= 20),
    CONSTRAINT ck_instituicao_emec     CHECK (codigo_emec IS NULL
                                              OR (codigo_emec = btrim(codigo_emec)
                                                  AND codigo_emec <> ''
                                                  AND length(codigo_emec) <= 20))
);

-- Sigla única entre todas as não excluídas (ativas E inativas): o motivo da
-- unicidade em P15 é a ambiguidade de tela, e a instituição inativa continua
-- aparecendo no grid (3.19).
CREATE UNIQUE INDEX uq_instituicao_sigla
    ON instituicao (sigla) WHERE excluido_em IS NULL;

-- Código e-MEC único entre todas as instituições não excluídas (ativas E
-- inativas) — decisão do dono do produto revendo P15: o código identifica
-- a IES no MEC, então duas instituições com o mesmo código é erro de dado
-- mesmo com uma delas inativa. Isso também elimina o caso de não conseguir
-- reativar uma instituição porque outra assumiu o código dela enquanto ela
-- estava inativa — o conflito passa a ser barrado no cadastro/edição.
CREATE UNIQUE INDEX uq_instituicao_codigo_emec
    ON instituicao (codigo_emec)
 WHERE excluido_em IS NULL AND codigo_emec IS NOT NULL;

-- Ordenação padrão do grid e da rota pública, em collation pt-BR.
CREATE INDEX idx_instituicao_nome
    ON instituicao (nome COLLATE "pt-BR-x-icu") WHERE excluido_em IS NULL;

CREATE TABLE usuario (
    id                          UUID        PRIMARY KEY,
    instituicao_id              UUID        REFERENCES instituicao (id),
    nome                        TEXT        NOT NULL,
    email                       TEXT        NOT NULL,
    senha_hash                  TEXT,
    senha_provisoria            BOOLEAN     NOT NULL DEFAULT true,
    sessoes_validas_a_partir_de TIMESTAMPTZ NOT NULL DEFAULT now(),
    provedor_identidade         TEXT        NOT NULL DEFAULT 'credencial_local',
    identificador_externo       TEXT,
    criado_em                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em               TIMESTAMPTZ,
    excluido_em                 TIMESTAMPTZ,
    versao                      INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_usuario_provedor CHECK (provedor_identidade IN ('credencial_local')),
    CONSTRAINT ck_usuario_nome     CHECK (btrim(nome) <> '' AND length(nome) <= 200),
    -- Normalização de e-mail é invariante de banco, não só do construtor.
    CONSTRAINT ck_usuario_email    CHECK (email = lower(btrim(email))
                                          AND email <> ''
                                          AND length(email) <= 320
                                          AND position('@' in email) > 1),
    -- A exclusão lógica anula a credencial, e só ela (3.8).
    CONSTRAINT ck_usuario_credencial CHECK ((senha_hash IS NOT NULL) = (excluido_em IS NULL))
);

-- Alvo da chave estrangeira composta de usuario_perfil (design.md §5.3).
CREATE UNIQUE INDEX uq_usuario_id_instituicao ON usuario (id, instituicao_id);

-- ARMADILHA 1 RESOLVIDA COM UM ÚNICO OBJETO.
-- NULLS NOT DISTINCT faz o vínculo nulo do Administrador do Sistema comparar
-- como igual, impedindo dois administradores com o mesmo e-mail. Sem isso o
-- Postgres trataria cada NULL como distinto e os dois passariam.
CREATE UNIQUE INDEX uq_usuario_instituicao_email_ativo
    ON usuario (instituicao_id, email) NULLS NOT DISTINCT
 WHERE excluido_em IS NULL;

-- ARMADILHA 3: instituicao_id é a primeira coluna, porque entra em toda
-- consulta de isolamento. Serve a ordenação padrão (Nome crescente).
CREATE INDEX idx_usuario_instituicao_nome
    ON usuario (instituicao_id, nome COLLATE "pt-BR-x-icu")
 WHERE excluido_em IS NULL;

-- Tabela de vínculo usuário-perfil (design.md §5.2, §5.3): um usuário tem um
-- CONJUNTO de perfis, nunca um só. instituicao_id é denormalizado de
-- propósito: a FK composta abaixo amarra este valor ao mesmo par (usuario,
-- instituição) da linha de usuario, e é isso que faz o CHECK abaixo garantir
-- as duas metades de "administrador não acumula" (U-15) sem trigger.
CREATE TABLE usuario_perfil (
    usuario_id     UUID        NOT NULL,
    perfil         TEXT        NOT NULL,
    instituicao_id UUID,
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (usuario_id, perfil),

    CONSTRAINT ck_usuario_perfil_valor CHECK (perfil IN (
        'administrador_sistema','pesquisador_institucional',
        'coordenador_curso','professor','aluno')),

    -- D-17: a regra "administrador não acumula" vira estado não representável.
    CONSTRAINT ck_usuario_perfil_estrutura
        CHECK ((perfil = 'administrador_sistema') = (instituicao_id IS NULL)),

    CONSTRAINT fk_usuario_perfil_usuario
        FOREIGN KEY (usuario_id, instituicao_id)
        REFERENCES usuario (id, instituicao_id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

-- "quem possui este perfil nesta instituição": filtro do grid (G-03) e as
-- duas invariantes de último detentor (E-09, E-17, AS-06).
CREATE INDEX idx_usuario_perfil_instituicao
    ON usuario_perfil (instituicao_id, perfil, usuario_id);

-- Sonda do EXISTS da rota pública do combo, que antes vivia em usuario.
CREATE INDEX idx_usuario_perfil_pi
    ON usuario_perfil (instituicao_id)
 WHERE perfil = 'pesquisador_institucional';

CREATE TABLE auditoria (
    id             UUID        PRIMARY KEY,
    acao           TEXT        NOT NULL,
    resultado      TEXT        NOT NULL,
    ator_id        UUID        REFERENCES usuario (id),
    instituicao_id UUID        REFERENCES instituicao (id),
    recurso_tipo   TEXT,
    recurso_id     UUID,
    detalhes       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    ip_origem      INET,
    executado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_auditoria_resultado CHECK (resultado IN ('sucesso','negado','falha','erro'))
);

CREATE INDEX idx_auditoria_instituicao_data ON auditoria (instituicao_id, executado_em DESC);
CREATE INDEX idx_auditoria_recurso          ON auditoria (recurso_tipo, recurso_id);
