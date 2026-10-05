"use client"

import { useEffect, useId, useRef, useState } from "react"
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
import { Field, FieldContent, FieldError, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import type { Plano } from "../types"

interface EncerrarDialogProps {
  plano: Plano | null
  onOpenChange: (open: boolean) => void
  confirmando: boolean
  onConfirmar: (motivo: string) => void
}

export function EncerrarDialog({
  plano,
  onOpenChange,
  confirmando,
  onConfirmar,
}: EncerrarDialogProps) {
  const idBase = useId()
  const campoMotivoRef = useRef<HTMLTextAreaElement>(null)
  const [motivo, setMotivo] = useState("")
  const [erro, setErro] = useState<string | undefined>()

  useEffect(() => {
    if (plano) {
      setMotivo("")
      setErro(undefined)
      requestAnimationFrame(() => campoMotivoRef.current?.focus())
    }
  }, [plano])

  if (!plano) return null

  function handleConfirmar() {
    if (!motivo.trim()) {
      setErro("Informe o motivo do encerramento.")
      return
    }
    onConfirmar(motivo)
  }

  return (
    <AlertDialog open={!!plano} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle>
            Encerrar antecipadamente o plano de {plano.curso.nome}?
          </AlertDialogTitle>
          <AlertDialogDescription>
            A partir de agora, nenhuma entrega nova é aceita — avaliar as pendentes continua
            possível.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <Field data-invalid={!!erro}>
          <FieldLabel htmlFor={`${idBase}-motivo`}>Motivo</FieldLabel>
          <FieldContent>
            <Textarea
              id={`${idBase}-motivo`}
              ref={campoMotivoRef}
              value={motivo}
              onChange={(e) => setMotivo(e.target.value)}
              disabled={confirmando}
              aria-required="true"
            />
            {erro && <FieldError>{erro}</FieldError>}
          </FieldContent>
        </Field>
        <AlertDialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={confirmando}>
            Cancelar
          </Button>
          <LoadingButton
            loading={confirmando}
            loadingText="Encerrando..."
            onClick={handleConfirmar}
          >
            Encerrar
          </LoadingButton>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
