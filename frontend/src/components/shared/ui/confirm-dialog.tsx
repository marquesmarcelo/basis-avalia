"use client"

import { LoadingButton } from "@/components/shared/ui/loading-button"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"

interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  titulo: string
  descricao: string
  rotuloConfirmar?: string
  rotuloConfirmando?: string
  rotuloCancelar?: string
  destrutivo?: boolean
  confirmando?: boolean
  onConfirmar: () => void
}

export function ConfirmDialog({
  open,
  onOpenChange,
  titulo,
  descricao,
  rotuloConfirmar = "Confirmar",
  rotuloConfirmando = "Confirmando...",
  rotuloCancelar = "Cancelar",
  destrutivo = false,
  confirmando = false,
  onConfirmar,
}: ConfirmDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{titulo}</AlertDialogTitle>
          <AlertDialogDescription>{descricao}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel autoFocus disabled={confirmando}>
            {rotuloCancelar}
          </AlertDialogCancel>
          <AlertDialogAction
            render={
              <LoadingButton
                variant={destrutivo ? "destructive" : "default"}
                loading={confirmando}
                loadingText={rotuloConfirmando}
                onClick={onConfirmar}
              />
            }
          >
            {rotuloConfirmar}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
