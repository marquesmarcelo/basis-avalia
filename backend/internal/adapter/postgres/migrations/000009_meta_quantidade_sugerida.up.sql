-- specs/indicadores/design.md — quantidade sugerida da meta, herdada pelo
-- item do plano na criação (uma vez só). Sugestão, nunca regra: a
-- apuração do relatório continua lendo exclusivamente item_plano.quantidade.
ALTER TABLE meta ADD COLUMN quantidade_sugerida INT;
ALTER TABLE meta ADD CONSTRAINT ck_meta_quantidade_sugerida
    CHECK (quantidade_sugerida IS NULL OR quantidade_sugerida >= 1);
