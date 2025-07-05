# API Endpoints - Silence VPN

## Обзор

Документация всех HTTP и gRPC эндпоинтов сервисов проекта Silence VPN.

## Архитектура API

```
Client → Gateway (API Gateway) → Микросервисы
                ↓
        ┌─────────────────┐
        │ Auth Service    │
        │ VPN Core        │
        │ DPI Bypass      │
        │ Analytics       │
        │ Notifications   │
        └─────────────────┘
                ↓
        ┌─────────────────┐
        │ InfluxDB        │
        │ Redis           │
        │ PostgreSQL      │
        │ RabbitMQ        │
        └─────────────────┘
```

## 1. Gateway Service (API Gateway)

**Базовый URL**: `http://localhost:8080`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния API Gateway  
**Ответ**:

```json
{
	"status": "ok",
	"service": "gateway",
	"version": "1.0.0"
}
```

### Root

```
GET /
```

**Описание**: Информация о сервисе  
**Ответ**:

```json
{
	"message": "Silence VPN Gateway Service",
	"version": "1.0.0"
}
```

### Интеграция VPN + Обфускация

```
POST /api/v1/connect
```

**Описание**: Создание VPN-соединения с обфускацией трафика  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Тело запроса**:

```json
{
	"bypass_method": "shadowsocks",
	"bypass_config": {
		"local_port": 1080,
		"remote_host": "proxy.example.com",
		"remote_port": 8388,
		"password": "secret",
		"encryption": "aes-256-gcm"
	},
	"vpn_config": {
		"name": "my-vpn",
		"listen_port": 51820,
		"mtu": 1420,
		"auto_recovery": true
	}
}
```

**Ответ**:

```json
{
	"bypass_id": "bypass-123",
	"bypass_port": 1080,
	"vpn_tunnel": "tunnel-456",
	"status": "connected",
	"created_at": "2024-01-01T12:00:00Z"
}
```

### Проксирование к Auth Service

```
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/me
```

**Описание**: Проксирование запросов к сервису аутентификации  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>` (кроме register/login)

### Проксирование к VPN Core

```
POST   /api/v1/vpn/tunnels
GET    /api/v1/vpn/tunnels/list
GET    /api/v1/vpn/tunnels/get
POST   /api/v1/vpn/tunnels/start
POST   /api/v1/vpn/tunnels/stop
GET    /api/v1/vpn/tunnels/stats
POST   /api/v1/vpn/peers/add
GET    /api/v1/vpn/peers/get
GET    /api/v1/vpn/peers/list
DELETE /api/v1/vpn/peers/remove
```

**Описание**: Проксирование запросов к VPN Core сервису  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

### Проксирование к DPI Bypass

```
POST   /api/v1/dpi-bypass/bypass
GET    /api/v1/dpi-bypass/bypass
GET    /api/v1/dpi-bypass/bypass/{id}
DELETE /api/v1/dpi-bypass/bypass/{id}
POST   /api/v1/dpi-bypass/bypass/{id}/start
POST   /api/v1/dpi-bypass/bypass/{id}/stop
GET    /api/v1/dpi-bypass/bypass/{id}/stats
```

**Описание**: Проксирование запросов к DPI Bypass сервису  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

### Проксирование к Analytics

```
GET    /api/v1/analytics/health
GET    /api/v1/analytics/metrics/connections
GET    /api/v1/analytics/metrics/bypass-effectiveness
GET    /api/v1/analytics/metrics/user-activity
GET    /api/v1/analytics/metrics/server-load
GET    /api/v1/analytics/metrics/errors
GET    /api/v1/analytics/dashboards
POST   /api/v1/analytics/dashboards
GET    /api/v1/analytics/dashboards/{id}
PUT    /api/v1/analytics/dashboards/{id}
DELETE /api/v1/analytics/dashboards/{id}
GET    /api/v1/analytics/alerts
POST   /api/v1/analytics/alerts
GET    /api/v1/analytics/alerts/{id}
PUT    /api/v1/analytics/alerts/{id}
DELETE /api/v1/analytics/alerts/{id}
```

**Описание**: Проксирование запросов к Analytics сервису  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

### Проксирование к Notifications

```
GET    /api/v1/notifications/health
POST   /api/v1/notifications/notifications
```

**Описание**: Проксирование запросов к Notifications сервису  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>` (для POST запросов)

### Проксирование к Server Manager

```
# Управление серверами
POST   /api/v1/server-manager/servers
GET    /api/v1/server-manager/servers
GET    /api/v1/server-manager/servers/{id}
PUT    /api/v1/server-manager/servers/{id}
DELETE /api/v1/server-manager/servers/{id}
POST   /api/v1/server-manager/servers/{id}/start
POST   /api/v1/server-manager/servers/{id}/stop
POST   /api/v1/server-manager/servers/{id}/restart
GET    /api/v1/server-manager/servers/{id}/stats
GET    /api/v1/server-manager/servers/{id}/health

# Масштабирование
GET    /api/v1/server-manager/scaling/policies
POST   /api/v1/server-manager/scaling/policies
PUT    /api/v1/server-manager/scaling/policies/{id}
DELETE /api/v1/server-manager/scaling/policies/{id}
POST   /api/v1/server-manager/scaling/evaluate

# Резервное копирование
GET    /api/v1/server-manager/backups/configs
POST   /api/v1/server-manager/backups/configs
PUT    /api/v1/server-manager/backups/configs/{id}
DELETE /api/v1/server-manager/backups/configs/{id}
POST   /api/v1/server-manager/servers/{id}/backup
POST   /api/v1/server-manager/servers/{id}/restore/{backup_id}

# Обновления
GET    /api/v1/server-manager/servers/{id}/update
POST   /api/v1/server-manager/servers/{id}/update
POST   /api/v1/server-manager/servers/{id}/update/cancel

# Мониторинг
GET    /api/v1/server-manager/health/all
```

