import { AvisoSessaoEncerrada } from "@/features/auth/components/aviso-sessao-encerrada"
import { LoginForm } from "@/features/auth/components/login-form"

export default function Home() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 px-4 py-10">
      <AvisoSessaoEncerrada />
      <LoginForm />
    </div>
  )
}
