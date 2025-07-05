# Гайд-лайн для разработчиков - Silence VPN

## Обзор проекта

Silence VPN — это микросервисная система для защищённого VPN с обфускацией трафика. Проект состоит из множества взаимосвязанных сервисов, каждый из которых отвечает за определённую функциональность.

## Архитектура системы

### Основные сервисы

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Gateway   │    │    Auth     │    │  VPN Core   │    │ DPI Bypass  │
│   :8080     │◄──►│   :8081     │    │   :8082     │◄──►│   :8083     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Прокси    │    │   JWT Auth  │    │ WireGuard   │    │ Обфускация  │
│   запросов  │    │   + DB      │    │   туннели   │    │   трафика   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Analytics   │    │ Notifications│   │Server Manager│   │   Shared    │
│   :8084     │    │   :8086     │    │   :8085     │    │  Libraries  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Связи между сервисами

1. **Gateway (8080)** — единая точка входа

   - Проксирует запросы ко всем сервисам
   - Обрабатывает JWT аутентификацию
   - Интегрирует VPN + обфускацию

2. **Auth (8081)** — аутентификация

   - Регистрация/вход пользователей
   - Генерация JWT токенов
   - Хранение в PostgreSQL

3. **VPN Core (8082)** — WireGuard туннели

   - Создание/управление туннелями
   - Мониторинг статистики
   - Интеграция с WireGuard

4. **DPI Bypass (8083)** — обфускация трафика

   - Shadowsocks, V2Ray, obfs4
   - Создание bypass-конфигураций
   - Статистика обфускации

5. **Analytics (8084)** — метрики и мониторинг

   - Сбор метрик от всех сервисов
   - Дашборды и алерты
   - Хранение в InfluxDB

6. **Notifications (8086)** — уведомления

   - Email, SMS, push, telegram, slack
   - Интеграция с RabbitMQ
   - Отправка метрик в Analytics

7. **Server Manager (8085)** — управление серверами
   - Создание/управление серверами
   - Масштабирование
   - Мониторинг здоровья

## Начало работы

### 1. Подготовка окружения

```bash
# Клонирование репозитория
git clone <repository-url>
cd silence

# Установка зависимостей
go mod download

# Установка Task (task runner)
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Проверка установки
task --version
```

### 2. Запуск зависимостей

```bash
# Запуск всех сервисов через Docker Compose
docker-compose up -d

# Проверка статуса
docker-compose ps
```

### 3. Сборка сервисов

```bash
# Сборка всех сервисов
task build:all

# Или отдельных сервисов
task build:auth
task build:gateway
task build:vpn-core
task build:dpi-bypass
task build:analytics
task build:notifications
task build:server-manager
```

## Структура проекта

### Основные директории

```
silence/
├── api/                    # API Gateway и Auth
│   ├── gateway/           # API Gateway (порт 8080)
│   └── auth/              # Auth Service (порт 8081)
├── rpc/                   # Микросервисы
│   ├── vpn-core/          # VPN Core (порт 8082)
│   ├── dpi-bypass/        # DPI Bypass (порт 8083)
│   ├── analytics/         # Analytics (порт 8084)
│   ├── server-manager/    # Server Manager (порт 8085)
│   └── notifications/     # Notifications (порт 8086)
├── shared/                # Общие библиотеки
├── docs/                  # Документация
├── deployments/           # Конфигурации развертывания
└── scripts/               # Скрипты
```

### Архитектура каждого сервиса

```
service/
├── cmd/
│   └── main.go           # Точка входа
├── internal/
│   ├── adapters/         # Внешние адаптеры (HTTP, DB, etc.)
│   ├── app/              # Приложение (DI, конфигурация)
│   ├── config/           # Конфигурация
│   ├── domain/           # Доменные модели
│   ├── ports/            # Интерфейсы (ports)
│   ├── services/         # Бизнес-логика
│   └── types/            # Типы данных
├── etc/                  # Конфигурационные файлы
├── go.mod
└── go.sum
```

## Процесс разработки новой фичи

### 1. Анализ требований

Перед началом разработки определите:

- **Какой сервис** отвечает за фичу?
- **Какие API** нужно создать/изменить?
- **Какие связи** с другими сервисами?
- **Какие данные** нужно хранить?

### 2. Создание ветки

```bash
# Создание feature ветки
git checkout -b feature/your-feature-name

# Следование naming convention
# feature/add-user-roles
# feature/vpn-auto-recovery
# bugfix/fix-auth-token
```

### 3. Разработка

#### Шаг 1: Доменные модели

Создайте модели в `internal/domain/`:

```go
// internal/domain/your_feature.go
package domain

type YourFeature struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type YourFeatureRepository interface {
    Create(feature *YourFeature) error
    GetByID(id string) (*YourFeature, error)
    Update(feature *YourFeature) error
    Delete(id string) error
    List() ([]*YourFeature, error)
}
```

#### Шаг 2: Интерфейсы (Ports)

