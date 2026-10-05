export interface Periodo {
  id: string
  nome: string
  data_inicio: string
  data_fim: string
  situacao: "nao_iniciado" | "aberto" | "encerrado"
  planos: number
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroPeriodos {
  [key: string]: unknown
  nome: string
  situacao: "todas" | "nao_iniciado" | "aberto" | "encerrado"
}

export interface PeriodoInput {
  nome: string
  data_inicio: string
  data_fim: string
}
