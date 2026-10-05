export interface IndicadorEmbutido {
  id: string
  codigo: string
  nome: string
  escopo: "plataforma" | "instituicao"
  referencia_instrumento: string
  situacao: "ativo" | "inativo"
}

export interface Meta {
  id: string
  nome: string
  descricao: string
  situacao: "ativo" | "inativo"
  planos: number
  indicadores: IndicadorEmbutido[]
  quantidade_sugerida: number | null
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroMetas {
  [key: string]: unknown
  busca: string
  indicador_id: string
  origem: "todas" | "plataforma" | "instituicao"
  situacao: "ativo" | "inativo" | "todas"
}

export interface MetaInput {
  nome: string
  descricao: string
  indicadores: string[]
  quantidade_sugerida: number | null
}

export interface SugestaoMeta {
  id: string
  nome: string
  indicadores: IndicadorEmbutido[]
  quantidade_sugerida: number | null
}
