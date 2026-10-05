#!/usr/bin/env node
// Reprova qualquer comentário em arquivo autoral de frontend (.ts/.tsx que
// vai para o bundle do navegador) — CLAUDE.md, "Comentários que chegam ao
// usuário final — proibidos". Exclui só duas coisas: diretivas de lint/tipo
// (eslint-disable, biome-ignore, @ts-expect-error, @ts-ignore) e o caminho
// vendorizado da biblioteca de componentes.
//
//   node examples/quality/validar-comentarios-frontend.mjs $(find frontend/src -regex '.*\.tsx?$' -not -regex '.*\.test\.tsx?$' -not -regex '.*\.spec\.tsx?$')
//
// Saída: lista de comentários reprovados com arquivo, linha e conteúdo.
// Código de saída 1 se algum arquivo violar — pronto para usar em CI.
//
// Escopo: só .ts/.tsx de frontend/src, e só os que realmente entram no
// bundle — arquivos de teste (.test.ts, .spec.ts) nunca são enviados ao
// navegador, então não são cobertos por esta regra (são código de
// desenvolvimento, como examples/quality/ é para o backend).

import fs from 'fs';

const CAMINHO_VENDORIZADO = /\/components\/ui\//;
const DIRETIVA_PERMITIDA = /^\s*\/\/\s*(eslint-disable|biome-ignore|@ts-expect-error|@ts-ignore)/;

function violacoesDoArquivo(caminho) {
  if (CAMINHO_VENDORIZADO.test(caminho)) return [];
  const linhas = fs.readFileSync(caminho, 'utf8').split('\n');
  const violacoes = [];
  let dentroDeBloco = false;

  linhas.forEach((linha, i) => {
    const trimmed = linha.trim();

    if (dentroDeBloco) {
      violacoes.push({ linha: i + 1, texto: trimmed });
      if (trimmed.includes('*/')) dentroDeBloco = false;
      return;
    }

    if (trimmed.startsWith('/*')) {
      violacoes.push({ linha: i + 1, texto: trimmed });
      if (!trimmed.includes('*/')) dentroDeBloco = true;
      return;
    }

    if (trimmed.startsWith('//')) {
      if (DIRETIVA_PERMITIDA.test(linha)) return;
      violacoes.push({ linha: i + 1, texto: trimmed });
    }
  });

  return violacoes;
}

const arquivos = process.argv.slice(2);
if (arquivos.length === 0) {
  console.error('uso: node validar-comentarios-frontend.mjs <arquivo.ts|.tsx> [...]');
  process.exit(2);
}

let totalViolacoes = 0;

for (const arquivo of arquivos) {
  const violacoes = violacoesDoArquivo(arquivo);
  if (violacoes.length === 0) continue;
  totalViolacoes += violacoes.length;
  console.log(`\n[FALHA] ${arquivo}`);
  for (const v of violacoes) {
    console.log(`   ${v.linha}: ${v.texto}`);
  }
}

if (totalViolacoes > 0) {
  console.log(
    `\n${totalViolacoes} comentário(s) em arquivo autoral de frontend — remova realocando ` +
      'para o design.md/ux.md da feature (CLAUDE.md, "Comentários que chegam ao usuário final").'
  );
} else {
  console.log(`\n${arquivos.length} arquivo(s) verificado(s), zero comentários.`);
}
process.exit(totalViolacoes > 0 ? 1 : 0);
