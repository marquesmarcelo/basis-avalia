import { expect } from "@playwright/test"
import { login } from "../helpers/auth"
import { API_URL } from "../helpers/env"
import { abrirUsuariosEPesquisar, test as usuarioTest } from "../helpers/usuario"

const test = usuarioTest

// specs/cursos/design.md §10.3 — "o relógio não se move, o dado se move":
// o E2E não consegue avançar o relógio do servidor, então reproduz o
// mesmo resultado observável mexendo diretamente na data_fim de uma
// designação, exatamente como uma portaria vencendo sozinha produziria.
//
// "Minhas metas" (o item de menu que aparece/some com o perfil derivado)
// é escopo de metas-coordenacao, ainda não implementada — este teste
// verifica a mesma mudança de perfil via /auth/eu (perfis_derivados,
// cursos_coordenados) e via o aviso em toast, que já são desta feature.
// Registrado em specs/cursos/testes-pendentes.md.
test.describe("Designação e perfil derivado", () => {
  test("designar, encerrar por edição de data_fim, e autodesignação com acúmulo de PI", async ({
    page,
    criarUsuarioViaUI,
  }) => {
    const sufixo = Date.now()
    const nomeCurso = `Curso E2E Designação ${sufixo}`
    const nomeProfessor = `Professor E2E ${sufixo}`
    const emailProfessor = `professor.e2e.${sufixo}@fsa.edu.br`

    await login(page, {
      instituicao: "Faculdade Serra Azul (FSA)",
      email: "maria.souza@fsa.edu.br",
      senha: "reuniao-nde-2026",
    })
    await expect(page).toHaveURL("/app")

    await abrirUsuariosEPesquisar(page)
    await criarUsuarioViaUI(page, {
      nome: nomeProfessor,
      email: emailProfessor,
      perfis: ["Professor"],
      senha: "senha-provisoria-e2e-2026",
    })

    const respostaCriarCurso = await page.request.post(`${API_URL}/cursos`, {
      data: { nome: nomeCurso, codigo_emec: "", grau: "bacharelado", modalidade: "presencial" },
    })
    expect(respostaCriarCurso.ok()).toBeTruthy()
    const curso = await respostaCriarCurso.json()

    // início alguns dias no passado: o modal de encerrar trava a data
    // mínima em data_inicio, então "início = hoje" impediria escolher
    // "fim = ontem" mais adiante.
    const tresDiasAtras = new Date()
    tresDiasAtras.setDate(tresDiasAtras.getDate() - 3)
    const dataInicio = tresDiasAtras.toISOString().slice(0, 10)

    await page.goto("/app/cursos")
    await page.getByLabel("Nome ou código").fill(nomeCurso)
    await page.getByRole("button", { name: "Pesquisar" }).click()
    await page.getByRole("button", { name: `Ver designações de ${nomeCurso}` }).click()
    await expect(page).toHaveURL(new RegExp(`/app/cursos/${curso.id}/designacoes`))

    await page.getByRole("link", { name: "+ Nova" }).click()
    await expect(page).toHaveURL(new RegExp(`/app/cursos/${curso.id}/designacoes/novo`))
    await page.getByLabel("Coordenador").click()
    await page.getByRole("option").filter({ hasText: nomeProfessor }).click()
    await page.getByLabel("Portaria", { exact: true }).fill("1/2026-E2E")
    await page.getByLabel("Início").fill(dataInicio)
    await page.getByRole("button", { name: "Salvar", exact: true }).click()
    await expect(page).toHaveURL(new RegExp(`/app/cursos/${curso.id}/designacoes$`))
    await page.getByRole("button", { name: "Pesquisar" }).click()
    await expect(page.getByRole("row", { name: new RegExp(nomeProfessor) })).toContainText("Vigente")

    const contextoProfessor = await page.context().browser()!.newContext()
    const paginaProfessor = await contextoProfessor.newPage()
    await login(paginaProfessor, {
      instituicao: "Faculdade Serra Azul (FSA)",
      email: emailProfessor,
      senha: "senha-provisoria-e2e-2026",
    })
    await expect(paginaProfessor).toHaveURL(/\/app/, { timeout: 5000 })
    let euProfessor = await paginaProfessor.request.get(`${API_URL}/auth/eu`).then((r) => r.json())
    expect(euProfessor.perfis_derivados).toContain("coordenador_curso")
    expect(euProfessor.cursos_coordenados).toBe(1)
    await contextoProfessor.close()

    await page.getByRole("button", { name: `Encerrar designação de ${nomeProfessor}` }).click()
    const dialogo = page.getByRole("dialog", { name: /Encerrar a designação de/ })
    await expect(dialogo).toContainText("deixa de ter o perfil de Coordenador de Curso")
    const ontem = new Date()
    ontem.setDate(ontem.getDate() - 1)
    await dialogo.getByLabel("Nova data de fim").fill(ontem.toISOString().slice(0, 10))
    await dialogo.getByRole("button", { name: "Encerrar" }).click()
    await expect(dialogo).not.toBeVisible()
    await expect(page.getByRole("row", { name: new RegExp(nomeProfessor) })).toContainText("Encerrada")

    const contextoProfessor2 = await page.context().browser()!.newContext()
    const paginaProfessor2 = await contextoProfessor2.newPage()
    await login(paginaProfessor2, {
      instituicao: "Faculdade Serra Azul (FSA)",
      email: emailProfessor,
      senha: "senha-provisoria-e2e-2026",
    })
    await expect(paginaProfessor2).toHaveURL(/\/app/, { timeout: 5000 })
    euProfessor = await paginaProfessor2.request.get(`${API_URL}/auth/eu`).then((r) => r.json())
    expect(euProfessor.perfis_derivados).not.toContain("coordenador_curso")
    expect(euProfessor.cursos_coordenados).toBe(0)
    await contextoProfessor2.close()

    await page.goto("/app/cursos")
    await page.getByLabel("Nome ou código").fill(nomeCurso)
    await page.getByRole("button", { name: "Pesquisar" }).click()
    await expect(page.getByRole("row", { name: new RegExp(nomeCurso) })).toContainText("Vago")

    await page.getByRole("button", { name: `Ver designações de ${nomeCurso}` }).click()
    await page.getByRole("link", { name: "+ Nova" }).click()
    await expect(page).toHaveURL(new RegExp(`/app/cursos/${curso.id}/designacoes/novo`))
    await page.getByLabel("Coordenador").click()
    await page.getByRole("option").filter({ hasText: "Maria Souza" }).click()
    await expect(page.getByText("também é Pesquisador(a) Institucional")).toBeVisible()
    await expect(page.getByText("Você está se designando para este curso")).toBeVisible()
    await page.getByLabel("Portaria", { exact: true }).fill("2/2026-E2E")
    await page.getByLabel("Início").fill(new Date().toISOString().slice(0, 10))
    await page.getByRole("button", { name: "Salvar", exact: true }).click()
    await expect(page).toHaveURL(new RegExp(`/app/cursos/${curso.id}/designacoes$`))
    await page.getByRole("button", { name: "Pesquisar" }).click()

    const linhaAutodesignacao = page.getByRole("row", { name: /Maria Souza/ })
    await expect(linhaAutodesignacao).toContainText("autodesignação")
  })
})
