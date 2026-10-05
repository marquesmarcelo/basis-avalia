import { UsersIcon } from "lucide-react"
import { describe, expect, it } from "vitest"
import { montarNav, type NavGroup } from "./nav-config"

const configTeste: NavGroup[] = [
  {
    label: "Metas",
    icon: UsersIcon,
    items: [
      {
        label: "Item de dois perfis",
        path: "/app/desempenho",
        permissoes: ["instituicao.listar", "administrador.gerenciar"],
        icon: UsersIcon,
      },
      {
        label: "Catálogo de metas",
        path: "/app/metas",
        permissoes: ["usuario.listar"],
        icon: UsersIcon,
      },
    ],
  },
]

describe("montarNav", () => {
  it("mostra o item quando o usuário tem qualquer uma das permissões declaradas", () => {
    const grupos = montarNav(configTeste, ["administrador.gerenciar"])
    expect(grupos).toHaveLength(1)
    expect(grupos[0].items.map((i) => i.path)).toEqual(["/app/desempenho"])
  })

  it("quem acumula dois perfis que dão o mesmo item vê o item uma vez", () => {
    // Duas permissões distintas do MESMO item presentes ao mesmo tempo —
    // simula um usuário que acumula dois perfis, cada um dando acesso ao
    // item por um motivo diferente. O path só pode aparecer uma vez.
    const grupos = montarNav(configTeste, [
      "instituicao.listar",
      "administrador.gerenciar",
      "usuario.listar",
    ])
    const paths = grupos[0].items.map((i) => i.path)
    expect(paths).toEqual(["/app/desempenho", "/app/metas"])
    expect(paths.filter((p) => p === "/app/desempenho")).toHaveLength(1)
  })

  it("remove o item quando o usuário não tem nenhuma das permissões", () => {
    const grupos = montarNav(configTeste, ["outra.permissao"])
    expect(grupos).toHaveLength(0)
  })

  it("remove o grupo inteiro quando nenhum item sobra", () => {
    const grupos = montarNav(configTeste, [])
    expect(grupos).toEqual([])
  })
})
