# Silence VPN Platform

[![CI/CD Pipeline](https://github.com/par1ram/silence/actions/workflows/ci.yml/badge.svg)](https://github.com/par1ram/silence/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/par1ram/silence)](https://goreportcard.com/report/github.com/par1ram/silence)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/docker/pulls/silence/gateway.svg)](https://hub.docker.com/r/silence/gateway)

**Silence VPN** - это современная, высокопроизводительная VPN платформа, построенная на основе микросервисной архитектуры. Платформа предоставляет полный стек решений для VPN сервисов, включая обход DPI блокировок, управление серверами, аналитику и мониторинг.

## 🚀 Особенности

- **Микросервисная архитектура** - масштабируемая и надежная система
- **Обход DPI блокировок** - продвинутые алгоритмы обхода блокировок
- **Автоматическое управление серверами** - Docker-based масштабирование
- **Реалтайм аналитика** - детальная статистика и мониторинг
- **Современный веб-интерфейс** - React + Next.js фронтенд
- **Мультипротокольная поддержка** - WireGuard, OpenVPN, IKEv2
- **Kubernetes ready** - готов для продакшн развертывания
- **Observability** - полный стек мониторинга с Prometheus, Grafana, Jaeger

## 📋 Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    Frontend     │    │   API Gateway   │    │  Auth Service   │
│   (Next.js)     │◄──►│   (Go + Gin)    │◄──►│   (Go + gRPC)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
     ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
     │  VPN Core       │ │  DPI Bypass     │ │  Server Manager │
     │  (Go + gRPC)    │ │  (Go + gRPC)    │ │  (Go + gRPC)    │
     └─────────────────┘ └─────────────────┘ └─────────────────┘
                │               │               │
     ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
     │   Analytics     │ │  Notifications  │ │   PostgreSQL    │
     │  (Go + gRPC)    │ │  (Go + gRPC)    │ │     Redis       │
     └─────────────────┘ └─────────────────┘ └─────────────────┘
```

## 🛠️ Технологический стек

### Backend
- **Go 1.21+** - основной язык разработки
- **gRPC** - межсервисное взаимодействие
- **PostgreSQL** - основная база данных
- **Redis** - кэширование и сессии
- **RabbitMQ** - асинхронная обработка сообщений
- **InfluxDB** - метрики и аналитика
- **ClickHouse** - большие данные и аналитика

### Frontend
- **Next.js 15** - React фреймворк
- **TypeScript** - типизация
- **Tailwind CSS** - стилизация
- **Zustand** - управление состоянием
- **React Query** - работа с API
- **Axios** - HTTP клиент

### DevOps & Monitoring
- **Docker** - контейнеризация
- **Kubernetes** - оркестрация
- **Prometheus** - метрики
- **Grafana** - дашборды
- **Jaeger** - распределенный трейсинг
- **Loki** - логирование

## 🚀 Быстрый старт

### Предварительные требования

- **Docker** и **Docker Compose**
- **Go 1.21+**
- **Node.js 18+**
- **Make**
- **Git**

### Установка

1. **Клонирование репозитория**
   ```bash
   git clone https://github.com/par1ram/silence.git
   cd silence
   ```

2. **Запуск инфраструктуры**
   ```bash
   make infra-up
   ```

3. **Генерация API документации**
   ```bash
   make generate-client-sdk
   ```

4. **Запуск всех сервисов**
   ```bash
   make dev-all
   ```

5. **Установка зависимостей фронтенда**
   ```bash
   make setup-frontend
   ```

6. **Запуск фронтенда**
   ```bash
   make frontend-dev
   ```

### Доступ к сервисам

- **Веб-интерфейс**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger
- **Grafana**: http://localhost:3000 (в observability стеке)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686

## 📖 Документация

### Основная документация
- [Архитектура системы](docs/ARCHITECTURE.md)
- [API документация](docs/API.md)
- [Руководство разработчика](docs/DEVELOPER_GUIDE.md)
- [Руководство по развертыванию](docs/DEPLOYMENT_GUIDE.md)

### Сервисы
- [Auth Service](api/auth/README.md)
- [API Gateway](api/gateway/README.md)
- [VPN Core](rpc/vpn-core/README.md)
- [DPI Bypass](rpc/dpi-bypass/README.md)
- [Server Manager](docs/SERVER_MANAGER_GUIDE.md)
- [Analytics](rpc/analytics/README.md)
- [Notifications](rpc/notifications/README.md)

### Дополнительно
- [Мультиплатформенные приложения](docs/MULTIPLATFORM_PLAN.md)
- [Kubernetes развертывание](deployments/k8s/README.md)
- [Observability](deployments/observability/README.md)

## 🔧 Разработка

### Makefile команды

```bash
# Разработка
make dev-all              # Запуск всех сервисов
make dev-single          # Запуск отдельного сервиса
make setup-frontend      # Настройка фронтенда

# Сборка
make build               # Сборка всех сервисов
make build-frontend      # Сборка фронтенда
make docker-build        # Сборка Docker образов

# Тестирование
make test                # Запуск всех тестов
make test-integration    # Интеграционные тесты
make lint                # Линтинг кода

# API
make swagger             # Генерация Swagger документации
make generate-client-sdk # Генерация клиентского SDK

# Инфраструктура
make infra-up            # Запуск инфраструктуры
make infra-down          # Остановка инфраструктуры
make status              # Статус сервисов
```

### Структура проекта

```
silence/
├── api/                    # HTTP API сервисы
│   ├── auth/              # Сервис авторизации
│   └── gateway/           # API Gateway
├── rpc/                   # gRPC сервисы
│   ├── analytics/         # Аналитика
│   ├── dpi-bypass/        # Обход DPI
│   ├── notifications/     # Уведомления
│   ├── server-manager/    # Управление серверами
│   └── vpn-core/          # VPN функциональность
├── shared/                # Общие библиотеки
├── frontend/              # Веб-интерфейс
│   ├── src/
│   │   ├── components/    # React компоненты
│   │   ├── pages/         # Страницы
│   │   ├── generated/     # Сгенерированные API хуки
│   │   └── lib/           # Утилиты
├── deployments/           # Конфигурации развертывания
│   ├── k8s/              # Kubernetes манифесты
│   └── observability/     # Мониторинг
├── docs/                  # Документация
└── scripts/               # Скрипты автоматизации
```

### Генерация API клиента

Проект использует автоматическую генерацию TypeScript клиента из Swagger документации:

```bash
# Генерация из proto файлов
make swagger

# Генерация React Query хуков
make generate-client-sdk

# Или в папке frontend
cd frontend
npm run generate:api
```

## 🐳 Docker развертывание

### Локальная разработка

```bash
# Запуск с Docker Compose
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

### Production сборка

```bash
# Сборка всех образов
make docker-build

# Запуск в production режиме
make dev-production
```

## ☸️ Kubernetes развертывание

### Single-server deployment

```bash
# Применение всех манифестов
kubectl apply -f deployments/k8s/single-server/

# Проверка статуса
kubectl get pods -n silence

# Просмотр логов
kubectl logs -f deployment/gateway -n silence
```

### Конфигурация

1. **Создание namespace**
   ```bash
   kubectl create namespace silence
   ```

2. **Настройка секретов**
   ```bash
   kubectl create secret generic silence-secrets \
     --from-literal=POSTGRES_PASSWORD=your-password \
     --from-literal=JWT_SECRET=your-jwt-secret \
     -n silence
   ```

3. **Применение конфигурации**
   ```bash
   kubectl apply -f deployments/k8s/single-server/
   ```

## 📊 Мониторинг

### Observability Stack

Проект включает полный стек мониторинга:

- **Prometheus** - сбор метрик
- **Grafana** - дашборды и визуализация
- **Jaeger** - распределенный трейсинг
- **Loki** - централизованное логирование
- **OpenTelemetry** - единый коллектор телеметрии

### Запуск мониторинга

```bash
cd deployments/observability
docker-compose up -d
```

### Доступ к интерфейсам

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **Loki**: http://localhost:3100

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Конкретный сервис
make test-auth
make test-gateway

# Интеграционные тесты
make test-integration

# Фронтенд тесты
make test-frontend
```

### Покрытие кода

```bash
# Генерация отчета о покрытии
make coverage

# Просмотр в браузере
open coverage.html
```

## 🚀 CI/CD

Проект использует GitHub Actions для автоматизации:

- **Линтинг** - проверка качества кода
- **Тестирование** - unit и integration тесты
- **Сборка** - Docker образы
- **Развертывание** - автоматический деплой в staging/production

### Переменные окружения

Для CI/CD необходимо настроить следующие секреты в GitHub:

```bash
KUBE_CONFIG_STAGING     # Kubeconfig для staging
KUBE_CONFIG_PRODUCTION  # Kubeconfig для production
DOCKER_REGISTRY_TOKEN   # Токен для Docker registry
```

## 🔐 Безопасность

### Основные принципы

- **JWT токены** для авторизации
- **TLS everywhere** - шифрование всех соединений
- **Rate limiting** - защита от злоупотреблений
- **Input validation** - валидация всех входных данных
- **CORS** - правильная настройка CORS политик

### Конфигурация безопасности

```yaml
# Пример конфигурации в Kubernetes
apiVersion: v1
kind: Secret
metadata:
  name: silence-secrets
type: Opaque
data:
  JWT_SECRET: <base64-encoded-secret>
  POSTGRES_PASSWORD: <base64-encoded-password>
```

## 🌐 Мультиплатформенные приложения

Проект включает планы для разработки приложений под различные платформы:

- **Desktop** - Electron приложения (Windows, macOS, Linux)
- **Mobile** - React Native приложения (iOS, Android)
- **CLI** - Консольные утилиты

Подробности в [документации](docs/MULTIPLATFORM_PLAN.md).

## 📈 Производительность

### Основные метрики

- **Latency** - < 50ms для API запросов
- **Throughput** - > 1000 RPS на сервис
- **Availability** - 99.9% uptime
- **Scalability** - горизонтальное масштабирование

### Мониторинг производительности

```bash
# Проверка нагрузки
kubectl top pods -n silence

# Метрики в Prometheus
curl http://localhost:9090/api/v1/query?query=silence_request_duration_seconds
```

## 🤝 Участие в разработке

### Внесение изменений

1. **Fork** репозитория
2. **Создайте** feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** изменения (`git commit -m 'Add amazing feature'`)
4. **Push** в branch (`git push origin feature/amazing-feature`)
5. **Создайте** Pull Request

### Стандарты кода

- **Go** - следуйте gofmt и golint
- **TypeScript** - используйте ESLint и Prettier
- **Commits** - используйте conventional commits
- **Tests** - покрытие > 80%

### Линтинг

```bash
# Go сервисы
make lint

# Frontend
make frontend-lint

# Исправление автоматически
make lint-fix
```

## 📄 Лицензия

Этот проект лицензирован под [MIT License](LICENSE).

## 🙏 Благодарности

- [Go](https://golang.org/) - за отличный язык программирования
- [Next.js](https://nextjs.org/) - за потрясающий React фреймворк
- [Docker](https://docker.com/) - за упрощение разработки
- [Kubernetes](https://kubernetes.io/) - за оркестрацию контейнеров
- [OpenTelemetry](https://opentelemetry.io/) - за observability

## 📞 Поддержка

- **GitHub Issues** - для багов и feature requests
- **GitHub Discussions** - для общих вопросов
- **Email** - support@silence-vpn.com
- **Documentation** - подробная документация в папке `docs/`

## 🔗 Ссылки

- [Официальный сайт](https://silence-vpn.com)
- [Документация](https://docs.silence-vpn.com)
- [API Reference](https://api.silence-vpn.com/docs)
- [Status Page](https://status.silence-vpn.com)

---

**Silence VPN** - надежное и безопасное VPN решение для современного мира. 🚀