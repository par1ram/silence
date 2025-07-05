# Frontend Quickstart для Silence VPN

## Как работает backend

Система Silence — это микросервисный backend для защищённого VPN с обфускацией трафика. Все основные действия пользователя реализованы через API Gateway (порт 8080), который проксирует запросы к нужным сервисам.

### Архитектура системы:

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
```

### Основные сценарии:

1. **Регистрация/авторизация** — POST /api/v1/auth/register, POST /api/v1/auth/login
2. **Подключение к VPN с обфускацией** — POST /api/v1/connect (core-фича)
3. **Управление серверами** — GET/POST /api/v1/server-manager/servers
4. **Аналитика и мониторинг** — GET /api/v1/analytics/metrics/\*
5. **Уведомления** — POST /api/v1/notifications/notifications

## Core-фича: подключение к VPN с обфускацией

Пользователь:

- Регистрируется (или логинится)
- Получает JWT токен
- Выбирает параметры (метод обфускации, сервер, порт и т.д.)
- Нажимает "Подключиться" — фронт отправляет POST /api/v1/connect с нужными параметрами и токеном
- Получает ответ с данными для подключения (id туннеля, порт, статус)

### Пример запроса на подключение

```js
fetch('/api/v1/connect', {
	method: 'POST',
	headers: {
		'Content-Type': 'application/json',
		Authorization: 'Bearer <JWT>',
	},
	body: JSON.stringify({
		bypass_method: 'shadowsocks',
		bypass_config: {
			local_port: 1080,
			remote_host: '127.0.0.1',
			remote_port: 8388,
			password: 'testpass',
			encryption: 'AES-256-GCM',
		},
		vpn_config: {
			name: 'my-vpn',
			listen_port: 51820,
			mtu: 1420,
			auto_recovery: true,
		},
	}),
})
	.then((res) => res.json())
	.then((data) => {
		// data.bypass_id, data.vpn_tunnel, data.status
	})
```

### Пример ответа

```json
{
	"bypass_id": "...",
	"bypass_port": 1080,
	"vpn_tunnel": "...",
	"status": "connected",
	"created_at": "..."
}
```

## Основные API для фронта

### Аутентификация

- **POST /api/v1/auth/register** — регистрация
- **POST /api/v1/auth/login** — логин (получить JWT)
- **GET /api/v1/auth/me** — профиль пользователя

### VPN и обфускация (Core)

- **POST /api/v1/connect** — создать VPN+обфускацию (core-фича)
- **GET /api/v1/dpi-bypass/bypass** — список bypass-конфигов
- **GET /api/v1/vpn/tunnels/list** — список туннелей
- **GET /api/v1/vpn/tunnels/stats** — статистика туннелей

### Управление серверами

- **GET /api/v1/server-manager/servers** — список серверов
- **POST /api/v1/server-manager/servers** — создать сервер
- **GET /api/v1/server-manager/servers/{id}/stats** — статистика сервера
- **GET /api/v1/server-manager/servers/{id}/health** — здоровье сервера

### Аналитика и мониторинг

- **GET /api/v1/analytics/metrics/connections** — метрики подключений
- **GET /api/v1/analytics/metrics/server-load** — нагрузка серверов
- **GET /api/v1/analytics/metrics/errors** — метрики ошибок
- **GET /api/v1/analytics/dashboards** — дашборды
- **GET /api/v1/analytics/alerts** — алерты

### Уведомления

- **POST /api/v1/notifications/notifications** — отправить уведомление
- **GET /api/v1/notifications/health** — состояние сервиса

## UI/UX рекомендации

### Безопасность

- После логина хранить JWT в памяти (или httpOnly cookie)
- Для защищённых запросов всегда отправлять Authorization: Bearer <JWT>
- Реализовать автоматическое обновление токена при истечении

### Основной интерфейс

- Показывать статус подключения (connected/disconnected)
- Отображать текущий сервер и метод обфускации
- Показывать статистику трафика (bytes in/out)

### Административная панель

- Дашборд с метриками и алертами
- Управление серверами (создание, мониторинг, масштабирование)
- Система уведомлений (настройка каналов, отправка тестовых уведомлений)
- Просмотр логов и ошибок

### Аналитика

- Графики подключений по времени
- Статистика эффективности обхода DPI
- Мониторинг нагрузки серверов
- История алертов и инцидентов

## Примеры использования API

### 1. Базовый flow подключения

```js
// 1. Регистрация
const register = async (email, password) => {
	const response = await fetch('/api/v1/auth/register', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password }),
	})
	return response.json()
}

// 2. Вход
const login = async (email, password) => {
	const response = await fetch('/api/v1/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password }),
	})
	const data = await response.json()
	localStorage.setItem('jwt', data.token)
	return data
}

// 3. Подключение к VPN
const connectVPN = async () => {
	const token = localStorage.getItem('jwt')
	const response = await fetch('/api/v1/connect', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
		},
		body: JSON.stringify({
			bypass_method: 'shadowsocks',
			bypass_config: {
				local_port: 1080,
				remote_host: '127.0.0.1',
				remote_port: 8388,
				password: 'testpass',
				encryption: 'AES-256-GCM',
			},
			vpn_config: {
				name: 'my-vpn',
				listen_port: 51820,
				mtu: 1420,
				auto_recovery: true,
			},
		}),
	})
	return response.json()
}
```

### 2. Мониторинг серверов

```js
// Получение списка серверов
const getServers = async () => {
	const token = localStorage.getItem('jwt')
	const response = await fetch('/api/v1/server-manager/servers', {
		headers: { Authorization: `Bearer ${token}` },
	})
	return response.json()
}

