import type { NextConfig } from "next"

const nextConfig: NextConfig = {
  // Output standalone — necessário para o build multi-stage do Dockerfile
  // de produção (imagem final não tem node_modules completo).
  output: "standalone",
}

export default nextConfig
