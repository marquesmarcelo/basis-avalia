"use client"

import { useParams, useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { ComboboxEntidade } from "@/components/shared/forms/combobox-entidade"
import { useDebounce } from "@/components/shared/hooks/use-debounce"
import { ErrorState } from "@/components/shared/ui/error-state"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Skeleton } from "@/components/ui/skeleton"
import { ResultadoCopiaDialog } from "@/features/plano/components/resultado-copia-dialog"
import { SelecaoCursosLote } from "@/features/plano/components/selecao-cursos-lote"
import { useCopiarEmLote } from "@/features/plano/hooks/use-copiar-em-lote"
import { useDestinosDeCopia } from "@/features/plano/hooks/use-destinos-de-copia"
import { usePeriodosSugestoes } from "@/features/plano/hooks/use-periodos-sugestoes"
import { usePlano } from "@/features/plano/hooks/use-plano"

export default function CopiarPlanoPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()

  const { data: plano, isLoading: carregandoPlano, error: erroPlano, carregar } = usePlano("planos")
  const { itens: periodos, buscar: buscarPeriodos } = usePeriodosSugestoes()
  const [periodoDestinoId, setPeriodoDestinoId] = useState<string | null>(null)
  const [busca, setBusca] = useState("")
  const buscaComDebounce = useDebounce(busca, 300)
  const [selecionados, setSelecionados] = useState<string[]>([])

  const {
    itens: destinos,
    isLoading: carregandoDestinos,
    buscar: buscarDestinos,
  } = useDestinosDeCopia()
  const { copiar, isSubmitting: copiando, error: erroCopia } = useCopiarEmLote()
  const [resultado, setResultado] = useState<Awaited<ReturnType<typeof copiar>>>(null)

  useEffect(() => {
    carregar(params.id)
    buscarPeriodos("")
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id])

  useEffect(() => {
    if (periodoDestinoId) {
      buscarDestinos(params.id, periodoDestinoId, buscaComDebounce)
      setSelecionados([])
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [periodoDestinoId, buscaComDebounce])

  async function handleCopiar() {
    if (!periodoDestinoId || selecionados.length === 0) return
    const res = await copiar(params.id, periodoDestinoId, selecionados)
    if (res) setResultado(res)
  }

  if (carregandoPlano || !plano) {
    return (
      <div className="w-full space-y-6">
        <Trilha
          itens={[
            { rotulo: "Início", href: "/app" },
            { rotulo: "Planos", href: "/app/planos" },
            { rotulo: "Copiar" },
          ]}
        />
        {erroPlano ? (
          <ErrorState
            titulo="Não foi possível carregar o plano de origem."
            onTentarNovamente={() => carregar(params.id)}
          />
        ) : (
          <Skeleton className="h-64 w-full" />
        )}
      </div>
    )
  }

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Planos", href: "/app/planos" },
          {
            rotulo: `${plano.curso.nome} — ${plano.periodo.nome}`,
            href: `/app/planos/${plano.id}`,
          },
          { rotulo: "Copiar" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Copiar plano para vários cursos</h1>

      <div className="max-w-3xl space-y-4">
        <div className="rounded-md border p-3 text-sm">
          <p className="font-medium">
            Origem: {plano.curso.nome} — {plano.periodo.nome}
          </p>
          <p className="text-muted-foreground">
            {plano.metas} meta(s), {plano.total_exigido} entrega(s) exigida(s).
          </p>
        </div>

        <div className="max-w-xs">
          <label htmlFor="periodo-destino" className="mb-1 block text-sm font-medium">
            Período de destino
          </label>
          <ComboboxEntidade
            id="periodo-destino"
            itens={periodos.map((p) => ({ value: p.id, label: p.nome }))}
            value={periodoDestinoId}
            onValueChange={setPeriodoDestinoId}
          />
        </div>

        {periodoDestinoId && (
          <SelecaoCursosLote
            destinos={destinos}
            carregando={carregandoDestinos}
            busca={busca}
            onBuscaChange={setBusca}
            selecionados={selecionados}
            onSelecionadosChange={setSelecionados}
          />
        )}

        <div className="rounded-md border p-3 text-sm">
          <p className="font-medium">O que vai junto</p>
          <p className="text-muted-foreground">
            Descrição, objetivo geral, resultados esperados, alinhamentos e os itens com as
            quantidades.
          </p>
          <p className="mt-2 font-medium">O que NÃO vai junto</p>
          <p className="text-muted-foreground">
            Entregas, anexos, avaliações, prazos e os dados de aprovação. Todo plano copiado nasce
            em rascunho.
          </p>
        </div>

        {erroCopia && (
          <p className="text-sm text-destructive">Não foi possível copiar o plano agora.</p>
        )}

        <div className="flex justify-end">
          <LoadingButton
            loading={copiando}
            loadingText="Copiando..."
            disabled={selecionados.length === 0}
            onClick={handleCopiar}
          >
            Copiar para {selecionados.length} curso(s)
          </LoadingButton>
        </div>
      </div>

      <ResultadoCopiaDialog
        resultado={resultado}
        onVerPlanosCriados={() =>
          router.push(`/app/planos?periodo_id=${periodoDestinoId}&situacao=rascunho`)
        }
      />
    </div>
  )
}
