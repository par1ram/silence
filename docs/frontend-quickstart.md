# Frontend Quickstart для Silence VPN

## Как работает backend

Система Silence — это микросервисный backend для защищённого VPN с обфускацией трафика. Все основные действия пользователя реализованы через API Gateway (порт 8080), который проксирует запросы к нужным сервисам (auth, vpn-core, dpi-bypass).

### Основные сценарии:

1. **Регистрация пользователя** — POST /api/v1/auth/register
2. **Вход пользователя** — POST /api/v1/auth/login
3. **Подключение к VPN с обфускацией** — POST /api/v1/connect (core-фича)

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

- **POST /api/v1/auth/register** — регистрация
- **POST /api/v1/auth/login** — логин (получить JWT)
- **POST /api/v1/connect** — создать VPN+обфускацию (core)
- **GET /api/v1/dpi-bypass/bypass** — список bypass-конфигов (для advanced UI)
- **GET /api/v1/vpn/tunnels/list** — список туннелей (для advanced UI)

## UI/UX рекомендации

- После логина хранить JWT в памяти (или httpOnly cookie)
- Для защищённых запросов всегда отправлять Authorization: Bearer <JWT>
- Показывать статус подключения (connected/disconnected)
- Для advanced UI: показывать статистику туннеля и bypass (через соответствующие GET)

## Пример минимального flow

1. Форма регистрации/логина
2. После логина — форма выбора метода обфускации и параметров
3. Кнопка "Подключиться" вызывает /api/v1/connect
4. Отображение статуса и параметров подключения

---

## Core-фича уже работает?

**Да!**

- Регистрация, логин, подключение к VPN с обфускацией через /api/v1/connect полностью реализованы и протестированы.
- Можно делать фронт для реального сценария: пользователь заходит, логинится, нажимает "Подключиться" — и получает защищённое VPN соединение с обфускацией.
