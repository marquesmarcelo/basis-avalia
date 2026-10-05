export interface MetaPaginacao {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface RespostaPaginada<T> {
  data: T[]
  meta: MetaPaginacao
}

export interface ParametrosListagem {
  page: number
  page_size: number
  sort: string
  order: "asc" | "desc"
}
