import { ArrowDownIcon, ArrowUpIcon } from "lucide-react"
import { TableHead } from "@/components/ui/table"

export type Direcao = "asc" | "desc"

interface CabecalhoOrdenavelProps {
  campo: string
  rotulo: string
  sortAtivo: string
  orderAtivo: Direcao
  onOrdenar: (campo: string) => void
  ordenavel?: boolean
  className?: string
}

export function CabecalhoOrdenavel({
  campo,
  rotulo,
  sortAtivo,
  orderAtivo,
  onOrdenar,
  ordenavel = true,
  className,
}: CabecalhoOrdenavelProps) {
  if (!ordenavel) {
    return <TableHead className={className}>{rotulo}</TableHead>
  }

  const ativo = sortAtivo === campo

  return (
    <TableHead
      className={className}
      aria-sort={ativo ? (orderAtivo === "asc" ? "ascending" : "descending") : "none"}
    >
      <button
        type="button"
        onClick={() => onOrdenar(campo)}
        className="flex items-center gap-1 font-medium hover:text-foreground"
      >
        {rotulo}
        {ativo &&
          (orderAtivo === "asc" ? (
            <ArrowUpIcon className="size-3.5" aria-hidden="true" />
          ) : (
            <ArrowDownIcon className="size-3.5" aria-hidden="true" />
          ))}
      </button>
    </TableHead>
  )
}
