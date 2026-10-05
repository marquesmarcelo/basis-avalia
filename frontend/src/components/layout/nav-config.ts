import {
  BarChart3Icon,
  Building2Icon,
  CalendarRangeIcon,
  CheckSquareIcon,
  ClipboardListIcon,
  FileTextIcon,
  GraduationCapIcon,
  type LucideIcon,
  ShieldIcon,
  TargetIcon,
  UserCogIcon,
  UsersIcon,
} from "lucide-react"
import { type Permissao, possuiAlgumaPermissao } from "@/lib/permissoes"

export interface NavItem {
  label: string
  path: string
  permissoes: Permissao[]
  icon: LucideIcon
  badge?: "pendentes_de_avaliacao" | "pendencias_nao_vistas"
}

export interface NavGroup {
  label: string
  icon: LucideIcon
  items: NavItem[]
}

export const NAV_CONFIG: NavGroup[] = [
  {
    label: "Metas",
    icon: TargetIcon,
    items: [
      {
        label: "Indicadores",
        path: "/app/indicadores",
        permissoes: ["indicador.listar"],
        icon: TargetIcon,
      },
      {
        label: "Catálogo de metas",
        path: "/app/metas",
        permissoes: ["meta.listar"],
        icon: ClipboardListIcon,
      },
      {
        label: "Períodos",
        path: "/app/periodos",
        permissoes: ["periodo.gerenciar"],
        icon: CalendarRangeIcon,
      },
      {
        label: "Planos",
        path: "/app/planos",
        permissoes: ["plano.listar", "plano.ler_proprio"],
        icon: FileTextIcon,
      },
      {
        label: "Minhas metas",
        path: "/app/minhas-metas",
        permissoes: ["entrega.registrar"],
        icon: CheckSquareIcon,
        badge: "pendencias_nao_vistas",
      },
      {
        label: "Avaliação de entregas",
        path: "/app/avaliacoes",
        permissoes: ["entrega.avaliar"],
        icon: ClipboardListIcon,
        badge: "pendentes_de_avaliacao",
      },
      {
        label: "Desempenho dos cursos",
        path: "/app/desempenho",
        permissoes: ["relatorio.ler", "relatorio.ler_proprio"],
        icon: BarChart3Icon,
      },
    ],
  },
  {
    label: "Administração",
    icon: UsersIcon,
    items: [
      { label: "Usuários", path: "/app/usuarios", permissoes: ["usuario.listar"], icon: UsersIcon },
      {
        label: "Cursos",
        path: "/app/cursos",
        permissoes: ["curso.listar"],
        icon: GraduationCapIcon,
      },
    ],
  },
  {
    label: "Sistema",
    icon: ShieldIcon,
    items: [
      {
        label: "Instituições",
        path: "/app/instituicoes",
        permissoes: ["instituicao.listar"],
        icon: Building2Icon,
      },
      {
        label: "Administradores",
        path: "/app/administradores",
        permissoes: ["administrador.gerenciar"],
        icon: UserCogIcon,
      },
      {
        label: "Indicadores do INEP",
        path: "/app/indicadores-inep",
        permissoes: ["indicador.plataforma.gerenciar"],
        icon: TargetIcon,
      },
    ],
  },
]

export function montarNav(config: NavGroup[], permissoes: string[]): NavGroup[] {
  return config
    .map((grupo) => {
      const vistos = new Set<string>()
      const items = grupo.items.filter((item) => {
        if (vistos.has(item.path)) return false
        if (!possuiAlgumaPermissao(permissoes, item.permissoes)) return false
        vistos.add(item.path)
        return true
      })
      return { ...grupo, items }
    })
    .filter((grupo) => grupo.items.length > 0)
}

export function navParaPerfil(permissoes: string[]): NavGroup[] {
  return montarNav(NAV_CONFIG, permissoes)
}

export const ROTAS_PROTEGIDAS: { path: string; permissoes: Permissao[] }[] = NAV_CONFIG.flatMap(
  (grupo) => grupo.items.map((item) => ({ path: item.path, permissoes: item.permissoes }))
)
