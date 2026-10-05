import {
  CopyIcon,
  DownloadIcon,
  PauseIcon,
  PlayIcon,
  RotateCcwIcon,
  SquareIcon,
} from "lucide-react"
import Link from "next/link"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { StatusBadge } from "@/components/shared/ui/status-badge"
import { Button } from "@/components/ui/button"
import type { Plano } from "../types"

const ROTULO_SITUACAO: Record<Plano["situacao"], string> = {
  rascunho: "Rascunho",
  vigente: "Vigente",
  encerrado: "Encerrado",
}

interface PlanoCabecalhoProps {
  plano: Plano
  somenteLeitura: boolean
  gerandoDocumento: boolean
  publicando: boolean
  despublicando: boolean
  reabrindo: boolean
  onGerarDocumento: () => void
  onPublicar: () => void
  onDespublicar: () => void
  onEncerrar: () => void
  onReabrir: () => void
}

export function PlanoCabecalho({
  plano,
  somenteLeitura,
  gerandoDocumento,
  publicando,
  despublicando,
  reabrindo,
  onGerarDocumento,
  onPublicar,
  onDespublicar,
  onEncerrar,
  onReabrir,
}: PlanoCabecalhoProps) {
  const periodoEncerradoDoPlano = plano.situacao === "encerrado"
  const notaDocumento =
    plano.situacao === "rascunho"
      ? "O documento trará a marca «RASCUNHO» visível, porque este plano ainda não foi publicado."
      : plano.situacao === "encerrado"
        ? `O documento trará a marca «ENCERRADO${plano.encerramento_motivo ? ` — ${plano.encerramento_motivo}` : ""}».`
        : plano.sem_aprovacao
          ? "O documento trará a observação «Aprovação ainda não registrada»."
          : null

  return (
    <div className="space-y-2 rounded-lg border p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-xl font-semibold">{plano.curso.nome}</h1>
          <span className="text-muted-foreground">· {plano.periodo.nome}</span>
          <StatusBadge
            label={ROTULO_SITUACAO[plano.situacao]}
            variant={plano.situacao === "vigente" ? "default" : "secondary"}
          />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {!somenteLeitura && (
            <Button
              variant="outline"
              size="sm"
              render={<Link href={`/app/planos/${plano.id}/copiar`} />}
            >
              <CopyIcon aria-hidden="true" /> Copiar
            </Button>
          )}
          <div className="flex flex-col items-end gap-1">
            <LoadingButton
              variant="outline"
              size="sm"
              loading={gerandoDocumento}
              loadingText="Gerando..."
              onClick={onGerarDocumento}
            >
              <DownloadIcon aria-hidden="true" /> Documento
            </LoadingButton>
            {notaDocumento && (
              <p className="max-w-xs text-right text-xs text-muted-foreground">{notaDocumento}</p>
            )}
          </div>
          {!somenteLeitura && plano.situacao === "rascunho" && (
            <LoadingButton loading={publicando} loadingText="Publicando..." onClick={onPublicar}>
              <PlayIcon aria-hidden="true" /> Publicar
            </LoadingButton>
          )}
          {!somenteLeitura && plano.situacao === "vigente" && (
            <LoadingButton
              variant="outline"
              loading={despublicando}
              loadingText="Despublicando..."
              disabled={plano.tem_entrega}
              aria-describedby={plano.tem_entrega ? "despublicar-bloqueado" : undefined}
              onClick={onDespublicar}
            >
              <PauseIcon aria-hidden="true" /> Despublicar
            </LoadingButton>
          )}
          {plano.tem_entrega && plano.situacao === "vigente" && !somenteLeitura && (
            <span id="despublicar-bloqueado" className="sr-only">
              Não é possível despublicar: este plano já tem entregas registradas.
            </span>
          )}
          {!somenteLeitura && plano.situacao === "vigente" && (
            <Button variant="outline" onClick={onEncerrar}>
              <SquareIcon aria-hidden="true" /> Encerrar
            </Button>
          )}
          {!somenteLeitura && periodoEncerradoDoPlano && !!plano.encerramento_motivo && (
            <LoadingButton
              variant="outline"
              loading={reabrindo}
              loadingText="Reabrindo..."
              onClick={onReabrir}
            >
              <RotateCcwIcon aria-hidden="true" /> Reabrir
            </LoadingButton>
          )}
        </div>
      </div>
      <p className="text-sm text-muted-foreground">
        Coordenador: {plano.coordenador_nome ?? "Sem responsável"}
      </p>
      {plano.situacao === "encerrado" && (
        <p className="text-sm text-muted-foreground">
          {plano.encerramento_motivo
            ? `Encerrado antecipadamente: ${plano.encerramento_motivo}`
            : "Encerrado — período encerrado."}
        </p>
      )}
    </div>
  )
}