**Описание**: Проксирование запросов к Server Manager сервису  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

---

## 2. Auth Service

**Базовый URL**: `http://localhost:8081`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния сервиса аутентификации  
**Ответ**:

```json
{
	"status": "ok",
	"service": "auth"
}
```

### Регистрация пользователя

```
POST /register
```

**Описание**: Создание нового аккаунта пользователя  
**Тело запроса**:

```json
{
	"email": "user@example.com",
	"password": "securepassword123"
}
```

**Ответ**:

```json
{
	"user_id": "123",
	"email": "user@example.com",
	"message": "User registered successfully"
}
```

### Вход пользователя

```
POST /login
```

**Описание**: Аутентификация пользователя  
**Тело запроса**:

```json
{
	"email": "user@example.com",
	"password": "securepassword123"
}
```

**Ответ**:

```json
{
	"user_id": "123",
	"email": "user@example.com",
	"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	"expires_in": 3600
}
```

### Получение профиля

```
GET /me
```

**Описание**: Получение информации о текущем пользователе  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Ответ**:

```json
{
	"user_id": "123",
	"email": "user@example.com",
	"created_at": "2024-01-01T12:00:00Z"
}
```

---

## 3. VPN Core Service

**Базовый URL**: `http://localhost:8082`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния VPN Core сервиса  
**Ответ**:

```json
{
	"status": "ok",
	"service": "vpn-core",
	"version": "1.0.0"
}
```

### Управление туннелями

#### Создание туннеля

```
POST /tunnels
```

**Описание**: Создание нового WireGuard туннеля  
**Тело запроса**:

```json
{
	"name": "my-tunnel",
	"listen_port": 51820,
	"mtu": 1420,
	"auto_recovery": true
}
```

**Ответ**:

```json
{
	"id": "tunnel-123",
	"name": "my-tunnel",
	"interface": "wg0",
	"status": "inactive",
	"public_key": "abc123...",
	"private_key": "xyz789...",
	"listen_port": 51820,
	"mtu": 1420,
	"created_at": "2024-01-01T12:00:00Z",
	"updated_at": "2024-01-01T12:00:00Z",
	"auto_recovery": true
}
```

#### Получение туннеля

```
GET /tunnels/get?id={tunnel_id}
```

**Описание**: Получение информации о туннеле по ID  
**Ответ**: Аналогично созданию туннеля

#### Список туннелей

```
GET /tunnels/list
```

**Описание**: Получение списка всех туннелей  
**Ответ**:

```json
{
	"tunnels": [
		{
			"id": "tunnel-123",
			"name": "my-tunnel",
			"status": "active",
			"listen_port": 51820,
			"created_at": "2024-01-01T12:00:00Z"
		}
	]
}
```

#### Запуск туннеля

```
POST /tunnels/start?id={tunnel_id}
```

**Описание**: Запуск туннеля  
**Ответ**: `200 OK`

#### Остановка туннеля

```
POST /tunnels/stop?id={tunnel_id}
```

**Описание**: Остановка туннеля  
**Ответ**: `200 OK`

#### Статистика туннеля

```
GET /tunnels/stats?id={tunnel_id}
```

**Описание**: Получение статистики туннеля  
**Ответ**:

```json
{
	"tunnel_id": "tunnel-123",
	"bytes_rx": 1024000,
	"bytes_tx": 512000,
	"peers_count": 5,
	"active_peers": 3,
	"last_updated": "2024-01-01T12:00:00Z",
	"uptime": 3600,
	"error_count": 0,
	"recovery_count": 1
}
```

### Управление пирами

#### Добавление пира

```
POST /peers/add
```

**Описание**: Добавление пира к туннелю  
**Тело запроса**:

```json
{
	"tunnel_id": "tunnel-123",
	"name": "client-1",
	"public_key": "client-public-key",
	"allowed_ips": "10.0.0.2/32",
	"endpoint": "192.168.1.100:51820",
	"keepalive": 25
}
```

**Ответ**:

```json
{
	"id": "peer-456",
	"tunnel_id": "tunnel-123",
	"name": "client-1",
	"public_key": "client-public-key",
	"allowed_ips": "10.0.0.2/32",
	"endpoint": "192.168.1.100:51820",
	"keepalive": 25,
	"status": "inactive",
	"created_at": "2024-01-01T12:00:00Z"
}
```

#### Получение пира

```
GET /peers/get?tunnel_id={tunnel_id}&peer_id={peer_id}
```

**Описание**: Получение информации о пире  
**Ответ**: Аналогично добавлению пира

#### Список пиров

```
GET /peers/list?tunnel_id={tunnel_id}
```

**Описание**: Получение списка пиров туннеля  
**Ответ**:

```json
{
	"peers": [
		{
			"id": "peer-456",
			"name": "client-1",
			"status": "active",
			"last_seen": "2024-01-01T12:00:00Z"
		}
	]
}
```

