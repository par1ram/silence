# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod download
RUN go build -o gateway api/gateway/cmd/main.go

FROM alpine:3.19
WORKDIR /app
RUN mkdir -p /app/etc
COPY --from=builder /app/gateway /app/gateway
COPY api/gateway/etc/config.env /app/etc/config.env
EXPOSE 8080
CMD ["/app/gateway"] 