import { describe, expect, it } from "vitest"
import { emailValido, obrigatorio, senhasCoincidem } from "./validacao"

describe("obrigatorio", () => {
  it("retorna a mensagem quando o valor está vazio", () => {
    expect(obrigatorio("", "Campo obrigatório.")).toBe("Campo obrigatório.")
    expect(obrigatorio("   ", "Campo obrigatório.")).toBe("Campo obrigatório.")
    expect(obrigatorio(null, "Campo obrigatório.")).toBe("Campo obrigatório.")
  })

  it("não retorna mensagem quando o valor está preenchido", () => {
    expect(obrigatorio("Maria Souza", "Campo obrigatório.")).toBeUndefined()
  })
})

describe("emailValido", () => {
  it("aceita um e-mail no formato correto", () => {
    expect(emailValido("maria.souza@fsa.edu.br", "E-mail inválido.")).toBeUndefined()
  })

  it("rejeita um e-mail sem @", () => {
    expect(emailValido("maria.souza-fsa.edu.br", "E-mail inválido.")).toBe("E-mail inválido.")
  })
})

describe("senhasCoincidem", () => {
  it("rejeita quando as senhas são diferentes", () => {
    expect(senhasCoincidem("abc123", "abc124", "As senhas não coincidem.")).toBe(
      "As senhas não coincidem."
    )
  })

  it("não retorna mensagem quando as senhas coincidem", () => {
    expect(senhasCoincidem("abc123", "abc123", "As senhas não coincidem.")).toBeUndefined()
  })
})
