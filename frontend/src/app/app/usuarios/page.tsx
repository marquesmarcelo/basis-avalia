"use client"

import { UsersIcon } from "lucide-react"
import Link from "next/link"
import { useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useEstadoDeGrid } from "@/components/shared/hooks/use-estado-de-grid"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Paginacao } from "@/components/shared/ui/paginacao"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { useEu } from "@/features/auth/hooks/use-eu"
import { RedefinirSenhaModal } from "@/features/usuario/components/redefinir-senha-modal"
import { UsuarioCardMobile } from "@/features/usuario/components/usuario-card-mobile"
import { UsuarioFiltro } from "@/features/usuario/components/usuario-filtro"
import { UsuarioTable } from "@/features/usuario/components/usuario-table"
import { useExcluirUsuario } from "@/features/usuario/hooks/use-excluir-usuario"
import { useUsuarios } from "@/features/usuario/hooks/use-usuarios"
import type { FiltroUsuarios, Usuario } from "@/features/usuario/types"

const BASE_PATH = "/usuarios"

export default function UsuariosPage() {
  const {
    estado,
    carregado,
    definirFiltros,
    ordenarPor,
    definirPagina,
    definirTamanhoDePagina,
    ajustarPaginaAoTotal,
  } = useEstadoDeGrid<FiltroUsuarios>("usuarios", { busca: "", perfil: "" }, "nome", "asc", 20)
  const { data, isLoading, error, pesquisar } = useUsuarios(BASE_PATH)
  const { excluir, idEmAndamento: idExcluindo } = useExcluirUsuario(BASE_PATH)
  const { data: eu } = useEu()

  const [jaPesquisou, setJaPesquisou] = useState(false)
  const [usuarioParaRedefinir, setUsuarioParaRedefinir] = useState<Usuario | null>(null)
  const [usuarioParaExcluir, setUsuarioParaExcluir] = useState<Usuario | null>(null)

  async function executarPesquisa() {
    setJaPesquisou(true)
    const resultado = await pesquisar(estado.filtros, {
      page: estado.page,
      page_size: estado.pageSize,
      sort: estado.sort,
      order: estado.order,
    })
    if (resultado) {
      ajustarPaginaAoTotal(resultado.meta.total_pages)
    }
  }

  useEffect(() => {
    if (carregado && jaPesquisou) {
      executarPesquisa()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.page, estado.pageSize, estado.sort, estado.order])

  useEffect(() => {
    if (jaPesquisou) {
      executarPesquisa()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [estado.filtros])

  function handlePesquisar(filtros: FiltroUsuarios) {
    definirFiltros(filtros)
    setJaPesquisou(true)
  }

  async function confirmarExclusao() {
    if (!usuarioParaExcluir) return
    const sucesso = await excluir(usuarioParaExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${usuarioParaExcluir.nome} foi excluído(a).`)
      setUsuarioParaExcluir(null)
      executarPesquisa()
    } else {
      setUsuarioParaExcluir(null)
    }
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Usuários" }]} />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Usuários</h1>
        <LoadingButton id="botao-novo" render={<Link href="/app/usuarios/novo" />}>
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      <UsuarioFiltro
        filtrosIniciais={estado.filtros}
        onPesquisar={handlePesquisar}
        pesquisando={isLoading}
      />

      {!jaPesquisou && (
        <EmptyState
          icon={UsersIcon}
          titulo="Use os filtros acima e clique em Pesquisar para ver os usuários."
        />
      )}

      {jaPesquisou && isLoading && (
        <SkeletonTable
          colunas={["Nome", "E-mail", "Perfis", "Situação", "Cadastrado em", "Ações"]}
        />
      )}

      {jaPesquisou && !isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os usuários agora."
          onTentarNovamente={executarPesquisa}
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={UsersIcon}
          titulo="Nenhum usuário encontrado."
          descricao="Revise os filtros e tente novamente."
        />
      )}

      {jaPesquisou && !isLoading && !error && itens.length > 0 && eu && (
        <div className="space-y-4">
          <p className="sr-only" aria-live="polite">
            {data?.meta.total} usuários encontrados.
          </p>
          <div className="hidden md:block">
            <UsuarioTable
              itens={itens}
              sort={estado.sort}
              order={estado.order}
              onOrdenar={ordenarPor}
              usuarioAtualId={eu.id}
              basePath={BASE_PATH}
              onRedefinirSenha={setUsuarioParaRedefinir}
              onExcluir={setUsuarioParaExcluir}
              idExcluindo={idExcluindo}
            />
          </div>
          <div className="grid gap-3 md:hidden">
            {itens.map((item) => (
              <UsuarioCardMobile
                key={item.id}
                item={item}
                usuarioAtualId={eu.id}
                basePath={BASE_PATH}
                onRedefinirSenha={setUsuarioParaRedefinir}
                onExcluir={setUsuarioParaExcluir}
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

      <RedefinirSenhaModal
        usuario={usuarioParaRedefinir}
        basePath={BASE_PATH}
        onOpenChange={(aberto) => !aberto && setUsuarioParaRedefinir(null)}
      />

      <ConfirmDialog
        open={!!usuarioParaExcluir}
        onOpenChange={(aberto) => !aberto && setUsuarioParaExcluir(null)}
        titulo={`Excluir ${usuarioParaExcluir?.nome}?`}
        descricao="Esta ação não pode ser desfeita. A pessoa perderá o acesso ao sistema."
        rotuloConfirmar="Excluir"
        rotuloConfirmando="Excluindo..."
        destrutivo
        confirmando={idExcluindo === usuarioParaExcluir?.id}
        onConfirmar={confirmarExclusao}
      />
    </div>
  )
}