#### Удаление пира

```
DELETE /peers/remove?tunnel_id={tunnel_id}&peer_id={peer_id}
```

**Описание**: Удаление пира из туннеля  
**Ответ**:

```json
{
	"success": true
}
```

---

## 4. DPI Bypass Service

**Базовый URL**: `http://localhost:8083`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния DPI Bypass сервиса  
**Ответ**:

```json
{
	"status": "ok",
	"service": "dpi-bypass",
	"version": "1.0.0"
}
```

### Root

```
GET /
```

**Описание**: Информация о сервисе  
**Ответ**:

```json
{
	"message": "Silence DPI Bypass Service"
}
```

### Управление обфускацией

#### Создание bypass

```
POST /api/v1/bypass
```

**Описание**: Создание новой конфигурации обфускации  
**Тело запроса**:

```json
{
	"name": "my-bypass",
	"method": "shadowsocks",
	"local_port": 1080,
	"remote_host": "proxy.example.com",
	"remote_port": 8388,
	"password": "secret",
	"encryption": "aes-256-gcm"
}
```

**Ответ**:

```json
{
	"id": "bypass-123",
	"name": "my-bypass",
	"method": "shadowsocks",
	"local_port": 1080,
	"remote_host": "proxy.example.com",
	"remote_port": 8388,
	"status": "created",
	"created_at": "2024-01-01T12:00:00Z"
}
```

#### Список bypass

```
GET /api/v1/bypass
```

**Описание**: Получение списка всех bypass конфигураций  
**Ответ**:

```json
[
	{
		"id": "bypass-123",
		"name": "my-bypass",
		"method": "shadowsocks",
		"status": "running",
		"local_port": 1080
	}
]
```

#### Получение bypass

```
GET /api/v1/bypass/{id}
```

**Описание**: Получение информации о bypass по ID  
**Ответ**: Аналогично созданию bypass

#### Удаление bypass

```
DELETE /api/v1/bypass/{id}
```

**Описание**: Удаление bypass конфигурации  
**Ответ**: `204 No Content`

#### Запуск bypass

```
POST /api/v1/bypass/{id}/start
```

**Описание**: Запуск bypass обфускации  
**Ответ**:

```json
{
	"status": "started"
}
```

#### Остановка bypass

```
POST /api/v1/bypass/{id}/stop
```

**Описание**: Остановка bypass обфускации  
**Ответ**:

```json
{
	"status": "stopped"
}
```

#### Статистика bypass

```
GET /api/v1/bypass/{id}/stats
```

**Описание**: Получение статистики bypass  
**Ответ**:

```json
{
	"id": "bypass-123",
	"bytes_processed": 1024000,
	"connections_count": 150,
	"successful_connections": 145,
	"failed_connections": 5,
	"uptime": 3600,
	"last_updated": "2024-01-01T12:00:00Z"
}
```

---

## 5. Server Manager Service

**Базовый URL**: `http://localhost:8085`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния Server Manager сервиса  
**Ответ**:

```json
{
	"status": "ok",
	"service": "server-manager",
	"version": "1.0.0",
	"timestamp": "2024-01-01T12:00:00Z"
}
```

### Root

```
GET /
```

**Описание**: Информация о сервисе  
**Ответ**:

```json
{
	"message": "Silence Server Manager Service",
	"version": "1.0.0"
}
```

### Управление серверами

#### Создание сервера

```
POST /api/v1/servers
```

**Описание**: Создание нового сервера  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Тело запроса**:

```json
{
	"name": "vpn-server-1",
	"type": "vpn",
	"region": "us-east-1",
	"config": {
		"environment": {
			"VPN_PORT": "51820",
			"JWT_SECRET": "secret"
		},
		"command": ["/app/vpn-core"]
	}
}
```

**Ответ**:

```json
{
	"id": "server-123",
	"name": "vpn-server-1",
	"type": "vpn",
	"status": "running",
	"region": "us-east-1",
	"ip": "192.168.1.100",
	"port": 51820,
	"cpu": 0.0,
	"memory": 0.0,
	"disk": 0.0,
	"network": 0.0,
	"created_at": "2024-01-01T12:00:00Z",
	"updated_at": "2024-01-01T12:00:00Z"
}
```

#### Получение списка серверов

```
GET /api/v1/servers?type=vpn&region=us-east-1&status=running
```

**Описание**: Получение списка серверов с фильтрами  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Query параметры**:

- `type` - тип сервера (vpn, dpi, gateway, analytics)
- `region` - регион сервера
- `status` - статус сервера (creating, running, stopped, error)

**Ответ**:

```json
[
	{
		"id": "server-123",
		"name": "vpn-server-1",
		"type": "vpn",
		"status": "running",
		"region": "us-east-1",
		"ip": "192.168.1.100",
		"port": 51820,
		"cpu": 0.25,
		"memory": 0.15,
		"disk": 0.05,
		"network": 0.1,
		"created_at": "2024-01-01T12:00:00Z",
		"updated_at": "2024-01-01T12:00:00Z"
	}
]
```

#### Управление жизненным циклом

```
POST /api/v1/servers/{id}/start
POST /api/v1/servers/{id}/stop
POST /api/v1/servers/{id}/restart
```

**Описание**: Запуск, остановка и перезапуск сервера  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

#### Статистика сервера

```
GET /api/v1/servers/{id}/stats
```

**Описание**: Получение статистики сервера  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Ответ**:

