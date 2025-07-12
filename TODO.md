# TODO: Комплексный план развития проекта Silence

## 🏗️ Инфраструктура и DevOps
- [x] Упростить observability stack (убрать дублирование)
  - [x] Оставить только: Prometheus (метрики), Loki (логи), Jaeger (трейсинг)
  - [x] Убрать: Zipkin, Tempo, дублирующий Redis
  - [x] Настроить единый OpenTelemetry Collector
- [x] Настроить Kubernetes для single-server deployment
  - [x] Создать манифесты для всех сервисов
  - [x] Настроить ingress controller
  - [x] Настроить persistent volumes
- [x] Доработать CI/CD pipeline
  - [x] Настроить GitHub Actions
  - [x] Добавить автоматическое тестирование
  - [x] Настроить автодеплой на сервер
- [x] Убрать air из всех сервисов
- [x] Настроить запуск всех сервисов через Makefile

## 🔧 Бэкенд сервисы
- [x] Пофиксить все ошибки компиляции
- [x] Настроить правильные порты для всех сервисов
- [x] Проверить работу всех Docker контейнеров
- [x] Настроить межсервисное взаимодействие
- [x] Добавить health checks для всех сервисов
- [ ] Настроить миграции баз данных
- [ ] Добавить логирование и мониторинг
- [x] Исправить проблемы с диагностикой проекта

## 📱 Фронтенд и API интеграция
- [x] Генерация Swagger документации из proto файлов
  - [x] Настроить protoc-gen-openapiv2
  - [x] Добавить HTTP аннотации во все proto файлы (71 эндпоинт)
  - [x] Создать unified API схему (201 определение типов)
  - [x] Настроить автогенерацию клиентского SDK
  - [x] Создать скрипт для объединения swagger файлов
  - [x] Исправить проблемы диагностики проекта
- [x] Интеграция фронтенда с бэкендом
  - [x] Настроить API клиент на основе Swagger
  - [x] Добавить авторизацию (JWT токены)
  - [x] Настроить управление состоянием (Zustand)
  - [x] Добавить React Query для API запросов
  - [x] Создать утилиты для работы с API
  - [x] Сгенерировать типобезопасные API хуки
- [ ] Исправить ошибки компиляции фронтенда
  - [x] Исправить проверки ролей пользователей (USER_ROLE_ADMIN, USER_ROLE_MODERATOR)
  - [x] Убрать несуществующее поле username из authUser
  - [ ] Заменить старые API вызовы на сгенерированные из Swagger
  - [ ] Исправить типы в компонентах админ панели
- [ ] Компоненты UI (используйте frontend/src/generated/requests для API)
  - [ ] Дашборд для управления серверами
    - [ ] Список серверов (GET /api/v1/servers)
    - [ ] Статистика серверов (GET /api/v1/servers/{id}/stats)
    - [ ] Управление серверами (POST /api/v1/servers/{id}/start|stop|restart)
    - [ ] Создание/редактирование серверов (POST/PUT /api/v1/servers)
  - [ ] Страница аналитики
    - [ ] Дашборд аналитики (GET /api/v1/analytics/dashboard)
    - [ ] Метрики системы (GET /api/v1/analytics/statistics/system)
    - [ ] История метрик (GET /api/v1/analytics/metrics/history)
    - [ ] Статистика пользователей (GET /api/v1/analytics/statistics/users/{user_id})
  - [ ] Настройки VPN
    - [ ] Управление туннелями (GET/POST /api/v1/vpn/tunnels)
    - [ ] Настройка пиров (GET/POST /api/v1/vpn/tunnels/{id}/peers)
    - [ ] Статистика туннелей (GET /api/v1/vpn/tunnels/{id}/stats)
    - [ ] Настройки DPI (GET/POST /api/v1/dpi/configs)
  - [ ] Управление уведомлениями
    - [ ] Список уведомлений (GET /api/v1/notifications)
    - [ ] Шаблоны уведомлений (GET/POST /api/v1/notifications/templates)
    - [ ] Настройки пользователя (GET/PUT /api/v1/notifications/preferences/{user_id})
  - [ ] Аутентификация и пользователи
    - [ ] Форма входа (POST /api/v1/auth/login)
    - [ ] Регистрация (POST /api/v1/auth/register)
    - [ ] Профиль пользователя (GET /api/v1/auth/me)
    - [ ] Управление пользователями (GET/POST/PUT/DELETE /api/v1/users)
- [x] Интеграция с сгенерированными API хуками
  - [x] Обновить существующие компоненты
  - [x] Заменить ручные API вызовы на сгенерированные хуки
  - [x] Добавить типобезопасность
  - [x] Настроить error handling
  - [x] Исправить TypeScript ошибки в сгенерированных типах
  - [x] Создать документацию по API генерации
  - [x] Создать пошаговый план интеграции (docs/FRONTEND_INTEGRATION_PLAN.md)
  - [x] Создать промт для AI-ассистента (docs/AI_INTEGRATION_PROMPT.md)
  - [ ] **ТЕКУЩАЯ ЗАДАЧА:** Интеграция фронтенда с готовыми API хуками
    - [ ] Настроить HTTP клиент для работы с Gateway (localhost:8080)
    - [ ] Заменить AuthService на AuthServiceService из generated/requests
    - [ ] Заменить serverService на ServerManagerServiceService
    - [ ] Заменить vpnService на VpnCoreServiceService
    - [ ] Заменить analyticsService на AnalyticsServiceService
    - [ ] Заменить notificationService на NotificationsServiceService
    - [ ] Создать новые React Query хуки в src/hooks/
    - [ ] Обновить компоненты для использования новых хуков
    - [ ] Добавить обработку ошибок авторизации (401/403)
    - [ ] Настроить JWT токен в заголовки Authorization
    - [ ] Протестировать все функции (логин, серверы, VPN, аналитика)

