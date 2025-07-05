# Интеграция Analytics сервиса с Gateway

## Обзор

Analytics сервис интегрирован в Gateway как единая точка входа для всех аналитических запросов. Сервис предоставляет метрики, дашборды и алерты для мониторинга системы Silence VPN.

## Архитектура интеграции

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Gateway   │    │  Analytics  │    │  InfluxDB   │
│   :8080     │◄──►│   :8084     │◄──►│   :8086     │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Прокси    │    │   Метрики   │    │   Хранение  │
│   запросов  │    │   + Алерты  │    │   данных    │
└─────────────┘    └─────────────┘    └─────────────┘
```

## API эндпоинты

### Через Gateway (с аутентификацией)

```
GET    /api/v1/analytics/health                    # Health check
GET    /api/v1/analytics/metrics/connections       # Метрики подключений
GET    /api/v1/analytics/metrics/bypass-effectiveness # Эффективность обхода DPI
GET    /api/v1/analytics/metrics/user-activity     # Активность пользователей
GET    /api/v1/analytics/metrics/server-load       # Нагрузка серверов
GET    /api/v1/analytics/metrics/errors            # Метрики ошибок
GET    /api/v1/analytics/dashboards                # Список дашбордов
POST   /api/v1/analytics/dashboards                # Создание дашборда
GET    /api/v1/analytics/dashboards/{id}           # Получение дашборда
PUT    /api/v1/analytics/dashboards/{id}           # Обновление дашборда
DELETE /api/v1/analytics/dashboards/{id}           # Удаление дашборда
GET    /api/v1/analytics/alerts                    # Список алертов
POST   /api/v1/analytics/alerts                    # Создание алерта
GET    /api/v1/analytics/alerts/{id}               # Получение алерта
PUT    /api/v1/analytics/alerts/{id}               # Обновление алерта
DELETE /api/v1/analytics/alerts/{id}               # Удаление алерта
```

### Прямой доступ к Analytics (без аутентификации)

```
GET    http://localhost:8084/health                # Health check
GET    http://localhost:8084/metrics/connections   # Метрики подключений
GET    http://localhost:8084/metrics/bypass-effectiveness # Эффективность обхода DPI
GET    http://localhost:8084/metrics/user-activity # Активность пользователей
GET    http://localhost:8084/metrics/server-load   # Нагрузка серверов
GET    http://localhost:8084/metrics/errors        # Метрики ошибок
GET    http://localhost:8084/dashboards            # Список дашбордов
POST   http://localhost:8084/dashboards            # Создание дашборда
GET    http://localhost:8084/dashboards/{id}       # Получение дашборда
PUT    http://localhost:8084/dashboards/{id}       # Обновление дашборда
DELETE http://localhost:8084/dashboards/{id}       # Удаление дашборда
GET    http://localhost:8084/alerts                # Список алертов
POST   http://localhost:8084/alerts                # Создание алерта
GET    http://localhost:8084/alerts/{id}           # Получение алерта
PUT    http://localhost:8084/alerts/{id}           # Обновление алерта
DELETE http://localhost:8084/alerts/{id}           # Удаление алерта
```

## Конфигурация

### Переменные окружения

```bash
# Analytics Service
HTTP_PORT=:8084
LOG_LEVEL=info
VERSION=1.0.0

# InfluxDB
INFLUXDB_URL=http://localhost:8086
INFLUXDB_TOKEN=your-influxdb-token
INFLUXDB_ORG=your-org
INFLUXDB_BUCKET=analytics

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# External Services
VPN_CORE_URL=http://localhost:8082
GATEWAY_URL=http://localhost:8080
DPI_BYPASS_URL=http://localhost:8083
AUTH_URL=http://localhost:8081

# Gateway (добавить)
ANALYTICS_URL=http://localhost:8084
```

### Docker Compose

```yaml
version: '3.8'

services:
  # ... существующие сервисы ...

  influxdb:
    image: influxdb:2.7-alpine
    container_name: silence_influxdb
    environment:
      DOCKER_INFLUXDB_INIT_MODE: setup
      DOCKER_INFLUXDB_INIT_USERNAME: admin
      DOCKER_INFLUXDB_INIT_PASSWORD: adminpassword
      DOCKER_INFLUXDB_INIT_ORG: silence
      DOCKER_INFLUXDB_INIT_BUCKET: analytics
      DOCKER_INFLUXDB_INIT_ADMIN_TOKEN: your-influxdb-token
    ports:
      - '8086:8086'
    volumes:
      - influxdb_data:/var/lib/influxdb2
