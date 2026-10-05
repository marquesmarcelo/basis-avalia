"use client"

import { usePathname, useRouter } from "next/navigation"
import { useEffect, useRef } from "react"
import { notificar } from "@/components/shared/ui/notificacoes"
import { apiClient } from "@/lib/api-client"
import type { ContextoDeSessao } from "../types"

function mesmoConjunto(a: string[], b: string[]): boolean {
  if (a.length !== b.length) return false
  const ordenadoA = [...a].sort()
  const ordenadoB = [...b].sort()
  return ordenadoA.every((p, i) => p === ordenadoB[i])
}

interface CursoResumo {
  id: string
  nome: string
}

async function buscarMeusCursos(): Promise<CursoResumo[]> {
  try {
    return await apiClient.get<CursoResumo[]>("/meus-cursos", { ignorarInterceptor401: true })
  } catch {
    return []
  }
}

async function buscarPortariaVigente(
  cursoId: string,
  coordenadorId: string
): Promise<string | null> {
  try {
    const query = new URLSearchParams({
      coordenador_id: coordenadorId,
      situacao: "vigente",
      page_size: "1",
    })
    const resposta = await apiClient.get<{ data: { portaria: string }[] }>(
      `/cursos/${cursoId}/designacoes?${query.toString()}`,
      { ignorarInterceptor401: true }
    )
    return resposta.data[0]?.portaria ?? null
  } catch {
    return null
  }
}

export function useSincronizacaoDePerfis(
  eu: ContextoDeSessao | null,
  recarregar: () => Promise<ContextoDeSessao | null>
) {
  const pathname = usePathname()
  const router = useRouter()
  const euRef = useRef(eu)
  const cursosCoordenadosRef = useRef<CursoResumo[]>([])
  const montadoRef = useRef(false)

  useEffect(() => {
    euRef.current = eu
  }, [eu])

  async function sincronizarCoordenacao(
    usuarioId: string,
    coordenaAntes: boolean,
    coordenaAgora: boolean
  ) {
    if (!coordenaAntes && !coordenaAgora) return

    const cursosAntes = cursosCoordenadosRef.current
    const cursosAgora = coordenaAgora ? await buscarMeusCursos() : []
    cursosCoordenadosRef.current = cursosAgora

    const ganhos = cursosAgora.filter((c) => !cursosAntes.some((a) => a.id === c.id))
    const perdidos = cursosAntes.filter((a) => !cursosAgora.some((c) => c.id === a.id))

    for (const curso of ganhos) {
      const portaria = await buscarPortariaVigente(curso.id, usuarioId)
      notificar.info(
        `Você agora coordena o curso ${curso.nome}${portaria ? `, pela Portaria ${portaria}` : ""}. A área Minhas metas está disponível no menu.`
      )
    }
    for (const curso of perdidos) {
      notificar.info(`Sua designação como coordenador de ${curso.nome} foi encerrada.`)
    }
  }

  async function verificar() {
    const anterior = euRef.current
    if (!anterior) return
    const coordenavaAntes = anterior.perfis.includes("coordenador_curso")

    const novo = await recarregar()
    if (!novo) return

    const coordenaAgora = novo.perfis.includes("coordenador_curso")
    await sincronizarCoordenacao(novo.id, coordenavaAntes, coordenaAgora)

    if (!mesmoConjunto(anterior.perfis, novo.perfis)) {
      notificar.info("Seus perfis foram alterados. A navegação foi atualizada.")
    }
    if (coordenavaAntes && !coordenaAgora && pathname.startsWith("/app/minhas-metas")) {
      router.replace("/app")
    }
  }

  useEffect(() => {
    if (!montadoRef.current) {
      montadoRef.current = true
      return
    }
    verificar()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname])

  useEffect(() => {
    function handler() {
      if (document.visibilityState === "visible") {
        verificar()
      }
    }
    window.addEventListener("focus", handler)
    document.addEventListener("visibilitychange", handler)
    return () => {
      window.removeEventListener("focus", handler)
      document.removeEventListener("visibilitychange", handler)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
}
