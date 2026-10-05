-- specs/indicadores/design.md §4 — catálogo do INEP (comum à instalação) e
-- catálogo de metas da instituição.

CREATE TABLE indicador (
    id                     UUID        PRIMARY KEY,
    escopo                 TEXT        NOT NULL,
    instituicao_id         UUID        REFERENCES instituicao (id),
    codigo                 TEXT        NOT NULL,
    nome                   TEXT        NOT NULL,
    descricao              TEXT        NOT NULL DEFAULT '',
    referencia_instrumento TEXT,
    situacao               TEXT        NOT NULL DEFAULT 'ativo',
    criado_em              TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em          TIMESTAMPTZ,
    excluido_em            TIMESTAMPTZ,
    versao                 INT         NOT NULL DEFAULT 1,

    CONSTRAINT ck_indicador_escopo    CHECK (escopo IN ('plataforma','instituicao')),
    CONSTRAINT ck_indicador_situacao  CHECK (situacao IN ('ativo','inativo')),
    -- IV-06: o par (escopo, instituição) não pode ser incoerente.
    CONSTRAINT ck_indicador_coerencia CHECK ((escopo = 'plataforma') = (instituicao_id IS NULL)),
    -- Referência obrigatória na plataforma, proibida na instituição.
    CONSTRAINT ck_indicador_referencia CHECK (
        (escopo = 'plataforma' AND referencia_instrumento IS NOT NULL
                               AND btrim(referencia_instrumento) <> '')
     OR (escopo = 'instituicao' AND referencia_instrumento IS NULL)),
    CONSTRAINT ck_indicador_codigo    CHECK (btrim(codigo) <> '' AND length(codigo) <= 50),
    CONSTRAINT ck_indicador_nome      CHECK (btrim(nome) <> '' AND length(nome) <= 300)
);

-- Um único índice, três regras (design.md §4.2): dois "1.4" de plataforma
-- colidem (NULLS NOT DISTINCT trata os NULL como iguais); dois GEST-01 na
-- mesma instituição colidem; o mesmo código em instituições diferentes, ou
-- em escopos diferentes, não colide.
CREATE UNIQUE INDEX uq_indicador_instituicao_codigo
    ON indicador (instituicao_id, codigo) NULLS NOT DISTINCT
 WHERE excluido_em IS NULL;

CREATE INDEX idx_indicador_listagem
    ON indicador (instituicao_id, codigo) WHERE excluido_em IS NULL;
CREATE INDEX idx_indicador_plataforma
    ON indicador (codigo) WHERE excluido_em IS NULL AND escopo = 'plataforma';

CREATE TABLE meta (
    id             UUID        PRIMARY KEY,
    instituicao_id UUID        NOT NULL REFERENCES instituicao (id),
    nome           TEXT        NOT NULL,
    descricao      TEXT        NOT NULL DEFAULT '',
    situacao       TEXT        NOT NULL DEFAULT 'ativo',
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em  TIMESTAMPTZ,
    excluido_em    TIMESTAMPTZ,
    versao         INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_meta_situacao CHECK (situacao IN ('ativo','inativo')),
    CONSTRAINT ck_meta_nome     CHECK (btrim(nome) <> '' AND length(nome) <= 300)
);

-- Único por instituição, comparado por forma normalizada (design.md §4.2):
-- "Engenharia Civil" e "engenharia civil" colidem; acento é distinto.
CREATE UNIQUE INDEX uq_meta_instituicao_nome
    ON meta (instituicao_id, lower(btrim(nome))) WHERE excluido_em IS NULL;

CREATE INDEX idx_meta_listagem
    ON meta (instituicao_id, nome COLLATE "pt-BR-x-icu") WHERE excluido_em IS NULL;

-- meta_indicador não tem instituicao_id de propósito (design.md §4.1): ela
-- nunca é alvo de AplicarEscopo — só é alcançada a partir de uma meta já
-- recortada, e a linha de indicador que ela aponta é recortada por
-- AplicarEscopo(AlvoIndicador) na junção.
CREATE TABLE meta_indicador (
    meta_id        UUID NOT NULL REFERENCES meta (id) ON DELETE CASCADE,
    indicador_id   UUID NOT NULL REFERENCES indicador (id),
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (meta_id, indicador_id)
);

-- As duas direções são consultadas: meta -> indicadores na tela (a PK
-- cobre), indicador -> metas na contagem de uso (este índice cobre).
CREATE INDEX idx_meta_indicador_inverso
    ON meta_indicador (indicador_id, meta_id);
