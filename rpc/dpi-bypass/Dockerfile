# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o dpi-bypass rpc/dpi-bypass/cmd/main.go

FROM alpine:3.19
WORKDIR /app
RUN mkdir -p /app/etc
COPY --from=builder /app/dpi-bypass /app/dpi-bypass
COPY rpc/dpi-bypass/etc/config.env /app/etc/config.env
EXPOSE 8080
CMD ["/app/dpi-bypass"] 