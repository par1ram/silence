FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
# Копируем только go.mod и go.sum для go mod download
COPY rpc/notifications/go.mod ./
COPY rpc/notifications/go.sum ./
COPY shared ./shared
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifications rpc/notifications/cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/notifications .

# Экспортируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./notifications"] 