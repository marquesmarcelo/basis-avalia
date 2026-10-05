"use client"

import Link from "next/link"
import { Fragment } from "react"
import { IndicadorDeNavegacao } from "@/components/shared/ui/indicador-de-navegacao"
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"

export interface TrilhaItem {
  rotulo: string
  href?: string
}

interface TrilhaProps {
  itens: TrilhaItem[]
}

function TrilhaCrumb({ item }: { item: TrilhaItem }) {
  if (!item.href) {
    return <BreadcrumbPage>{item.rotulo}</BreadcrumbPage>
  }
  return (
    <BreadcrumbLink
      render={
        <Link href={item.href}>
          {item.rotulo}
          <IndicadorDeNavegacao className="ml-1 inline size-3" />
        </Link>
      }
    />
  )
}

export function Trilha({ itens }: TrilhaProps) {
  const primeiro = itens[0]
  const ultimo = itens[itens.length - 1]
  const intermediarios = itens.slice(1, -1)

  return (
    <Breadcrumb className="mb-4">
      <BreadcrumbList>
        <BreadcrumbItem>
          <TrilhaCrumb item={primeiro} />
        </BreadcrumbItem>

        {intermediarios.length > 0 && (
          <>
            <BreadcrumbSeparator className="sm:hidden" />
            <BreadcrumbItem className="sm:hidden">
              <BreadcrumbEllipsis />
            </BreadcrumbItem>
          </>
        )}

        {intermediarios.map((item) => (
          <Fragment key={item.rotulo}>
            <BreadcrumbSeparator className="hidden sm:inline-flex" />
            <BreadcrumbItem className="hidden sm:inline-flex">
              <TrilhaCrumb item={item} />
            </BreadcrumbItem>
          </Fragment>
        ))}

        {itens.length > 1 && (
          <>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage>{ultimo.rotulo}</BreadcrumbPage>
            </BreadcrumbItem>
          </>
        )}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
