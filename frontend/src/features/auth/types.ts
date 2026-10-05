export interface InstituicaoResumo {
  id: string
  nome: string
  sigla: string
}

export interface ContextoDeSessao {
  id: string
  nome: string
  email: string
  perfis: string[]
  perfis_rotulos: string[]
  senha_provisoria: boolean
  instituicao: InstituicaoResumo | null
  permissoes: string[]
}

export interface InstituicaoPublica {
  id: string
  nome: string
  sigla: string
}
