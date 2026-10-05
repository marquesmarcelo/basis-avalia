ALTER TABLE meta DROP CONSTRAINT IF EXISTS ck_meta_quantidade_sugerida;
ALTER TABLE meta DROP COLUMN IF EXISTS quantidade_sugerida;
