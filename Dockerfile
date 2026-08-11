# ── STAGE 1: Build Frontend (SvelteKit) ─────────────
FROM node:22-alpine AS frontend-builder
ENV CI=true
WORKDIR /app/frontend

RUN corepack enable && corepack prepare pnpm@latest --activate

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm build

# ── STAGE 2: Build Backend (Go with embedded Frontend) ──
FROM golang:alpine AS backend-builder
WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Copy compiled frontend assets into Go embed folder
COPY --from=frontend-builder /app/frontend/build ./internal/static/build

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pulsewatch ./cmd/server

# ── STAGE 3: Minimal Single Container Runner ────────
FROM alpine:3.20 AS runner

WORKDIR /app

COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=backend-builder /app/backend/pulsewatch ./pulsewatch

EXPOSE 8080

ENV PORT=8080
ENV APP_ENV=production
ENV APP_MODE=selfhosted

CMD ["./pulsewatch"]
