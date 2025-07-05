# CI/CD Guide - Silence VPN

## Обзор

Проект Silence VPN использует современную CI/CD систему на основе GitHub Actions для автоматизации процессов разработки, тестирования и развертывания.

## Архитектура CI/CD

### Workflows

1. **ci.yml** - Основной CI/CD pipeline
2. **security.yml** - Проверки безопасности
3. **release.yml** - Автоматические релизы

### Этапы Pipeline

```
Push/PR → Lint → Test → Build → Docker Build → Deploy
```

## Настройка

### Предварительные требования

1. **GitHub Repository** с настроенными secrets
2. **Kubernetes Cluster** для деплоя
3. **Docker Registry** (GitHub Container Registry)
4. **Helm** для управления Kubernetes

### GitHub Secrets

Настройте следующие secrets в вашем GitHub repository:

```bash
# Для доступа к Docker Registry
GITHUB_TOKEN (автоматически доступен)

# Для деплоя в Kubernetes (опционально)
KUBECONFIG_BASE64
KUBE_CONFIG_DATA

# Для уведомлений (опционально)
SLACK_WEBHOOK
DISCORD_WEBHOOK
```

## Workflows

### 1. Основной CI/CD Pipeline (ci.yml)

**Триггеры:**

- Push в `main` и `develop` ветки
- Pull Request в `main` и `develop` ветки

**Jobs:**

#### Lint

- Проверка кода с помощью `golangci-lint`
- Проверка всех сервисов параллельно
- Выход из pipeline при ошибках

#### Test

- Unit тесты для всех сервисов
- Integration тесты
- Запуск тестовых баз данных (PostgreSQL, Redis, InfluxDB)

#### Build

- Сборка бинарных файлов для всех сервисов
- Матричная сборка для разных сервисов
- Загрузка артефактов

#### Docker Build

- Сборка Docker образов
- Публикация в GitHub Container Registry
- Кэширование слоев для ускорения сборки

#### Deploy

- Автоматический деплой в staging (develop ветка)
- Автоматический деплой в production (main ветка)

### 2. Security Scan (security.yml)

**Триггеры:**

- Push в `main` и `develop` ветки
- Pull Request в `main` и `develop` ветки
- Еженедельно по расписанию

**Jobs:**

#### Dependency Check

- `govulncheck` - проверка уязвимостей в зависимостях
- `gosec` - статический анализ безопасности

#### Container Scan

- `Trivy` - сканирование Docker образов
- Проверка уязвимостей в базовых образах

#### Secret Detection

- `TruffleHog` - поиск секретов в коде
- Проверка истории коммитов

#### License Check

- Проверка лицензий зависимостей
- Соответствие политике лицензирования

### 3. Release (release.yml)

**Триггеры:**

- Push тегов вида `v*` (например, `v1.0.0`)

**Jobs:**

#### Create Release

- Автоматическое создание GitHub Release
- Генерация changelog из коммитов
- Создание release notes

#### Build and Push

- Сборка Docker образов с тегом версии
- Публикация в registry
- Множественные теги (latest, major.minor, version)

#### Build Binaries

- Сборка бинарных файлов для разных платформ
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Деплой

### Окружения

1. **Staging** - автоматический деплой из `develop` ветки
2. **Production** - автоматический деплой из `main` ветки

### Стратегия деплоя

- **Blue-Green Deployment** для production
- **Rolling Update** для staging
- Автоматический rollback при ошибках

### Kubernetes

Используется Helm для управления Kubernetes ресурсами:

```bash
# Деплой в staging
./scripts/deploy.sh staging v1.0.0

# Деплой в production
./scripts/deploy.sh production v1.0.0

# Проверка статуса
./scripts/deploy.sh status

# Health check
./scripts/deploy.sh health

# Rollback
./scripts/deploy.sh rollback
```

## Мониторинг

### Метрики

- Prometheus для сбора метрик
- Grafana для визуализации
- ServiceMonitor для автоматического обнаружения сервисов

### Логирование

- Централизованное логирование через ELK stack
- Структурированные логи в JSON формате
- Уровни логирования: DEBUG, INFO, WARN, ERROR

### Алерты

- Prometheus AlertManager
- Уведомления в Slack/Discord
- Email уведомления для критических ошибок

## Безопасность

### Сканирование

- Автоматическое сканирование зависимостей
- Проверка Docker образов
- Поиск секретов в коде
- Анализ уязвимостей

### Секреты

- Kubernetes Secrets для хранения секретов
- Автоматическая генерация паролей
- Ротация секретов

### Сетевая безопасность

- Network Policies в Kubernetes
- Ingress с TLS
- mTLS между сервисами

## Troubleshooting

### Частые проблемы

1. **Build failures**

   ```bash
   # Проверка логов сборки
   task build

   # Очистка кэша
   task clean
   ```

2. **Deployment failures**

   ```bash
   # Проверка статуса подов
   kubectl get pods -n silence-production

   # Просмотр логов
   kubectl logs -f deployment/auth-service -n silence-production
   ```

3. **Security scan failures**

   ```bash
   # Обновление зависимостей
   go mod tidy
   go mod download

   # Проверка уязвимостей
   govulncheck ./...
   ```

### Полезные команды

```bash
# Локальная проверка линтера
task lint

# Локальное тестирование
task test

# Сборка Docker образов
task docker:build

# Деплой в локальный Kubernetes
./scripts/deploy.sh staging latest
```

## Best Practices

### Разработка

1. **Feature Branches** - работа в отдельных ветках
2. **Pull Requests** - обязательный код-ревью
3. **Semantic Versioning** - использование семантического версионирования
4. **Conventional Commits** - структурированные сообщения коммитов

### Тестирование

1. **Unit Tests** - покрытие кода тестами >80%
2. **Integration Tests** - тестирование взаимодействия сервисов
3. **E2E Tests** - тестирование полного пользовательского сценария

### Деплой

1. **Immutable Infrastructure** - неизменяемая инфраструктура
2. **Infrastructure as Code** - управление инфраструктурой через код
3. **Blue-Green Deployment** - безопасное обновление production

### Безопасность

1. **Least Privilege** - минимальные права доступа
2. **Secret Management** - безопасное управление секретами
3. **Regular Updates** - регулярное обновление зависимостей

## Автоматизация

### GitHub Actions

Все процессы автоматизированы через GitHub Actions:

- Автоматическая сборка при push
- Автоматическое тестирование
- Автоматический деплой
- Автоматические релизы

### Скрипты

Дополнительные скрипты для автоматизации:

- `scripts/deploy.sh` - деплой в Kubernetes
- `scripts/test-integration.sh` - интеграционные тесты
- `scripts/backup.sh` - резервное копирование

## Мониторинг и алерты

### Метрики

- CPU и Memory использование
- Latency и throughput
- Error rates
- Custom business metrics

### Алерты

- High CPU/Memory usage
- High error rates
- Service unavailability
- Security incidents

### Дашборды

- Grafana дашборды для каждого сервиса
- Общий дашборд системы
- Business metrics дашборд

## Заключение

CI/CD система Silence VPN обеспечивает:

- Быструю и надежную доставку кода
- Высокое качество и безопасность
- Автоматизацию рутинных задач
- Масштабируемость и отказоустойчивость

Для получения дополнительной информации обратитесь к документации конкретных компонентов или создайте issue в GitHub repository.
