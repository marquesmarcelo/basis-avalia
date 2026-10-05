"use client"

import { TargetIcon } from "lucide-react"
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
import { ExcluirIndicadorDialog } from "@/features/indicador/components/excluir-indicador-dialog"
import { InativarIndicadorDialog } from "@/features/indicador/components/inativar-indicador-dialog"
import { IndicadorCardMobile } from "@/features/indicador/components/indicador-card-mobile"
import { IndicadorFiltro } from "@/features/indicador/components/indicador-filtro"
import { IndicadorTable } from "@/features/indicador/components/indicador-table"
import { useAlterarSituacaoIndicador } from "@/features/indicador/hooks/use-alterar-situacao-indicador"
import { useExcluirIndicador } from "@/features/indicador/hooks/use-excluir-indicador"
import { useIndicadores } from "@/features/indicador/hooks/use-indicadores"
import type { FiltroIndicadores, Indicador } from "@/features/indicador/types"

export default function IndicadoresPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroIndicadores>(
    "indicadores",
    { busca: "", origem: "todos", situacao: "ativo" },
    "codigo",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useIndicadores()
  const { alterarSituacao, idEmAndamento } = useAlterarSituacaoIndicador()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirIndicador()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraInativar, setParaInativar] = useState<Indicador | null>(null)
  const [paraExcluir, setParaExcluir] = useState<Indicador | null>(null)

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

  function handlePesquisar(filtros: FiltroIndicadores) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handleAlterarSituacao(item: Indicador) {
    if (item.situacao === "ativo") {
      setParaInativar(item)
      return
    }
    reativar(item)
  }

  async function reativar(item: Indicador) {
    const resultado = await alterarSituacao(item.id, "ativo", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.codigo} foi reativado.`)
      executarPesquisa()
    }
  }

  async function confirmarInativacao() {
    if (!paraInativar) return
    const resultado = await alterarSituacao(paraInativar.id, "inativo", paraInativar.versao)
    if (resultado) {
      notificar.sucesso(`${paraInativar.codigo} foi inativado.`)
      setParaInativar(null)
      executarPesquisa()
    }
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${paraExcluir.codigo} foi excluído.`)
      setParaExcluir(null)
      executarPesquisa()
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Metas" }, { rotulo: "Indicadores" }]}
      />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Indicadores</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/indicadores/novo" />}>
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <IndicadorFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={TargetIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os indicadores."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={["Código", "Nome", "Origem", "Referência", "Situação", "Metas", "Ações"]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os indicadores agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={TargetIcon}
          titulo="Nenhum indicador encontrado."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} indicadores encontrados.
          </p>
          <div className="hidden md:block">
            <IndicadorTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              onAlterarSituacao={handleAlterarSituacao}
              onExcluir={setParaExcluir}
              idAlterandoSituacao={idEmAndamento ?? idExcluindo}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <IndicadorCardMobile
                key={item.id}
                item={item}
                onAlterarSituacao={handleAlterarSituacao}
                onExcluir={setParaExcluir}
                idAlterandoSituacao={idEmAndamento ?? idExcluindo}
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
          <p className="text-sm text-muted-foreground">
            🔒 Indicadores do INEP são mantidos pelo Administrador do Sistema e são somente leitura
            aqui. Você pode usá-los em metas normalmente.
          </p>
          <p className="text-sm text-muted-foreground">
            &quot;Metas&quot; conta apenas as da sua instituição.
          </p>
        </div>
      )}

      <InativarIndicadorDialog
        indicador={paraInativar}
        onOpenChange={(aberto) => !aberto && setParaInativar(null)}
        confirmando={idEmAndamento === paraInativar?.id}
        onConfirmar={confirmarInativacao}
      />

      <ExcluirIndicadorDialog
        indicador={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