```json
{
	"server_id": "server-123",
	"cpu": 0.25,
	"memory": 0.15,
	"disk": 0.05,
	"network": 0.1,
	"connections": 150,
	"timestamp": "2024-01-01T12:00:00Z"
}
```

#### Здоровье сервера

```
GET /api/v1/servers/{id}/health
```

**Описание**: Получение информации о здоровье сервера  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Ответ**:

```json
{
	"server_id": "server-123",
	"status": "healthy",
	"message": "Server is running normally",
	"timestamp": "2024-01-01T12:00:00Z"
}
```

### Масштабирование

#### Политики масштабирования

```
GET /api/v1/scaling/policies
POST /api/v1/scaling/policies
PUT /api/v1/scaling/policies/{id}
DELETE /api/v1/scaling/policies/{id}
```

**Описание**: Управление политиками масштабирования  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

**Пример создания политики**:

```json
{
	"name": "vpn-auto-scaling",
	"min_servers": 2,
	"max_servers": 10,
	"cpu_threshold": 0.8,
	"memory_threshold": 0.8,
	"scale_up_cooldown": "300s",
	"scale_down_cooldown": "600s",
	"enabled": true
}
```

#### Оценка масштабирования

```
POST /api/v1/scaling/evaluate
```

**Описание**: Запуск оценки необходимости масштабирования  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

### Резервное копирование

#### Конфигурации резервного копирования

```
GET /api/v1/backups/configs
POST /api/v1/backups/configs
PUT /api/v1/backups/configs/{id}
DELETE /api/v1/backups/configs/{id}
```

**Описание**: Управление конфигурациями резервного копирования  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

**Пример конфигурации**:

```json
{
	"server_id": "server-123",
	"schedule": "0 2 * * *",
	"retention": 30,
	"type": "full",
	"destination": "s3://backups/silence/",
	"enabled": true
}
```

#### Создание и восстановление резервных копий

```
POST /api/v1/servers/{id}/backup
POST /api/v1/servers/{id}/restore/{backup_id}
```

**Описание**: Создание и восстановление резервных копий  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

### Обновления

#### Управление обновлениями

```
GET /api/v1/servers/{id}/update
POST /api/v1/servers/{id}/update
POST /api/v1/servers/{id}/update/cancel
```

**Описание**: Управление обновлениями серверов  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`

**Пример запуска обновления**:

```json
{
	"version": "1.1.0",
	"force": false
}
```

### Мониторинг

#### Здоровье всех серверов

```
GET /api/v1/health/all
```

**Описание**: Получение информации о здоровье всех серверов  
**Заголовки**: `Authorization: Bearer <JWT_TOKEN>`  
**Ответ**:

```json
[
	{
		"server_id": "server-123",
		"status": "healthy",
		"message": "Server is running normally",
		"timestamp": "2024-01-01T12:00:00Z"
	},
	{
		"server_id": "server-456",
		"status": "warning",
		"message": "High CPU usage detected",
		"timestamp": "2024-01-01T12:00:00Z"
	}
]
```

---

## 6. Analytics Service

**Базовый URL**: `http://localhost:8084`

### Health Check

```
GET /health
```

**Описание**: Проверка состояния Analytics сервиса  
**Ответ**:

```json
{
	"status": "ok",
	"service": "analytics",
	"version": "1.0.0"
}
```

### Метрики

#### Метрики подключений

```
GET /metrics/connections
```

**Описание**: Получение метрик подключений VPN  
**Параметры запроса**:

- `start` (string) - Начальное время (RFC3339)
- `end` (string) - Конечное время (RFC3339)
- `user_id` (string) - ID пользователя (опционально)
- `server_id` (string) - ID сервера (опционально)
- `region` (string) - Регион (опционально)

**Ответ**:

```json
{
	"metrics": [
		{
			"name": "connection",
			"type": "counter",
			"value": 1.0,
			"timestamp": "2024-01-01T12:00:00Z",
			"user_id": "user-123",
			"server_id": "server-456",
			"protocol": "wireguard",
			"bypass_type": "shadowsocks",
			"region": "eu-west",
			"duration": 5000,
			"bytes_in": 1024,
			"bytes_out": 2048
		}
	],
	"total": 100,
	"has_more": false
}
```

#### Эффективность обхода DPI

```
GET /metrics/bypass-effectiveness
```

**Описание**: Получение метрик эффективности обхода DPI  
**Параметры запроса**:

- `start` (string) - Начальное время (RFC3339)
- `end` (string) - Конечное время (RFC3339)
- `bypass_type` (string) - Тип обхода (опционально)

**Ответ**:

```json
{
	"metrics": [
		{
			"name": "bypass_effectiveness",
			"type": "gauge",
			"value": 0.95,
			"timestamp": "2024-01-01T12:00:00Z",
			"bypass_type": "shadowsocks",
			"success_rate": 0.95,
			"latency": 50,
			"throughput": 100.0,
			"blocked_count": 5,
			"total_attempts": 100
		}
	],
	"total": 50,
	"has_more": false
}
```

#### Активность пользователей

```
GET /metrics/user-activity
```

**Описание**: Получение метрик активности пользователей  
**Параметры запроса**:

- `start` (string) - Начальное время (RFC3339)
- `end` (string) - Конечное время (RFC3339)
- `user_id` (string) - ID пользователя (опционально)

**Ответ**:

