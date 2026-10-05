"use client"

import { GraduationCapIcon } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
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
import { CursoCardMobile } from "@/features/curso/components/curso-card-mobile"
import { CursoFiltro } from "@/features/curso/components/curso-filtro"
import { CursoTable } from "@/features/curso/components/curso-table"
import { ExcluirCursoDialog } from "@/features/curso/components/excluir-curso-dialog"
import { InativarCursoDialog } from "@/features/curso/components/inativar-curso-dialog"
import { useAlterarSituacaoCurso } from "@/features/curso/hooks/use-alterar-situacao-curso"
import { useCursos } from "@/features/curso/hooks/use-cursos"
import { useExcluirCurso } from "@/features/curso/hooks/use-excluir-curso"
import type { Curso, FiltroCursos } from "@/features/curso/types"

export default function CursosPage() {
  const router = useRouter()
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroCursos>(
    "cursos",
    { busca: "", grau: "todos", modalidade: "todas", coordenador_id: "", situacao: "ativo" },
    "nome",
    "asc",
    20
  )
  const { data, isLoading, error, pesquisar } = useCursos()
  const { alterarSituacao, idEmAndamento } = useAlterarSituacaoCurso()
  const { excluir, idEmAndamento: idExcluindo } = useExcluirCurso()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [paraInativar, setParaInativar] = useState<Curso | null>(null)
  const [paraExcluir, setParaExcluir] = useState<Curso | null>(null)

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

  function handlePesquisar(filtros: FiltroCursos) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  useEffect(() => {
    if (jaPesquisou) executarPesquisa()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function abrirDesignacoes(item: Curso) {
    router.push(`/app/cursos/${item.id}/designacoes`)
  }

  function handleAlterarSituacao(item: Curso) {
    if (item.situacao === "ativo") {
      setParaInativar(item)
      return
    }
    reativar(item)
  }

  async function reativar(item: Curso) {
    const resultado = await alterarSituacao(item.id, "ativo", item.versao)
    if (resultado) {
      notificar.sucesso(`${item.nome} foi reativado.`)
      executarPesquisa()
    }
  }

  async function confirmarInativacao() {
    if (!paraInativar) return
    const resultado = await alterarSituacao(paraInativar.id, "inativo", paraInativar.versao)
    if (resultado) {
      notificar.sucesso(`${paraInativar.nome} foi inativado.`)
      setParaInativar(null)
      executarPesquisa()
    }
  }

  async function confirmarExclusao() {
    if (!paraExcluir) return
    const sucesso = await excluir(paraExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${paraExcluir.nome} foi excluído.`)
      setParaExcluir(null)
      executarPesquisa()
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Cursos" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Cursos</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/cursos/novo" />}>
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <CursoFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={GraduationCapIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os cursos."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={[
            "Nome",
            "Código e-MEC",
            "Grau",
            "Modalidade",
            "Coordenador",
            "Situação",
            "Ações",
          ]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os cursos agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={GraduationCapIcon}
          titulo="Nenhum curso encontrado."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && (
        <div className="space-y-4" aria-busy={isLoading}>
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} cursos encontrados.
          </p>
          <div className="hidden md:block">
            <CursoTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              onVerDesignacoes={abrirDesignacoes}
              onAlterarSituacao={handleAlterarSituacao}
              onExcluir={setParaExcluir}
              idAlterandoSituacao={idEmAndamento ?? idExcluindo}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <CursoCardMobile
                key={item.id}
                item={item}
                onVerDesignacoes={abrirDesignacoes}
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
            ⚠ Curso sem designação vigente não recebe entregas, e os planos vigentes dele continuam
            sendo cobrados.
          </p>
        </div>
      )}

      <InativarCursoDialog
        curso={paraInativar}
        onOpenChange={(aberto) => !aberto && setParaInativar(null)}
        confirmando={idEmAndamento === paraInativar?.id}
        onConfirmar={confirmarInativacao}
      />

      <ExcluirCursoDialog
        curso={paraExcluir}
        onOpenChange={(aberto) => !aberto && setParaExcluir(null)}
        confirmando={idExcluindo === paraExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
