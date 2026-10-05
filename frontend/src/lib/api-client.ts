"use client"

import { concluirCarregamento, iniciarCarregamento } from "@/lib/carregamento"

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:3001/api/v1"

export interface ErroApi {
  status: number
  code: string
  message: string
  campo?: string
}

interface OpcoesRequisicao {
  ignorarInterceptor401?: boolean
}

function motivoEncerramentoValido(codigo: string): boolean {
  return (
    codigo === "SESSAO_EXPIRADA" || codigo === "CONTA_EXCLUIDA" || codigo === "INSTITUICAO_INATIVA"
  )
}

async function lerJsonSeguro(
  resposta: Response
): Promise<{ error?: { code?: string; message?: string; campo?: string } } | null> {
  return resposta.json().catch(() => null)
}

async function requisitar<T>(
  caminho: string,
  init?: RequestInit,
  opcoes?: OpcoesRequisicao
): Promise<T> {
  iniciarCarregamento()
  try {
    const resposta = await fetch(`${BASE_URL}${caminho}`, {
      ...init,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    })

    if (resposta.status === 401 && !opcoes?.ignorarInterceptor401) {
      const corpo = await lerJsonSeguro(resposta)
      const codigo = corpo?.error?.code as string | undefined
      if (typeof window !== "undefined") {
        window.sessionStorage.setItem(
          "auth:motivo-encerramento",
          codigo && motivoEncerramentoValido(codigo) ? codigo : "SESSAO_EXPIRADA"
        )
        window.location.replace("/")
      }
      const erro: ErroApi = {
        status: 401,
        code: codigo ?? "SESSAO_EXPIRADA",
        message: corpo?.error?.message ?? "",
      }
      throw erro
    }

    if (resposta.status === 403) {
      const corpo = await lerJsonSeguro(resposta)
      if (corpo?.error?.code === "SENHA_PROVISORIA" && typeof window !== "undefined") {
        window.location.href = "/app/alterar-senha"
      }
      const erro: ErroApi = {
        status: 403,
        code: corpo?.error?.code ?? "PERMISSAO_NEGADA",
        message: corpo?.error?.message ?? "",
      }
      throw erro
    }

    if (!resposta.ok) {
      const corpo = await lerJsonSeguro(resposta)
      const erro: ErroApi = {
        status: resposta.status,
        code: corpo?.error?.code ?? "ERRO_INTERNO",
        message: corpo?.error?.message ?? "Não foi possível concluir a operação agora.",
        campo: corpo?.error?.campo,
      }
      throw erro
    }

    if (resposta.status === 204) {
      return undefined as T
    }
    return (await resposta.json()) as T
  } finally {
    concluirCarregamento()
  }
}

export const apiClient = {
  get: <T>(caminho: string, opcoes?: OpcoesRequisicao) =>
    requisitar<T>(caminho, { method: "GET" }, opcoes),
  post: <T>(caminho: string, corpo?: unknown, opcoes?: OpcoesRequisicao) =>
    requisitar<T>(
      caminho,
      { method: "POST", body: corpo !== undefined ? JSON.stringify(corpo) : undefined },
      opcoes
    ),
  put: <T>(caminho: string, corpo?: unknown, opcoes?: OpcoesRequisicao) =>
    requisitar<T>(
      caminho,
      { method: "PUT", body: corpo !== undefined ? JSON.stringify(corpo) : undefined },
      opcoes
    ),
  patch: <T>(caminho: string, corpo?: unknown, opcoes?: OpcoesRequisicao) =>
    requisitar<T>(
      caminho,
      { method: "PATCH", body: corpo !== undefined ? JSON.stringify(corpo) : undefined },
      opcoes
    ),
  delete: <T>(caminho: string, opcoes?: OpcoesRequisicao) =>
    requisitar<T>(caminho, { method: "DELETE" }, opcoes),
}
