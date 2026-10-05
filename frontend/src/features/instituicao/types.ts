export interface Instituicao {
  id: string
  nome: string
  sigla: string
  codigo_emec: string | null
  situacao: "ativa" | "inativa"
  pesquisadores_ativos: number
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroInstituicoes {
  [key: string]: unknown
  busca: string
  situacao: "todas" | "ativa" | "inativa"
}

export interface InstituicaoInput {
  nome: string
  sigla: string
  codigo_emec: string
}
