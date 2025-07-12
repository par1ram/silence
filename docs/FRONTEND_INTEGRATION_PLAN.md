# Пошаговый план интеграции фронтенда с готовыми API

## 🎯 Цель задачи
Интегрировать существующий React фронтенд с готовыми API хуками, заменив самописные сервисы на сгенерированные из Swagger схемы.

## 📋 Текущее состояние фронтенда

### ✅ Что уже есть:
- **Next.js 15** с App Router и TypeScript
- **Компоненты UI**: дашборд, VPN подключения, аутентификация
- **Zustand store** для управления состоянием
- **React Query** для кэширования данных
- **Самописные API сервисы** в `src/services/api.ts`
- **Кастомные типы** в `src/types/api.ts`

### ✅ Что сгенерировано:
- **71 API эндпоинт** в `src/generated/requests/`
- **201 определение типов** в `src/generated/types.gen.ts`
- **HTTP клиент** в `src/generated/requests/core/`
- **Swagger схема** в `docs/swagger/unified-api.json`

## 🚀 Пошаговый план интеграции

### Шаг 1: Анализ существующего кода
```bash
# Файлы для изучения:
silence/frontend/src/services/api.ts         # Самописные API сервисы
silence/frontend/src/stores/auth.ts          # Zustand store для авторизации
silence/frontend/src/types/api.ts            # Кастомные типы API
silence/frontend/src/generated/requests/     # Сгенерированные API хуки
silence/frontend/src/generated/types.gen.ts  # Сгенерированные типы
```

**Задача:** Понять какие API вызовы используются в текущем коде и как они связаны с компонентами.

### Шаг 2: Настройка HTTP клиента
```typescript
// Обновить src/lib/api-client.ts
import { OpenAPI } from '@/generated/requests';

OpenAPI.BASE = 'http://localhost:8080';
OpenAPI.WITH_CREDENTIALS = true;

// Настроить JWT токены
OpenAPI.TOKEN = async () => {
  return localStorage.getItem('auth-token') || '';
};

// Добавить обработку ошибок
OpenAPI.HEADERS = {
  'Content-Type': 'application/json',
};
```

**Файлы для изменения:**
- `src/lib/api-client.ts` - создать новый файл для настройки
- `src/generated/requests/core/OpenAPI.ts` - проверить конфигурацию

### Шаг 3: Замена AuthService
```typescript
// Заменить в src/stores/auth.ts
import { AuthService } from '@/services/api';
// НА:
import { AuthServiceService } from '@/generated/requests';

// Старый вызов:
const res = await AuthService.login(data);
// Новый вызов:
const res = await AuthServiceService.authServiceLogin({
  requestBody: data
});
```

**Соответствие эндпоинтов:**
- `AuthService.login()` → `AuthServiceService.authServiceLogin()`
- `AuthService.register()` → `AuthServiceService.authServiceRegister()`
- `AuthService.getProfile()` → `AuthServiceService.authServiceGetMe()`

**Файлы для изменения:**
- `src/stores/auth.ts` - обновить все API вызовы
- `src/components/auth/AuthProvider.tsx` - проверить типы

### Шаг 4: Замена ServerService
```typescript
// Заменить в компонентах
import { serverService } from '@/services/api';
// НА:
import { ServerManagerServiceService } from '@/generated/requests';

// Старый вызов:
const servers = await serverService.getServers();
// Новый вызов:
const servers = await ServerManagerServiceService.serverManagerServiceListServers({
  limit: 10,
  offset: 0
});
```

**Соответствие эндпоинтов:**
- `serverService.getServers()` → `ServerManagerServiceService.serverManagerServiceListServers()`
- `serverService.getServer(id)` → `ServerManagerServiceService.serverManagerServiceGetServer()`
- `serverService.createServer()` → `ServerManagerServiceService.serverManagerServiceCreateServer()`

**Файлы для изменения:**
- `src/components/vpn/VPNDashboard.tsx` - обновить загрузку серверов
- `src/app/dashboard/page.tsx` - обновить статус серверов

### Шаг 5: Замена VPNService
```typescript
// Заменить VPN API вызовы
import { VpnCoreServiceService } from '@/generated/requests';

// Старый вызов:
const tunnels = await vpnService.getTunnels();
// Новый вызов:
const tunnels = await VpnCoreServiceService.vpnCoreServiceListTunnels({});
```

**Соответствие эндпоинтов:**
- `vpnService.getTunnels()` → `VpnCoreServiceService.vpnCoreServiceListTunnels()`
- `vpnService.getTunnelStats()` → `VpnCoreServiceService.vpnCoreServiceGetTunnelStats()`

**Файлы для изменения:**
- `src/components/vpn/VPNCore.tsx` - обновить управление туннелями
- `src/components/vpn/VPNDashboard.tsx` - обновить статистику

### Шаг 6: Замена AnalyticsService
```typescript
// Заменить аналитику
import { AnalyticsServiceService } from '@/generated/requests';

// Новые вызовы:
const dashboard = await AnalyticsServiceService.analyticsServiceGetDashboardData({
  timeRange: '24h'
});
const systemStats = await AnalyticsServiceService.analyticsServiceGetSystemStats({});
```

**Файлы для изменения:**
- `src/components/analytics/` - все компоненты аналитики
- `src/components/dashboard/StatsCards.tsx` - обновить метрики

