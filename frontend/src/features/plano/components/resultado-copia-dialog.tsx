"use client"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import type { ResultadoCopia } from "../types"

interface ResultadoCopiaDialogProps {
  resultado: ResultadoCopia | null
  onVerPlanosCriados: () => void
}

export function ResultadoCopiaDialog({ resultado, onVerPlanosCriados }: ResultadoCopiaDialogProps) {
  if (!resultado) return null
  return (
    <AlertDialog open={!!resultado}>
      <AlertDialogContent className="max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle role="status">
            ✓ {resultado.criados} planos criados, em rascunho.
          </AlertDialogTitle>
          {resultado.pulados.length > 0 && (
            <AlertDialogDescription className="space-y-1 text-left">
              <span className="block font-medium">⚠ {resultado.pulados.length} cursos pulados</span>
              <ul className="list-inside list-disc">
                {resultado.pulados.map((p) => (
                  <li key={p.curso.id}>{p.motivo}</li>
                ))}
              </ul>
            </AlertDialogDescription>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction onClick={onVerPlanosCriados}>Ver planos criados</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
