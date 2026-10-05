"use client"

import { Trilha } from "@/components/layout/trilha"
import { AlterarSenhaForm } from "@/features/auth/components/alterar-senha-form"
import { useEu } from "@/features/auth/hooks/use-eu"

export default function AlterarSenhaPage() {
  const { data: eu } = useEu()
  if (!eu) return null

  const obrigatoria = eu.senha_provisoria

  return (
    <div className="w-full space-y-6">
      {!obrigatoria && (
        <Trilha itens={[{ rotulo: "Início", href: "/app" }, { rotulo: "Alterar minha senha" }]} />
      )}

      <div>
        <h1 className="text-2xl font-semibold">
          {obrigatoria ? "Defina sua senha" : "Alterar minha senha"}
        </h1>
        {obrigatoria && (
          <p className="text-muted-foreground">
            É necessário definir uma senha própria para continuar.
          </p>
        )}
      </div>

      <AlterarSenhaForm variante={obrigatoria ? "obrigatoria" : "opcional"} />
    </div>
  )
}
