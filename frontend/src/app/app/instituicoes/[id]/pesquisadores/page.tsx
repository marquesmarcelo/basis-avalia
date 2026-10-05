"use client"

import { UsersIcon } from "lucide-react"
import Link from "next/link"
import { use, useCallback, useEffect, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { ErrorState } from "@/components/shared/ui/error-state"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { SkeletonTable } from "@/components/shared/ui/skeleton-table"
import { useEu } from "@/features/auth/hooks/use-eu"
import { useInstituicao } from "@/features/instituicao/hooks/use-instituicao"
import { RedefinirSenhaModal } from "@/features/usuario/components/redefinir-senha-modal"
import { UsuarioCardMobile } from "@/features/usuario/components/usuario-card-mobile"
import { UsuarioTable } from "@/features/usuario/components/usuario-table"
import { useExcluirUsuario } from "@/features/usuario/hooks/use-excluir-usuario"
import { useUsuarios } from "@/features/usuario/hooks/use-usuarios"
import type { Usuario } from "@/features/usuario/types"

export default function PesquisadoresDaInstituicaoPage({
  params,
}: PageProps<"/app/instituicoes/[id]/pesquisadores">) {
  const { id } = use(params)
  const basePath = `/instituicoes/${id}/pesquisadores`

  const { data: eu } = useEu()
  const { data: instituicao } = useInstituicao(id)
  const { data, isLoading, error, pesquisar } = useUsuarios(basePath)
  const { excluir, idEmAndamento: idExcluindo } = useExcluirUsuario(basePath)

  const [sort, setSort] = useState("nome")
  const [order, setOrder] = useState<"asc" | "desc">("asc")
  const [usuarioParaRedefinir, setUsuarioParaRedefinir] = useState<Usuario | null>(null)
  const [usuarioParaExcluir, setUsuarioParaExcluir] = useState<Usuario | null>(null)

  const carregar = useCallback(() => {
    pesquisar({ busca: "", perfil: "" }, { page: 1, page_size: 100, sort, order })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sort, order])

  useEffect(() => {
    carregar()
  }, [carregar])

  function ordenarPor(campo: string) {
    if (sort === campo) {
      setOrder(order === "asc" ? "desc" : "asc")
    } else {
      setSort(campo)
      setOrder("asc")
    }
  }

  async function confirmarExclusao() {
    if (!usuarioParaExcluir) return
    const sucesso = await excluir(usuarioParaExcluir.id)
    if (sucesso) {
      notificar.sucesso(`${usuarioParaExcluir.nome} foi excluído(a).`)
      carregar()
    }
    setUsuarioParaExcluir(null)
  }

  const itens = data?.data ?? []

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Instituições", href: "/app/instituicoes" },
          { rotulo: instituicao?.nome ?? "Pesquisadores Institucionais" },
        ]}
      />

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">
          Pesquisadores Institucionais {instituicao ? `de ${instituicao.nome}` : ""}
        </h1>
        <LoadingButton
          id="botao-novo"
          render={<Link href={`/app/instituicoes/${id}/pesquisadores/novo`} />}
        >
          + Novo
          <IndicadorDeNavegacao />
        </LoadingButton>
      </div>

      {isLoading && (
        <SkeletonTable
          colunas={["Nome", "E-mail", "Perfis", "Situação", "Cadastrado em", "Ações"]}
        />
      )}

      {!isLoading && error && (
        <ErrorState
          titulo="Não foi possível carregar os Pesquisadores Institucionais agora."
          onTentarNovamente={carregar}
        />
      )}

      {!isLoading && !error && itens.length === 0 && (
        <EmptyState
          icon={UsersIcon}
          titulo="Nenhum Pesquisador Institucional cadastrado ainda."
          descricao="Cadastre o primeiro Pesquisador Institucional desta instituição."
        />
      )}

      {!isLoading && !error && itens.length > 0 && eu && (
        <div className="space-y-4">
          <div className="hidden md:block">
            <UsuarioTable
              itens={itens}
              sort={sort}
              order={order}
              onOrdenar={ordenarPor}
              usuarioAtualId={eu.id}
              basePath={basePath}
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
                basePath={basePath}
                onRedefinirSenha={setUsuarioParaRedefinir}
                onExcluir={setUsuarioParaExcluir}
                idExcluindo={idExcluindo}
              />
            ))}
          </div>
        </div>
      )}

      <RedefinirSenhaModal
        usuario={usuarioParaRedefinir}
        basePath={basePath}
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