```

## Запуск и тестирование

### 1. Запуск зависимостей

```bash
# Запуск PostgreSQL, Redis, InfluxDB
docker-compose up -d postgres redis influxdb

# Ожидание запуска
sleep 10
```

### 2. Сборка сервисов

```bash
# Сборка Analytics
task build:analytics

# Сборка Gateway
task build:gateway
```

### 3. Запуск сервисов

```bash
# Запуск Analytics
cd rpc/analytics
./bin/analytics &

# Запуск Gateway
cd api/gateway
./bin/gateway &
```

### 4. Тестирование

```bash
# Тест health check
curl http://localhost:8080/health
curl http://localhost:8084/health

# Тест проксирования
curl http://localhost:8080/api/v1/analytics/health

# Тест метрик (требует JWT токен)
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/v1/analytics/metrics/connections
```

### 5. Автоматическое тестирование

```bash
# Запуск скрипта тестирования
./scripts/test-integration.sh
```

## Компоненты Analytics сервиса

### 1. MetricsCollector

Собирает метрики из других сервисов:

- **VPN Core**: нагрузка серверов, статистика туннелей
- **Gateway**: количество запросов, время ответа
- **DPI Bypass**: эффективность обхода, статистика bypass
- **Auth**: активность пользователей, логины

### 2. AlertService

Управляет алертами:

- Создание правил алертов
- Оценка условий
- Отправка уведомлений
- История алертов

### 3. DashboardRepository

Управляет дашбордами:

- Создание/редактирование дашбордов
- Хранение конфигураций в Redis
- Виджеты для отображения метрик

### 4. InfluxDBRepository

Хранит метрики:

- Временные серии данных
- Агрегация метрик
- Запросы для анализа

## Примеры использования

### Получение метрик подключений

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  "http://localhost:8080/api/v1/analytics/metrics/connections?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z"
```

### Создание дашборда

```bash
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "VPN Overview",
    "description": "Обзор VPN метрик",
    "widgets": [
      {
        "type": "chart",
        "title": "Подключения",
        "query": {
          "metric": "connections",
          "time_range": "1h"
        }
      }
    ]
  }' \
  http://localhost:8080/api/v1/analytics/dashboards
```

### Создание алерта

```bash
curl -X POST -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Server Load",
    "description": "Высокая нагрузка на сервер",
    "condition": "cpu_usage > 90",
    "severity": "high",
    "message": "CPU usage превышает 90%"
  }' \
  http://localhost:8080/api/v1/analytics/alerts
```

## Мониторинг

### Логи

```bash
# Логи Analytics
tail -f rpc/analytics/build-errors.log

# Логи Gateway
tail -f api/gateway/build-errors.log
```

### Метрики

- **Количество запросов к Analytics**
- **Время ответа API**
- **Количество собранных метрик**
- **Количество активных алертов**

### Health Checks

```bash
# Проверка Analytics
curl http://localhost:8084/health

# Проверка через Gateway
curl http://localhost:8080/api/v1/analytics/health
```

## Troubleshooting

### Проблемы с подключением

```bash
# Проверка портов
netstat -tlnp | grep :8084
netstat -tlnp | grep :8080

# Проверка логов
tail -f rpc/analytics/build-errors.log
```

### Проблемы с InfluxDB

```bash
# Проверка InfluxDB
curl http://localhost:8086/health

# Проверка токена
curl -H "Authorization: Token your-influxdb-token" \
  http://localhost:8086/api/v2/buckets
```

### Проблемы с Redis

```bash
# Проверка Redis
redis-cli ping

# Проверка данных
redis-cli keys "*dashboard*"
```

## Будущие улучшения

1. **Веб-интерфейс** для управления дашбордами
2. **Экспорт метрик** в Prometheus
3. **Интеграция с Grafana** для визуализации
4. **Машинное обучение** для прогнозирования
5. **Уведомления** через email/SMS/Telegram
6. **Автоматическое масштабирование** на основе метрик
