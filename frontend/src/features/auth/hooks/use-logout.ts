"use client"

import { useState } from "react"
import { apiClient } from "@/lib/api-client"

const CHAVE_MOTIVO_ENCERRAMENTO = "auth:motivo-encerramento"

export function useLogout() {
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function logout() {
    setIsSubmitting(true)
    try {
      await apiClient.post<void>("/auth/logout", undefined, { ignorarInterceptor401: true })
    } catch {
    } finally {
      window.sessionStorage.removeItem(CHAVE_MOTIVO_ENCERRAMENTO)
      setIsSubmitting(false)
      window.location.replace("/")
    }
  }

  return { logout, isSubmitting }
}
