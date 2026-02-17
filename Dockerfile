# ── Frontend build stage ──
FROM node:22-alpine AS frontend

WORKDIR /app/web

# Cache de dependências npm
COPY web/package.json web/package-lock.json* ./
RUN npm ci --no-audit --no-fund

# Build da SPA React
COPY web/ .
RUN npm run build

# ── Go build stage ──
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add ca-certificates git gcc musl-dev

WORKDIR /app

# Cache de dependências Go
COPY go.mod go.sum ./
RUN go mod download

# Copiar código Go
COPY . .

# Copiar dist do frontend para o embed directory
COPY --from=frontend /app/web/dist ./pkg/devclaw/webui/dist/

# Build do binário com SQLite FTS5
RUN CGO_ENABLED=1 GOOS=linux go build \
    -tags 'sqlite_fts5' \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o devclaw ./cmd/devclaw

# ── Runtime stage ──
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

# Cria usuário não-root para segurança
RUN addgroup -S devclaw && adduser -S devclaw -G devclaw

USER devclaw
WORKDIR /home/devclaw

# Copia binário e config de exemplo
COPY --from=builder /app/devclaw /usr/local/bin/devclaw
COPY --from=builder /app/configs/devclaw.example.yaml /etc/devclaw/config.example.yaml

# Volumes para persistência de sessões e dados
VOLUME ["/home/devclaw/sessions", "/home/devclaw/data"]

# Expor portas: gateway (8080) e webui (8090)
EXPOSE 8080 8090

# Health check via comando `devclaw health`.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["devclaw", "health"]

ENTRYPOINT ["devclaw"]
CMD ["serve", "--config", "/etc/devclaw/config.yaml"]
