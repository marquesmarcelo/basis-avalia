-- specs/cursos/design.md §4 — curso e designação de coordenação, com
-- vigência garantida pelo banco (EXCLUDE USING gist).

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE curso (
    id             UUID        PRIMARY KEY,
    instituicao_id UUID        NOT NULL REFERENCES instituicao (id),
    nome           TEXT        NOT NULL,
    codigo_emec    TEXT,
    grau           TEXT        NOT NULL,
    modalidade     TEXT        NOT NULL,
    situacao       TEXT        NOT NULL DEFAULT 'ativo',
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em  TIMESTAMPTZ,
    excluido_em    TIMESTAMPTZ,
    versao         INT         NOT NULL DEFAULT 1,
    CONSTRAINT ck_curso_grau       CHECK (grau IN ('bacharelado','licenciatura','tecnologo')),
    CONSTRAINT ck_curso_modalidade CHECK (modalidade IN ('presencial','a_distancia')),
    CONSTRAINT ck_curso_situacao   CHECK (situacao IN ('ativo','inativo')),
    CONSTRAINT ck_curso_nome       CHECK (btrim(nome) <> '' AND length(nome) <= 300)
);

-- F-05 (fundacao-metas.md §3.5): a chave que designacao referencia em FK
-- composta.
CREATE UNIQUE INDEX uq_curso_id_instituicao ON curso (id, instituicao_id);

-- CU-02: o conflito de nome vale mesmo com o existente inativo — só
-- exclusão lógica libera o nome, nunca a situação.
CREATE UNIQUE INDEX uq_curso_instituicao_nome
    ON curso (instituicao_id, lower(btrim(nome))) WHERE excluido_em IS NULL;
-- CU-03: sem código é permitido quantas vezes for preciso.
CREATE UNIQUE INDEX uq_curso_instituicao_emec
    ON curso (instituicao_id, codigo_emec)
 WHERE excluido_em IS NULL AND codigo_emec IS NOT NULL;
CREATE INDEX idx_curso_listagem
    ON curso (instituicao_id, nome COLLATE "pt-BR-x-icu") WHERE excluido_em IS NULL;

CREATE TABLE designacao (
    id              UUID        PRIMARY KEY,
    curso_id        UUID        NOT NULL,
    instituicao_id  UUID        NOT NULL,
    coordenador_id  UUID        NOT NULL REFERENCES usuario (id),
    portaria        TEXT        NOT NULL,
    data_inicio     DATE        NOT NULL,
    data_fim        DATE,
    autodesignacao  BOOLEAN     NOT NULL DEFAULT false,
    criado_em       TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em   TIMESTAMPTZ,
    excluido_em     TIMESTAMPTZ,
    versao          INT         NOT NULL DEFAULT 1,

    CONSTRAINT ck_designacao_datas    CHECK (data_fim IS NULL OR data_fim >= data_inicio),
    CONSTRAINT ck_designacao_portaria CHECK (btrim(portaria) <> '' AND length(portaria) <= 100),
    CONSTRAINT fk_designacao_curso
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id),

    -- DG-03: garantia de banco, não de aplicação. '[]' fecha os dois lados
    -- do intervalo (o dia da data de fim entra inteiro, DG-06) e
    -- COALESCE(data_fim, 'infinity') faz o prazo indeterminado colidir
    -- com qualquer coisa depois do início.
    CONSTRAINT ex_designacao_sem_sobreposicao EXCLUDE USING gist (
        curso_id WITH =,
        daterange(data_inicio, COALESCE(data_fim, 'infinity'::date), '[]') WITH &&
    ) WHERE (excluido_em IS NULL)
);

CREATE UNIQUE INDEX uq_designacao_id_curso ON designacao (id, curso_id);

-- Sustenta a derivação do perfil em TODA requisição autenticada
-- (fundacao-metas.md §4.1) — não junta curso, de propósito (C-07).
CREATE INDEX idx_designacao_coordenador
    ON designacao (coordenador_id, data_inicio, data_fim) WHERE excluido_em IS NULL;
CREATE INDEX idx_designacao_curso
    ON designacao (curso_id, data_inicio DESC) WHERE excluido_em IS NULL;
