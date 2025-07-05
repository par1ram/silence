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
