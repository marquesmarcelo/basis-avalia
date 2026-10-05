-- specs/plano-acao/design.md §4 — período, plano de ação curso/coordenador,
-- itens do plano e documento gerado.

CREATE TABLE periodo (
    id             UUID        PRIMARY KEY,
    instituicao_id UUID        NOT NULL REFERENCES instituicao (id),
    nome           TEXT        NOT NULL,
    data_inicio    DATE        NOT NULL,
    data_fim       DATE        NOT NULL,
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em  TIMESTAMPTZ,
    excluido_em    TIMESTAMPTZ,
    versao         INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_periodo_datas CHECK (data_fim >= data_inicio),
    CONSTRAINT ck_periodo_nome  CHECK (btrim(nome) <> '' AND length(nome) <= 100)
);
CREATE UNIQUE INDEX uq_periodo_instituicao_nome
    ON periodo (instituicao_id, lower(btrim(nome))) WHERE excluido_em IS NULL;
CREATE INDEX idx_periodo_listagem
    ON periodo (instituicao_id, data_inicio DESC) WHERE excluido_em IS NULL;

CREATE TABLE plano (
    id                   UUID        PRIMARY KEY,
    instituicao_id       UUID        NOT NULL,
    curso_id             UUID        NOT NULL,
    periodo_id           UUID        NOT NULL REFERENCES periodo (id),
    descricao            TEXT        NOT NULL,
    objetivo_geral       TEXT        NOT NULL,
    resultados_esperados TEXT        NOT NULL,
    alinhamento_pdi      TEXT        NOT NULL DEFAULT '',
    alinhamento_ppc      TEXT        NOT NULL DEFAULT '',
    aprovacao_data       DATE,
    aprovacao_orgao      TEXT,
    situacao_publicacao  TEXT        NOT NULL DEFAULT 'rascunho',
    encerrado_em         TIMESTAMPTZ,
    encerramento_motivo  TEXT,
    criado_em            TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em        TIMESTAMPTZ,
    excluido_em          TIMESTAMPTZ,
    versao               INT         NOT NULL DEFAULT 1,

    CONSTRAINT ck_plano_situacao CHECK (situacao_publicacao IN ('rascunho','vigente')),
    CONSTRAINT ck_plano_orgao    CHECK (aprovacao_orgao IS NULL
                                     OR aprovacao_orgao IN ('nde','colegiado_curso')),
    -- PL-03: os dois ou nenhum, garantido pelo banco também.
    CONSTRAINT ck_plano_aprovacao CHECK ((aprovacao_data IS NULL) = (aprovacao_orgao IS NULL)),
    CONSTRAINT ck_plano_encerramento CHECK (encerrado_em IS NULL
                                         OR btrim(coalesce(encerramento_motivo,'')) <> ''),
    CONSTRAINT ck_plano_textos CHECK (btrim(descricao) <> ''
                                  AND btrim(objetivo_geral) <> ''
                                  AND btrim(resultados_esperados) <> ''),
    CONSTRAINT fk_plano_curso
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id)
);

-- PP-1 / PL-02 / a corrida de CP-05: a garantia está AQUI, não na
-- pré-verificação da tela.
CREATE UNIQUE INDEX uq_plano_curso_periodo
    ON plano (curso_id, periodo_id) WHERE excluido_em IS NULL;

CREATE UNIQUE INDEX uq_plano_id_curso ON plano (id, curso_id);
CREATE INDEX idx_plano_listagem
    ON plano (instituicao_id, periodo_id, curso_id) WHERE excluido_em IS NULL;

CREATE TABLE item_plano (
    id             UUID        PRIMARY KEY,
    plano_id       UUID        NOT NULL,
    curso_id       UUID        NOT NULL,
    instituicao_id UUID        NOT NULL,
    meta_id        UUID        NOT NULL REFERENCES meta (id),
    quantidade     INT         NOT NULL,
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em  TIMESTAMPTZ,
    excluido_em    TIMESTAMPTZ,
    versao         INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_item_quantidade CHECK (quantidade >= 1),
    CONSTRAINT fk_item_plano
        FOREIGN KEY (plano_id, curso_id) REFERENCES plano (id, curso_id)
);
CREATE UNIQUE INDEX uq_item_plano_meta
    ON item_plano (plano_id, meta_id) WHERE excluido_em IS NULL;
CREATE UNIQUE INDEX uq_item_id_curso ON item_plano (id, curso_id);
CREATE INDEX idx_item_plano ON item_plano (plano_id) WHERE excluido_em IS NULL;
CREATE INDEX idx_item_meta  ON item_plano (meta_id)  WHERE excluido_em IS NULL;

CREATE TABLE documento (
    id              UUID        PRIMARY KEY,
    plano_id        UUID        NOT NULL,
    curso_id        UUID        NOT NULL,
    instituicao_id  UUID        NOT NULL,
    chave_objeto    TEXT        NOT NULL,
    nome_arquivo    TEXT        NOT NULL,
    situacao_no_ato TEXT        NOT NULL,
    gerado_por      UUID        NOT NULL REFERENCES usuario (id),
    gerado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    criado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    excluido_em     TIMESTAMPTZ,
    CONSTRAINT fk_documento_plano
        FOREIGN KEY (plano_id, curso_id) REFERENCES plano (id, curso_id)
);
-- Sustenta a lateral de "último documento" que substitui plano.documento_id
-- (design.md P-11, §7.4) — a única forma de ler "o mais recente" sem
-- coluna dedicada nem varredura sequencial.
CREATE INDEX idx_documento_plano ON documento (plano_id, gerado_em DESC);
