import type { IndicadorEmbutido } from "@/features/meta/types"

export interface CursoEmbutido {
  id: string
  nome: string
  grau: string
  modalidade: string
  codigo_emec?: string
  vago: boolean
}

export interface PeriodoEmbutido {
  id: string
  nome: string
}

export interface Aprovacao {
  data: string
  orgao: "nde" | "colegiado_curso"
}

export interface UltimoDocumento {
  id: string
  gerado_em: string
}

export interface ItemDoPlano {
  id: string
  meta_id: string
  meta_nome: string
  indicadores: IndicadorEmbutido[]
  quantidade: number
  tem_entrega: boolean
  versao: number
}

export type SituacaoPlano = "rascunho" | "vigente" | "encerrado"

export interface Plano {
  id: string
  curso: CursoEmbutido
  periodo: PeriodoEmbutido
  descricao: string
  objetivo_geral: string
  resultados_esperados: string
  alinhamento_pdi: string
  alinhamento_ppc: string
  situacao: SituacaoPlano
  metas: number
  total_exigido: number
  aprovacao: Aprovacao | null
  sem_aprovacao: boolean
  tem_entrega: boolean
  encerramento_motivo?: string
  ultimo_documento?: UltimoDocumento | null
  coordenador_nome?: string | null
  itens?: ItemDoPlano[]
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroPlanos {
  [key: string]: unknown
  periodo_id: string
  curso_id: string
  situacao: "todas" | SituacaoPlano
  aprovacao: "todos" | "aprovados" | "sem_aprovacao"
}

export interface PlanoInput {
  curso_id: string
  periodo_id: string
  descricao: string
  objetivo_geral: string
  resultados_esperados: string
  alinhamento_pdi: string
  alinhamento_ppc: string
  aprovacao_data: string
  aprovacao_orgao: string
}

export interface ItemPlanoInput {
  meta_id: string
  quantidade: number
}

export interface DestinoDeCopia {
  curso_id: string
  curso_nome: string
  coordenador_nome: string | null
  vago: boolean
  ja_tem_plano: boolean
}

export interface CursoPulado {
  curso: { id: string }
  motivo: string
}

export interface ResultadoCopia {
  criados: number
  pulados: CursoPulado[]
}
