"use client"

import { Loader2Icon } from "lucide-react"
import { useLinkStatus } from "next/link"
import { useEffect } from "react"
import { concluirCarregamento, iniciarCarregamento } from "@/lib/carregamento"

interface IndicadorDeNavegacaoProps {
  className?: string
}

export function IndicadorDeNavegacao({ className = "size-3.5" }: IndicadorDeNavegacaoProps) {
  const { pending } = useLinkStatus()

  useEffect(() => {
    if (!pending) return
    iniciarCarregamento()
    return () => concluirCarregamento()
  }, [pending])

  if (!pending) return null
  return (
    <Loader2Icon
      className={`animate-spin motion-reduce:animate-none ${className}`}
      aria-hidden="true"
    />
  )
}
