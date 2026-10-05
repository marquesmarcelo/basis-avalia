"use client"

import { CheckSquareIcon } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { useLocalStorage } from "@/components/shared/hooks/use-local-storage"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { useMinhasMetas } from "@/features/entrega/hooks/use-minhas-metas"
import { usePeriodosParaMinhasMetas } from "@/features/entrega/hooks/use-periodos-para-minhas-metas"

export default function MinhasMetasPage() {
  const { itens: periodos, carregar: carregarPeriodos } = usePeriodosParaMinhasMetas()
  const { data, isLoading, error, carregar } = useMinhasMetas()
  const {
    valor: periodoId,
    setValor: setPeriodoId,
    carregado,
  } = useLocalStorage<string>("grid-state:minhas-metas:periodo", "")
  const [inicializado, setInicializado] = useState(false)

  useEffect(() => {
    carregarPeriodos()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!carregado || inicializado || periodos.length === 0) return
    setInicializado(true)
    if (periodoId && periodos.some((p) => p.id === periodoId)) {
      carregar(periodoId)
      return
    }
    const aberto = periodos.find((p) => p.aberto)
    const escolhido = aberto?.id ?? periodos[0]?.id ?? ""
    if (escolhido) {
      setPeriodoId(escolhido)
      carregar(escolhido)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [carregado, inicializado, periodos])

  function trocarPeriodo(novoId: string | null) {
    if (!novoId) return
    setPeriodoId(novoId)
    carregar(novoId)
  }

  const semNenhumCurso = !isLoading && !error && data !== null && data.length === 0

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Minhas metas" }]} />

      <div className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold">Minhas metas</h1>
        <div className="w-64">
          <SelectComRotulo
            value={periodoId}
            onValueChange={trocarPeriodo}
            itens={periodos.map((p) => ({
              value: p.id,
              label: `${p.nome} (até ${new Date(p.data_fim).toLocaleDateString("pt-BR")})`,
            }))}
            placeholder="Selecione o período"
            aria-label="Selecionar período"
          />
        </div>
      </div>

      {isLoading && (
        <div className="space-y-4" aria-busy="true">
          {[1, 2].map((i) => (
            <div key={i} className="animate-pulse space-y-2 rounded-lg border p-4">
              <div className="h-5 w-48 rounded bg-muted" />
              <div className="h-16 rounded bg-muted" />
            </div>
          ))}
        </div>
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar suas metas agora."
          onTentarNovamente={() => periodoId && carregar(periodoId)}
        />
      )}

      {!isLoading && !error && periodos.length === 0 && (
        <EmptyState
          icon={CheckSquareIcon}
          titulo="Você ainda não responde por nenhum curso — quando uma designação vigente existir, suas metas aparecem aqui."
        />
      )}

      {!isLoading && !error && semNenhumCurso && periodos.length > 0 && (
        <EmptyState
          icon={CheckSquareIcon}
          titulo="Nenhum plano de ação vigente para os seus cursos neste período."
        />
      )}

      {!isLoading && !error && data && data.length > 0 && (
        <div className="space-y-6">
          {data.map((grupo) => (
            <section key={grupo.curso_id} aria-labelledby={`curso-${grupo.curso_id}`}>
              <h2 id={`curso-${grupo.curso_id}`} className="mb-3 text-lg font-semibold">
                {grupo.curso_nome}
              </h2>
              <div className="grid gap-3 xl:grid-cols-2">
                {grupo.itens.map((item) => {
                  const percentual = Math.min(
                    100,
                    item.quantidade > 0 ? Math.round((item.aceitas / item.quantidade) * 100) : 0
                  )
                  return (
                    <div key={item.item_plano_id} className="space-y-2 rounded-lg border p-4">
                      <p className="font-medium">{item.meta_nome}</p>
                      {item.indicadores.length > 0 && (
                        <p className="text-xs text-muted-foreground">
                          {item.indicadores
                            .map(
                              (i) =>
                                `${i.codigo} (${i.escopo === "plataforma" ? "Do INEP" : "Da instituição"})`
                            )
                            .join(" · ")}
                        </p>
                      )}
                      <div
                        role="progressbar"
                        aria-valuenow={percentual}
                        aria-valuemin={0}
                        aria-valuemax={100}
                        className="h-2 w-full overflow-hidden rounded-full bg-muted"
                      >
                        <div className="h-full bg-primary" style={{ width: `${percentual}%` }} />
                      </div>
                      <p className="text-sm text-muted-foreground">
                        {item.aceitas} de {item.quantidade} aceitas
                        {item.pendentes > 0 &&
                          ` · ${item.pendentes} pendente${item.pendentes > 1 ? "s" : ""}`}
                        {item.em_correcao > 0 && ` · ${item.em_correcao} em correção`}
                      </p>
                      <div className="flex gap-2 pt-1">
                        <Link
                          href={`/app/itens/${item.item_plano_id}/entregas`}
                          className="text-sm font-medium text-primary hover:underline"
                        >
                          Ver entregas
                        </Link>
                        <Link
                          href={`/app/itens/${item.item_plano_id}/entregas/nova`}
                          className="text-sm font-medium text-primary hover:underline"
                        >
                          + Prestar contas
                        </Link>
                      </div>
                    </div>
                  )
                })}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
