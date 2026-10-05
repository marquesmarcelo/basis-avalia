-- specs/metas-coordenacao/design.md §4.7 (revisão pós code-review, T-123,
-- T-124, T-127) — defesa contra injeção na saída, e a coluna que fecha
-- duas pendências: reivindicação atômica do relê e a última tentativa que
-- faltava para o backoff exponencial.

-- T-123/T-125: nome é uma linha. `motivo` e `observacao` são texto livre
-- do usuário que atravessa fronteiras de saída (e-mail, CSV) onde
-- caractere de controle deixa de ser dado e vira instrução. Validado no
-- Value Object (camada que decide antes de qualquer escrita) e aqui no
-- CHECK (camada que sobrevive a um caminho de escrita futuro que não
-- passe pelo Go) — classe POSIX [:cntrl:] cobre \r, \n e os demais C0/DEL.
ALTER TABLE entrega
    ADD CONSTRAINT ck_entrega_motivo_sem_controle CHECK (motivo !~ '[[:cntrl:]]'),
    ADD CONSTRAINT ck_entrega_observacao_sem_controle CHECK (observacao !~ '[[:cntrl:]]');

-- T-127: a coluna de reivindicação do relê. Substitui a garantia de
-- `FOR UPDATE SKIP LOCKED` em autocommit (que libera os locks ao fim do
-- SELECT, antes do envio de e-mail, sem proteger nada) por uma ESCRITA
-- que sobrevive à chamada SMTP — a reivindicação é feita por um único
-- UPDATE ... WHERE id IN (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING,
-- atômico mesmo sem transação explícita. A mesma coluna serve de "última
-- tentativa" para o backoff exponencial (design.md §8.4 simplificação
-- registrada, agora fechada).
ALTER TABLE entrega ADD COLUMN notificacao_reivindicada_em TIMESTAMPTZ;
