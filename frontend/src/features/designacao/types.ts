export interface CoordenadorDaDesignacao {
  id: string
  nome: string
}

export type SituacaoDesignacao = "futura" | "vigente" | "encerrada"

export interface Designacao {
  id: string
  coordenador: CoordenadorDaDesignacao
  portaria: string
  data_inicio: string
  data_fim: string | null
  situacao: SituacaoDesignacao
  autodesignacao: boolean
  outras_designacoes_vigentes: number
  versao: number
}

export interface FiltroDesignacoes {
  [key: string]: unknown
  situacao: SituacaoDesignacao | "todas"
  coordenador_id: string
}

export interface DesignacaoInput {
  coordenador_id: string
  portaria: string
  data_inicio: string
  data_fim: string
}

export interface Candidato {
  id: string
  nome: string
  email: string
  perfis: string[]
}
