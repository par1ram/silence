# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod download
RUN go build -o server-manager rpc/server-manager/cmd/main.go

FROM alpine:3.19
WORKDIR /app
RUN mkdir -p /app/etc
COPY --from=builder /app/server-manager /app/server-manager
COPY rpc/server-manager/etc/config.env /app/etc/config.env
COPY rpc/server-manager/internal/adapters/database/migrations /app/internal/adapters/database/migrations
EXPOSE 8080
CMD ["/app/server-manager"] 