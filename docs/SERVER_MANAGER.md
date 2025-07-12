# Server Manager Service Documentation

## Обзор

Server Manager Service - это микросервис для управления серверами в системе Silence VPN. Он предоставляет функциональность для создания, управления, мониторинга и масштабирования серверов различных типов.

## Архитектура

### Основные компоненты

- **GRPC API** - основной интерфейс для взаимодействия с сервисом
- **Database Layer** - PostgreSQL для хранения метаданных серверов
- **Docker Integration** - управление контейнерами через Docker API
- **Monitoring** - сбор метрик и состояния серверов

### Типы серверов

1. **VPN Server** - основные VPN серверы
2. **DPI Bypass Server** - серверы для обхода блокировок
3. **Gateway Server** - шлюзовые серверы
4. **Analytics Server** - серверы аналитики

## API Методы

### Health Check

```protobuf
rpc Health(HealthRequest) returns (HealthResponse);
```

Проверка состояния сервиса.

### Управление серверами

#### CreateServer
```protobuf
rpc CreateServer(CreateServerRequest) returns (Server);
```

Создание нового сервера.

**Параметры:**
- `name` - имя сервера
- `type` - тип сервера (VPN, DPI, Gateway, Analytics)
- `region` - регион размещения
- `config` - конфигурация сервера

**Пример запроса:**
```json
{
  "name": "vpn-server-eu-1",
  "type": "SERVER_TYPE_VPN",
  "region": "eu-central-1",
  "config": {
    "cpu": "2",
    "memory": "4096",
    "disk": "50",
    "vpn_protocol": "wireguard"
  }
}
```

#### GetServer
```protobuf
rpc GetServer(GetServerRequest) returns (Server);
```

Получение информации о сервере по ID.

#### ListServers
```protobuf
rpc ListServers(ListServersRequest) returns (ListServersResponse);
```

Получение списка серверов с фильтрацией.

**Фильтры:**
- По типу сервера
- По статусу
- По региону
- Пагинация (limit/offset)

#### UpdateServer
```protobuf
rpc UpdateServer(UpdateServerRequest) returns (Server);
```

Обновление конфигурации сервера.

#### DeleteServer
```protobuf
rpc DeleteServer(DeleteServerRequest) returns (DeleteServerResponse);
```

Удаление сервера.

### Операции с серверами

#### StartServer
```protobuf
rpc StartServer(StartServerRequest) returns (StartServerResponse);
```

Запуск сервера.

#### StopServer
```protobuf
rpc StopServer(StopServerRequest) returns (StopServerResponse);
```

Остановка сервера.

#### RestartServer
```protobuf
rpc RestartServer(RestartServerRequest) returns (RestartServerResponse);
```

Перезапуск сервера.

### Мониторинг

#### GetServerStats
```protobuf
rpc GetServerStats(GetServerStatsRequest) returns (ServerStats);
```

Получение статистики сервера (CPU, память, диск, сеть).

#### GetServerHealth
```protobuf
rpc GetServerHealth(GetServerHealthRequest) returns (ServerHealth);
```

Проверка здоровья сервера.

#### MonitorServer
```protobuf
rpc MonitorServer(MonitorServerRequest) returns (stream ServerMonitorEvent);
```

Потоковый мониторинг сервера (WebSocket-подобный).

### Масштабирование

#### ScaleServer
```protobuf
rpc ScaleServer(ScaleServerRequest) returns (ScaleServerResponse);
```

Масштабирование ресурсов сервера.

**Действия:**
- `SCALE_ACTION_UP` - увеличение ресурсов
- `SCALE_ACTION_DOWN` - уменьшение ресурсов
- `SCALE_ACTION_AUTO` - автоматическое масштабирование

#### CreateBackup
```protobuf
rpc CreateBackup(CreateBackupRequest) returns (CreateBackupResponse);
```

Создание резервной копии сервера.

#### RestoreBackup
```protobuf
rpc RestoreBackup(RestoreBackupRequest) returns (RestoreBackupResponse);
```

Восстановление из резервной копии.

## Конфигурация

### Переменные окружения

```bash
# GRPC порт
GRPC_PORT=8085

# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=silence_server_manager
DB_SSLMODE=disable

# Docker
DOCKER_HOST=unix:///var/run/docker.sock
DOCKER_API_VERSION=1.41
DOCKER_TIMEOUT=30s

# Миграции
MIGRATIONS_DIR=/app/migrations

# Логирование
LOG_LEVEL=info
```

### Конфигурация сервера

Пример конфигурации для VPN сервера:

