# Промт для AI-ассистента: Интеграция фронтенда Silence VPN

## 🎯 Ваша задача
Интегрировать существующий React фронтенд с готовыми API хуками, заменив самописные сервисы на сгенерированные из Swagger схемы.

## 📋 Контекст проекта Silence VPN

### Что уже готово:
- ✅ **7 микросервисов** (auth, server-manager, vpn-core, analytics, dpi-bypass, notifications, gateway)
- ✅ **71 API эндпоинт** полностью работают на `localhost:8080`
- ✅ **Swagger документация** сгенерирована из proto файлов
- ✅ **TypeScript API клиент** сгенерирован в `frontend/src/generated/requests/`
- ✅ **React фронтенд** с базовыми компонентами

### Что нужно сделать:
Заменить самописные API сервисы на сгенерированные хуки, сохранив функциональность UI.

## 🔍 Где искать информацию

### Обязательно прочитайте:
1. **`docs/FRONTEND_INTEGRATION_PLAN.md`** - ГЛАВНЫЙ файл с пошаговым планом
2. **`docs/FRONTEND_INTEGRATION_GUIDE.md`** - справочник по архитектуре
3. **`TODO.md`** - текущие задачи с деталями

### Изучите структуру фронтенда:
```
frontend/src/
├── generated/requests/        # Сгенерированные API (готовые к использованию)
├── services/api.ts           # Самописные сервисы (заменить)
├── stores/auth.ts            # Zustand store (обновить)
├── types/api.ts              # Кастомные типы (обновить)
├── components/               # React компоненты (обновить)
└── hooks/                    # React Query хуки (создать новые)
```

### Ключевые файлы для изменения:
- `src/services/api.ts` - заменить на сгенерированные API
- `src/stores/auth.ts` - обновить API вызовы
- `src/components/vpn/VPNDashboard.tsx` - обновить загрузку данных
- `src/app/dashboard/page.tsx` - обновить статистику

## 📡 Соответствие API

### Замена сервисов:
```typescript
// СТАРОЕ (удалить):
import { AuthService } from '@/services/api';
await AuthService.login(data);

// НОВОЕ (использовать):
import { AuthServiceService } from '@/generated/requests';
await AuthServiceService.authServiceLogin({
  requestBody: data
});
```

### Основные замены:
| Старый сервис | Новый сервис | Примечания |
|---------------|--------------|------------|
| `AuthService` | `AuthServiceService` | Логин, регистрация, профиль |
| `serverService` | `ServerManagerServiceService` | Управление серверами |
| `vpnService` | `VpnCoreServiceService` | VPN туннели |
| `analyticsService` | `AnalyticsServiceService` | Метрики и статистика |
| `notificationService` | `NotificationsServiceService` | Уведомления |

## 🔧 Пошаговый план работы

### Шаг 1: Настройка HTTP клиента
```typescript
// Создать src/lib/api-client.ts
import { OpenAPI } from '@/generated/requests';

OpenAPI.BASE = 'http://localhost:8080';
OpenAPI.TOKEN = () => localStorage.getItem('auth-token') || '';
```

### Шаг 2: Обновление AuthStore
```typescript
// Обновить src/stores/auth.ts
import { AuthServiceService } from '@/generated/requests';

// Заменить API вызовы:
const response = await AuthServiceService.authServiceLogin({
  requestBody: { email, password }
});
```

### Шаг 3: Создание React Query хуков
```typescript
// Создать src/hooks/useServers.ts
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

### Шаг 4: Обновление компонентов
```typescript
// Обновить компоненты для использования новых хуков
import { useServers } from '@/hooks/useServers';

export function VPNDashboard() {
  const { data: servers, isLoading } = useServers();
  // ...
}
```

## ⚠️ Важные особенности

### 1. Формат токена
```typescript
// Токен приходит в поле token, не accessToken
localStorage.setItem('auth-token', response.token);
```

### 2. Формат запросов
```typescript
// Все параметры передаются в объекте
AuthServiceService.authServiceLogin({
  requestBody: { email, password }
});

ServerManagerServiceService.serverManagerServiceListServers({
  limit: 10,
  offset: 0
});
```

### 3. Типы данных
```typescript
// Использовать сгенерированные типы
import type {
  authUser,
  serverServer,
  vpnTunnel
} from '@/generated/requests';
```

## 🧪 Тестирование

### Проверочный список:
- [ ] Логин/регистрация работают
- [ ] Загрузка списка серверов
- [ ] Создание VPN туннеля
- [ ] Получение статистики
- [ ] Отображение уведомлений
- [ ] Обработка ошибок

### Команды для тестирования:
```bash
# Запуск фронтенда
cd frontend && npm run dev

# Запуск бэкенда (в отдельном терминале)
make dev-all

# Проверка API
curl http://localhost:8080/api/v1/auth/health
```

## 📚 Справочные материалы

1. **Архитектура**: `docs/FRONTEND_INTEGRATION_GUIDE.md`
2. **Пошаговый план**: `docs/FRONTEND_INTEGRATION_PLAN.md`
3. **Сгенерированные типы**: `frontend/src/generated/types.gen.ts`
4. **Swagger UI**: `http://localhost:8080/swagger`
5. **API Gateway**: `http://localhost:8080`

## 🎯 Ожидаемый результат

После интеграции фронтенд будет:
- Использовать **типобезопасные API вызовы**
- Работать с **актуальными типами данных**
- Иметь **централизованную обработку ошибок**
- Использовать **кэширование через React Query**
- Автоматически **синхронизироваться** при изменении API

## 💡 Советы для успешной интеграции

1. **Начните с изучения** `docs/FRONTEND_INTEGRATION_PLAN.md`
2. **Сравните старые и новые** API вызовы
3. **Тестируйте каждый шаг** - запускайте фронтенд после каждого изменения
4. **Используйте TypeScript** - он покажет ошибки типов
5. **Проверяйте Network tab** в DevTools для отладки API вызовов

---

**Статус**: Готов к выполнению
**Приоритет**: Высокий
**Время**: 2-3 дня
**Результат**: Полная интеграция фронтенда с готовыми API хуками

Удачи в интеграции! 🚀

ОБЯЗАТЕЛЬНО СОБЛЮДАЙ ОБЩИЙ СТЕКЛЯННЫЙ ДИЗАЙН ФРОНТА