Создайте интерфейсы в `internal/ports/`:

```go
// internal/ports/your_feature.go
package ports

import "your-service/internal/domain"

type YourFeatureService interface {
    CreateFeature(name string) (*domain.YourFeature, error)
    GetFeature(id string) (*domain.YourFeature, error)
    UpdateFeature(id, name string) (*domain.YourFeature, error)
    DeleteFeature(id string) error
    ListFeatures() ([]*domain.YourFeature, error)
}
```

#### Шаг 3: Сервис (Business Logic)

Создайте сервис в `internal/services/`:

```go
// internal/services/your_feature.go
package services

import (
    "your-service/internal/domain"
    "your-service/internal/ports"
)

type yourFeatureService struct {
    repo domain.YourFeatureRepository
}

func NewYourFeatureService(repo domain.YourFeatureRepository) ports.YourFeatureService {
    return &yourFeatureService{repo: repo}
}

func (s *yourFeatureService) CreateFeature(name string) (*domain.YourFeature, error) {
    feature := &domain.YourFeature{
        ID:        generateID(),
        Name:      name,
        Status:    "active",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    err := s.repo.Create(feature)
    if err != nil {
        return nil, err
    }

    return feature, nil
}
```

#### Шаг 4: HTTP Handlers

Создайте handlers в `internal/adapters/http/`:

```go
// internal/adapters/http/your_feature_handlers.go
package http

import (
    "encoding/json"
    "net/http"
    "your-service/internal/ports"
)

type yourFeatureHandler struct {
    service ports.YourFeatureService
}

func NewYourFeatureHandler(service ports.YourFeatureService) *yourFeatureHandler {
    return &yourFeatureHandler{service: service}
}

func (h *yourFeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name string `json:"name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    feature, err := h.service.CreateFeature(req.Name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(feature)
}
```

#### Шаг 5: Интеграция в main.go

Добавьте в `cmd/main.go`:

```go
// cmd/main.go
func main() {
    // ... existing code ...

    // Your Feature
    yourFeatureRepo := database.NewYourFeatureRepository(db)
    yourFeatureService := services.NewYourFeatureService(yourFeatureRepo)
    yourFeatureHandler := http.NewYourFeatureHandler(yourFeatureService)

    // Routes
    router.HandleFunc("/api/v1/your-features", yourFeatureHandler.CreateFeature).Methods("POST")
    router.HandleFunc("/api/v1/your-features/{id}", yourFeatureHandler.GetFeature).Methods("GET")

    // ... existing code ...
}
```

### 4. Тестирование

#### Unit тесты

```go
// internal/services/your_feature_test.go
package services

import (
    "testing"
    "your-service/internal/domain"
)

func TestYourFeatureService_CreateFeature(t *testing.T) {
    // Arrange
    mockRepo := &MockYourFeatureRepository{}
    service := NewYourFeatureService(mockRepo)

    // Act
    feature, err := service.CreateFeature("test-feature")

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if feature.Name != "test-feature" {
        t.Errorf("Expected name 'test-feature', got %s", feature.Name)
    }
}
```

#### Интеграционные тесты

```bash
# Запуск тестов
go test ./...

# Запуск с coverage
go test -cover ./...

# Запуск конкретного теста
go test -v ./internal/services -run TestYourFeatureService
```

### 5. Интеграция с Gateway

Если фича требует доступа через Gateway, добавьте проксирование:

```go
// api/gateway/internal/adapters/http/handlers.go
func (h *Handler) ProxyYourFeature(w http.ResponseWriter, r *http.Request) {
    // Проксирование на соответствующий сервис
    targetURL := h.config.YourFeatureURL + r.URL.Path
    h.proxyRequest(w, r, targetURL)
}
```

### 6. Документация

Обновите документацию:

```bash
# API endpoints
docs/api-endpoints.md

