import type { VariantProps } from "class-variance-authority"
import { Badge, type badgeVariants } from "@/components/ui/badge"

interface StatusBadgeProps {
  label: string
  variant?: VariantProps<typeof badgeVariants>["variant"]
  className?: string
}

export function StatusBadge({ label, variant = "default", className }: StatusBadgeProps) {
  return (
    <Badge variant={variant} className={className}>
      {label}
    </Badge>
  )
}
