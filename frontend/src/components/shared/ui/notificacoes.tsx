"use client"

import { Toaster, toast } from "@/components/ui/toast"

export { Toaster }

export const notificar = {
  sucesso: (titulo: string, descricao?: string) =>
    toast.add({ title: titulo, description: descricao, type: "success", timeout: 4000 }),
  erro: (titulo: string, descricao?: string) =>
    toast.add({
      title: titulo,
      description: descricao,
      type: "error",
      timeout: 8000,
      priority: "high",
    }),
  aviso: (titulo: string, descricao?: string) =>
    toast.add({ title: titulo, description: descricao, type: "warning", timeout: 6000 }),
  info: (titulo: string, descricao?: string) =>
    toast.add({ title: titulo, description: descricao, type: "info", timeout: 4000 }),
}
