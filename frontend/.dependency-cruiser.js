/** @type {import('dependency-cruiser').IConfiguration} */
module.exports = {
  forbidden: [
    {
      name: "shared-nao-importa-features",
      comment:
        "components/shared e components/ui nunca importam de features/ — viola o sentido das dependências",
      severity: "error",
      from: { path: "^src/components/(shared|ui)/" },
      to: { path: "^src/features/" },
    },
    // dependency-cruiser não correlaciona grupo de captura entre "from" e
    // "to" — cada feature precisa da própria regra explícita (negative
    // lookahead literal), uma por pasta em src/features/. Adicionar uma
    // linha aqui ao criar uma feature nova.
    {
      name: "auth-nao-importa-outra-feature",
      comment: "features não se importam diretamente — usar components/shared/ como intermediário",
      severity: "warn",
      from: { path: "^src/features/auth/" },
      to: { path: "^src/features/(?!auth/)" },
    },
    {
      name: "instituicao-nao-importa-outra-feature",
      comment: "features não se importam diretamente — usar components/shared/ como intermediário",
      severity: "warn",
      from: { path: "^src/features/instituicao/" },
      to: { path: "^src/features/(?!instituicao/)" },
    },
    {
      name: "usuario-nao-importa-outra-feature",
      comment: "features não se importam diretamente — usar components/shared/ como intermediário",
      severity: "warn",
      from: { path: "^src/features/usuario/" },
      to: { path: "^src/features/(?!usuario/)" },
    },
    {
      name: "sem-import-relativo-acima-de-src",
      severity: "error",
      from: { path: "^src/" },
      to: { path: "^\\.\\./\\.\\." },
    },
    {
      name: "sem-dependencias-ciclicas",
      severity: "error",
      from: {},
      to: { circular: true },
    },
  ],

  options: {
    doNotFollow: {
      path: "node_modules",
    },
    tsPreCompilationDeps: true,
    tsConfig: {
      fileName: "tsconfig.json",
    },
    reporterOptions: {
      dot: {
        collapsePattern: "^(node_modules|src/components|src/features)/[^/]+",
      },
    },
  },
}
