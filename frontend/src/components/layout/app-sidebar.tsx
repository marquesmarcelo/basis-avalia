import { navParaPerfil } from "./nav-config"
import { SidebarGroup } from "./sidebar-group"

interface AppSidebarProps {
  permissoes: string[]
  colapsado?: boolean
  badges?: Record<string, number>
}

export function AppSidebar({ permissoes, colapsado = false, badges }: AppSidebarProps) {
  const grupos = navParaPerfil(permissoes)

  if (grupos.length === 0) return null

  return (
    <nav aria-label="Navegação principal" className="flex flex-col gap-2 p-2">
      {grupos.map((grupo) => (
        <SidebarGroup key={grupo.label} grupo={grupo} colapsado={colapsado} badges={badges} />
      ))}
    </nav>
  )
}
