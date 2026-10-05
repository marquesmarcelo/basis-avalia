export type Permissao =
  | "instituicao.listar"
  | "instituicao.criar"
  | "instituicao.editar"
  | "instituicao.inativar"
  | "pi.gerenciar"
  | "usuario.listar"
  | "usuario.criar"
  | "usuario.editar"
  | "usuario.excluir"
  | "usuario.redefinir_senha"
  | "administrador.gerenciar"
  | "indicador.plataforma.gerenciar"
  | "indicador.listar"
  | "indicador.gerenciar"
  | "meta.listar"
  | "meta.gerenciar"
  | "periodo.gerenciar"
  | "plano.listar"
  | "plano.gerenciar"
  | "plano.ler_proprio"
  | "curso.listar"
  | "curso.gerenciar"
  | "curso.ler_proprio"
  | "designacao.gerenciar"
  | "entrega.registrar"
  | "entrega.listar"
  | "entrega.avaliar"
  | "relatorio.ler"
  | "relatorio.exportar"
  | "relatorio.ler_proprio"

export function possuiPermissao(permissoes: string[], permissao: Permissao): boolean {
  return permissoes.includes(permissao)
}

export function possuiAlgumaPermissao(permissoes: string[], candidatas: Permissao[]): boolean {
  return candidatas.some((p) => permissoes.includes(p))
}
