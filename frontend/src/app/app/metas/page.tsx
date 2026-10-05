"use client"

import { ClipboardListIcon } from "lucide-react"
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
import { ExcluirMetaDialog } from "@/features/meta/components/excluir-meta-dialog"
import { InativarMetaDialog } from "@/features/meta/components/inativar-meta-dialog"
import { MetaCardMobile } from "@/features/meta/components/meta-card-mobile"
import { MetaFiltro } from "@/features/meta/components/meta-filtro"
import { MetaTable } from "@/features/meta/components/meta-table"
import { useAlterarSituacaoMeta } from "@/features/meta/hooks/use-alterar-situacao-meta"
import { useExcluirMeta } from "@/features/meta/hooks/use-excluir-meta"
import { useMetas } from "@/features/meta/hooks/use-metas"
import type { FiltroMetas, Meta } from "@/features/meta/types"

export default function MetasPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroMetas>(
    "metas",
    { busca: "", indicador_id: "", origem: "todas", situacao: "ativo" },
    "nome",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useMetas()
  const { alterarSituacao, idEmAndamento } = useAlterarSituacaoMeta()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirMeta()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraInativar, setParaInativar] = useState<Meta | null>(null)
  const [paraExcluir, setParaExcluir] = useState<Meta | null>(null)

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

  function handlePesquisar(filtros: FiltroMetas) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handleAlterarSituacao(item: Meta) {
    if (item.situacao === "ativo") {
      setParaInativar(item)
      return
    }
    reativar(item)
  }

  async function reativar(item: Meta) {
    const resultado = await alterarSituacao(item.id, "ativo", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.nome} foi reativada.`)
      executarPesquisa()
    }
  }

  async function confirmarInativacao() {
    if (!paraInativar) return
    const resultado = await alterarSituacao(paraInativar.id, "inativo", paraInativar.versao)
    if (resultado) {
      notificar.sucesso(`${paraInativar.nome} foi inativada.`)
      setParaInativar(null)
      executarPesquisa()
    }
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${paraExcluir.nome} foi excluída.`)
      setParaExcluir(null)
      executarPesquisa()
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Metas" },
          { rotulo: "Catálogo de metas" },
        ]}
      />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Metas</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/metas/novo" />}>
          + Nova
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <MetaFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={ClipboardListIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver as metas."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable colunas={["Nome", "Indicadores", "Situação", "Planos", "Ações"]} />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar as metas agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={ClipboardListIcon}
          titulo="Nenhuma meta encontrada."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} metas encontradas.
          </p>
          <div className="hidden md:block">
            <MetaTable
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
              <MetaCardMobile
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
            Uma entrega desta meta atende todos os indicadores dela — a quantidade exigida NÃO é
            multiplicada pelo número de indicadores.
          </p>
          <p className="text-sm text-muted-foreground">
            A quantidade de entregas é do item do plano, não da meta.
          </p>
        </div>
      )}

      <InativarMetaDialog
        meta={paraInativar}
        onOpenChange={(aberto) => !aberto && setParaInativar(null)}
        confirmando={idEmAndamento === paraInativar?.id}
        onConfirmar={confirmarInativacao}
      />

      <ExcluirMetaDialog
        meta={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