## 🖥️ Мультиплатформенные приложения
- [x] Core функционал (общий для всех платформ)
  - [x] VPN подключение и управление
  - [x] Обход DPI блокировок
  - [x] Мониторинг соединения
  - [x] Синхронизация настроек
- [x] Desktop приложения
  - [x] Electron приложение (Windows, Mac, Linux)
  - [x] Системный трей
  - [x] Автозапуск
  - [x] Нативные уведомления
- [x] Mobile приложения
  - [x] React Native или Flutter
  - [x] iOS: WireGuard integration
  - [x] Android: VPN Service
  - [x] Push уведомления
- [x] CLI утилита
  - [x] Управление через командную строку
  - [x] Автоматизация и скрипты

## 🗄️ Архитектура данных
- [ ] Унифицировать схемы баз данных
- [ ] Настроить миграции
- [ ] Оптимизировать запросы
- [ ] Добавить кэширование (Redis)
- [ ] Настроить бэкапы

## 📊 Сервис Server Manager
- [x] Документация по работе с сервисом
  - [x] API методы для управления серверами
  - [x] Схема создания/удаления серверов
  - [x] Мониторинг состояния серверов
- [x] Интеграция с Docker
- [ ] Автоматическое масштабирование
- [ ] Backup и восстановление

## 🔐 Безопасность
- [ ] Настроить HTTPS для всех сервисов
- [ ] Добавить rate limiting
- [ ] Настроить CORS
- [ ] Валидация входных данных
- [ ] Аудит безопасности

## 📈 Мониторинг и аналитика
- [ ] Настроить сбор метрик
- [ ] Создать дашборды в Grafana
- [ ] Настроить алерты
- [ ] Логирование пользовательских действий
- [ ] Аналитика использования VPN

## 🧪 Тестирование
- [ ] Unit тесты для всех сервисов
- [ ] Integration тесты
- [ ] E2E тесты для фронтенда
- [ ] Load testing
- [ ] Security testing

## 📚 Документация
- [ ] README для каждого сервиса
- [x] API документация (docs/swagger/unified-api.json)
- [x] Справочник интеграции фронтенда (docs/FRONTEND_INTEGRATION_GUIDE.md)
- [x] Пошаговый план интеграции (docs/FRONTEND_INTEGRATION_PLAN.md)
- [x] Промт для AI-ассистента (docs/AI_INTEGRATION_PROMPT.md)
- [x] Отчет о генерации API (docs/API_GENERATION_REPORT.md)
- [x] Отчет о решении диагностики (docs/DIAGNOSTIC_RESOLUTION.md)
- [x] Быстрый старт для AI (docs/QUICK_START_FOR_AI.md)
- [ ] Deployment guide
- [ ] User manual
- [ ] Developer guide

## 🚀 Развертывание
- [ ] Развертывание
- [ ] Single-server deployment
- [x] Docker Compose для разработки
- [ ] Kubernetes для продакшена
- [ ] CDN для статики
- [ ] Мониторинг в продакшене
- [x] Исправить проблемы с Docker сетями

## 🔄 Оптимизация
- [ ] Профилирование производительности
- [ ] Оптимизация Docker образов
- [ ] Кэширование на разных уровнях
- [ ] Асинхронная обработка задач
- [ ] Оптимизация сетевого трафика

## 🎯 Приоритетные задачи для фронтенда (для AI-ассистента)

### 📋 Готовые материалы для интеграции:
- **Пошаговый план**: `docs/FRONTEND_INTEGRATION_PLAN.md`
- **Промт для AI**: `docs/AI_INTEGRATION_PROMPT.md`
- **Справочник API**: `docs/FRONTEND_INTEGRATION_GUIDE.md`
- **Сгенерированные хуки**: `frontend/src/generated/requests/`

### 🚀 Этапы интеграции:
1. **Изучить документацию** - прочитать FRONTEND_INTEGRATION_PLAN.md
2. **Настроить HTTP клиент** - создать src/lib/api-client.ts
3. **Заменить AuthService** - обновить src/stores/auth.ts
4. **Заменить остальные сервисы** - serverService, vpnService, analyticsService
5. **Создать React Query хуки** - в src/hooks/
6. **Обновить компоненты** - использовать новые хуки
7. **Протестировать** - проверить все функции

## 📋 Справочная информация для AI-ассистента
- **API Gateway:** localhost:8080 (единая точка входа)
- **Swagger UI:** http://localhost:8080/swagger
- **Главная документация:** docs/FRONTEND_INTEGRATION_PLAN.md
- **Промт для AI:** docs/AI_INTEGRATION_PROMPT.md
- **Сгенерированные API:** frontend/src/generated/requests/
- **Команды:** make swagger, make generate-client-sdk, make frontend-dev

## 🎯 Следующий шаг
**Для новой AI-нейронки:** Прочитай `docs/AI_INTEGRATION_PROMPT.md` и следуй пошаговому плану интеграции фронтенда с готовыми API хуками!
