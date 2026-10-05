"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { PasswordInput } from "@/components/shared/forms/password-input"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Button } from "@/components/ui/button"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { useAlterarSenha } from "../hooks/use-alterar-senha"
import { definirSessaoAtual } from "../hooks/use-eu"

interface AlterarSenhaFormProps {
  variante: "opcional" | "obrigatoria"
}

export function AlterarSenhaForm({ variante }: AlterarSenhaFormProps) {
  const router = useRouter()
  const { alterarSenha, isSubmitting, error } = useAlterarSenha()

  const [senhaAtual, setSenhaAtual] = useState("")
  const [senhaNova, setSenhaNova] = useState("")
  const [confirmacao, setConfirmacao] = useState("")
  const [erroConfirmacao, setErroConfirmacao] = useState<string | undefined>()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setErroConfirmacao(undefined)

    if (senhaNova !== confirmacao) {
      setErroConfirmacao("As senhas não coincidem.")
      return
    }

    const contexto = await alterarSenha({ senhaAtual, senhaNova })
    if (!contexto) {
      return
    }

    definirSessaoAtual(contexto)
    notificar.sucesso("Senha alterada com sucesso.")
    router.replace("/app")
  }

  function handleCancelar() {
    router.back()
  }

  return (
    <form
      onSubmit={handleSubmit}
      aria-busy={isSubmitting}
      noValidate
      className="w-full max-w-md space-y-4"
    >
      <FieldGroup>
        <Field data-invalid={error?.status === 401}>
          <FieldLabel htmlFor="senha-atual">Senha atual</FieldLabel>
          <FieldContent>
            <PasswordInput
              id="senha-atual"
              autoComplete="current-password"
              value={senhaAtual}
              onChange={(e) => setSenhaAtual(e.target.value)}
              disabled={isSubmitting}
              aria-invalid={error?.status === 401}
              aria-describedby={error?.status === 401 ? "senha-atual-erro" : undefined}
            />
            {error?.status === 401 && (
              <FieldError id="senha-atual-erro">A senha atual está incorreta.</FieldError>
            )}
          </FieldContent>
        </Field>

        <Field data-invalid={error?.campo === "senha_nova"}>
          <FieldLabel htmlFor="senha-nova">Nova senha</FieldLabel>
          <FieldContent>
            <PasswordInput
              id="senha-nova"
              autoComplete="new-password"
              value={senhaNova}
              onChange={(e) => setSenhaNova(e.target.value)}
              disabled={isSubmitting}
              aria-invalid={error?.campo === "senha_nova"}
              aria-describedby={error?.campo === "senha_nova" ? "senha-nova-erro" : undefined}
            />
            {error?.campo === "senha_nova" && (
              <FieldError id="senha-nova-erro">Informe a senha.</FieldError>
            )}
          </FieldContent>
        </Field>

        <Field data-invalid={!!erroConfirmacao}>
          <FieldLabel htmlFor="senha-confirmacao">Confirme a nova senha</FieldLabel>
          <FieldContent>
            <PasswordInput
              id="senha-confirmacao"
              autoComplete="new-password"
              value={confirmacao}
              onChange={(e) => setConfirmacao(e.target.value)}
              disabled={isSubmitting}
              aria-invalid={!!erroConfirmacao}
              aria-describedby={erroConfirmacao ? "senha-confirmacao-erro" : undefined}
            />
            {erroConfirmacao && (
              <FieldError id="senha-confirmacao-erro">{erroConfirmacao}</FieldError>
            )}
          </FieldContent>
        </Field>

        <div className="flex justify-end gap-2">
          {variante === "opcional" && (
            <Button
              type="button"
              variant="outline"
              onClick={handleCancelar}
              disabled={isSubmitting}
            >
              Cancelar
            </Button>
          )}
          <LoadingButton type="submit" loading={isSubmitting} loadingText="Salvando...">
            Salvar
          </LoadingButton>
        </div>
      </FieldGroup>
    </form>
  )
}
