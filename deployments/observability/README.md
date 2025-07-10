# Silence VPN Observability Stack

Полный стек наблюдаемости для Silence VPN на базе OpenTelemetry, включающий метрики, трейсинг и логирование.

## Компоненты

### 🔍 Сбор и маршрутизация данных
- **OpenTelemetry Collector** - Центральный компонент для сбора, обработки и маршрутизации телеметрии
- **Prometheus** - Система мониторинга и сбора метрик
- **Node Exporter** - Экспортер системных метрик
- **cAdvisor** - Мониторинг контейнеров

### 📊 Метрики и визуализация
- **Grafana** - Дашборды и визуализация метрик
- **AlertManager** - Система алертов и уведомлений

### 🔬 Распределенный трейсинг
- **Jaeger** - Система распределенного трейсинга
- **Zipkin** - Альтернативная система трейсинга
- **Tempo** - Высокопроизводительная система трейсинга от Grafana

### 📝 Логирование
- **Loki** - Система агрегации логов
- **Promtail** - Агент для сбора логов

### 🗄️ Хранилище
- **Redis** - Кэширование и сессии

## Быстрый старт

### 1. Запуск стека наблюдаемости

```bash
cd silence/deployments/observability
docker-compose up -d
```

### 2. Проверка статуса сервисов

```bash
docker-compose ps
```

### 3. Доступ к интерфейсам

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger UI**: http://localhost:16686
- **Zipkin UI**: http://localhost:9411
- **Loki**: http://localhost:3100
- **AlertManager**: http://localhost:9093

## Конфигурация

### OpenTelemetry Collector

Основная конфигурация в `otel-collector-config.yaml`:

```yaml
receivers:
  - otlp (gRPC/HTTP)
  - prometheus
  - jaeger
  - zipkin
  - filelog

processors:
  - batch
  - memory_limiter
  - resource
  - attributes
  - span
  - probabilistic_sampler

exporters:
  - jaeger
  - zipkin
  - otlp/tempo
  - prometheus
  - loki
```

### Prometheus

Конфигурация в `prometheus.yml` включает:
- Scraping целей всех сервисов Silence VPN
- Recording rules для общих запросов
- Интеграцию с AlertManager
- Remote write в Tempo

### Grafana

Предустановленные дашборды:
- VPN Analytics Overview
- Server Performance
- User Activity
- Error Tracking
- Infrastructure Metrics

## Метрики приложения

### VPN Specific Metrics

```
# Подключения
silence_active_connections{server_id, region}
silence_total_connections{user_id, server_id, protocol, region}
silence_connection_duration_seconds{server_id, region}

# Серверы
silence_server_cpu_usage_percent{server_id, region}
silence_server_memory_usage_percent{server_id, region}
silence_server_connections{server_id, region}

# Пользователи
silence_active_users{region}
silence_user_sessions{user_id, region}
silence_user_data_transferred_bytes{user_id, region}

# Обход блокировок
silence_bypass_attempts{bypass_type, region}
silence_bypass_success{bypass_type, region}
silence_bypass_latency_seconds{bypass_type, region}

# Система
silence_system_load
silence_system_errors{error_type, service}
silence_system_response_time_seconds{method, endpoint}
```

### Analytics Service Metrics

```
# Обработка метрик
silence_metrics_processed{metric_type}
silence_metrics_errors{error_type, metric_type}
silence_metrics_latency_seconds{metric_type}

# Дашборды
silence_dashboard_requests{dashboard_type}
silence_predictions_generated{prediction_type}
```

## Трейсинг

### Конфигурация трейсинга в приложении

```go
// Инициализация OpenTelemetry
telemetryManager, err := telemetry.NewTelemetryManager(cfg.OpenTelemetry, logger)

// Создание трейсера
tracer := telemetryManager.GetTracer()

// Создание span
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// Добавление атрибутов
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.String("server.id", serverID),
)
```

### Распределенный трейсинг

Все сервисы Silence VPN настроены для передачи trace context:
- HTTP headers (трейсинг между сервисами)
- gRPC metadata (трейсинг RPC вызовов)
- Логи с trace_id и span_id

## Алерты

### Критические алерты

```yaml
# Сервис недоступен
- alert: ServiceDown
  expr: up == 0
  for: 1m
  labels:
    severity: critical

# Высокая частота ошибок
- alert: HighErrorRate
  expr: rate(silence_system_errors[5m]) > 0.1
  for: 5m
  labels:
    severity: critical

# Высокая задержка VPN
- alert: VPNHighLatency
  expr: histogram_quantile(0.95, rate(silence_bypass_latency_seconds_bucket[5m])) > 1.0
  for: 5m
  labels:
    severity: warning
```

