"use client"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { PasswordInput } from "@/components/shared/forms/password-input"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Field, FieldContent, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useLogin } from "../hooks/use-login"
import { InstituicaoLoginCombo, VALOR_ADMINISTRACAO_SISTEMA } from "./instituicao-login-combo"

const CHAVE_ULTIMA_INSTITUICAO = "auth:ultima-instituicao"

function lerUltimaInstituicaoSalva(): string | null {
  const salva = window.localStorage.getItem(CHAVE_ULTIMA_INSTITUICAO)
  if (!salva) {
    return null
  }
  try {
    const { id } = JSON.parse(salva) as { id: string }
    return id
  } catch {
    return null
  }
}

export function LoginForm() {
  const router = useRouter()
  const { login, isSubmitting, error } = useLogin()

  const [instituicaoId, setInstituicaoId] = useState<string | null>(null)
  const [email, setEmail] = useState("")
  const [senha, setSenha] = useState("")

  const [erroInstituicao, setErroInstituicao] = useState<string | undefined>()
  const [erroEmail, setErroEmail] = useState<string | undefined>()
  const [erroSenha, setErroSenha] = useState<string | undefined>()

  useEffect(() => {
    const id = lerUltimaInstituicaoSalva()
    if (id) {
      setInstituicaoId(id)
    }
  }, [])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    const semInstituicao = instituicaoId === null ? "Selecione sua instituição." : undefined
    const semEmail = email.trim() === "" ? "O e-mail é obrigatório." : undefined
    const semSenha = senha === "" ? "A senha é obrigatória." : undefined
    setErroInstituicao(semInstituicao)
    setErroEmail(semEmail)
    setErroSenha(semSenha)
    if (semInstituicao || semEmail || semSenha) {
      return
    }

    window.localStorage.setItem(CHAVE_ULTIMA_INSTITUICAO, JSON.stringify({ id: instituicaoId }))

    const idParaApi = instituicaoId === VALOR_ADMINISTRACAO_SISTEMA ? null : instituicaoId
    const contexto = await login({ instituicaoId: idParaApi, email, senha })

    if (!contexto) {
      setSenha("")
      document.getElementById("login-senha")?.focus()
      return
    }

    if (contexto.senha_provisoria) {
      router.replace("/app/alterar-senha")
    } else {
      router.replace("/app")
    }
  }

  return (
    <Card className="w-full max-w-[420px]">
      <CardHeader>
        <CardTitle className="text-xl">Entrar</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} aria-busy={isSubmitting} noValidate>
          <FieldGroup>
            {error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {error.status === 401
                    ? "Instituição, e-mail ou senha inválidos."
                    : "Não foi possível entrar agora. Tente novamente."}
                </AlertDescription>
              </Alert>
            )}

            <Field data-invalid={!!erroInstituicao}>
              <FieldLabel htmlFor="login-instituicao">Instituição</FieldLabel>
              <FieldContent>
                <InstituicaoLoginCombo
                  id="login-instituicao"
                  value={instituicaoId}
                  onValueChange={setInstituicaoId}
                  invalido={!!erroInstituicao}
                  descricaoId="login-instituicao-erro"
                />
                {erroInstituicao && (
                  <FieldError id="login-instituicao-erro">{erroInstituicao}</FieldError>
                )}
              </FieldContent>
            </Field>

            <Field data-invalid={!!erroEmail}>
              <FieldLabel htmlFor="login-email">E-mail</FieldLabel>
              <FieldContent>
                <Input
                  id="login-email"
                  type="email"
                  autoComplete="username"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  disabled={isSubmitting}
                  aria-invalid={!!erroEmail}
                  aria-describedby={erroEmail ? "login-email-erro" : undefined}
                />
                {erroEmail && <FieldError id="login-email-erro">{erroEmail}</FieldError>}
              </FieldContent>
            </Field>

            <Field data-invalid={!!erroSenha}>
              <FieldLabel htmlFor="login-senha">Senha</FieldLabel>
              <FieldContent>
                <PasswordInput
                  id="login-senha"
                  autoComplete="current-password"
                  value={senha}
                  onChange={(e) => setSenha(e.target.value)}
                  disabled={isSubmitting}
                  aria-invalid={!!erroSenha}
                  aria-describedby={erroSenha ? "login-senha-erro" : undefined}
                />
                {erroSenha && <FieldError id="login-senha-erro">{erroSenha}</FieldError>}
              </FieldContent>
            </Field>

            <LoadingButton
              type="submit"
              loading={isSubmitting}
              loadingText="Entrando..."
              className="w-full"
            >
              Entrar
            </LoadingButton>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
