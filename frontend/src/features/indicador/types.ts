export interface Indicador {
  id: string
  escopo: "plataforma" | "instituicao"
  codigo: string
  nome: string
  descricao: string
  referencia_instrumento: string | null
  situacao: "ativo" | "inativo"
  metas: number
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroIndicadores {
  [key: string]: unknown
  busca: string
  origem: "todos" | "plataforma" | "instituicao"
  situacao: "ativo" | "inativo" | "todos"
}

export interface IndicadorInput {
  codigo: string
  nome: string
  descricao: string
}

export interface SugestaoIndicador {
  id: string
  codigo: string
  nome: string
  escopo: "plataforma" | "instituicao"
  referencia_instrumento: string
}
