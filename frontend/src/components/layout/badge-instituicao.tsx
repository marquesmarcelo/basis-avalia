"use client"

import { Badge } from "@/components/ui/badge"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import type { ContextoDeSessao } from "@/features/auth/types"

interface BadgeInstituicaoProps {
  eu: ContextoDeSessao
}

export function BadgeInstituicao({ eu }: BadgeInstituicaoProps) {
  const nomeCompleto = eu.instituicao
    ? `${eu.instituicao.nome} (${eu.instituicao.sigla})`
    : "Administração do sistema"
  const sigla = eu.instituicao ? eu.instituicao.sigla : "Sistema"

  return (
    <>
      <Badge variant="outline" className="hidden xl:inline-flex">
        {nomeCompleto}
      </Badge>
      <Tooltip>
        <TooltipTrigger
          render={
            <Badge
              variant="outline"
              tabIndex={0}
              className="hidden max-w-40 truncate md:inline-flex xl:hidden"
            />
          }
        >
          {nomeCompleto}
        </TooltipTrigger>
        <TooltipContent>{nomeCompleto}</TooltipContent>
      </Tooltip>
      <Badge variant="outline" className="inline-flex md:hidden">
        {sigla}
      </Badge>
    </>
  )
}