### Каналы уведомлений

- **Email** - критические алерты
- **Slack** - все алерты
- **Webhook** - интеграция с внешними системами

## Логирование

### Структурированные логи

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "analytics",
  "message": "User activity recorded",
  "trace_id": "abc123",
  "span_id": "def456",
  "user_id": "user123",
  "server_id": "server-us-1",
  "region": "us"
}
```

### Запросы в Loki

```logql
# Ошибки в analytics сервисе
{service="analytics"} |= "ERROR"

# Логи пользователя
{user_id="user123"} | json

# Логи по trace_id
{trace_id="abc123"} | json
```

## Производительность

### Настройки для production

```yaml
# OpenTelemetry Collector
processors:
  batch:
    send_batch_size: 1024
    timeout: 10s
  memory_limiter:
    limit_mib: 500
  probabilistic_sampler:
    sampling_percentage: 10

# Prometheus
global:
  scrape_interval: 15s
  evaluation_interval: 15s

# Loki
limits_config:
  ingestion_rate_mb: 4
  ingestion_burst_size_mb: 6
  retention_period: 744h
```

### Ресурсы

Рекомендуемые ресурсы для production:
- **OpenTelemetry Collector**: 1 CPU, 1GB RAM
- **Prometheus**: 2 CPU, 4GB RAM
- **Grafana**: 1 CPU, 512MB RAM
- **Jaeger**: 1 CPU, 1GB RAM
- **Loki**: 1 CPU, 2GB RAM
- **Tempo**: 1 CPU, 2GB RAM

## Мониторинг системы наблюдаемости

### Health checks

```bash
# OpenTelemetry Collector
curl http://localhost:13133/

# Prometheus
curl http://localhost:9090/-/healthy

# Grafana
curl http://localhost:3000/api/health

# Jaeger
curl http://localhost:16686/

# Loki
curl http://localhost:3100/ready
```

### Внутренние метрики

Все компоненты экспортируют собственные метрики:
- otel_collector_* - метрики коллектора
- prometheus_* - метрики Prometheus
- grafana_* - метрики Grafana
- jaeger_* - метрики Jaeger
- loki_* - метрики Loki

## Разработка

### Локальная разработка

```bash
# Запуск только необходимых сервисов
docker-compose up -d prometheus grafana jaeger loki

# Просмотр логов
docker-compose logs -f otel-collector

# Перезапуск с новой конфигурацией
docker-compose restart otel-collector
```

### Тестирование

```bash
# Отправка тестовых трейсов
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d @test-trace.json

# Отправка тестовых метрик
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d @test-metrics.json
```

## Устранение неполадок

### Общие проблемы

1. **Коллектор не получает данные**
   - Проверьте конфигурацию receivers
   - Убедитесь, что приложение отправляет данные на правильные endpoints

2. **Метрики не отображаются в Grafana**
   - Проверьте настройки datasource
   - Убедитесь, что Prometheus scraping работает

3. **Трейсы не видны в Jaeger**
   - Проверьте конфигурацию exporters
   - Убедитесь, что sampling настроен правильно

### Логи для отладки

```bash
# OpenTelemetry Collector
docker-compose logs otel-collector

# Prometheus
docker-compose logs prometheus

# Grafana
docker-compose logs grafana
```

## Безопасность

### Рекомендации для production

1. **Аутентификация**
   - Настройте аутентификацию в Grafana
   - Используйте TLS для всех подключений
   - Настройте RBAC для доступа к дашбордам

2. **Сеть**
   - Используйте internal networks
   - Настройте firewall правила
   - Ограничьте доступ к административным портам

3. **Данные**
   - Настройте retention policies
   - Используйте шифрование для sensitive данных
   - Регулярно архивируйте старые данные

## Масштабирование

### Горизонтальное масштабирование

```yaml
# Несколько экземпляров коллектора
otel-collector-1:
  image: otel/opentelemetry-collector-contrib:latest
  ports:
    - "4317:4317"

otel-collector-2:
  image: otel/opentelemetry-collector-contrib:latest
  ports:
    - "4318:4318"

# Load balancer
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
  depends_on:
    - otel-collector-1
    - otel-collector-2
```

### Вертикальное масштабирование

```yaml
services:
  prometheus:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '1'
          memory: 2G
```

## Поддержка

Для вопросов и проблем:
- Создайте issue в репозитории
- Проверьте документацию OpenTelemetry
- Обратитесь к команде разработки

## Ссылки

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)