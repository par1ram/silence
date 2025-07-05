# Система Silence VPN с обфускацией

## Обзор проекта

Silence - это микросервисная система VPN с интегрированной обфускацией трафика для обхода DPI (Deep Packet Inspection). Система состоит из четырех основных сервисов, работающих в едином пайплайне.

## Архитектура системы

### Сервисы

1. **Gateway** (порт 8080) - единая точка входа для пользователей
2. **Auth** (порт 8081) - аутентификация и авторизация пользователей
3. **VPN Core** (порт 8082) - управление WireGuard туннелями
4. **DPI Bypass** (порт 8083) - обфускация трафика

### Диаграмма системы

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Gateway   │    │    Auth     │    │  VPN Core   │    │ DPI Bypass  │
│   :8080     │◄──►│   :8081     │    │   :8082     │◄──►│   :8083     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Прокси    │    │   JWT Auth  │    │ WireGuard   │    │ Обфускация  │
│   запросов  │    │   + DB      │    │   туннели   │    │   трафика   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

## Пайплайн подключения пользователя

### 1. Регистрация/Аутентификация

```bash
# Регистрация пользователя
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass", "email": "user@example.com"}'

# Вход пользователя
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'
```

**Результат**: JWT токен для аутентифицированных запросов

### 2. Подключение к VPN с обфускацией

```bash
curl -X POST http://localhost:8080/api/v1/connect \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "bypass_method": "shadowsocks",
    "bypass_config": {
      "local_port": 1080,
      "remote_host": "127.0.0.1",
      "remote_port": 8388,
      "password": "testpass",
      "encryption": "AES-256-GCM"
    },
    "vpn_config": {
      "name": "my-vpn",
      "listen_port": 51820,
      "mtu": 1420,
      "auto_recovery": true
    }
  }'
```

**Результат**:

```json
{
	"bypass_id": "1751741446062401000",
	"bypass_port": 1080,
	"vpn_tunnel": "1751741446063320000",
	"status": "connected",
	"created_at": "2025-07-05T21:50:46.063482+03:00"
}
```

## Что происходит внутри

### Gateway (единая точка входа)

1. **Проксирование запросов** на соответствующие сервисы
2. **JWT аутентификация** для защищенных эндпоинтов
3. **Интеграция VPN + обфускация** через `/api/v1/connect`

### Auth (аутентификация)

1. **Регистрация пользователей** с хешированием паролей
2. **JWT токены** для сессий
3. **PostgreSQL** для хранения пользователей

### VPN Core (WireGuard туннели)

1. **Создание WireGuard интерфейсов** через `wgctrl`
2. **Управление туннелями** (создание/запуск/остановка)
3. **Мониторинг статистики** туннелей
4. **Управление пирами** в туннелях

### DPI Bypass (обфускация)

1. **Поддерживаемые методы**:

   - **Shadowsocks** - SOCKS5 прокси с шифрованием
   - **V2Ray** - многофункциональный прокси
   - **Obfs4** - обфускация трафика
   - **Custom** - кастомные методы (chaff, fragment, timing, hybrid)

2. **Создание bypass-конфигураций** с уникальными ID
3. **Запуск/остановка** обфусцированных соединений
4. **Мониторинг статистики** bypass-соединений

## Методы обфускации

### Shadowsocks

- **Протокол**: SOCKS5 с шифрованием
- **Алгоритмы**: AES-256-GCM, ChaCha20-Poly1305
- **Использование**: для обхода простых DPI

### V2Ray

- **Протокол**: VMess, VLESS, Trojan
- **Транспорт**: WebSocket, HTTP/2, QUIC
- **Использование**: для обхода продвинутых DPI

### Obfs4

- **Протокол**: обфускация Tor
- **Методы**: obfs4, meek
- **Использование**: для обхода блокировок Tor

### Custom

- **Chaff**: добавление ложного трафика
- **Fragment**: фрагментация пакетов
- **Timing**: изменение временных интервалов
- **Hybrid**: комбинация методов

## API эндпоинты

### Gateway (прокси)

```
GET    /health                    # Health check
POST   /api/v1/auth/register     # Регистрация
POST   /api/v1/auth/login        # Вход
POST   /api/v1/connect           # VPN + обфускация
GET    /api/v1/dpi-bypass/*      # Прокси на DPI Bypass
GET    /api/v1/vpn/*             # Прокси на VPN Core
```

### DPI Bypass

