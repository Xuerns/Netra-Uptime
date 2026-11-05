# ── STAGE 1: Build Client (SvelteKit) ─────────────
FROM node:22.23.4-alpine AS client-builder
ENV CI=true
WORKDIR /app/client

COPY client/package.json client/pnpm-lock.yaml client/pnpm-workspace.yaml client/.npmrc ./
RUN corepack enable && corepack install

RUN pnpm install --frozen-lockfile

COPY client/ ./
RUN pnpm build

# ── STAGE 2: Build Server (Go with embedded Client) ──
FROM golang:1.27.3-alpine AS server-builder
WORKDIR /app/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

# Copy compiled client assets into Go embed folder
COPY --from=client-builder /app/client/build ./internal/static/build

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o netra-uptime ./cmd/netra-uptime

# ── STAGE 3: Minimal Single Container Runner ────────
FROM alpine:3.20 AS runner

WORKDIR /app

COPY --from=server-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=server-builder /app/server/netra-uptime ./netra-uptime

EXPOSE 27346

ENV PORT=27346
ENV HOST=0.0.0.0
ENV APP_ENV=production

CMD ["./netra-uptime"]
