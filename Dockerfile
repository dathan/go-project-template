# ── Stage 1: Build frontend ────────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --prefer-offline
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Build Go binary ───────────────────────────────────────────────────
FROM golang:1.25-alpine AS go-builder
ENV CGO_ENABLED=0
RUN apk add --no-cache git make

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Embed the built frontend so the Go binary can serve it
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN make build-linux

# ── Stage 3: CA certs ─────────────────────────────────────────────────────────
FROM alpine:3.21 AS certs
RUN apk add -U --no-cache ca-certificates

# ── Stage 4: Final image ───────────────────────────────────────────────────────
FROM alpine:3.21 AS release
LABEL org.opencontainers.image.source="https://github.com/dathan/go-project-template"

COPY --from=certs      /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /app/bin/server                    /app/server
COPY --from=go-builder /app/conf/config.yaml               /app/conf/config.yaml

WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
