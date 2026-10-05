-- specs/metas-coordenacao/design.md §4 — entrega, anexo e a tabela técnica
-- de idempotência do registro de entrega.

CREATE TABLE entrega (
    id             UUID        PRIMARY KEY,
    item_plano_id  UUID        NOT NULL,
    curso_id       UUID        NOT NULL,
    instituicao_id UUID        NOT NULL,
    enviada_por    UUID        NOT NULL REFERENCES usuario (id),
    corrigida_por  UUID        REFERENCES usuario (id),
    observacao     TEXT        NOT NULL DEFAULT '',
    situacao       TEXT        NOT NULL DEFAULT 'pendente_avaliacao',
    rodadas_de_recusa  INT     NOT NULL DEFAULT 0,
    prazo_correcao_ate TIMESTAMPTZ,
    avaliada_por   UUID        REFERENCES usuario (id),
    avaliada_em    TIMESTAMPTZ,
    motivo         TEXT        NOT NULL DEFAULT '',
    avaliador_era_coordenador BOOLEAN NOT NULL DEFAULT false,
    pendencia_vista_em TIMESTAMPTZ,

    notificacao_evento     TEXT,
    notificacao_gerada_em  TIMESTAMPTZ,
    notificacao_enviada_em TIMESTAMPTZ,
    notificacao_tentativas INT NOT NULL DEFAULT 0,
    notificacao_ultimo_erro TEXT,

    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    atualizado_em  TIMESTAMPTZ,
    excluido_em    TIMESTAMPTZ,
    versao         INT         NOT NULL DEFAULT 1,

    CONSTRAINT ck_entrega_situacao CHECK (situacao IN
        ('pendente_avaliacao','aceita','recusada')),
    CONSTRAINT ck_entrega_rodadas  CHECK (rodadas_de_recusa BETWEEN 0 AND 3),
    CONSTRAINT ck_entrega_recusa   CHECK (situacao <> 'recusada' OR btrim(motivo) <> ''),
    CONSTRAINT ck_entrega_notificacao CHECK (notificacao_evento IS NULL
        OR notificacao_evento IN ('recusa','desfazimento','prazo_restaurado')),
    CONSTRAINT fk_entrega_item
        FOREIGN KEY (item_plano_id, curso_id) REFERENCES item_plano (id, curso_id)
);

CREATE UNIQUE INDEX uq_entrega_id_curso ON entrega (id, curso_id);
CREATE INDEX idx_entrega_item ON entrega (item_plano_id, situacao) WHERE excluido_em IS NULL;
CREATE INDEX idx_entrega_fila ON entrega (instituicao_id, criado_em)
 WHERE excluido_em IS NULL AND situacao = 'pendente_avaliacao';
CREATE INDEX idx_entrega_enviada_por ON entrega (enviada_por) WHERE excluido_em IS NULL;

-- o índice parcial que mantém a varredura do relê barata, mesmo com a
-- tabela grande (design.md §4.1, V-1).
CREATE INDEX idx_entrega_notificacao_pendente
    ON entrega (notificacao_gerada_em)
 WHERE notificacao_enviada_em IS NULL AND notificacao_evento IS NOT NULL;

-- o candidato à restauração de prazo é um subconjunto pequeno (V-5).
CREATE INDEX idx_entrega_recusada_por_curso
    ON entrega (curso_id, prazo_correcao_ate)
 WHERE excluido_em IS NULL AND situacao = 'recusada';

CREATE TABLE anexo (
    id             UUID        PRIMARY KEY,
    entrega_id     UUID        NOT NULL,
    curso_id       UUID        NOT NULL,
    instituicao_id UUID        NOT NULL,
    nome_original  TEXT        NOT NULL,
    tipo           TEXT        NOT NULL,
    tamanho_bytes  BIGINT      NOT NULL,
    chave_objeto   TEXT        NOT NULL,
    hash_sha256    TEXT        NOT NULL,
    criado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
    excluido_em    TIMESTAMPTZ,
    CONSTRAINT ck_anexo_tipo CHECK (tipo IN ('pdf','jpeg','png','docx','odt')),
    CONSTRAINT ck_anexo_tamanho CHECK (tamanho_bytes > 0 AND tamanho_bytes <= 10485760),
    CONSTRAINT fk_anexo_entrega
        FOREIGN KEY (entrega_id, curso_id) REFERENCES entrega (id, curso_id)
);
CREATE INDEX idx_anexo_entrega ON anexo (entrega_id) WHERE excluido_em IS NULL;

-- Tabela puramente técnica, sem rota de API própria — a única do projeto
-- em que a chave primária não é UUID (design.md §4): a garantia de
-- idempotência É a chave primária composta, dentro da mesma transação do
-- INSERT da entrega (§7). Chaveada por usuario_id, não por instituição: é
-- o recorte mais estreito possível.
CREATE TABLE idempotencia (
    usuario_id  UUID        NOT NULL REFERENCES usuario (id),
    rota        TEXT        NOT NULL,
    chave       TEXT        NOT NULL,
    recurso_id  UUID        NOT NULL,
    criado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (usuario_id, rota, chave)
);
CREATE INDEX idx_idempotencia_expurgo ON idempotencia (criado_em);
