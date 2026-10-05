import { InfoIcon } from "lucide-react"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Progress } from "@/components/ui/progress"
import { Skeleton } from "@/components/ui/skeleton"
import type { ErroApi } from "@/lib/api-client"
import type { ItemDesempenhoPorCurso, MetaDesempenhoPorCurso } from "../types"

interface GraficoDesempenhoPorCursoProps {
  itens: ItemDesempenhoPorCurso[]
  cursosSemCoordenador: string[]
  meta: MetaDesempenhoPorCurso | null
  isLoading: boolean
  error: ErroApi | null
  onTentarNovamente: () => void
}

const LARGURAS_ESQUELETO = [95, 70, 85, 55, 90, 65, 80, 45, 75, 60]

export function GraficoDesempenhoPorCurso({
  itens,
  cursosSemCoordenador,
  meta,
  isLoading,
  error,
  onTentarNovamente,
}: GraficoDesempenhoPorCursoProps) {
  return (
    <section aria-labelledby="grafico-cursos-titulo" className="space-y-3 rounded-lg border p-4">
      <h2 id="grafico-cursos-titulo" className="text-base font-semibold">
        Cumprimento de metas por curso — comprovantes entregues
      </h2>

      {isLoading && (
        <ul aria-busy="true" className="space-y-3">
          {LARGURAS_ESQUELETO.map((largura) => (
            <li key={`linha-esqueleto-${largura}`} className="space-y-1.5">
              <Skeleton className="h-4 w-40" />
              <Skeleton className="h-4" style={{ width: `${largura}%` }} />
            </li>
          ))}
        </ul>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar o ranking de cursos agora."
          onTentarNovamente={onTentarNovamente}
        />
      )}

      {!isLoading && !error && itens.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Nenhum curso com coordenador para ranquear com estes filtros.
        </p>
      )}

      {!isLoading && !error && itens.length > 0 && meta && (
        <div className="space-y-4">
          <p aria-live="polite" className="text-sm text-muted-foreground">
            Mostrando os {itens.length} melhores de {meta.total_cursos_com_coordenador} cursos com
            coordenador neste filtro
          </p>
          <ol className="space-y-3">
            {itens.map((item, indice) => (
              <li
                key={item.curso_id}
                className="flex flex-col gap-1 md:flex-row md:items-center md:gap-3"
              >
                <div className="flex items-baseline justify-between gap-2 md:w-48 md:shrink-0 md:flex-col md:items-start md:gap-0">
                  <span className="truncate font-medium md:w-full" title={item.curso_nome}>
                    {indice + 1}. {item.curso_nome}
                  </span>
                  <span className="shrink-0 text-xs text-muted-foreground md:hidden">
                    {item.responsavel_nome}
                  </span>
                </div>
                <span
                  className="hidden text-xs text-muted-foreground md:block md:w-32 md:shrink-0 md:truncate"
                  title={item.responsavel_nome}
                >
                  {item.responsavel_nome}
                </span>
                <div className="flex flex-1 items-center gap-2">
                  <Progress
                    value={item.percentual}
                    aria-valuetext={`${item.curso_nome}: ${item.percentual}% dos comprovantes exigidos entregues, ${item.aceitas_total} de ${item.exigido_total}`}
                    className="flex-1"
                  />
                  <span
                    aria-hidden="false"
                    className="w-10 shrink-0 text-right text-xs font-medium"
                  >
                    {item.percentual}%
                  </span>
                </div>
                <span className="text-right text-xs text-muted-foreground md:w-20 md:shrink-0">
                  {item.aceitas_total} de {item.exigido_total}
                </span>
              </li>
            ))}
          </ol>
          {cursosSemCoordenador.length > 0 && (
            <p className="flex items-start gap-2 text-xs text-muted-foreground">
              <InfoIcon aria-hidden="true" className="mt-0.5 size-3.5 shrink-0" />
              {cursosSemCoordenador.length} curso(s) sem coordenador neste filtro não entram no
              ranking: {cursosSemCoordenador.join(", ")} — consulte-os na tabela abaixo.
            </p>
          )}
        </div>
      )}
    </section>
  )
}
