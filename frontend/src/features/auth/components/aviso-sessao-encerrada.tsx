"use client"

import { useEffect, useState } from "react"
import { Alert, AlertDescription } from "@/components/ui/alert"

const CHAVE_MOTIVO = "auth:motivo-encerramento"

const MENSAGENS: Record<string, string> = {
  INSTITUICAO_INATIVA:
    "Sua instituição foi desativada. Entre em contato com o Administrador do Sistema.",
  CONTA_EXCLUIDA: "Sua conta foi removida. Entre em contato com quem administra o sistema.",
  SESSAO_EXPIRADA: "Sua sessão expirou. Entre novamente.",
}

export function AvisoSessaoEncerrada() {
  const [mensagem, setMensagem] = useState<string | null>(null)

  useEffect(() => {
    const codigo = window.sessionStorage.getItem(CHAVE_MOTIVO)
    if (codigo) {
      setMensagem(MENSAGENS[codigo] ?? MENSAGENS.SESSAO_EXPIRADA)
      window.sessionStorage.removeItem(CHAVE_MOTIVO)
    }
  }, [])

  if (!mensagem) {
    return null
  }

  return (
    <Alert variant="destructive" className="mb-4 w-full max-w-[420px]">
      <AlertDescription>{mensagem}</AlertDescription>
    </Alert>
  )
}
