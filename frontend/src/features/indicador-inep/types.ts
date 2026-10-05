export interface IndicadorInep {
  id: string
  escopo: "plataforma"
  codigo: string
  nome: string
  descricao: string
  referencia_instrumento: string
  situacao: "ativo" | "inativo"
  metas: number
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroIndicadoresInep {
  [key: string]: unknown
  busca: string
  situacao: "ativo" | "inativo" | "todos"
}

export interface IndicadorInepInput {
  codigo: string
  nome: string
  descricao: string
  referencia_instrumento: string
}
