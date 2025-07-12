# Справочник интеграции фронтенда с бэкендом Silence VPN

## 🏗️ Архитектура проекта

### Структура сервисов
```
silence/
├── api/gateway/          # API Gateway (HTTP прокси, порт 8080)
├── api/auth/            # Сервис аутентификации (gRPC + HTTP, порт 50051)
├── rpc/analytics/       # Аналитика (gRPC, порт 50052)
├── rpc/server-manager/  # Управление серверами (gRPC, порт 50053)
├── rpc/vpn-core/       # VPN туннели (gRPC, порт 50054)
├── rpc/dpi-bypass/     # Обход DPI (gRPC, порт 50055)
├── rpc/notifications/  # Уведомления (gRPC, порт 50056)
└── frontend/           # React фронтенд (порт 3000)
```

### Сетевое взаимодействие
- **Фронтенд** → **Gateway** (HTTP REST API на порту 8080)
- **Gateway** → **Сервисы** (gRPC на портах 50051-50056)
- **Сервисы** → **Базы данных** (PostgreSQL, Redis, ClickHouse, InfluxDB)

## 📡 API Endpoints

### Аутентификация (Auth Service)
```
POST   /api/v1/auth/login        # Вход пользователя
POST   /api/v1/auth/register     # Регистрация
GET    /api/v1/auth/me          # Профиль пользователя
POST   /api/v1/users            # Создание пользователя (админ)
GET    /api/v1/users/{id}       # Получение пользователя
PUT    /api/v1/users/{id}       # Обновление пользователя
DELETE /api/v1/users/{id}       # Удаление пользователя
GET    /api/v1/users            # Список пользователей
POST   /api/v1/users/{id}/block  # Блокировка пользователя
POST   /api/v1/users/{id}/unblock # Разблокировка
POST   /api/v1/users/{id}/role  # Изменение роли
```

### Управление серверами (Server Manager)
```
GET    /api/v1/servers          # Список серверов
POST   /api/v1/servers          # Создание сервера
GET    /api/v1/servers/{id}     # Детали сервера
PUT    /api/v1/servers/{id}     # Обновление сервера
DELETE /api/v1/servers/{id}     # Удаление сервера
POST   /api/v1/servers/{id}/start    # Запуск сервера
POST   /api/v1/servers/{id}/stop     # Остановка сервера
POST   /api/v1/servers/{id}/restart  # Перезапуск сервера
GET    /api/v1/servers/{id}/stats    # Статистика сервера
GET    /api/v1/servers/{id}/health   # Состояние сервера
POST   /api/v1/servers/{id}/scale    # Масштабирование
```

### VPN туннели (VPN Core)
```
GET    /api/v1/vpn/tunnels      # Список туннелей
POST   /api/v1/vpn/tunnels      # Создание туннеля
GET    /api/v1/vpn/tunnels/{id} # Детали туннеля
DELETE /api/v1/vpn/tunnels/{id} # Удаление туннеля
POST   /api/v1/vpn/tunnels/{id}/start # Запуск туннеля
POST   /api/v1/vpn/tunnels/{id}/stop  # Остановка туннеля
GET    /api/v1/vpn/tunnels/{id}/stats # Статистика туннеля
GET    /api/v1/vpn/tunnels/{id}/peers # Список пиров
POST   /api/v1/vpn/tunnels/{id}/peers # Добавление пира
```

### Аналитика (Analytics)
```
GET    /api/v1/analytics/dashboard     # Данные дашборда
GET    /api/v1/analytics/metrics       # Метрики
POST   /api/v1/analytics/metrics       # Отправка метрики
GET    /api/v1/analytics/metrics/history # История метрик
GET    /api/v1/analytics/statistics    # Статистика
GET    /api/v1/analytics/statistics/system # Системная статистика
GET    /api/v1/analytics/statistics/users/{user_id} # Статистика пользователя
```

### Обход DPI (DPI Bypass)
```
GET    /api/v1/dpi/configs      # Конфигурации обхода
POST   /api/v1/dpi/configs      # Создание конфигурации
GET    /api/v1/dpi/configs/{id} # Детали конфигурации
PUT    /api/v1/dpi/configs/{id} # Обновление конфигурации
DELETE /api/v1/dpi/configs/{id} # Удаление конфигурации
POST   /api/v1/dpi/bypass/start # Запуск обхода
POST   /api/v1/dpi/bypass/stop  # Остановка обхода
GET    /api/v1/dpi/bypass/{session_id}/status # Статус сессии
```

### Уведомления (Notifications)
```
GET    /api/v1/notifications    # Список уведомлений
POST   /api/v1/notifications/dispatch # Отправка уведомления
GET    /api/v1/notifications/{id} # Детали уведомления
PATCH  /api/v1/notifications/{id}/status # Обновление статуса
GET    /api/v1/notifications/templates # Шаблоны
POST   /api/v1/notifications/templates # Создание шаблона
GET    /api/v1/notifications/preferences/{user_id} # Настройки пользователя
```

## 🔐 Авторизация

### JWT токены
- **Формат:** `Authorization: Bearer <token>`
- **Получение:** `POST /api/v1/auth/login`
- **Обновление:** `POST /api/v1/auth/refresh` (если реализовано)
- **Исключения:** Health check эндпоинты не требуют авторизации

### Роли пользователей
```typescript
enum UserRole {
  USER_ROLE_USER = "USER_ROLE_USER",           // Обычный пользователь
  USER_ROLE_MODERATOR = "USER_ROLE_MODERATOR", // Модератор  
  USER_ROLE_ADMIN = "USER_ROLE_ADMIN"          // Администратор
}
```

