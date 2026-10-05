export interface CoordenadorDoCurso {
  id: string
  nome: string
  data_fim: string | null
  tambem_pesquisador_institucional: boolean
}

export type GrauDeCurso = "bacharelado" | "licenciatura" | "tecnologo"
export type ModalidadeDeCurso = "presencial" | "a_distancia"
export type SituacaoCurso = "ativo" | "inativo"

export interface Curso {
  id: string
  nome: string
  codigo_emec: string | null
  grau: GrauDeCurso
  modalidade: ModalidadeDeCurso
  situacao: SituacaoCurso
  coordenador: CoordenadorDoCurso | null
  tem_vinculo: boolean
  plano_do_periodo: string | null
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroCursos {
  [key: string]: unknown
  busca: string
  grau: GrauDeCurso | "todos"
  modalidade: ModalidadeDeCurso | "todas"
  coordenador_id: string
  situacao: SituacaoCurso | "todas"
}

export interface CursoInput {
  nome: string
  codigo_emec: string
  grau: GrauDeCurso
  modalidade: ModalidadeDeCurso
}
