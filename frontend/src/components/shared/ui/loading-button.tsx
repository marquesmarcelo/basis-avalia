"use client"

import type { VariantProps } from "class-variance-authority"
import { Loader2Icon } from "lucide-react"
import { Button, type buttonVariants } from "@/components/ui/button"

interface LoadingButtonProps
  extends React.ComponentProps<typeof Button>,
    VariantProps<typeof buttonVariants> {
  loading?: boolean
  loadingText?: string
}

export function LoadingButton({
  loading = false,
  loadingText,
  children,
  disabled,
  className,
  ...props
}: LoadingButtonProps) {
  return (
    <Button
      disabled={disabled || loading}
      aria-busy={loading}
      className={loading ? `opacity-75 ${className ?? ""}` : className}
      {...props}
    >
      {loading && (
        <Loader2Icon className="animate-spin motion-reduce:animate-none" aria-hidden="true" />
      )}
      {loading && loadingText ? loadingText : children}
    </Button>
  )
}