```
POST   /api/v1/bypass            # Создать bypass
GET    /api/v1/bypass            # Список bypass
GET    /api/v1/bypass/{id}       # Получить bypass
DELETE /api/v1/bypass/{id}       # Удалить bypass
POST   /api/v1/bypass/{id}/start # Запустить bypass
POST   /api/v1/bypass/{id}/stop  # Остановить bypass
GET    /api/v1/bypass/{id}/stats # Статистика bypass
```

### VPN Core

```
POST   /tunnels                  # Создать туннель
GET    /tunnels/list             # Список туннелей
GET    /tunnels/get?id={id}      # Получить туннель
POST   /tunnels/start?id={id}    # Запустить туннель
POST   /tunnels/stop?id={id}     # Остановить туннель
GET    /tunnels/stats?id={id}    # Статистика туннеля
POST   /peers/add                # Добавить пира
GET    /peers/list?tunnel_id={id} # Список пиров
DELETE /peers/remove             # Удалить пира
```

## Запуск системы

### Разработка

```bash
# Запуск всех сервисов
task dev:auth      # Auth сервис
task dev:dpi-bypass # DPI Bypass сервис
task dev:vpn-core   # VPN Core сервис
task dev:gateway    # Gateway сервис
```

### Продакшн

```bash
# Docker Compose
docker-compose up -d

# Kubernetes
kubectl apply -f deployments/kubernetes/
```

## Конфигурация

### Переменные окружения

```bash
# Gateway
GATEWAY_PORT=8080
GATEWAY_JWT_SECRET=your-secret
GATEWAY_AUTH_URL=http://localhost:8081
GATEWAY_VPN_URL=http://localhost:8082
GATEWAY_DPI_URL=http://localhost:8083

# Auth
AUTH_PORT=8081
AUTH_JWT_SECRET=your-secret
AUTH_DB_HOST=localhost
AUTH_DB_PORT=5432
AUTH_DB_NAME=silence_auth

# VPN Core
VPN_CORE_PORT=8082
VPN_CORE_GRPC_PORT=9092
WIREGUARD_DIR=/etc/wireguard

# DPI Bypass
DPI_BYPASS_PORT=8083
```

## Мониторинг

### Логи

Все сервисы используют структурированное логирование с полями:

- `service` - имя сервиса
- `level` - уровень логирования
- `timestamp` - временная метка
- `message` - сообщение

### Метрики

- **Gateway**: количество запросов, время ответа
- **Auth**: количество пользователей, успешные/неуспешные логины
- **VPN Core**: количество туннелей, пиров, трафик
- **DPI Bypass**: количество bypass-соединений, статистика обфускации

## Безопасность

### Аутентификация

- **JWT токены** с временем жизни
- **Хеширование паролей** bcrypt
- **HTTPS** в продакшне

### Шифрование

- **WireGuard**: Curve25519, ChaCha20, Poly1305
- **Shadowsocks**: AES-256-GCM, ChaCha20-Poly1305
- **V2Ray**: TLS 1.3, различные шифры

### Сетевая безопасность

- **Изоляция сервисов** в отдельных контейнерах
- **Firewall правила** для ограничения доступа
- **VPN туннели** для шифрования трафика

## Troubleshooting

### Проблемы с подключением

```bash
# Проверка сервисов
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health

# Проверка логов
tail -f logs/gateway.log
tail -f logs/auth.log
tail -f logs/vpn-core.log
tail -f logs/dpi-bypass.log
```

### Проблемы с WireGuard

```bash
# Проверка интерфейсов
sudo wg show

# Проверка модуля ядра
lsmod | grep wireguard

# Перезапуск WireGuard
sudo systemctl restart wg-quick@wg0
```

### Проблемы с обфускацией

```bash
# Проверка bypass-соединений
curl http://localhost:8080/api/v1/dpi-bypass/bypass

# Проверка статистики
curl http://localhost:8080/api/v1/dpi-bypass/bypass/{id}/stats
```

## Будущие улучшения

1. **Мобильные приложения** для iOS/Android
2. **Веб-интерфейс** для управления
3. **Кластерный режим** для высокой доступности
4. **Интеграция с CDN** для глобального покрытия
5. **Машинное обучение** для выбора оптимального метода обфускации
6. **Поддержка IPv6** и новых протоколов
7. **Интеграция с DNS-over-HTTPS**
8. **Автоматическое переключение** между методами обфускации

---

# Интеграция WireGuard в VPN Core

## Обзор

VPN Core сервис интегрирован с WireGuard для создания и управления VPN туннелями. Интеграция использует официальную Go библиотеку `wgctrl` для управления WireGuard интерфейсами.

## Архитектура

### Компоненты

