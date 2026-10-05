"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { Trilha } from "@/components/layout/trilha"
import { useGuardaDeSaida } from "@/components/shared/hooks/use-guarda-de-saida"
import { ConfirmDialog } from "@/components/shared/ui/confirm-dialog"
import { useEu } from "@/features/auth/hooks/use-eu"
import { UsuarioForm } from "@/features/usuario/components/usuario-form"

const BASE_PATH = "/administradores"

export default function NovoAdministradorPage() {
  const router = useRouter()
  const { data: eu } = useEu()
  const [sujo, setSujo] = useState(false)
  const { confirmando, confirmarDescarte, cancelarDescarte, solicitarSaida } =
    useGuardaDeSaida(sujo)

  return (
    <div className="w-full space-y-6">
      <Trilha
        itens={[
          { rotulo: "Início", href: "/app" },
          { rotulo: "Administradores do Sistema", href: "/app/administradores" },
          { rotulo: "Novo" },
        ]}
      />

      <h1 className="text-2xl font-semibold">Novo administrador</h1>

      {eu && (
        <UsuarioForm
          basePath={BASE_PATH}
          usuario={null}
          usuarioAtualId={eu.id}
          perfilFixo="administrador_sistema"
          rotuloPerfilFixo="Administrador do Sistema"
          mensagemSucesso="Administrador cadastrado com sucesso."
          criarOutro
          onDirtyChange={setSujo}
          onCancelar={() => solicitarSaida("/app/administradores")}
          onSalvo={() => router.push("/app/administradores")}
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
