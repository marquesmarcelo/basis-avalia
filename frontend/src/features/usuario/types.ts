export interface Usuario {
  id: string
  nome: string
  email: string
  perfis: string[]
  perfis_rotulos: string[]
  senha_provisoria: boolean
  criado_em: string
  atualizado_em: string | null
  versao: number
}

export interface FiltroUsuarios {
  [key: string]: unknown
  busca: string
  perfil: string
}

export interface UsuarioInput {
  nome: string
  email: string
  perfis?: string[]
  senha: string
}

export interface UsuarioAtualizarInput {
  nome: string
  email: string
  perfis?: string[]
}

export const PERFIS_ESCOLHIVEIS: { value: string; label: string }[] = [
  { value: "aluno", label: "Aluno" },
  { value: "professor", label: "Professor" },
  { value: "coordenador_curso", label: "Coordenador de Curso" },
  { value: "pesquisador_institucional", label: "Pesquisador Institucional" },
]
