"use client"

import { Building2Icon, InboxIcon, UsersIcon } from "lucide-react"
import Link from "next/link"
import { Trilha } from "@/components/layout/trilha"
import { EmptyState } from "@/components/shared/ui/empty-state"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useEu } from "@/features/auth/hooks/use-eu"
import { possuiPermissao } from "@/lib/permissoes"

function primeiroNome(nome: string): string {
  return nome.trim().split(/\s+/)[0] ?? nome
}

export default function InicioPage() {
  const { data: eu } = useEu()
  if (!eu) return null

  const podeUsuarios = possuiPermissao(eu.permissoes, "usuario.listar")
  const podeInstituicoes = possuiPermissao(eu.permissoes, "instituicao.listar")

  return (
    <div className="w-full space-y-6">
      <Trilha itens={[{ rotulo: "Início" }]} />

      <div>
        <h1 className="text-2xl font-semibold">Bem-vindo(a), {primeiroNome(eu.nome)}.</h1>
        <p className="text-muted-foreground">
          {eu.instituicao
            ? `Você está conectado(a) como ${eu.perfis_rotulos.join(", ")} em ${eu.instituicao.nome}.`
            : `Você está conectado(a) como ${eu.perfis_rotulos.join(", ")}.`}
        </p>
      </div>

      {podeUsuarios && (
        <Link href="/app/usuarios" className="block max-w-sm">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <UsersIcon className="size-4" aria-hidden="true" />
                Usuários
              </CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Cadastrar pessoas e definir perfis. Abrir →
            </CardContent>
          </Card>
        </Link>
      )}

      {podeInstituicoes && (
        <Link href="/app/instituicoes" className="block max-w-sm">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Building2Icon className="size-4" aria-hidden="true" />
                Instituições
              </CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              Cadastrar instituições e o primeiro Pesquisador Institucional. Abrir →
            </CardContent>
          </Card>
        </Link>
      )}

      {!podeUsuarios && !podeInstituicoes && (
        <EmptyState
          icon={InboxIcon}
          titulo="Nenhuma funcionalidade disponível para o seu perfil nesta versão."
          descricao="As telas de cursos e metas serão liberadas em breve."
        />
      )}
    </div>
  )
}