// Статистика сервера
const getServerStats = async (serverId) => {
	const token = localStorage.getItem('jwt')
	const response = await fetch(
		`/api/v1/server-manager/servers/${serverId}/stats`,
		{
			headers: { Authorization: `Bearer ${token}` },
		}
	)
	return response.json()
}
```

### 3. Аналитика

```js
// Метрики подключений
const getConnectionMetrics = async (start, end) => {
	const token = localStorage.getItem('jwt')
	const response = await fetch(
		`/api/v1/analytics/metrics/connections?start=${start}&end=${end}`,
		{ headers: { Authorization: `Bearer ${token}` } }
	)
	return response.json()
}

// Нагрузка серверов
const getServerLoad = async () => {
	const token = localStorage.getItem('jwt')
	const response = await fetch('/api/v1/analytics/metrics/server-load', {
		headers: { Authorization: `Bearer ${token}` },
	})
	return response.json()
}
```

### 4. Уведомления

```js
// Отправка уведомления
const sendNotification = async (type, title, message, recipients, channels) => {
	const token = localStorage.getItem('jwt')
	const response = await fetch('/api/v1/notifications/notifications', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${token}`,
		},
		body: JSON.stringify({
			type,
			priority: 'normal',
			title,
			message,
			recipients,
			channels,
			metadata: { source: 'frontend' },
		}),
	})
	return response.json()
}

// Пример: уведомление о высокой нагрузке
sendNotification(
	'system_alert',
	'Высокая нагрузка на сервер',
	'CPU usage превышает 90%',
	['admin@silence.com'],
	['email', 'slack']
)
```

## Рекомендуемая структура интерфейса

### Для обычных пользователей:

1. **Страница входа/регистрации**
2. **Главная панель**:
   - Статус подключения (подключен/отключен)
   - Кнопка подключения/отключения
   - Выбор сервера и метода обфускации
   - Статистика трафика
3. **Настройки**:
   - Профиль пользователя
   - Предпочтения уведомлений
   - История подключений

### Для администраторов:

1. **Дашборд**:
   - Обзор всех серверов
   - Активные алерты
   - Ключевые метрики
2. **Управление серверами**:
   - Список серверов с статусами
   - Создание новых серверов
   - Мониторинг производительности
3. **Аналитика**:
   - Графики метрик
   - Настройка дашбордов
   - Управление алертами
4. **Уведомления**:
   - Отправка уведомлений
   - Настройка каналов доставки
   - История уведомлений

---

## Что уже работает?

**Все основные компоненты!**

### ✅ Полностью реализовано и протестировано:

1. **Аутентификация** - регистрация, логин, управление пользователями
2. **VPN Core** - создание WireGuard туннелей, управление пирами
3. **DPI Bypass** - обфускация трафика (Shadowsocks, V2Ray, obfs4)
4. **API Gateway** - единая точка входа для всех сервисов
5. **Analytics** - сбор метрик, дашборды, алерты
6. **Notifications** - отправка уведомлений через email, SMS, push, telegram, slack, webhook
7. **Server Manager** - управление серверами, масштабирование, мониторинг

### 🎯 Готовые сценарии для фронтенда:

1. **Пользовательский сценарий**: регистрация → логин → подключение к VPN с обфускацией
2. **Административный сценарий**: мониторинг серверов → анализ метрик → отправка уведомлений
3. **Операционный сценарий**: создание серверов → настройка алертов → масштабирование

### 🚀 Можно сразу делать фронтенд!

Все API endpoints работают, протестированы и документированы. Фронтенд может использовать любой современный фреймворк (React, Vue, Angular) для создания полнофункционального интерфейса.

## Технические требования

### Минимальные требования:

- **JavaScript ES6+** или **TypeScript**
- **HTTP клиент** (fetch API или axios)
- **Роутинг** для SPA
- **Управление состоянием** (Redux, Zustand, Context API)

### Рекомендуемые технологии:

- **React 18+** с **TypeScript**
- **Vite** для сборки
- **React Router** для роутинга
- **TanStack Query** для кэширования API
- **Tailwind CSS** для стилизации
- **Recharts** или **Chart.js** для графиков
- **React Hook Form** для форм

### Структура проекта:

```
frontend/
├── src/
│   ├── components/          # Переиспользуемые компоненты
│   │   ├── ui/             # Базовые UI компоненты
│   │   ├── auth/           # Компоненты аутентификации
│   │   ├── vpn/            # VPN интерфейс
│   │   ├── analytics/      # Аналитика и дашборды
│   │   ├── servers/        # Управление серверами
│   │   └── notifications/  # Система уведомлений
│   ├── hooks/              # Кастомные хуки
│   ├── services/           # API сервисы
│   ├── stores/             # Управление состоянием
│   ├── types/              # TypeScript типы
│   └── utils/              # Утилиты
```

### Ключевые компоненты для реализации:

1. **AuthProvider** - управление аутентификацией и JWT
2. **VPNConnection** - подключение/отключение VPN
3. **ServerDashboard** - мониторинг серверов
4. **AnalyticsCharts** - графики метрик
5. **NotificationCenter** - отправка и настройка уведомлений
6. **AlertManager** - управление алертами

### Безопасность:

- Валидация всех форм на клиенте и сервере
- Защита от XSS и CSRF атак
- Безопасное хранение JWT токенов
- Rate limiting на клиенте
- Логирование действий пользователей
