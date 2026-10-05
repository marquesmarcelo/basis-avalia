"use client"

import { useParams, useRouter } from "next/navigation"
import { useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { AreaDeAnexos } from "@/components/shared/forms/area-de-anexos"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Textarea } from "@/components/ui/textarea"
import { useRegistrarEntrega } from "@/features/entrega/hooks/use-registrar-entrega"

export default function RegistrarEntregaPage() {
  const params = useParams<{ itemId: string }>()
  const router = useRouter()
  const { registrar, cancelar, isSubmitting, error } = useRegistrarEntrega()
  const [observacao, setObservacao] = useState("")
  const [arquivos, setArquivos] = useState<File[]>([])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (arquivos.length === 0) {
      notificar.erro("Anexe pelo menos um comprovante.")
      return
    }
    const resultado = await registrar(params.itemId, observacao, arquivos)
    if (resultado) {
      notificar.sucesso("Entrega registrada.")
      router.push(`/app/itens/${params.itemId}/entregas`)
    }
  }

  return (
    <div className="w-full max-w-2xl space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Minhas metas", href: "/app/minhas-metas" },
          { rotulo: "Nova entrega" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Prestar contas</h1>

      {error && (
        <p
          role="alert"
          className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
        >
          {error.message || "Não foi possível registrar a entrega agora."}
        </p>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <fieldset disabled={isSubmitting} className="space-y-2">
          <label htmlFor="observacao" className="text-sm font-medium">
            Observação (opcional)
          </label>
          <Textarea
            id="observacao"
            value={observacao}
            onChange={(e) => setObservacao(e.target.value)}
            placeholder="Ex.: Reunião ordinária do NDE de 12/03/2026."
          />
        </fieldset>

        <AreaDeAnexos
          modo="agregado"
          arquivosNaFila={arquivos}
          onArquivosNaFilaChange={setArquivos}
          enviandoAgregado={isSubmitting}
          disabled={isSubmitting}
        />

        <div className="flex gap-3">
          <LoadingButton type="submit" loading={isSubmitting} loadingText="Enviando...">
            Registrar
          </LoadingButton>
          {isSubmitting && (
            <LoadingButton type="button" variant="outline" onClick={cancelar}>
              Cancelar envio
            </LoadingButton>
          )}
        </div>
      </form>
    </div>
  )
}
