import type { Metadata } from "next"
import { Geist, Geist_Mono } from "next/font/google"
import "./globals.css"
import { Toaster } from "@/components/shared/ui/notificacoes"
import { TopProgressBar } from "@/components/shared/ui/top-progress-bar"
import { TooltipProvider } from "@/components/ui/tooltip"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "basis-avalia",
  description:
    "Acompanhamento das metas da coordenação, apoiado nos instrumentos de avaliação do INEP.",
}

export const dynamic = "force-dynamic"

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="pt-BR" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col">
        <TooltipProvider>
          <Toaster>
            <TopProgressBar />
            {children}
          </Toaster>
        </TooltipProvider>
      </body>
    </html>
  )
}