```json
{
	"metrics": [
		{
			"name": "user_activity",
			"type": "counter",
			"value": 1.0,
			"timestamp": "2024-01-01T12:00:00Z",
			"user_id": "user-123",
			"session_count": 3,
			"total_time": 120,
			"data_usage": 512,
			"login_count": 5
		}
	],
	"total": 25,
	"has_more": false
}
```

#### Нагрузка серверов

```
GET /metrics/server-load
```

**Описание**: Получение метрик нагрузки серверов  
**Параметры запроса**:

- `start` (string) - Начальное время (RFC3339)
- `end` (string) - Конечное время (RFC3339)
- `server_id` (string) - ID сервера (опционально)
- `region` (string) - Регион (опционально)

**Ответ**:

```json
{
	"metrics": [
		{
			"name": "server_load",
			"type": "gauge",
			"value": 0.75,
			"timestamp": "2024-01-01T12:00:00Z",
			"server_id": "server-456",
			"region": "eu-west",
			"cpu_usage": 75.0,
			"memory_usage": 60.0,
			"network_in": 50.0,
			"network_out": 30.0,
			"connections": 100
		}
	],
	"total": 10,
	"has_more": false
}
```

#### Метрики ошибок

```
GET /metrics/errors
```

**Описание**: Получение метрик ошибок  
**Параметры запроса**:

- `start` (string) - Начальное время (RFC3339)
- `end` (string) - Конечное время (RFC3339)
- `error_type` (string) - Тип ошибки (опционально)
- `service` (string) - Сервис (опционально)

**Ответ**:

```json
{
	"metrics": [
		{
			"name": "error",
			"type": "counter",
			"value": 1.0,
			"timestamp": "2024-01-01T12:00:00Z",
			"error_type": "connection_timeout",
			"service": "vpn_core",
			"user_id": "user-123",
			"server_id": "server-456",
			"status_code": 500,
			"description": "Connection timeout error"
		}
	],
	"total": 15,
	"has_more": false
}
```

### Дашборды

#### Список дашбордов

```
GET /dashboards
```

**Описание**: Получение списка всех дашбордов  
**Ответ**:

```json
[
	{
		"id": "dashboard-123",
		"name": "VPN Overview",
		"description": "Обзор VPN метрик",
		"created_at": "2024-01-01T12:00:00Z",
		"updated_at": "2024-01-01T12:00:00Z"
	}
]
```

#### Создание дашборда

```
POST /dashboards
```

**Описание**: Создание нового дашборда  
**Тело запроса**:

```json
{
	"name": "VPN Overview",
	"description": "Обзор VPN метрик",
	"widgets": [
		{
			"id": "widget-1",
			"type": "chart",
			"title": "Подключения",
			"query": {
				"time_range": {
					"start": "2024-01-01T00:00:00Z",
					"end": "2024-01-02T00:00:00Z"
				},
				"aggregation": "sum",
				"group_by": ["region"]
			},
			"config": {
				"chart_type": "line"
			},
			"position": {
				"x": 0,
				"y": 0,
				"width": 6,
				"height": 4
			}
		}
	],
	"layout": {
		"columns": 12,
		"rows": 8
	}
}
```

**Ответ**:

```json
{
	"id": "dashboard-123",
	"name": "VPN Overview",
	"description": "Обзор VPN метрик",
	"widgets": [...],
	"layout": {...},
	"created_at": "2024-01-01T12:00:00Z",
	"updated_at": "2024-01-01T12:00:00Z"
}
```

#### Получение дашборда

```
GET /dashboards/{id}
```

**Описание**: Получение дашборда по ID  
**Ответ**: Аналогично созданию дашборда

#### Обновление дашборда

```
PUT /dashboards/{id}
```

**Описание**: Обновление дашборда  
**Тело запроса**: Аналогично созданию дашборда  
**Ответ**: Аналогично созданию дашборда

#### Удаление дашборда

```
DELETE /dashboards/{id}
```

**Описание**: Удаление дашборда  
**Ответ**: `204 No Content`

### Алерты

#### Список алертов

```
GET /alerts
```

**Описание**: Получение списка всех алертов  
**Ответ**:

```json
[
	{
		"id": "alert-123",
		"name": "High Server Load",
		"description": "Высокая нагрузка на сервер",
		"condition": "cpu_usage > 90",
		"severity": "high",
		"status": "active",
		"enabled": true,
		"created_at": "2024-01-01T12:00:00Z",
		"updated_at": "2024-01-01T12:00:00Z"
	}
]
```

#### Создание алерта

```
POST /alerts
```

**Описание**: Создание нового алерта  
**Тело запроса**:

```json
{
	"name": "High Server Load",
	"description": "Высокая нагрузка на сервер",
	"condition": "cpu_usage > 90",
	"severity": "high",
	"message": "CPU usage превышает 90%",
	"enabled": true
}
```

**Ответ**:

```json
{
	"id": "alert-123",
	"name": "High Server Load",
	"description": "Высокая нагрузка на сервер",
	"condition": "cpu_usage > 90",
	"severity": "high",
	"message": "CPU usage превышает 90%",
	"status": "active",
	"enabled": true,
	"created_at": "2024-01-01T12:00:00Z",
	"updated_at": "2024-01-01T12:00:00Z"
}
```

#### Получение алерта

```
GET /alerts/{id}
```

**Описание**: Получение алерта по ID  
**Ответ**: Аналогично созданию алерта