# Если нужно, создайте новую документацию
docs/your-feature-guide.md
```

## Интеграция между сервисами

### 1. HTTP API

Основной способ коммуникации — HTTP REST API:

```go
// Пример вызова другого сервиса
func (s *service) CallOtherService() error {
    resp, err := http.Get("http://other-service:port/api/endpoint")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Обработка ответа
    return nil
}
```

### 2. RabbitMQ (для уведомлений)

Notifications сервис использует RabbitMQ для получения событий:

```go
// Отправка события в RabbitMQ
func (s *service) SendNotification(event *domain.NotificationEvent) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return s.rabbitMQ.Publish("notifications", "notifications.*", body)
}
```

### 3. InfluxDB (для метрик)

Analytics сервис собирает метрики через HTTP API:

```go
// Отправка метрики в Analytics
func (s *service) SendMetric(metric *domain.Metric) error {
    body, err := json.Marshal(metric)
    if err != nil {
        return err
    }

    resp, err := http.Post("http://analytics:8084/metrics", "application/json", bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### 4. PostgreSQL (для данных)

Auth и Server Manager используют PostgreSQL:

```go
// Пример работы с базой данных
func (r *repository) Create(user *domain.User) error {
    query := `INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)`
    _, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
    return err
}
```

## Конфигурация

### Переменные окружения

Каждый сервис использует переменные окружения для конфигурации:

```bash
# Общие переменные
LOG_LEVEL=info
VERSION=1.0.0

# Специфичные для сервиса
HTTP_PORT=:8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=service_name
DB_USER=postgres
DB_PASSWORD=password

# URL других сервисов
AUTH_URL=http://localhost:8081
VPN_CORE_URL=http://localhost:8082
ANALYTICS_URL=http://localhost:8084
```

### Конфигурационные файлы

```yaml
# etc/config.yaml
server:
  port: 8080
  host: localhost

database:
  host: localhost
  port: 5432
  name: service_name
  user: postgres
  password: password

services:
  auth: http://localhost:8081
  vpn_core: http://localhost:8082
  analytics: http://localhost:8084
```

## Логирование

Все сервисы используют структурированное логирование:

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("Service started",
    zap.String("service", "your-service"),
    zap.String("version", "1.0.0"),
    zap.Int("port", 8080),
)
```

## Мониторинг

### Health Checks

Каждый сервис предоставляет health check endpoint:

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

### Метрики

Analytics сервис собирает метрики от всех сервисов:

```bash
# Метрики подключений
curl http://localhost:8084/metrics/connections

# Нагрузка серверов
curl http://localhost:8084/metrics/server-load
```

## Безопасность

### JWT Аутентификация

Большинство API требуют JWT токен:

```go
// Middleware для проверки JWT
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Проверка токена
        claims, err := validateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Добавление claims в контекст
        ctx := context.WithValue(r.Context(), "user", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Rate Limiting

Gateway использует rate limiting:

```go
// Rate limiting middleware
func RateLimitMiddleware(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Every(time.Second), 10)

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too many requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Отладка

### Логи

```bash
# Просмотр логов сервиса
tail -f logs/service.log

# Логи Docker контейнера
docker logs -f container_name
```

### Отладка API

```bash
# Тестирование API
curl -X POST http://localhost:8080/api/v1/endpoint \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{"key": "value"}'
```

### Отладка базы данных

```bash
# Подключение к PostgreSQL
psql -h localhost -U postgres -d service_name

# Просмотр таблиц
\dt

# Выполнение запроса
SELECT * FROM users LIMIT 10;
```

## Развертывание

### Docker

```bash
# Сборка образа
docker build -t silence/your-service .

# Запуск контейнера
docker run -p 8080:8080 silence/your-service
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  your-service:
    build: ./rpc/your-service
    ports:
      - '8080:8080'
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
    depends_on:
      - postgres
```

### Kubernetes

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: your-service
  template:
    metadata:
      labels:
        app: your-service
    spec:
      containers:
        - name: your-service
          image: silence/your-service:latest
          ports:
            - containerPort: 8080
```

## Лучшие практики

### 1. Код

- Следуйте принципам Clean Architecture
- Используйте интерфейсы для абстракции
- Пишите тесты для бизнес-логики
- Документируйте публичные API

### 2. Безопасность

- Всегда валидируйте входные данные
- Используйте prepared statements для SQL
- Храните секреты в переменных окружения
- Логируйте подозрительную активность

### 3. Производительность

- Используйте connection pooling для БД
- Кэшируйте часто используемые данные
- Оптимизируйте запросы к базе данных
- Мониторьте производительность

### 4. Мониторинг

- Добавляйте метрики для важных операций
- Логируйте ошибки с контекстом
- Настройте алерты для критических событий
- Отслеживайте время ответа API

## Полезные команды

### Разработка

```bash
# Запуск в режиме разработки
task dev:all

# Сборка конкретного сервиса
task build:auth

# Запуск тестов
task test:all

# Проверка кода
task lint:all
```

### Отладка

```bash
# Просмотр логов
task logs:all

# Проверка здоровья сервисов
task health:check

# Тестирование API
task test:api
```

### Развертывание

```bash
# Сборка всех образов
task docker:build

# Запуск в Docker
task docker:up

# Развертывание в Kubernetes
task k8s:deploy
```

## Получение помощи

### Документация

- `docs/api-endpoints.md` — API документация
- `docs/frontend-quickstart.md` — гайд для фронтенда
- `docs/analytics-integration.md` — интеграция аналитики
- `docs/wireguard-integration.md` — интеграция WireGuard

### Логи и отладка

- Проверьте логи сервиса
- Используйте health check endpoints
- Тестируйте API через curl
- Проверьте конфигурацию

### Команда

- Создайте issue в репозитории
- Опишите проблему с контекстом
- Приложите логи и конфигурацию
- Предложите возможное решение

## Заключение

Этот гайд поможет вам быстро интегрироваться в проект Silence VPN. Помните о принципах микросервисной архитектуры и следуйте установленным практикам. При возникновении вопросов обращайтесь к документации и команде.
