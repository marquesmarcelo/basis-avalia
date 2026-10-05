"use client"

import { useState } from "react"
import { SelectComRotulo } from "@/components/shared/forms/select-com-rotulo"
import { LoadingButton } from "@/components/shared/ui/loading-button"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { FiltroUsuarios } from "../types"
import { PERFIS_ESCOLHIVEIS } from "../types"

interface UsuarioFiltroProps {
  filtrosIniciais: FiltroUsuarios
  onPesquisar: (filtros: FiltroUsuarios) => void
  pesquisando: boolean
  mostrarFiltroPerfil?: boolean
  idCampoBusca?: string
}

export function UsuarioFiltro({
  filtrosIniciais,
  onPesquisar,
  pesquisando,
  mostrarFiltroPerfil = true,
  idCampoBusca = "usuarios-filtro-busca",
}: UsuarioFiltroProps) {
  const [busca, setBusca] = useState(filtrosIniciais.busca)
  const [perfil, setPerfil] = useState(filtrosIniciais.perfil)

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onPesquisar({ busca, perfil })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-3 sm:flex-row sm:items-end sm:flex-wrap"
    >
      <Field className="sm:w-64">
        <FieldLabel htmlFor={idCampoBusca}>Nome ou e-mail</FieldLabel>
        <FieldContent>
          <Input id={idCampoBusca} value={busca} onChange={(e) => setBusca(e.target.value)} />
        </FieldContent>
      </Field>
      {mostrarFiltroPerfil && (
        <Field className="sm:w-56">
          <FieldLabel htmlFor="usuarios-filtro-perfil">Perfil</FieldLabel>
          <FieldContent>
            <SelectComRotulo
              id="usuarios-filtro-perfil"
              value={perfil || "todos"}
              onValueChange={(v) => setPerfil(!v || v === "todos" ? "" : v)}
              itens={[{ value: "todos", label: "Todos" }, ...PERFIS_ESCOLHIVEIS]}
            />
          </FieldContent>
        </Field>
      )}
      <LoadingButton type="submit" loading={pesquisando} loadingText="Pesquisando...">
        Pesquisar
      </LoadingButton>
    </form>
  )
}
