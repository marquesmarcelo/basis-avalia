-- LIMITAÇÃO DECLARADA (design.md §7, T-182): este down recria o CHECK
-- largo (coordenador_curso volta a ser um valor aceito), mas NÃO restaura
-- as linhas de usuario_perfil apagadas pelo up — DELETE de migration não
-- é reversível na prática, e fingir que é seria pior do que admitir que
-- não é.
--
-- Isso não impede rollback de incidente: pela regra de compatibilidade
-- para trás do CLAUDE.md, a versão anterior da aplicação continua
-- funcionando com o schema já estreitado (ela só nunca vê ninguém com
-- coordenador_curso atribuído, o que é exatamente o estado real depois do
-- up) — reverter a IMAGEM nunca exige reverter o BANCO. Este down existe
-- para simetria de ferramenta (golang-migrate down completo), não como
-- caminho de recuperação de dado.

ALTER TABLE usuario_perfil
    DROP CONSTRAINT ck_usuario_perfil_valor,
    ADD CONSTRAINT ck_usuario_perfil_valor CHECK (perfil IN (
        'administrador_sistema', 'pesquisador_institucional', 'coordenador_curso', 'professor', 'aluno'
    ));
