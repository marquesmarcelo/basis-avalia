import type { ContextoDeSessao } from "@/features/auth/types"

interface AppFooterProps {
  eu: ContextoDeSessao
}

export function AppFooter({ eu }: AppFooterProps) {
  const versao = process.env.NEXT_PUBLIC_APP_VERSION ?? "dev"
  const ano = new Date().getFullYear()
  const nomeInstituicao = eu.instituicao ? eu.instituicao.nome : null

  return (
    <footer className="flex h-10 shrink-0 items-center justify-center gap-2 border-t bg-background px-4 text-xs text-muted-foreground">
      <span>basis-avalia</span>
      <span>·</span>
      <span>v{versao}</span>
      <span>·</span>
      <span>{ano}</span>
      {nomeInstituicao && (
        <>
          <span>·</span>
          <span>{nomeInstituicao}</span>
        </>
      )}
    </footer>
  )
}
