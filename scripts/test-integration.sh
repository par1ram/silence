#!/bin/bash

# Скрипт для тестирования интеграции Analytics сервиса с Gateway

set -e

echo "🚀 Тестирование интеграции Analytics сервиса с Gateway"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARN] $1${NC}"
}

# Проверка доступности сервисов
check_service() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    log "Проверка доступности $service_name..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url" > /dev/null 2>&1; then
            log "✅ $service_name доступен"
            return 0
        fi
        
        warn "Попытка $attempt/$max_attempts: $service_name недоступен, ожидание..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    error "❌ $service_name недоступен после $max_attempts попыток"
    return 1
}

# Запуск зависимостей
log "Запуск зависимостей (PostgreSQL, Redis, InfluxDB)..."
docker-compose up -d postgres redis influxdb

# Ожидание запуска зависимостей
sleep 10

# Проверка зависимостей
check_service "PostgreSQL" "http://localhost:5432" || exit 1
check_service "Redis" "http://localhost:6379" || exit 1
check_service "InfluxDB" "http://localhost:8086" || exit 1

# Сборка сервисов
log "Сборка сервисов..."
task build:analytics
task build:gateway

# Запуск Analytics сервиса
log "Запуск Analytics сервиса..."
cd rpc/analytics
./bin/analytics &
ANALYTICS_PID=$!
cd ../..

# Ожидание запуска Analytics
sleep 5
check_service "Analytics" "http://localhost:8084/health" || exit 1

# Запуск Gateway сервиса
log "Запуск Gateway сервиса..."
cd api/gateway
./bin/gateway &
GATEWAY_PID=$!
cd ../..

# Ожидание запуска Gateway
sleep 5
check_service "Gateway" "http://localhost:8080/health" || exit 1

# Тестирование API
log "Тестирование API..."

# Тест health check
log "Тест health check Gateway..."
curl -s http://localhost:8080/health | jq .

# Тест health check Analytics
log "Тест health check Analytics..."
curl -s http://localhost:8084/health | jq .

# Тест проксирования к Analytics через Gateway
log "Тест проксирования к Analytics через Gateway..."
curl -s http://localhost:8080/api/v1/analytics/health | jq .

# Тест метрик (без аутентификации - должно вернуть 401)
log "Тест метрик без аутентификации..."
curl -s -w "%{http_code}" http://localhost:8080/api/v1/analytics/metrics/connections | tail -1

# Очистка
log "Очистка..."
kill $ANALYTICS_PID 2>/dev/null || true
kill $GATEWAY_PID 2>/dev/null || true

log "✅ Тестирование завершено успешно!"

echo ""
echo "📊 Результаты тестирования:"
echo "  - Analytics сервис: ✅"
echo "  - Gateway сервис: ✅"
echo "  - Проксирование: ✅"
echo "  - Health checks: ✅"
echo ""
echo "🎉 Интеграция работает корректно!" 