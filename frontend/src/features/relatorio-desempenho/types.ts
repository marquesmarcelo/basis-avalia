import type { IndicadorEmbutido } from "@/features/meta/types"

export interface LinhaRelatorio {
  item_plano_id: string
  curso_nome: string
  curso_vago: boolean
  responsavel_nome: string | null
  responsavel_desde: string | null
  vago_desde: string | null
  meta_nome: string
  indicadores: IndicadorEmbutido[]
  exigido: number
  aceitas: number
  pendentes: number
  em_correcao: number
  cumprimento: number
  situacao: "cumprida" | "sem_responsavel" | "em_andamento" | "em_correcao" | "nao_cumprida"
  inclui_entregas_anteriores: boolean
  avaliacao_pelo_proprio_coordenador: boolean
}

export interface ResumoRelatorio {
  cursos_sem_coordenador: number
  metas_nao_cumpridas_de_vagos: number
}

export interface ItemDesempenhoPorCurso {
  curso_id: string
  curso_nome: string
  responsavel_nome: string
  exigido_total: number
  aceitas_total: number
  percentual: number
}

export interface MetaDesempenhoPorCurso {
  total_cursos_com_coordenador: number
  total_cursos_sem_coordenador: number
}

export interface FiltroRelatorio {
  [key: string]: unknown
  periodo_id: string
  curso_id: string
  meta_id: string
  indicador_id: string
  origem: string
  situacao: string
  autoavaliado: string
  incluir_inativos: string
}
