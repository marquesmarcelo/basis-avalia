"use client"

import { useRouter } from "next/navigation"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useLogout } from "@/features/auth/hooks/use-logout"
import type { ContextoDeSessao } from "@/features/auth/types"

interface MenuUsuarioProps {
  eu: ContextoDeSessao
  apenasSair?: boolean
}

function iniciais(nome: string): string {
  const partes = nome.trim().split(/\s+/)
  const primeira = partes[0]?.[0] ?? ""
  const ultima = partes.length > 1 ? partes[partes.length - 1][0] : ""
  return (primeira + ultima).toUpperCase()
}

function contextoDaSessao(eu: ContextoDeSessao): { visivel: string; completo: string } {
  const contexto = eu.instituicao ? eu.instituicao.nome : "Administração do sistema"
  if (!eu.instituicao) {
    return { visivel: contexto, completo: contexto }
  }
  const perfisVisiveis = eu.perfis_rotulos.slice(0, 2).join(", ")
  const resumo = eu.perfis_rotulos.length > 2 ? ` +${eu.perfis_rotulos.length - 2}` : ""
  return {
    visivel: `${contexto} · ${perfisVisiveis}${resumo}`,
    completo: `${contexto} · ${eu.perfis_rotulos.join(", ")}`,
  }
}

export function MenuUsuario({ eu, apenasSair = false }: MenuUsuarioProps) {
  const router = useRouter()
  const { logout, isSubmitting } = useLogout()
  const contexto = contextoDaSessao(eu)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <button type="button" aria-label="Abrir menu do usuário" disabled={isSubmitting}>
            <Avatar>
              <AvatarFallback>{iniciais(eu.nome)}</AvatarFallback>
            </Avatar>
          </button>
        }
      />
      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuGroup>
          <DropdownMenuLabel className="flex flex-col gap-1 py-2">
            <span className="text-sm font-medium">{eu.nome}</span>
            <span className="text-xs text-muted-foreground">{eu.email}</span>
            <span className="text-xs text-muted-foreground">
              <span aria-hidden="true">{contexto.visivel}</span>
              <span className="sr-only">{contexto.completo}</span>
            </span>
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        {!apenasSair && (
          <DropdownMenuItem onClick={() => router.push("/app/alterar-senha")}>
            Alterar minha senha
          </DropdownMenuItem>
        )}
        <DropdownMenuItem onClick={() => logout()}>Sair</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
