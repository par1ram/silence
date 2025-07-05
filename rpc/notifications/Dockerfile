FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifications cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем собранное приложение
COPY --from=builder /app/notifications .

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./notifications"] 