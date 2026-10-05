-- specs/cursos/design.md §7 — a migration que altera privilégio de
-- pessoas reais: coordenador_curso deixa de ser um perfil atribuível,
-- passando a ser sempre derivado de designação (fundacao-metas.md §4).
--
-- Nota de numeração: design.md descreve estes três passos como parte da
-- MESMA migration que cria as tabelas (ali chamada "000003_cursos.up.sql",
-- passos 4-6). Migration já aplicada nunca é editada (regra do projeto) —
-- as tabelas de curso/designação já foram criadas e aplicadas em
-- 000003_cursos. Esta migration separada, 000005, produz exatamente o
-- mesmo estado final do banco, na mesma ordem relativa (depois de
-- 000003_cursos existir, e depois do seed ter criado as designações que
-- sustentam a derivação — DBA valida CP-15 com linhas presentes).
--
-- ORDEM OBRIGATÓRIA, não incidental: auditar (1) antes de apagar (2) —
-- depois do DELETE não há mais o que registrar; o CHECK estreitado (3) só
-- depois do DELETE — a restrição não é aceita com linhas que a violam.

-- 1. Uma linha de auditoria por usuário afetado.
INSERT INTO auditoria (id, acao, resultado, ator_id, instituicao_id, recurso_tipo, recurso_id, detalhes, executado_em)
SELECT gen_random_uuid(), 'normalizar_perfil_coordenador', 'sucesso', NULL, u.instituicao_id, 'Usuario', u.id,
       jsonb_build_object('motivo', 'perfil de coordenador passou a ser derivado de designação'), now()
  FROM usuario u
  JOIN usuario_perfil up ON up.usuario_id = u.id
 WHERE up.perfil = 'coordenador_curso';

-- 2. Remove o perfil atribuído — quem tem designação vigente continua
-- coordenador pela derivação; quem não tem mantém os demais perfis (o
-- DELETE é por (usuario_id, perfil), nunca a linha inteira do usuário).
DELETE FROM usuario_perfil WHERE perfil = 'coordenador_curso';

-- 3. O CHECK estreitado, só agora.
ALTER TABLE usuario_perfil
    DROP CONSTRAINT ck_usuario_perfil_valor,
    ADD CONSTRAINT ck_usuario_perfil_valor CHECK (perfil IN (
        'administrador_sistema', 'pesquisador_institucional', 'professor', 'aluno'
    ));
