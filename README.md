# Silence VPN

Современный VPN сервис на Go с микросервисной архитектурой и интеграцией WireGuard.

## Архитектура

Проект построен на микросервисной архитектуре с использованием:

- **Go** - основной язык разработки
- **gRPC** - межсервисное взаимодействие
- **WireGuard** - VPN протокол
- **PostgreSQL** - база данных
- **Zap** - структурированное логирование
- **Clean Architecture** - чистая архитектура

### Сервисы

1. **Auth Service** - аутентификация и авторизация
2. **Gateway Service** - API Gateway с проксированием
3. **VPN Core Service** - управление VPN туннелями
4. **DPI Bypass Service** - обход DPI (в разработке)

## WireGuard Интеграция

VPN Core сервис интегрирован с WireGuard для создания и управления VPN туннелями:

### Возможности

- ✅ Создание и удаление WireGuard интерфейсов
- ✅ Управление пирами (добавление/удаление)
- ✅ Генерация ключей WireGuard
- ✅ Мониторинг статистики туннелей
- ✅ HTTP и gRPC API
- ✅ Graceful shutdown

### Технологии

- **wgctrl** - официальная Go библиотека для управления WireGuard
- **Mock адаптер** - для тестирования без прав root
- **Curve25519** - криптографические ключи
- **ChaCha20/Poly1305** - шифрование и аутентификация

## Быстрый старт

### Требования

- Go 1.21+
- WireGuard tools
- PostgreSQL (для Auth сервиса)

### Установка

```bash
# Клонирование
git clone <repository>
cd silence

# Установка зависимостей
task install

# Генерация gRPC кода
task proto:generate

# Сборка всех сервисов
task build:all
```

### Запуск

```bash
# Запуск VPN Core (с mock адаптером)
cd rpc/vpn-core
./bin/vpn-core

# Запуск с реальным WireGuard (требует root)
sudo ./bin/vpn-core
```

### API Примеры

```bash
# Создание туннеля
curl -X POST "http://localhost:8082/tunnels" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-tunnel","listen_port":51820,"mtu":1420}'

# Запуск туннеля
curl -X POST "http://localhost:8082/tunnels/start?id=<tunnel-id>"

# Получение списка туннелей
curl -X GET "http://localhost:8082/tunnels/list"
```

## Разработка

### Структура проекта

```
silence/
├── api/                    # API определения
├── rpc/                    # Микросервисы
│   ├── auth/              # Auth сервис
│   ├── gateway/           # Gateway сервис
│   └── vpn-core/          # VPN Core сервис
├── shared/                # Общий код
├── scripts/               # Скрипты
├── tests/                 # Тесты
└── docs/                  # Документация
```

### Команды разработки

```bash
# Hot reload для разработки
task dev:vpn-core

# Запуск тестов
task test:all

# Проверка кода
task lint:all

# Сборка Docker образов
task docker:build
```

### Принципы разработки

- **Clean Architecture** - разделение на слои
- **SOLID принципы** - объектно-ориентированный дизайн
- **Dependency Injection** - управление зависимостями
- **Interface Segregation** - минимальные интерфейсы
- **Single Responsibility** - одна ответственность

## Документация

- [WireGuard Интеграция](docs/wireguard-integration.md)
- [API Документация](docs/api.md)
- [Архитектура](docs/architecture.md)
- [Развертывание](docs/deployment.md)

## Лицензия

MIT License

## Вклад в проект

1. Fork репозитория
2. Создайте feature branch
3. Внесите изменения
4. Добавьте тесты
5. Создайте Pull Request

## Статус проекта

- [x] Auth Service (базовая функциональность)
- [x] Gateway Service (проксирование)
- [x] VPN Core Service (WireGuard интеграция)
- [ ] DPI Bypass Service
- [ ] Frontend
- [ ] Мониторинг и метрики
- [ ] CI/CD пайплайны
