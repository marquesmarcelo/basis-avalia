# Dockerfile de produção — backend Go + Gin
# Build multi-stage: imagem final sem toolchain Go, só o binário compilado.

FROM golang:1.27-alpine AS builder
WORKDIR /app
# Copiar go.mod e go.sum primeiro para aproveitar cache de layer quando
# só o código mudar (sem alteração de dependências).
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/api

FROM alpine:3.20 AS runner
WORKDIR /app
# Certificados TLS para chamadas HTTPS externas e fuso horário correto
RUN apk --no-cache add ca-certificates tzdata
# Usuário não-root (design.md §16 T-104, achado M-4) — uid 1000, mesmo
# padrão já usado em Dockerfile.dev e no frontend.
RUN adduser -D -u 1000 appuser
COPY --from=builder --chown=appuser:appuser /server .
COPY --from=builder --chown=appuser:appuser /app/internal/adapter/postgres/migrations ./migrations
USER appuser
EXPOSE 3001 9090
# /healthz e /readyz vivem na porta administrativa (design.md §4.5, §16
# T-100) — nunca na pública.
HEALTHCHECK --interval=10s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:9090/healthz || exit 1
CMD ["./server"]