#### Обновление алерта

```
PUT /alerts/{id}
```

**Описание**: Обновление алерта  
**Тело запроса**: Аналогично созданию алерта  
**Ответ**: Аналогично созданию алерта

#### Удаление алерта

```
DELETE /alerts/{id}
```

**Описание**: Удаление алерта  
**Ответ**: `204 No Content`

#### История алертов

```
GET /alerts/{id}/history?limit=10
```

**Описание**: Получение истории срабатываний алерта  
**Параметры запроса**:

- `limit` (int) - Количество записей (по умолчанию 10)

**Ответ**:

```json
[
	{
		"id": "alert-instance-456",
		"rule_id": "alert-123",
		"severity": "high",
		"message": "CPU usage превышает 90%",
		"status": "triggered",
		"created_at": "2024-01-01T12:00:00Z",
		"metric_value": 95.5,
		"server_id": "server-456"
	}
]
```

#### Подтверждение алерта

```
POST /alerts/{id}/acknowledge
```

**Описание**: Подтверждение алерта  
**Ответ**: `200 OK`

#### Разрешение алерта

```
POST /alerts/{id}/resolve
```

**Описание**: Разрешение алерта  
**Ответ**: `200 OK`

---

## 6. Notifications Service

**Базовый URL**: `http://localhost:8086`

### Health Check

```
GET /healthz
```

**Описание**: Проверка состояния Notifications сервиса  
**Ответ**:

```json
"ok"
```

### Отправка уведомлений

```
POST /notifications
```

**Описание**: Отправка уведомления через один или несколько каналов  
**Тело запроса**:

```json
{
	"type": "alert",
	"priority": "high",
	"title": "Важное уведомление",
	"message": "Текст уведомления",
	"recipients": ["user@example.com", "+1234567890"],
	"channels": ["email", "sms"],
	"metadata": {
		"source": "vpn-core",
		"user_id": "123",
		"server_id": "456"
	}
}
```

**Параметры**:

- `type` (string) - Тип уведомления (alert, warning, info, system_alert, etc.)
- `priority` (string) - Приоритет (low, normal, high, urgent)
- `title` (string) - Заголовок уведомления
- `message` (string) - Текст уведомления
- `recipients` (array) - Список получателей (email, телефон, chat_id, etc.)
- `channels` (array) - Каналы доставки (email, sms, push, telegram, slack, webhook)
- `metadata` (object) - Дополнительные данные

**Ответ**:

```json
"ok"
```

### Поддерживаемые типы уведомлений

#### Системные уведомления

- `system_alert` - Системные алерты
- `server_down` - Сервер недоступен
- `server_up` - Сервер восстановлен
- `high_load` - Высокая нагрузка
- `low_disk_space` - Недостаточно места на диске
- `backup_failed` - Ошибка резервного копирования
- `backup_success` - Успешное резервное копирование
- `update_failed` - Ошибка обновления
- `update_success` - Успешное обновление

#### Пользовательские уведомления

- `user_login` - Вход пользователя
- `user_logout` - Выход пользователя
- `user_registered` - Регистрация пользователя
- `user_blocked` - Блокировка пользователя
- `user_unblocked` - Разблокировка пользователя
- `password_reset` - Сброс пароля
- `subscription_expired` - Истечение подписки
- `subscription_renewed` - Продление подписки

#### VPN уведомления

- `vpn_connected` - Подключение к VPN
- `vpn_disconnected` - Отключение от VPN
- `vpn_error` - Ошибка VPN
- `bypass_blocked` - Блокировка обхода DPI
- `bypass_success` - Успешный обход DPI

#### Аналитика уведомления

- `metrics_alert` - Алерт метрик
- `anomaly_detected` - Обнаружена аномалия
- `threshold_exceeded` - Превышен порог

### Поддерживаемые каналы доставки

#### Email

- **Формат получателей**: `user@example.com`
- **Описание**: Отправка по электронной почте

#### SMS

- **Формат получателей**: `+1234567890`
- **Описание**: Отправка SMS сообщений

#### Push

- **Формат получателей**: `device_token_123`
- **Описание**: Push уведомления для мобильных устройств

#### Telegram

- **Формат получателей**: `@username` или `chat_id`
- **Описание**: Отправка через Telegram бота

#### Slack

- **Формат получателей**: `#channel` или `@username`
- **Описание**: Отправка в Slack канал или пользователю

#### Webhook

- **Формат получателей**: `http://example.com/webhook`
- **Описание**: HTTP POST запрос на указанный URL

### Примеры использования

#### Отправка email уведомления

```bash
curl -X POST http://localhost:8086/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "alert",
    "priority": "high",
    "title": "VPN подключение потеряно",
    "message": "Соединение с VPN сервером было прервано",
    "recipients": ["admin@silence.com"],
    "channels": ["email"],
    "metadata": {
      "source": "vpn-core",
      "server_id": "server-123"
    }
  }'
```

#### Отправка SMS уведомления

```bash
curl -X POST http://localhost:8086/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "warning",
    "priority": "medium",
    "title": "Высокая нагрузка",
    "message": "Нагрузка на сервер превышает 80%",
    "recipients": ["+1234567890"],
    "channels": ["sms"],
    "metadata": {
      "source": "analytics",
      "server_id": "server-456"
    }
  }'
```

#### Отправка через несколько каналов

