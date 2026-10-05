ALTER TABLE entrega DROP COLUMN IF EXISTS notificacao_reivindicada_em;
ALTER TABLE entrega DROP CONSTRAINT IF EXISTS ck_entrega_observacao_sem_controle;
ALTER TABLE entrega DROP CONSTRAINT IF EXISTS ck_entrega_motivo_sem_controle;
