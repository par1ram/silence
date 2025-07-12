# Быстрый старт для AI-ассистента - Проект Silence VPN

## 🚀 Что нужно знать сразу

### Статус проекта
- ✅ Бэкенд: 7 микросервисов готовы и работают
- ✅ API: 71 эндпоинт сгенерирован, документация готова
- ✅ Инфраструктура: Docker, базы данных настроены
- 🔄 Фронтенд: React + TypeScript, нужна интеграция с API

### Архитектура
```
Фронтенд (React) → API Gateway (8080) → Микросервисы (gRPC)
                                      ↓
                                 Базы данных (PostgreSQL, Redis, ClickHouse)
```

## 📁 Ключевые файлы для понимания

### Документация (читай в первую очередь)
- `docs/FRONTEND_INTEGRATION_GUIDE.md` - ГЛАВНЫЙ файл для интеграции
- `docs/API_GENERATION_REPORT.md` - отчет о готовой API
- `docs/DIAGNOSTIC_RESOLUTION.md` - решение проблем
- `TODO.md` - текущие задачи и прогресс

### Сгенерированные API хуки
- `frontend/src/generated/requests/` - готовые API функции
- `frontend/src/generated/types.ts` - TypeScript типы
- `docs/swagger/unified-api.json` - схема API (71 эндпоинт)

### Конфигурация
- `Makefile` - все команды для разработки
- `docker-compose.yml` - инфраструктура
- `go.work` - Go workspace для всех сервисов

## 🎯 Текущая задача

**Интеграция фронтенда с готовыми API хуками**

### Что готово:
1. **71 API эндпоинт** полностью работают
2. **TypeScript типы** автогенерированы
3. **HTTP клиент** настроен
4. **JWT авторизация** готова

### Что нужно сделать:
1. **Импортировать** API функции из `frontend/src/generated/requests`
2. **Настроить** JWT токены через `POST /api/v1/auth/login`
3. **Создать** React компоненты для дашборда, серверов, туннелей
4. **Добавить** React Query для кэширования данных

## 🛠️ Основные команды

```bash
# Запуск всего проекта
make dev-all

# Только фронтенд
make frontend-dev

# Проверка API
make swagger

# Диагностика
make diagnostics

# Статус инфраструктуры
make infra-status
```

## 📡 API Endpoints (все работают!)

### Аутентификация
```
POST /api/v1/auth/login     # Логин → JWT токен
GET  /api/v1/auth/me        # Профиль пользователя
POST /api/v1/users          # Создание пользователя
```

### Серверы
```
GET  /api/v1/servers        # Список серверов
POST /api/v1/servers/{id}/start  # Запуск сервера
GET  /api/v1/servers/{id}/stats  # Статистика сервера
```

### VPN Туннели
```
GET  /api/v1/vpn/tunnels    # Список туннелей
POST /api/v1/vpn/tunnels    # Создание туннеля
GET  /api/v1/vpn/tunnels/{id}/peers  # Пиры туннеля
```

### Аналитика
```
GET  /api/v1/analytics/dashboard     # Дашборд
GET  /api/v1/analytics/statistics/system  # Системная статистика
```

## 🔧 Пример использования API

```typescript
import { AuthService, ServerManagerService } from '@/generated/requests';

// Логин
const login = await AuthService.authServiceLogin({
  email: "admin@example.com",
  password: "password"
});

// Получение серверов
const servers = await ServerManagerService.serverManagerServiceListServers({
  limit: 10
});
```

## ⚠️ Важно знать

1. **Единая точка входа:** Все API запросы идут через `localhost:8080`
2. **JWT обязателен:** Кроме эндпоинтов `/health` и `/login`
3. **Типы автогенерированы:** Не редактировать `frontend/src/generated/`
4. **Swagger UI:** Доступен на `http://localhost:8080/swagger`

## 🎯 Следующие шаги

1. **Прочитай** `docs/FRONTEND_INTEGRATION_GUIDE.md` - там все детали
2. **Изучи** сгенерированные типы в `frontend/src/generated/`
3. **Посмотри** на структуру проекта в `frontend/src/`
4. **Начни** с простых компонентов (логин, список серверов)
5. **Используй** готовые API функции вместо axios/fetch

## 📋 Статус задач

- ✅ Бэкенд сервисы (7 микросервисов)
- ✅ API документация (71 эндпоинт)
- ✅ Клиентский SDK (TypeScript)
- ✅ Инфраструктура (Docker, БД)
- 🔄 UI компоненты (React)
- 🔄 Интеграция с API хуками

**Все готово для интеграции фронтенда!** 🚀