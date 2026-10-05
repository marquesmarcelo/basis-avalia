"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export interface OpcaoSelect {
  value: string
  label: string
}

interface SelectComRotuloProps {
  id?: string
  value: string
  onValueChange: (value: string) => void
  itens: OpcaoSelect[]
  placeholder?: string
  disabled?: boolean
  className?: string
  size?: "sm" | "default"
  "aria-label"?: string
}

export function SelectComRotulo({
  id,
  value,
  onValueChange,
  itens,
  placeholder,
  disabled,
  className,
  size,
  "aria-label": ariaLabel,
}: SelectComRotuloProps) {
  return (
    <Select
      items={itens}
      value={value}
      onValueChange={(v) => onValueChange(v as string)}
      disabled={disabled}
    >
      <SelectTrigger id={id} className={className ?? "w-full"} size={size} aria-label={ariaLabel}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {itens.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
