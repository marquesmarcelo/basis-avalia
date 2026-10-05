"use client"

import { LoadingButton } from "@/components/shared/ui/loading-button"
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import type { Plano } from "../types"

interface PublicarDialogProps {
  plano: Plano | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: () => void
}

export function PublicarDialog({
  plano,
  onOpenChange,
  confirmando,
  onConfirmar,
}: PublicarDialogProps) {
  if (!plano) return null
  return (
    <AlertDialog open={!!plano} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle>Publicar plano de {plano.curso.nome}?</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2 text-left">
            <span className="block">
              {plano.metas} meta(s) · {plano.total_exigido} entrega(s) exigida(s).
            </span>
            {plano.curso.vago && (
              <span className="block">
                Este curso está sem coordenador designado — ninguém poderá registrar entrega
                enquanto estiver vago, mas as metas continuam sendo cobradas.
              </span>
            )}
            {plano.sem_aprovacao && (
              <span className="block">
                Ainda não há aprovação registrada. Você pode publicar mesmo assim e informar depois.
              </span>
            )}
            <span className="block">
              A partir da publicação, o coordenador passa a ver este plano em Minhas metas.
            </span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button
            variant="outline"
            autoFocus
            onClick={() => onOpenChange(false)}
            disabled={confirmando}
          >
            Cancelar
          </Button>
          <LoadingButton loading={confirmando} loadingText="Publicando..." onClick={onConfirmar}>
            Publicar
          </LoadingButton>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
