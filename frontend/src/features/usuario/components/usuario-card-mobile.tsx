import { KeyIcon, PencilIcon, Trash2Icon } from "lucide-react"
import Link from "next/link"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import type { Usuario } from "../types"

interface UsuarioCardMobileProps {
  item: Usuario
  usuarioAtualId: string
  basePath: string
  onRedefinirSenha: (item: Usuario) => void
  onExcluir: (item: Usuario) => void
  idExcluindo: string | null
}

export function UsuarioCardMobile({
  item,
  usuarioAtualId,
  basePath,
  onRedefinirSenha,
  onExcluir,
  idExcluindo,
}: UsuarioCardMobileProps) {
  const propriaLinha = item.id === usuarioAtualId

  return (
    <Card>
      <CardContent className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-2">
          <div>
            <p className="font-medium">{item.nome}</p>
            <p className="text-sm text-muted-foreground">{item.perfis_rotulos.join(", ")}</p>
          </div>
          {item.senha_provisoria ? (
            <StatusBadge label="Primeiro acesso pendente" variant="secondary" />
          ) : (
            <StatusBadge label="Ativo" variant="default" />
          )}
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            render={<Link href={`/app${basePath}/${item.id}`} aria-label={`Editar ${item.nome}`} />}
          >
            <PencilIcon aria-hidden="true" />
            <IndicadorDeNavegacao />
          </Button>
          {!propriaLinha && (
            <>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={`Redefinir senha de ${item.nome}`}
                onClick={() => onRedefinirSenha(item)}
              >
                <KeyIcon aria-hidden="true" />
              </Button>
              <LoadingButton
                variant="ghost"
                size="icon-sm"
                loading={idExcluindo === item.id}
                aria-label={`Excluir ${item.nome}`}
                onClick={() => onExcluir(item)}
              >
                <Trash2Icon aria-hidden="true" />
              </LoadingButton>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
