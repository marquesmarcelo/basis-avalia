import type { NextRequest } from "next/server"
import { NextResponse } from "next/server"

function origemDaApi(): string | null {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL
  if (!apiUrl) {
    return null
  }
  try {
    return new URL(apiUrl).origin
  } catch {
    return null
  }
}

export function proxy(request: NextRequest) {
  const isDev = process.env.APP_ENV !== "production"
  const nonce = Buffer.from(crypto.randomUUID()).toString("base64")
  const apiOrigin = origemDaApi()
  const connectSrc = ["'self'", apiOrigin].filter(Boolean).join(" ")

  const cspHeader = isDev
    ? `default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src ${connectSrc} ws:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'`
    : `default-src 'self'; script-src 'self' 'nonce-${nonce}'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self'; connect-src ${connectSrc}; frame-ancestors 'none'; base-uri 'self'; form-action 'self'`

  const requestHeaders = new Headers(request.headers)
  requestHeaders.set("x-nonce", nonce)
  requestHeaders.set("Content-Security-Policy", cspHeader)

  const response = NextResponse.next({ request: { headers: requestHeaders } })
  response.headers.set("Content-Security-Policy", cspHeader)

  if (!isDev) {
    response.headers.set("X-Frame-Options", "DENY")
    response.headers.set("X-Content-Type-Options", "nosniff")
    response.headers.set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin")
    response.headers.set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
  }

  return response
}

export const config = {
  matcher: [
    {
      source: "/((?!_next/static|_next/image|favicon.ico).*)",
      missing: [
        { type: "header", key: "next-router-prefetch" },
        { type: "header", key: "purpose", value: "prefetch" },
      ],
    },
  ],
}
