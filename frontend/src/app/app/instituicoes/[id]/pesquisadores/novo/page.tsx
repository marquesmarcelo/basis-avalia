"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { use, useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { useEu } from "@/features/auth/hooks/use-eu"
import { useInstituicao } from "@/features/instituicao/hooks/use-instituicao"
import { UsuarioForm } from "@/features/usuario/components/usuario-form"

export default function NovoPesquisadorPage({
  params,
}: PageProps<"/app/instituicoes/[id]/pesquisadores/novo">) {
  const { id } = use(params)
  const basePath = `/instituicoes/${id}/pesquisadores`
  const router = useRouter()
  const searchParams = useSearchParams()
  const primeiro = searchParams.get("primeiro") === "true"

  const { data: eu } = useEu()
  const { data: instituicao } = useInstituicao(id)
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  const destinoPadrao = primeiro ? "/app/instituicoes" : `/app/instituicoes/${id}/pesquisadores`

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Instituições", href: "/app/instituicoes" },
          {
            rotulo: instituicao?.nome ?? "Pesquisadores",
            href: `/app/instituicoes/${id}/pesquisadores`,
          },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">
        {primeiro
          ? `Cadastrar o primeiro Pesquisador Institucional${instituicao ? ` de ${instituicao.nome}` : ""}`
          : "Novo Pesquisador Institucional"}
      </h1>

      {eu && (
        <UsuarioForm
          basePath={basePath}
          usuario={null}
          usuarioAtualId={eu.id}
          perfilFixo="pesquisador_institucional"
          rotuloSalvar={primeiro ? "Cadastrar" : "Salvar"}
          rotuloSalvando={primeiro ? "Cadastrando..." : "Salvando..."}
          mensagemSucesso="Pesquisador Institucional cadastrado com sucesso."
          criarOutro={!primeiro}
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida(destinoPadrao)}
          onSalvo={() => router.push(destinoPadrao)}
        />
      )}

      <ConfirmDialog
        open={confirmando}
        onOpenChange={(aberto) => !aberto && cancelarDescarte()}
        titulo="Descartar alterações?"
        descricao="Você tem alterações não salvas. Deseja descartá-las?"
        rotuloConfirmar="Descartar alterações"
        rotuloCancelar="Continuar editando"
        destrutivo
        onConfirmar={confirmarDescarte}
      />
    </div>
  )
}