```yaml
server:
  name: "vpn-server-eu-1"
  type: "VPN"
  region: "eu-central-1"
  
resources:
  cpu: "2"
  memory: "4096"
  disk: "50"
  
vpn:
  protocol: "wireguard"
  port: 51820
  max_clients: 100
  
network:
  subnet: "10.0.0.0/24"
  dns:
    - "1.1.1.1"
    - "8.8.8.8"
```

## Развертывание

### Docker

```bash
# Сборка образа
docker build -t silence/server-manager:latest .

# Запуск контейнера
docker run -d \
  --name server-manager \
  -p 8085:8085 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=password \
  silence/server-manager:latest
```

### Kubernetes

```bash
# Применение манифестов
kubectl apply -f deployments/kubernetes/11-server-manager.yaml

# Проверка статуса
kubectl get pods -n silence -l component=server-manager

# Просмотр логов
kubectl logs -f deployment/server-manager -n silence
```

## Мониторинг

### Метрики

Сервис экспортирует следующие метрики:

- `server_manager_servers_total` - общее количество серверов
- `server_manager_servers_by_status` - количество серверов по статусу
- `server_manager_servers_by_type` - количество серверов по типу
- `server_manager_operations_total` - общее количество операций
- `server_manager_operation_duration` - время выполнения операций

### Health Checks

```bash
# Проверка здоровья через grpc_health_probe
grpc_health_probe -addr=localhost:8085

# Проверка через curl (если есть HTTP endpoint)
curl http://localhost:8085/health
```

## Клиентские SDK

### Go Client

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/par1ram/silence/rpc/server-manager/api/proto"
)

func main() {
    conn, err := grpc.Dial("localhost:8085", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewServerManagerServiceClient(conn)
    
    // Создание сервера
    req := &pb.CreateServerRequest{
        Name:   "test-server",
        Type:   pb.ServerType_SERVER_TYPE_VPN,
        Region: "eu-central-1",
        Config: map[string]string{
            "cpu":    "2",
            "memory": "4096",
        },
    }
    
    server, err := client.CreateServer(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created server: %s", server.Id)
}
```

### JavaScript Client

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const packageDefinition = protoLoader.loadSync('server.proto');
const proto = grpc.loadPackageDefinition(packageDefinition);

const client = new proto.server.ServerManagerService(
  'localhost:8085',
  grpc.credentials.createInsecure()
);

// Создание сервера
const request = {
  name: 'test-server',
  type: 'SERVER_TYPE_VPN',
  region: 'eu-central-1',
  config: {
    cpu: '2',
    memory: '4096'
  }
};

client.createServer(request, (error, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('Created server:', response.id);
  }
});
```

## Безопасность

### Аутентификация

Сервис использует JWT токены для аутентификации запросов:

```bash
# Добавление токена в metadata
grpcurl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"name":"test-server"}' \
  localhost:8085 server.ServerManagerService/CreateServer
```

### Авторизация

Проверка прав доступа происходит на уровне методов:

- `CreateServer` - требует роль `admin`
- `DeleteServer` - требует роль `admin`
- `GetServer` - требует роль `user` или `admin`
- `ListServers` - требует роль `user` или `admin`

## Troubleshooting

### Частые проблемы

1. **Сервис не запускается**
   ```bash
   # Проверка логов
   kubectl logs deployment/server-manager -n silence
   
   # Проверка переменных окружения
   kubectl describe pod server-manager-xxx -n silence
   ```

2. **Ошибка подключения к Docker**
   ```bash
   # Проверка доступности Docker socket
   ls -la /var/run/docker.sock
   
   # Проверка прав доступа
   docker ps
   ```

3. **Ошибка подключения к БД**
   ```bash
   # Проверка подключения к PostgreSQL
   kubectl exec -it postgres-xxx -n silence -- psql -U postgres -c "\l"
   ```

### Логирование

Настройка уровня логирования:

```bash
# Для отладки
LOG_LEVEL=debug

# Для продакшена
LOG_LEVEL=info
```

## Roadmap

### Запланированные функции

- [ ] Автоматическое масштабирование на основе нагрузки
- [ ] Интеграция с облачными провайдерами (AWS, GCP, Azure)
- [ ] Поддержка Kubernetes deployments
- [ ] Расширенный мониторинг и алерты
- [ ] Backup в S3/MinIO
- [ ] Rolling updates без простоя

### Известные ограничения

- Поддержка только Docker контейнеров
- Один сервер на один контейнер
- Отсутствие автоматического failover
- Ограниченная поддержка сетевых конфигураций

## Поддержка

Для получения поддержки:

1. Создайте issue в GitHub репозитории
2. Приложите логи сервиса
3. Опишите шаги для воспроизведения проблемы
4. Укажите версию сервиса и окружение

## Лицензия

MIT License - см. файл LICENSE в корне проекта.