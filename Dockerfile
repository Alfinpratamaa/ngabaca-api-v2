# =========================
# Build stage
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (dibutuhkan untuk ambil dependencies)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (disable debug info biar lebih kecil & cepat)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ngabaca-api ./cmd/web

# =========================
# Final stage
# =========================
FROM alpine:latest

WORKDIR /

# Copy binary dari builder
COPY --from=builder /app/ngabaca-api /ngabaca-api

# Copy asset/env
COPY app.env .
COPY public ./public
COPY api-docs ./api-docs

EXPOSE 3000

CMD ["/ngabaca-api"]
