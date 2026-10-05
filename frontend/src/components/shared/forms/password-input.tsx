"use client"

import { EyeIcon, EyeOffIcon } from "lucide-react"
import { useId, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

interface PasswordInputProps extends Omit<React.ComponentProps<typeof Input>, "type"> {
  autoComplete?: "current-password" | "new-password"
}

export function PasswordInput({
  autoComplete = "current-password",
  className,
  ...props
}: PasswordInputProps) {
  const [visivel, setVisivel] = useState(false)
  const botaoId = useId()

  return (
    <div className="relative">
      <Input
        type={visivel ? "text" : "password"}
        autoComplete={autoComplete}
        className={`pr-9 ${className ?? ""}`}
        {...props}
      />
      <Button
        id={botaoId}
        type="button"
        variant="ghost"
        size="icon-sm"
        tabIndex={0}
        className="absolute top-1/2 right-1 -translate-y-1/2"
        aria-label={visivel ? "Ocultar senha" : "Mostrar senha"}
        aria-pressed={visivel}
        onClick={() => setVisivel((v) => !v)}
      >
        {visivel ? <EyeOffIcon aria-hidden="true" /> : <EyeIcon aria-hidden="true" />}
      </Button>
    </div>
  )
}
