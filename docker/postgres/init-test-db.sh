#!/bin/sh
# Executado automaticamente pela imagem oficial do Postgres na primeira
# inicialização (volume de dados vazio) — cria o banco de teste isolado
# exigido pelo CLAUDE.md ("Banco de teste isolado"), no mesmo servidor do
# banco de desenvolvimento.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE "${POSTGRES_TEST_DB}"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${POSTGRES_TEST_DB}')\gexec
    SELECT 'CREATE DATABASE "${POSTGRES_E2E_DB}"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${POSTGRES_E2E_DB}')\gexec
EOSQL
