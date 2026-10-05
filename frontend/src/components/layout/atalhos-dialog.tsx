"use client"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

interface AtalhosDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const ATALHOS = [
  { tecla: "Ctrl+N", acao: "Abrir Novo usuário / Nova instituição" },
  { tecla: "Ctrl+S", acao: "Salvar o formulário atual" },
  { tecla: "Ctrl+F", acao: "Focar o campo de busca" },
  { tecla: "Esc", acao: "Fechar modal aberto" },
  { tecla: "?", acao: "Abrir esta lista de atalhos" },
]

export function AtalhosDialog({ open, onOpenChange }: AtalhosDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Atalhos de teclado</DialogTitle>
        </DialogHeader>
        <dl className="flex flex-col gap-2 text-sm">
          {ATALHOS.map((atalho) => (
            <div key={atalho.tecla} className="flex items-center justify-between gap-4">
              <dt className="rounded border px-1.5 py-0.5 font-mono text-xs">{atalho.tecla}</dt>
              <dd className="text-muted-foreground">{atalho.acao}</dd>
            </div>
          ))}
        </dl>
      </DialogContent>
    </Dialog>
  )
}
