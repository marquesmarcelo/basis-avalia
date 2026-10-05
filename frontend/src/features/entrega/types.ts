import type { IndicadorEmbutido } from "@/features/meta/types"

export type SituacaoEntrega = "pendente_avaliacao" | "aceita" | "recusada"

export interface Anexo {
  id: string
  nome_original: string
  tipo: "pdf" | "jpeg" | "png" | "docx" | "odt"
  tamanho_bytes: number
  hash_sha256: string
  criado_em: string
}

export interface Entrega {
  id: string
  item_plano_id: string
  curso_id: string
  curso_nome: string
  meta_nome?: string
  quantidade?: number
  indicadores?: IndicadorEmbutido[]
  situacao: SituacaoEntrega
  rodadas: number
  prazo_correcao_ate: string | null
  enviada_por_id: string
  enviada_por_nome: string
  corrigida_por_nome: string | null
  observacao: string
  motivo: string
  avaliada_por_nome: string | null
  avaliada_em: string | null
  avaliador_era_coordenador: boolean
  pendencia_vista_em: string | null
  coordenado_pelo_avaliador: boolean
  anexos: Anexo[]
  criado_em: string
  versao: number
}

export interface ItemMinhasMetas {
  item_plano_id: string
  meta_nome: string
  indicadores: IndicadorEmbutido[]
  quantidade: number
  aceitas: number
  pendentes: number
  em_correcao: number
}

export interface GrupoMinhasMetas {
  curso_id: string
  curso_nome: string
  itens: ItemMinhasMetas[]
}

export interface Pendencias {
  pendentes_de_avaliacao: number
  pendencias_nao_vistas: number
}

export interface FiltroFilaAvaliacao {
  [key: string]: unknown
  periodo_id: string
  curso_id: string
  meta_id: string
  situacao: string
}
