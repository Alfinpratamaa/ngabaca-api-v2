
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ngabaca-api ./cmd/web/main.go


FROM alpine:latest

WORKDIR /
COPY --from=builder /ngabaca-api /ngabaca-api

COPY app.env .
COPY public ./public

copy api-docs ./api-docs



EXPOSE 3000

CMD ["/ngabaca-api"]