```bash
curl -X POST http://localhost:8086/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "system_alert",
    "priority": "urgent",
    "title": "Критическая ошибка",
    "message": "Обнаружена критическая ошибка в системе",
    "recipients": ["admin@silence.com", "+1234567890", "@admin_user"],
    "channels": ["email", "sms", "telegram"],
    "metadata": {
      "source": "system",
      "error_code": "CRIT_001"
    }
  }'
```

#### Отправка в Slack

```bash
curl -X POST http://localhost:8086/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "info",
    "priority": "normal",
    "title": "Обновление системы",
    "message": "Система будет обновлена в 02:00 UTC",
    "recipients": ["#general"],
    "channels": ["slack"],
    "metadata": {
      "source": "system",
      "maintenance": true
    }
  }'
```

#### Отправка webhook уведомления

```bash
curl -X POST http://localhost:8086/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "type": "webhook_test",
    "priority": "low",
    "title": "Тест интеграции",
    "message": "Проверка webhook интеграции",
    "recipients": ["http://example.com/webhook"],
    "channels": ["webhook"],
    "metadata": {
      "source": "test",
      "test_id": "webhook-001"
    }
  }'
```

### RabbitMQ интеграция

Notifications сервис также поддерживает получение уведомлений через RabbitMQ:

- **Exchange**: `notifications`
- **Queue**: `notifications`
- **Routing Key**: `notifications.*`
- **Consumer Tag**: `notifications-consumer`

#### Формат сообщения RabbitMQ

```json
{
	"id": "event-123",
	"type": "alert",
	"priority": "high",
	"title": "Уведомление из RabbitMQ",
	"message": "Это уведомление получено через RabbitMQ",
	"source": "vpn-core",
	"recipients": ["user@example.com"],
	"channels": ["email"],
	"timestamp": "2024-01-01T12:00:00Z"
}
```

### Интеграция с Analytics

Notifications сервис автоматически отправляет метрики в Analytics сервис:

- **Успешные доставки**: `POST /metrics/delivery`
- **Ошибки доставки**: `POST /metrics/errors`

---

## 7. VPN Core gRPC Service

**Порт**: `50051`

### Сервис: VpnCoreService

#### Health Check

```protobuf
rpc Health(HealthRequest) returns (HealthResponse)
```

#### Управление туннелями

```protobuf
rpc CreateTunnel(CreateTunnelRequest) returns (Tunnel)
rpc GetTunnel(GetTunnelRequest) returns (Tunnel)
rpc ListTunnels(ListTunnelsRequest) returns (ListTunnelsResponse)
rpc DeleteTunnel(DeleteTunnelRequest) returns (DeleteTunnelResponse)
rpc StartTunnel(StartTunnelRequest) returns (StartTunnelResponse)
rpc StopTunnel(StopTunnelRequest) returns (StopTunnelResponse)
rpc GetTunnelStats(GetTunnelStatsRequest) returns (TunnelStats)
```

#### Мониторинг и восстановление

```protobuf
rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse)
rpc EnableAutoRecovery(EnableAutoRecoveryRequest) returns (EnableAutoRecoveryResponse)
rpc DisableAutoRecovery(DisableAutoRecoveryRequest) returns (DisableAutoRecoveryResponse)
rpc RecoverTunnel(RecoverTunnelRequest) returns (RecoverTunnelResponse)
```

#### Управление пирами

```protobuf
rpc AddPeer(AddPeerRequest) returns (Peer)
rpc GetPeer(GetPeerRequest) returns (Peer)
rpc ListPeers(ListPeersRequest) returns (ListPeersResponse)
rpc RemovePeer(RemovePeerRequest) returns (RemovePeerResponse)
```

---

## Аутентификация

### JWT Token

Большинство эндпоинтов требуют аутентификации через JWT токен в заголовке:

```
Authorization: Bearer <JWT_TOKEN>
```

### Получение токена

1. Зарегистрируйтесь: `POST /api/v1/auth/register`
2. Войдите в систему: `POST /api/v1/auth/login`
3. Используйте полученный токен в заголовке Authorization

---

## Коды ответов

- `200 OK` - Успешный запрос
- `201 Created` - Ресурс создан
- `204 No Content` - Успешный запрос без тела ответа
- `400 Bad Request` - Неверный запрос
- `401 Unauthorized` - Не авторизован
- `404 Not Found` - Ресурс не найден
- `405 Method Not Allowed` - Метод не поддерживается
- `500 Internal Server Error` - Внутренняя ошибка сервера

---

## Примеры использования

### Полный пайплайн подключения к VPN с обфускацией

1. **Регистрация пользователя**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

2. **Вход в систему**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

3. **Создание VPN-соединения с обфускацией**

```bash
curl -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "bypass_method": "shadowsocks",
    "bypass_config": {
      "local_port": 1080,
      "remote_host": "proxy.example.com",
      "remote_port": 8388,
      "password": "secret",
      "encryption": "aes-256-gcm"
    },
    "vpn_config": {
      "name": "my-vpn",
      "listen_port": 51820,
      "mtu": 1420,
      "auto_recovery": true
    }
  }'
```

### Управление туннелями

**Создание туннеля**

```bash
curl -X POST http://localhost:8080/api/v1/vpn/tunnels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "name": "my-tunnel",
    "listen_port": 51820,
    "mtu": 1420,
    "auto_recovery": true
  }'
```

**Запуск туннеля**

