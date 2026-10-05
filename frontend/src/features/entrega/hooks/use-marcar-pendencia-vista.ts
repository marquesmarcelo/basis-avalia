"use client"

import { apiClient } from "@/lib/api-client"

export function useMarcarPendenciaVista() {
  async function marcarVista(id: string): Promise<void> {
    await apiClient.post(`/entregas/${id}/pendencia-vista`).catch(() => {})
  }
  return { marcarVista }
}
