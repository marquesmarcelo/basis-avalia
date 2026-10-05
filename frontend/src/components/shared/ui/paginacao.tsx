import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { Button } from "@/components/ui/button"

interface PaginacaoProps {
  page: number
  pageSize: number
  total: number
  totalPages: number
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
  opcoesTamanho?: number[]
}

export function Paginacao({
  page,
  pageSize,
  total,
  totalPages,
  onPageChange,
  onPageSizeChange,
  opcoesTamanho = [20, 50, 100],
}: PaginacaoProps) {
  return (
    <div className="flex flex-col items-center justify-between gap-3 sm:flex-row">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>{total} resultado(s)</span>
        <SelectComRotulo
          value={String(pageSize)}
          onValueChange={(v) => onPageSizeChange(Number(v))}
          size="sm"
          className="w-fit"
          aria-label="Itens por página"
          itens={opcoesTamanho.map((opcao) => ({
            value: String(opcao),
            label: `${opcao} por página`,
          }))}
        />
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="icon-sm"
          aria-label="Página anterior"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeftIcon aria-hidden="true" />
        </Button>
        <span className="text-sm text-muted-foreground">
          Página {page} de {Math.max(totalPages, 1)}
        </span>
        <Button
          variant="outline"
          size="icon-sm"
          aria-label="Próxima página"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          <ChevronRightIcon aria-hidden="true" />
        </Button>
      </div>
    </div>
  )
}
