-- Reverte a consolidação inteira: auditoria -> usuario_perfil/usuario
-- -> instituicao -> extensões, na ordem inversa de dependência de FK.

DROP TABLE IF EXISTS auditoria;

DROP TABLE IF EXISTS usuario_perfil;
DROP TABLE IF EXISTS usuario;

DROP TABLE IF EXISTS instituicao;

DROP EXTENSION IF EXISTS unaccent;
