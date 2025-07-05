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

## 5. VPN Core gRPC Service

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

---

## Мониторинг

### Health Checks

Все сервисы предоставляют эндпоинты health check:

- Gateway: `GET /health`
- Auth: `GET /health`
- VPN Core: `GET /health`
- DPI Bypass: `GET /health`

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
