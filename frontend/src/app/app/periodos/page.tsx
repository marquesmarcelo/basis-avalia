"use client"

import { CalendarRangeIcon } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { ExcluirPeriodoDialog } from "@/features/periodo/components/excluir-periodo-dialog"
import { PeriodoCardMobile } from "@/features/periodo/components/periodo-card-mobile"
import { PeriodoFiltro } from "@/features/periodo/components/periodo-filtro"
import { PeriodoTable } from "@/features/periodo/components/periodo-table"
import { useExcluirPeriodo } from "@/features/periodo/hooks/use-excluir-periodo"
import { usePeriodos } from "@/features/periodo/hooks/use-periodos"
import type { FiltroPeriodos, Periodo } from "@/features/periodo/types"

export default function PeriodosPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroPeriodos>(
    "periodos",
    { nome: "", situacao: "todas" },
    "data_inicio",
    "desc",
    20
  )
  const { data, isLoading, error, pesquisar } = usePeriodos()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirPeriodo()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraExcluir, setParaExcluir] = useState<Periodo | null>(null)

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

  function handlePesquisar(filtros: FiltroPeriodos) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso("Período excluído.")
      setParaExcluir(null)
      executarPesquisa()
    } else {
      notificar.erro("Este período tem planos vinculados e não pode ser excluído.")
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Períodos" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Períodos</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/periodos/novo" />}>
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <PeriodoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={CalendarRangeIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os períodos."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable colunas={["Nome", "Início", "Fim", "Situação", "Planos", "Ações"]} />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os períodos agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState icon={CalendarRangeIcon} titulo="Nenhum período encontrado." />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} períodos encontrados.
          </p>
          <div className="hidden md:block">
            <PeriodoTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              onExcluir={setParaExcluir}
              idExcluindo={idExcluindo}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <PeriodoCardMobile
                key={item.id}
                item={item}
                onExcluir={setParaExcluir}
                idExcluindo={idExcluindo}
              />
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

      <ExcluirPeriodoDialog
        periodo={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
