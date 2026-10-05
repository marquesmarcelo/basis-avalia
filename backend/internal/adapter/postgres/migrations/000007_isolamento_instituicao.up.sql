-- T-129 (autenticacao-usuarios/design.md §5.11) — achado da revisão de
-- segurança: F-05 afirmava que FK composta torna irrepresentável a
-- divergência de instituição entre uma linha e seu pai. Verdade em
-- `plano` e `designacao`; falsa em `item_plano`, `documento`, `entrega` e
-- `anexo` — as FKs dessas quatro ancoravam só o curso, e o isolamento de
-- instituição repousava em coluna mantida pela aplicação, sem garantia de
-- escrita (AplicarEscopo cobre leitura, nada cobria INSERT).
--
-- Mesmo padrão já usado em usuario_perfil e já usado por plano/designacao
-- para ancorar em curso: FK composta (curso_id, instituicao_id) contra
-- curso (id, instituicao_id) — reaproveita o índice único
-- uq_curso_id_instituicao, criado em 000003_cursos. A instituição do
-- curso é imutável (spec de cursos), então a redundância não deriva com o
-- tempo.
--
-- As quatro entram: nenhuma tem volume ou padrão de escrita que
-- justifique exceção neste projeto pré-release (dba, T-129).

ALTER TABLE item_plano
    ADD CONSTRAINT fk_item_plano_curso_instituicao
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id);

ALTER TABLE documento
    ADD CONSTRAINT fk_documento_curso_instituicao
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id);

ALTER TABLE entrega
    ADD CONSTRAINT fk_entrega_curso_instituicao
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id);

ALTER TABLE anexo
    ADD CONSTRAINT fk_anexo_curso_instituicao
        FOREIGN KEY (curso_id, instituicao_id) REFERENCES curso (id, instituicao_id);
