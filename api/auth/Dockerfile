# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o auth api/auth/cmd/main.go

FROM alpine:3.19
WORKDIR /app
RUN mkdir -p /app/etc
COPY --from=builder /app/auth /app/auth
COPY api/auth/etc/config.env /app/etc/config.env
EXPOSE 8080
CMD ["/app/auth"] 