## 📦 Сгенерированные API хуки

### Расположение файлов
```
frontend/src/generated/
├── requests/
│   ├── index.ts              # Экспорт всех API функций
│   ├── services.gen.ts       # Сгенерированные API сервисы
│   ├── types.gen.ts          # Типы из Swagger схемы
│   └── core/                 # Базовые классы HTTP клиента
├── queries/                  # React Query хуки (если сгенерированы)
└── types.ts                  # Дополнительные типы проекта
```

### Пример использования
```typescript
import { AuthService, ServerManagerService } from '@/generated/requests';

// Логин пользователя
const loginResponse = await AuthService.authServiceLogin({
  email: "user@example.com",
  password: "password"
});

// Получение списка серверов
const servers = await ServerManagerService.serverManagerServiceListServers({
  limit: 10,
  offset: 0
});
```

## 🗄️ Базы данных и хранилища

### PostgreSQL (порт 5432)
- **silence_auth** - пользователи, роли, сессии
- **silence_vpn** - туннели, пиры, конфигурации
- **silence_server_manager** - серверы, их состояние

### Redis (порт 6379)
- Кэширование сессий
- Rate limiting
- WebSocket сессии
- Временные данные

### ClickHouse (порт 8123, 9000)
- Аналитические данные
- Метрики использования
- Логи событий

### InfluxDB (порт 8086)
- Временные ряды
- Метрики производительности
- Мониторинг серверов

## 🛠️ Команды разработки

### Запуск проекта
```bash
# Запуск инфраструктуры
make infra-up

# Запуск всех сервисов
make dev-all

# Запуск только нужных сервисов
make dev-auth      # Только auth
make dev-gateway   # Только gateway
make dev-single    # Минимальный набор

# Запуск фронтенда
make frontend-dev
# или
cd frontend && npm run dev
```

### Работа с API
```bash
# Генерация Swagger документации
make swagger

# Генерация клиентского SDK
make generate-client-sdk

# Проверка API схемы
npm run api:validate

# Автоматическая регенерация при изменениях
npm run generate:api:watch
```

### Тестирование
```bash
# Все тесты
make test

# Тесты конкретного сервиса
make test-auth
make test-gateway

# Проверка состояния
make health
make infra-status
```

## 🔧 Настройка окружения

### Переменные окружения
```bash
# Gateway
GATEWAY_PORT=8080
GATEWAY_HOST=0.0.0.0

# Auth Service  
AUTH_GRPC_PORT=50051
AUTH_HTTP_PORT=8081
AUTH_DB_HOST=localhost
AUTH_DB_PORT=5432
AUTH_DB_NAME=silence_auth

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=silence
POSTGRES_PASSWORD=silence
```

### Docker Compose
- **development:** `docker-compose.dev.yml`
- **production:** `docker-compose.yml` 
- **unified:** `docker-compose.unified.yml`

## 📋 Типы данных

### Основные сущности
```typescript
// Пользователь
interface User {
  id: string;
  email: string;
  role: UserRole;
  status: UserStatus;
  created_at: string;
  updated_at: string;
}

// VPN Сервер
interface Server {
  id: string;
  name: string;
  type: ServerType;
  status: ServerStatus;
  region: string;
  ip: string;
  port: number;
  cpu: number;
  memory: number;
  config: Record<string, string>;
}

// VPN Туннель
interface Tunnel {
  id: string;
  name: string;
  interface: string;
  status: TunnelStatus;
  public_key: string;
  listen_port: number;
  mtu: number;
  auto_recovery: boolean;
}
```

## 🎯 Этапы интеграции

### 1. Настройка HTTP клиента
- Импорт сгенерированных API функций
- Настройка базового URL (http://localhost:8080)
- Добавление JWT токена в заголовки
- Обработка ошибок авторизации

### 2. Управление состоянием
- Настройка React Query для кэширования
- Создание store для глобального состояния (Zustand)
- Синхронизация данных между компонентами

### 3. Создание UI компонентов
- **Дашборд** - главная страница с аналитикой
- **Серверы** - управление VPN серверами
- **Туннели** - настройка WireGuard туннелей
- **Пользователи** - администрирование (для админов)
- **Настройки** - конфигурация DPI и уведомлений

### 4. Обработка данных реального времени
- WebSocket подключения через Gateway
- Обновление метрик в реальном времени
- Уведомления о событиях системы

## 🔍 Отладка и мониторинг

### Логи сервисов
```bash
# Логи всех сервисов
make logs

# Логи конкретного сервиса  
make logs-service SERVICE=auth
make logs-service SERVICE=gateway
```

### Health checks
```bash
# Быстрая проверка
make health-quick

# Подробная проверка
make health

# Статус инфраструктуры
make infra-status
```

### Swagger UI
- Документация доступна на `http://localhost:8080/swagger`
- Unified API схема: `docs/swagger/unified-api.json`

## ⚠️ Важные особенности

1. **Gateway как единая точка входа** - все HTTP запросы идут через порт 8080
2. **gRPC сервисы недоступны напрямую** - только через Gateway
3. **JWT обязателен** для всех эндпоинтов кроме health и login
4. **TypeScript типы автогенерированы** - не редактировать вручную
5. **Docker сети** - все сервисы в единой сети `silence_network`
6. **Порты строго зафиксированы** - не изменять без обновления конфигураций

---

**Дата создания:** 11 июля 2025  
**Версия API:** 1.0.0  
**Статус:** Готово к интеграции