1. **WireGuardManager** - интерфейс для управления WireGuard интерфейсами
2. **WGAdapter** - реализация с использованием `wgctrl`
3. **MockWGAdapter** - mock реализация для тестирования
4. **TunnelService** - сервис туннелей с интеграцией WireGuard

### Диаграмма

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP/gRPC     │    │   TunnelService  │    │   WGAdapter     │
│   Handlers      │───▶│                  │───▶│   (wgctrl)      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │   WireGuard      │
                       │   Kernel Module  │
                       └──────────────────┘
```

## Реализация

### WireGuardManager Interface

```go
type WireGuardManager interface {
    CreateInterface(name, privateKey string, listenPort, mtu int) error
    DeleteInterface(name string) error
    AddPeer(deviceName, publicKey string, allowedIPs []net.IPNet, endpoint *net.UDPAddr, keepalive int) error
    RemovePeer(deviceName, publicKey string) error
    GetDeviceStats(deviceName string) (interface{}, error)
    Close() error
}
```

### WGAdapter (wgctrl)

Основная реализация использует `golang.zx2c4.com/wireguard/wgctrl`:

- **Создание интерфейса**: `wgctrl.Client.ConfigureDevice()`
- **Управление пирами**: добавление/удаление через конфигурацию
- **Статистика**: получение статистики устройства
- **Безопасность**: использование приватных/публичных ключей

### MockWGAdapter

Mock реализация для тестирования без создания реальных интерфейсов:

- Логирует все операции
- Не создает реальные интерфейсы
- Возвращает mock статистику

## Использование

### Создание туннеля

```bash
curl -X POST "http://localhost:8082/tunnels" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-tunnel",
    "listen_port": 51820,
    "mtu": 1420
  }'
```

### Запуск туннеля

```bash
curl -X POST "http://localhost:8082/tunnels/start?id=<tunnel-id>"
```

### Остановка туннеля

```bash
curl -X POST "http://localhost:8082/tunnels/stop?id=<tunnel-id>"
```

### Получение статистики

```bash
curl -X GET "http://localhost:8082/tunnels/stats?id=<tunnel-id>"
```

## Конфигурация

### Переменные окружения

```bash
# WireGuard конфигурация
WIREGUARD_DIR=/etc/wireguard          # Директория конфигураций
WIREGUARD_INTERFACE=wg0               # Имя интерфейса по умолчанию
WIREGUARD_LISTEN_PORT=51820           # Порт по умолчанию
WIREGUARD_MTU=1420                    # MTU по умолчанию
```

### Требования

1. **WireGuard установлен**:

   ```bash
   # macOS
   brew install wireguard-tools

   # Ubuntu/Debian
   sudo apt install wireguard

   # CentOS/RHEL
   sudo yum install wireguard-tools
   ```

2. **Права root** (для создания интерфейсов):
   ```bash
   sudo ./bin/vpn-core
   ```

## Безопасность

### Ключи

- **Приватные ключи**: генерируются автоматически, хранятся в памяти
- **Публичные ключи**: передаются между пирами
- **Алгоритм**: Curve25519 для ключей

### Сетевая безопасность

- **Шифрование**: ChaCha20 для шифрования
- **Аутентификация**: Poly1305 для MAC
- **Протокол**: UDP с надежной доставкой

## Тестирование

### Mock режим

Для тестирования без прав root используйте MockWGAdapter:

```go
wgAdapter := wireguard.NewMockWGAdapter(logger)
```

### Интеграционные тесты

```bash
# Запуск с mock адаптером
./bin/vpn-core

# Тестирование API
curl -X POST "http://localhost:8082/tunnels" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","listen_port":51820,"mtu":1420}'
```

## Мониторинг

### Логи

Все операции WireGuard логируются:

```
{"level":"info","msg":"wireguard interface created","name":"wg0","port":51820}
{"level":"info","msg":"peer added to wireguard interface","device":"wg0","public_key":"..."}
```

### Метрики

- Количество активных туннелей
- Количество пиров на туннель
- Трафик (bytes rx/tx)
- Статус интерфейсов

## Troubleshooting

### Проблемы с правами

```bash
# Ошибка: permission denied
sudo ./bin/vpn-core
```

### Интерфейс не создается

```bash
# Проверка WireGuard
which wg

# Проверка модуля ядра
lsmod | grep wireguard
```

### Проблемы с сетью

```bash
# Проверка интерфейса
ip link show wg0

# Проверка маршрутов
ip route show
```

## Будущие улучшения

1. **Поддержка конфигурационных файлов**
2. **Интеграция с системой управления ключами**
3. **Мониторинг в реальном времени**
4. **Автоматическое восстановление соединений**
5. **Поддержка IPv6**
6. **Интеграция с DNS**