```bash
curl -X POST "http://localhost:8080/api/v1/vpn/tunnels/start?id=tunnel-123" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Получение статистики**

```bash
curl -X GET "http://localhost:8080/api/v1/vpn/tunnels/stats?id=tunnel-123" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Управление обфускацией

**Создание bypass**

```bash
curl -X POST http://localhost:8080/api/v1/dpi-bypass/bypass \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "name": "my-bypass",
    "method": "shadowsocks",
    "local_port": 1080,
    "remote_host": "proxy.example.com",
    "remote_port": 8388,
    "password": "secret",
    "encryption": "aes-256-gcm"
  }'
```

**Запуск bypass**

```bash
curl -X POST "http://localhost:8080/api/v1/dpi-bypass/bypass/bypass-123/start" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Аналитика и мониторинг

**Получение метрик подключений**

```bash
curl -X GET "http://localhost:8080/api/v1/analytics/metrics/connections?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Получение метрик нагрузки серверов**

```bash
curl -X GET "http://localhost:8080/api/v1/analytics/metrics/server-load?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Создание дашборда**

```bash
curl -X POST http://localhost:8080/api/v1/analytics/dashboards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "name": "VPN Overview",
    "description": "Обзор VPN метрик",
    "widgets": [
      {
        "id": "widget-1",
        "type": "chart",
        "title": "Подключения",
        "query": {
          "time_range": {
            "start": "2024-01-01T00:00:00Z",
            "end": "2024-01-02T00:00:00Z"
          },
          "aggregation": "sum",
          "group_by": ["region"]
        },
        "config": {
          "chart_type": "line"
        },
        "position": {
          "x": 0,
          "y": 0,
          "width": 6,
          "height": 4
        }
      }
    ],
    "layout": {
      "columns": 12,
      "rows": 8
    }
  }'
```

**Создание алерта**

```bash
curl -X POST http://localhost:8080/api/v1/analytics/alerts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "name": "High Server Load",
    "description": "Высокая нагрузка на сервер",
    "condition": "cpu_usage > 90",
    "severity": "high",
    "message": "CPU usage превышает 90%",
    "enabled": true
  }'
```

**Получение списка дашбордов**

```bash
curl -X GET http://localhost:8080/api/v1/analytics/dashboards \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Получение списка алертов**

```bash
curl -X GET http://localhost:8080/api/v1/analytics/alerts \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Уведомления

**Отправка email уведомления**

```bash
curl -X POST http://localhost:8080/api/v1/notifications/notifications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "type": "alert",
    "priority": "high",
    "title": "VPN подключение потеряно",
    "message": "Соединение с VPN сервером было прервано",
    "recipients": ["admin@silence.com"],
    "channels": ["email"],
    "metadata": {
      "source": "vpn-core",
      "server_id": "server-123"
    }
  }'
```

**Отправка SMS уведомления**

```bash
curl -X POST http://localhost:8080/api/v1/notifications/notifications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "type": "warning",
    "priority": "medium",
    "title": "Высокая нагрузка",
    "message": "Нагрузка на сервер превышает 80%",
    "recipients": ["+1234567890"],
    "channels": ["sms"],
    "metadata": {
      "source": "analytics",
      "server_id": "server-456"
    }
  }'
```

**Отправка через несколько каналов**

```bash
curl -X POST http://localhost:8080/api/v1/notifications/notifications \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "type": "system_alert",
    "priority": "urgent",
    "title": "Критическая ошибка",
    "message": "Обнаружена критическая ошибка в системе",
    "recipients": ["admin@silence.com", "+1234567890", "@admin_user"],
    "channels": ["email", "sms", "telegram"],
    "metadata": {
      "source": "system",
      "error_code": "CRIT_001"
    }
  }'
```

**Проверка состояния notifications сервиса**

```bash
curl -X GET http://localhost:8080/api/v1/notifications/health
```

---

## Мониторинг

### Health Checks

Все сервисы предоставляют эндпоинты health check:

- Gateway: `GET /health`
- Auth: `GET /health`
- VPN Core: `GET /health`
- DPI Bypass: `GET /health`
- Analytics: `GET /health`
- Notifications: `GET /healthz`

### Логирование

Все сервисы используют структурированное логирование через zap logger с уровнями:

- `INFO` - Общая информация
- `WARN` - Предупреждения
- `ERROR` - Ошибки
- `DEBUG` - Отладочная информация

### Метрики

VPN Core предоставляет детальную статистику туннелей и пиров через эндпоинты:

- `GET /tunnels/stats` - Статистика туннеля
- `GET /api/v1/bypass/{id}/stats` - Статистика bypass

Analytics сервис предоставляет комплексные метрики и мониторинг:

- `GET /api/v1/analytics/metrics/connections` - Метрики подключений
- `GET /api/v1/analytics/metrics/bypass-effectiveness` - Эффективность обхода DPI
- `GET /api/v1/analytics/metrics/user-activity` - Активность пользователей
- `GET /api/v1/analytics/metrics/server-load` - Нагрузка серверов
- `GET /api/v1/analytics/metrics/errors` - Метрики ошибок
- `GET /api/v1/analytics/dashboards` - Управление дашбордами
- `GET /api/v1/analytics/alerts` - Управление алертами

Notifications сервис предоставляет API для отправки уведомлений:

- `POST /api/v1/notifications/notifications` - Отправка уведомлений
- `GET /api/v1/notifications/health` - Проверка состояния сервиса
