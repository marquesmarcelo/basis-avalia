export function focarPrimeiroCampoComErro(form: HTMLFormElement | null) {
  if (!form) return
  requestAnimationFrame(() => {
    const primeiro = form.querySelector<HTMLElement>('[aria-invalid="true"]')
    primeiro?.focus()
  })
}
