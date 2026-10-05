import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface SkeletonTableProps {
  colunas: string[]
  linhas?: number
}

export function SkeletonTable({ colunas, linhas = 5 }: SkeletonTableProps) {
  return (
    <Table aria-busy="true">
      <TableHeader>
        <TableRow>
          {colunas.map((coluna) => (
            <TableHead key={coluna}>{coluna}</TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: linhas }, (_, n) => `linha-esqueleto-${n}`).map((chaveLinha) => (
          <TableRow key={chaveLinha}>
            {colunas.map((coluna) => (
              <TableCell key={coluna}>
                <Skeleton className="h-4 w-full max-w-32" />
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
