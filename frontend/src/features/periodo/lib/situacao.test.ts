import { describe, expect, it } from "vitest"
import { calcularSituacao } from "./situacao"

describe("calcularSituacao", () => {
  it("retorna nao_iniciado quando hoje é antes da data de início", () => {
    expect(calcularSituacao("2026-03-01", "2026-06-30", "2026-02-15")).toBe("nao_iniciado")
  })

  it("retorna aberto quando hoje está entre início e fim, inclusive nas bordas", () => {
    expect(calcularSituacao("2026-03-01", "2026-06-30", "2026-03-01")).toBe("aberto")
    expect(calcularSituacao("2026-03-01", "2026-06-30", "2026-04-15")).toBe("aberto")
    expect(calcularSituacao("2026-03-01", "2026-06-30", "2026-06-30")).toBe("aberto")
  })

  it("retorna encerrado quando hoje é depois da data de fim", () => {
    expect(calcularSituacao("2026-03-01", "2026-06-30", "2026-07-01")).toBe("encerrado")
  })
})
