# Dockerfile de produção — frontend Next.js
# Build multi-stage: imagem final só com o output standalone do `next build`.

FROM node:26-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:26-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV APP_ENV=production
RUN npm run build

FROM node:26-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
# Imagem node:alpine já vem com o usuário não-root "node" — reaproveitar
# em vez de criar outro (evita conflito de uid/gid 1000 já em uso).
USER node
EXPOSE 3000
HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3000/ || exit 1
CMD ["node", "server.js"]