### Шаг 7: Замена NotificationService
```typescript
// Заменить уведомления
import { NotificationsServiceService } from '@/generated/requests';

// Новые вызовы:
const notifications = await NotificationsServiceService.notificationsServiceListNotifications({
  limit: 20
});
```

**Файлы для изменения:**
- `src/components/notifications/` - все компоненты уведомлений

### Шаг 8: Обновление типов
```typescript
// Заменить кастомные типы
import type { User, Server, Tunnel } from '@/types/api';
// НА:
import type { 
  authUser, 
  serverServer, 
  vpnTunnel 
} from '@/generated/requests';

// Создать мапперы для совместимости
export const mapBackendUser = (user: authUser): User => ({
  id: user.id || '',
  email: user.email || '',
  role: user.role || 'USER_ROLE_USER',
  // ... остальные поля
});
```

**Файлы для изменения:**
- `src/types/api.ts` - добавить мапперы типов
- Все компоненты, использующие кастомные типы

### Шаг 9: Обновление React Query хуков
```typescript
// Создать новые хуки с сгенерированными API
// src/hooks/useServers.ts
import { useQuery } from '@tanstack/react-query';
import { ServerManagerServiceService } from '@/generated/requests';

export const useServers = () => {
  return useQuery({
    queryKey: ['servers'],
    queryFn: () => ServerManagerServiceService.serverManagerServiceListServers({
      limit: 100
    })
  });
};
```

**Файлы для создания:**
- `src/hooks/useAuth.ts` - хуки для авторизации
- `src/hooks/useServers.ts` - хуки для серверов
- `src/hooks/useVPN.ts` - хуки для VPN
- `src/hooks/useAnalytics.ts` - хуки для аналитики

### Шаг 10: Обновление компонентов
```typescript
// Обновить использование API в компонентах
// src/components/vpn/VPNDashboard.tsx
import { useServers } from '@/hooks/useServers';
import { useVPNStatus } from '@/hooks/useVPN';

export function VPNDashboard() {
  const { data: servers, isLoading } = useServers();
  const { data: status } = useVPNStatus();
  
  // Использовать новые данные...
}
```

**Файлы для обновления:**
- `src/components/vpn/VPNDashboard.tsx`
- `src/components/dashboard/StatsCards.tsx`
- `src/app/dashboard/page.tsx`
- Все компоненты, использующие API

## 📁 Структура файлов после интеграции

```
src/
├── generated/              # Сгенерированные API хуки (не трогать)
│   ├── requests/
│   └── types.gen.ts
├── hooks/                  # Новые React Query хуки
│   ├── useAuth.ts
│   ├── useServers.ts
│   ├── useVPN.ts
│   └── useAnalytics.ts
├── lib/
│   └── api-client.ts       # Конфигурация HTTP клиента
├── services/
│   └── api.ts             # Удалить после миграции
├── stores/
│   └── auth.ts            # Обновить с новыми API
├── types/
│   └── api.ts             # Добавить мапперы типов
└── components/            # Обновить использование API
```

## 🔧 Команды для работы

```bash
# Регенерация API при изменениях
npm run generate:api

# Запуск в режиме разработки
npm run dev

# Проверка типов
npm run type-check

# Валидация API схемы
npm run api:validate
```

## ⚠️ Важные моменты

### 1. Формат ответов API
```typescript
// Старый формат (самописный):
interface ApiResponse<T> {
  success: boolean;
  data: T;
}

// Новый формат (сгенерированный):
// Прямой возврат данных без обертки
```

### 2. Обработка ошибок
```typescript
// Настроить глобальную обработку ошибок
OpenAPI.HEADERS = {
  'Content-Type': 'application/json',
};

// Обработка 401 ошибок
OpenAPI.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.status === 401) {
      // Редирект на логин
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### 3. JWT токены
```typescript
// Сохранение токена после логина
const loginResponse = await AuthServiceService.authServiceLogin({
  requestBody: { email, password }
});

// Токен приходит в поле token, не accessToken
localStorage.setItem('auth-token', loginResponse.token);
```

## 📊 Тестирование интеграции

### Проверочный список:
- [ ] Логин/регистрация работают
- [ ] Загрузка списка серверов
- [ ] Создание VPN туннеля
- [ ] Получение статистики
- [ ] Отображение уведомлений
- [ ] Обработка ошибок авторизации
- [ ] Кэширование данных через React Query

### Команды для тестирования:
```bash
# Запуск фронтенда
make frontend-dev

# Запуск бэкенда
make dev-all

# Проверка API
curl http://localhost:8080/api/v1/auth/health
```

## 🎯 Результат
После выполнения всех шагов фронтенд будет использовать:
- **Типобезопасные API вызовы** из сгенерированных хуков
- **Актуальные типы данных** из Swagger схемы
- **Централизованную обработку ошибок**
- **Кэширование данных** через React Query
- **Автоматическую синхронизацию** при изменении API

## 📚 Справочные материалы

1. **Архитектура API**: `docs/FRONTEND_INTEGRATION_GUIDE.md`
2. **Сгенерированные типы**: `frontend/src/generated/types.gen.ts`
3. **Swagger схема**: `docs/swagger/unified-api.json`
4. **API Gateway**: `http://localhost:8080`
5. **Swagger UI**: `http://localhost:8080/swagger`

---

**Статус**: Готов к выполнению  
**Приоритет**: Высокий  
**Время выполнения**: 2-3 дня  
**Результат**: Полная интеграция фронтенда с готовыми API хуками