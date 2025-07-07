# Silence VPN

## Обзор проекта

Silence VPN — это микросервисная система для защищённого VPN с обфускацией трафика. Проект состоит из множества взаимосвязанных сервисов, каждый из которых отвечает за определённую функциональность.

**Язык**: Go 1.23
**Архитектура**: Микросервисы с Docker контейнерами
**Всего сервисов**: 11 (7 application + 4 infrastructure)

## Быстрый старт

### Требования к системе
- Docker 20.10+ с Docker Compose 2.0+
- 8GB RAM минимум (16GB рекомендуется)
- 20GB свободного места на диске
- macOS/Linux/Windows с WSL2

### Команды
```bash
# Запустить все сервисы
./manage.sh start

# Проверить состояние системы
./manage.sh health

# Посмотреть статус сервисов
./manage.sh status

# Посмотреть логи
./manage.sh logs [service-name]

# Остановить систему
./manage.sh stop
```

### URL сервисов (после запуска)
- Gateway: http://localhost:8080
- Auth: http://localhost:8081
- Analytics: http://localhost:8082
- DPI Bypass: http://localhost:8083
- VPN Core: http://localhost:8084
- Server Manager: http://localhost:8085
- Notifications: http://localhost:8087
- RabbitMQ UI: http://localhost:15672 (admin/admin)
- InfluxDB: http://localhost:8086

## Архитектура

### Схема системы
```
Frontend → API Gateway (8080) → Микросервисы
                ↓
        ┌─────────────────┐
        │ Auth Service    │ - Аутентификация, пользователи
        │ VPN Core        │ - WireGuard туннели
        │ DPI Bypass      │ - Обфускация трафика
        │ Analytics       │ - Метрики, дашборды, алерты
        │ Notifications   │ - Уведомления (email, SMS, push)
        │ Server Manager  │ - Управление серверами
        └─────────────────┘
                ↓
        ┌─────────────────┐
        │ InfluxDB        │
        │ Redis           │
        │ PostgreSQL      │
        │ RabbitMQ        │
        └─────────────────┘
```

### Сервисы приложения (Go)
- `api/gateway` - API Gateway и балансировщик нагрузки (Port 8080)
- `api/auth` - Сервис аутентификации с JWT (Port 8081)
- `rpc/analytics` - Сервис аналитики и отчетов (Port 8082)
- `rpc/dpi-bypass` - Сервис обхода DPI (Port 8083)
- `rpc/vpn-core` - Ядро управления VPN (Port 8084)
- `rpc/server-manager` - Управление инфраструктурой (Port 8085)
- `rpc/notifications` - Сервис уведомлений (Port 8087)

### Инфраструктурные сервисы
- PostgreSQL (Port 5432) - Основная база данных
- Redis (Port 6379) - Кэширование и сессии
- RabbitMQ (Ports 5672/15672) - Брокер сообщений
- InfluxDB (Port 8086) - База данных временных рядов для аналитики

### Поток данных
```
Клиент → Gateway (8080) → Auth/Analytics/VPN Core/DPI Bypass/Server Manager/Notifications
                     ↓
Инфраструктура: PostgreSQL, Redis, RabbitMQ, InfluxDB
```

## Развертывание (Docker)

Система полностью контейнеризирована с использованием Docker.

**Ключевые файлы:**
- `docker-compose.yml` - Основной файл оркестрации
- `.env` + `.env.example` - Конфигурация окружения
- `.dockerignore` - Оптимизация сборки
- `Dockerfile` для каждого сервиса (стандартизированные многоступенчатые сборки)
- `manage.sh` - Главный скрипт управления проектом

**Возможности скриптов управления:**
- Запуск/остановка/перезапуск сервисов
- Мониторинг состояния и статуса
- Просмотр логов и отладка
- Миграции базы данных
- Резервное копирование и очистка

Для получения более подробной информации см. `DEVELOPER_GUIDE.md` и `API_REFERENCE.md`.
