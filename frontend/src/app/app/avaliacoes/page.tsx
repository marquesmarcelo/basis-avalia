"use client"

import { ClipboardListIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { AvaliacaoCardMobile } from "@/features/avaliacao/components/avaliacao-card-mobile"
import { AvaliacaoFiltro } from "@/features/avaliacao/components/avaliacao-filtro"
import { AvaliacaoTable } from "@/features/avaliacao/components/avaliacao-table"
import type { FiltroFila } from "@/features/avaliacao/hooks/use-fila-de-avaliacao"
import { useFilaDeAvaliacao } from "@/features/avaliacao/hooks/use-fila-de-avaliacao"

export default function AvaliacoesPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroFila>(
    "avaliacoes",
    { periodo_id: "", curso_id: "", meta_id: "", situacao: "pendente_avaliacao" },
    "criado_em",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useFilaDeAvaliacao()
  const [jaPesquisou, setJaPesquisou] = useState(false)

  async function executarPesquisa() {
    setJaPesquisou(true)
    const resultado = await pesquisar(estado.filtros, {
      page: estado.page,
      page_size: estado.pageSize,
      sort: estado.sort,
      order: estado.order,
    })
    if (resultado) ajustarPaginaAoTotal(resultado.meta.total_pages)
  }

  useEffect(() => {
    if (carregado && jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  function handlePesquisar(filtros: FiltroFila) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Avaliação de entregas" }]} />

      <h1 className="text-2xl font-semibold">Avaliação de entregas</h1>

      <AvaliacaoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={ClipboardListIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver a fila."
        />
      )}
      {jaPesquisou && isLoading && (
        <SkeletonTable colunas={["Curso", "Meta", "Enviada por", "Data", "Situação", ""]} />
      )}
      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar a fila agora."
          onTentarNovamente={executarPesquisa}
        />
      )}
      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState icon={ClipboardListIcon} titulo="Nenhuma entrega encontrada." />
      )}
      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} entregas encontradas.
          </p>
          <div className="hidden md:block">
            <AvaliacaoTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <AvaliacaoCardMobile key={item.id} item={item} />
            ))}
          </div>
          {data && (
            <Paginacao
              page={data.meta.page}
              pageSize={data.meta.page_size}
              total={data.meta.total}
              totalPages={data.meta.total_pages}
              onPageChange={definirPagina}
              onPageSizeChange={definirTamanhoDePagina}
            />
          )}
        </div>
      )}
    </div>
  )
}
