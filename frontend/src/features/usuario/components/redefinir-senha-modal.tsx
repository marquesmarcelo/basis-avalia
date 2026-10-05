"use client"

import { useEffect, useId, useState } from "react"
import { PasswordInput } from "@/components/shared/forms/password-input"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { notificar } from "@/components/shared/ui/notificacoes"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { obrigatorio, senhasCoincidem } from "@/lib/validacao"
import { useRedefinirSenha } from "../hooks/use-redefinir-senha"
import type { Usuario } from "../types"

interface RedefinirSenhaModalProps {
  usuario: Usuario | null
  basePath: string
  onOpenChange: (open: boolean) => void
}

export function RedefinirSenhaModal({ usuario, basePath, onOpenChange }: RedefinirSenhaModalProps) {
  const idBase = useId()
  const [senhaNova, setSenhaNova] = useState("")
  const [confirmacao, setConfirmacao] = useState("")
  const [erroSenha, setErroSenha] = useState<string | undefined>()
  const [erroConfirmacao, setErroConfirmacao] = useState<string | undefined>()

  const { redefinir, isSubmitting, error } = useRedefinirSenha(basePath)

  useEffect(() => {
    if (usuario) {
      setSenhaNova("")
      setConfirmacao("")
      setErroSenha(undefined)
      setErroConfirmacao(undefined)
    }
  }, [usuario])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const semSenha = obrigatorio(senhaNova, "A senha é obrigatória.")
    const naoCoincide = senhasCoincidem(senhaNova, confirmacao, "As senhas não coincidem.")
    setErroSenha(semSenha)
    setErroConfirmacao(naoCoincide)
    if (semSenha || naoCoincide || !usuario) return

    const sucesso = await redefinir(usuario.id, senhaNova)
    if (sucesso) {
      notificar.sucesso(`Senha de ${usuario.nome} redefinida com sucesso.`)
      onOpenChange(false)
    }
  }

  return (
    <Dialog open={!!usuario} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Redefinir senha de {usuario?.nome}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} aria-busy={isSubmitting}>
          <FieldGroup>
            <Alert>
              <AlertDescription>
                A pessoa precisará definir uma senha própria no primeiro acesso após a redefinição.
              </AlertDescription>
            </Alert>
            {error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {error.message || "Não foi possível redefinir a senha agora."}
                </AlertDescription>
              </Alert>
            )}
            <Field data-invalid={!!erroSenha}>
              <FieldLabel htmlFor={`${idBase}-senha`}>Nova senha</FieldLabel>
              <FieldContent>
                <PasswordInput
                  id={`${idBase}-senha`}
                  autoComplete="new-password"
                  value={senhaNova}
                  onChange={(e) => setSenhaNova(e.target.value)}
                  disabled={isSubmitting}
                  aria-invalid={!!erroSenha}
                />
                {erroSenha && <FieldError>{erroSenha}</FieldError>}
              </FieldContent>
            </Field>
            <Field data-invalid={!!erroConfirmacao}>
              <FieldLabel htmlFor={`${idBase}-confirmacao`}>Confirmar senha</FieldLabel>
              <FieldContent>
                <PasswordInput
                  id={`${idBase}-confirmacao`}
                  autoComplete="new-password"
                  value={confirmacao}
                  onChange={(e) => setConfirmacao(e.target.value)}
                  disabled={isSubmitting}
                  aria-invalid={!!erroConfirmacao}
                />
                {erroConfirmacao && <FieldError>{erroConfirmacao}</FieldError>}
              </FieldContent>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancelar
            </Button>
            <LoadingButton type="submit" loading={isSubmitting} loadingText="Redefinindo...">
              Redefinir
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
