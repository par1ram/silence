# Server Manager Service - Руководство по использованию

## 📋 Обзор

Server Manager Service - это микросервис для управления VPN серверами в платформе Silence VPN. Он предоставляет API для создания, управления и мониторинга серверов, а также автоматического масштабирования.

## 🏗️ Архитектура

### Основные компоненты
- **gRPC Server**: Основной API сервер
- **Docker Integration**: Управление контейнерами серверов
- **PostgreSQL**: Хранение данных серверов
- **Health Monitoring**: Мониторинг состояния серверов
- **Auto Scaling**: Автоматическое масштабирование

### Схема базы данных
```sql
-- Таблица серверов
CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    ipv4_address INET NOT NULL,
    ipv6_address INET,
    port INTEGER NOT NULL DEFAULT 51820,
    protocol VARCHAR(50) NOT NULL DEFAULT 'wireguard',
    max_connections INTEGER NOT NULL DEFAULT 1000,
    current_connections INTEGER NOT NULL DEFAULT 0,
    cpu_cores INTEGER NOT NULL DEFAULT 2,
    memory_gb INTEGER NOT NULL DEFAULT 4,
    disk_gb INTEGER NOT NULL DEFAULT 50,
    bandwidth_mbps INTEGER NOT NULL DEFAULT 1000,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_ping TIMESTAMP WITH TIME ZONE,
    config JSONB,
    metadata JSONB
);

-- Таблица статистики серверов
CREATE TABLE server_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cpu_usage_percent DECIMAL(5,2) NOT NULL,
    memory_usage_percent DECIMAL(5,2) NOT NULL,
    disk_usage_percent DECIMAL(5,2) NOT NULL,
    network_in_mbps DECIMAL(10,2) NOT NULL DEFAULT 0,
    network_out_mbps DECIMAL(10,2) NOT NULL DEFAULT 0,
    active_connections INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER,
    packet_loss_percent DECIMAL(5,2) DEFAULT 0
);

-- Таблица политик масштабирования
CREATE TABLE scaling_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region VARCHAR(100) NOT NULL,
    min_servers INTEGER NOT NULL DEFAULT 1,
    max_servers INTEGER NOT NULL DEFAULT 10,
    target_cpu_percent DECIMAL(5,2) NOT NULL DEFAULT 70.0,
    target_memory_percent DECIMAL(5,2) NOT NULL DEFAULT 80.0,
    target_connections_percent DECIMAL(5,2) NOT NULL DEFAULT 85.0,
    scale_up_cooldown_minutes INTEGER NOT NULL DEFAULT 10,
    scale_down_cooldown_minutes INTEGER NOT NULL DEFAULT 30,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## 🚀 Быстрый старт

### Запуск сервиса
```bash
# Запуск через Docker Compose
docker-compose up server-manager

# Запуск через Make
make dev-server-manager

# Проверка статуса
curl http://localhost:9085/health
```

### Переменные окружения
```bash
# Основные настройки
GRPC_PORT=8085
LOG_LEVEL=info

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
```

## 📡 API Endpoints

### gRPC методы

#### CreateServer
Создание нового сервера
```protobuf
rpc CreateServer(CreateServerRequest) returns (CreateServerResponse);

message CreateServerRequest {
    string name = 1;
    string region = 2;
    string country = 3;
    string city = 4;
    string protocol = 5;
    int32 max_connections = 6;
    int32 cpu_cores = 7;
    int32 memory_gb = 8;
    int32 disk_gb = 9;
    int32 bandwidth_mbps = 10;
    google.protobuf.Struct config = 11;
    google.protobuf.Struct metadata = 12;
}
```

Пример использования:
```bash
grpcurl -plaintext -d '{
  "name": "US-East-1",
  "region": "us-east",
  "country": "United States",
  "city": "New York",
  "protocol": "wireguard",
  "max_connections": 1000,
  "cpu_cores": 4,
  "memory_gb": 8,
  "disk_gb": 100,
  "bandwidth_mbps": 1000
}' localhost:9085 server.ServerManagerService/CreateServer
```

#### GetServer
Получение информации о сервере
```protobuf
rpc GetServer(GetServerRequest) returns (GetServerResponse);

message GetServerRequest {
    string server_id = 1;
}
```

Пример использования:
```bash
grpcurl -plaintext -d '{
  "server_id": "123e4567-e89b-12d3-a456-426614174000"
}' localhost:9085 server.ServerManagerService/GetServer
```

#### ListServers
Получение списка серверов
```protobuf
rpc ListServers(ListServersRequest) returns (ListServersResponse);

message ListServersRequest {
    string region = 1;
    string country = 2;
    string status = 3;
    int32 limit = 4;
    int32 offset = 5;
}
```

Пример использования:
```bash
grpcurl -plaintext -d '{
  "region": "us-east",
  "status": "active",
  "limit": 10,
  "offset": 0
}' localhost:9085 server.ServerManagerService/ListServers
```

#### UpdateServer
Обновление сервера
```protobuf
rpc UpdateServer(UpdateServerRequest) returns (UpdateServerResponse);

message UpdateServerRequest {
    string server_id = 1;
    string name = 2;
    int32 max_connections = 3;
    string status = 4;
    google.protobuf.Struct config = 5;
    google.protobuf.Struct metadata = 6;
}
```

#### DeleteServer
Удаление сервера
```protobuf
rpc DeleteServer(DeleteServerRequest) returns (DeleteServerResponse);

message DeleteServerRequest {
    string server_id = 1;
    bool force = 2;
}
```

#### StartServer
Запуск сервера
```protobuf
rpc StartServer(StartServerRequest) returns (StartServerResponse);

message StartServerRequest {
    string server_id = 1;
}
```

#### StopServer
Остановка сервера
```protobuf
rpc StopServer(StopServerRequest) returns (StopServerResponse);

message StopServerRequest {
    string server_id = 1;
    bool graceful = 2;
}
```

#### RestartServer
Перезапуск сервера
```protobuf
rpc RestartServer(RestartServerRequest) returns (RestartServerResponse);

message RestartServerRequest {
    string server_id = 1;
    bool graceful = 2;
}
```

#### GetServerStats
Получение статистики сервера
```protobuf
rpc GetServerStats(GetServerStatsRequest) returns (GetServerStatsResponse);

message GetServerStatsRequest {
    string server_id = 1;
    google.protobuf.Timestamp start_time = 2;
    google.protobuf.Timestamp end_time = 3;
    int32 limit = 4;
}
```

#### GetServerLogs
Получение логов сервера
```protobuf
rpc GetServerLogs(GetServerLogsRequest) returns (GetServerLogsResponse);

message GetServerLogsRequest {
    string server_id = 1;
    int32 lines = 2;
    bool follow = 3;
    google.protobuf.Timestamp since = 4;
}
```

## 🐳 Docker интеграция

### Создание сервера
При создании сервера автоматически создается Docker контейнер:

```go
func (s *ServerManager) createDockerContainer(server *Server) error {
    config := &container.Config{
        Image: "silence/vpn-server:latest",
        Env: []string{
            fmt.Sprintf("SERVER_ID=%s", server.ID),
            fmt.Sprintf("SERVER_NAME=%s", server.Name),
            fmt.Sprintf("REGION=%s", server.Region),
            fmt.Sprintf("PROTOCOL=%s", server.Protocol),
            fmt.Sprintf("MAX_CONNECTIONS=%d", server.MaxConnections),
            fmt.Sprintf("PORT=%d", server.Port),
        },
        ExposedPorts: nat.PortSet{
            nat.Port(fmt.Sprintf("%d/udp", server.Port)): struct{}{},
        },
        Labels: map[string]string{
            "silence.server.id":     server.ID,
            "silence.server.name":   server.Name,
            "silence.server.region": server.Region,
            "silence.service":       "vpn-server",
        },
    }

    hostConfig := &container.HostConfig{
        PortBindings: nat.PortMap{
            nat.Port(fmt.Sprintf("%d/udp", server.Port)): []nat.PortBinding{
                {
                    HostIP:   "0.0.0.0",
                    HostPort: fmt.Sprintf("%d", server.Port),
                },
            },
        },
        RestartPolicy: container.RestartPolicy{
            Name: "unless-stopped",
        },
        NetworkMode: "silence_network",
    }

    resp, err := s.dockerClient.ContainerCreate(
        ctx,
        config,
        hostConfig,
        nil,
        nil,
        fmt.Sprintf("silence-vpn-server-%s", server.ID),
    )
    
    if err != nil {
        return fmt.Errorf("failed to create container: %w", err)
    }

    server.ContainerID = resp.ID
    return nil
}
```

### Мониторинг контейнеров
```go
func (s *ServerManager) monitorContainers() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            servers, err := s.repository.ListServers(context.Background(), nil)
            if err != nil {
                s.logger.Error("failed to list servers", "error", err)
                continue
            }

            for _, server := range servers {
                if server.ContainerID == "" {
                    continue
                }

                stats, err := s.getContainerStats(server.ContainerID)
                if err != nil {
                    s.logger.Error("failed to get container stats", 
                        "server_id", server.ID, "error", err)
                    continue
                }

                err = s.repository.SaveServerStats(context.Background(), &ServerStats{
                    ServerID:              server.ID,
                    CPUUsagePercent:      stats.CPUUsagePercent,
                    MemoryUsagePercent:   stats.MemoryUsagePercent,
                    DiskUsagePercent:     stats.DiskUsagePercent,
                    NetworkInMbps:        stats.NetworkInMbps,
                    NetworkOutMbps:       stats.NetworkOutMbps,
                    ActiveConnections:    stats.ActiveConnections,
                    Timestamp:            time.Now(),
                })
                
                if err != nil {
                    s.logger.Error("failed to save server stats", 
                        "server_id", server.ID, "error", err)
                }
            }
        case <-s.ctx.Done():
            return
        }
    }
}
```

## 🔄 Автоматическое масштабирование

### Настройка политик масштабирования
```bash
# Создание политики масштабирования
grpcurl -plaintext -d '{
  "region": "us-east",
  "min_servers": 2,
  "max_servers": 10,
  "target_cpu_percent": 70.0,
  "target_memory_percent": 80.0,
  "target_connections_percent": 85.0,
  "scale_up_cooldown_minutes": 10,
  "scale_down_cooldown_minutes": 30,
  "enabled": true
}' localhost:9085 server.ServerManagerService/CreateScalingPolicy
```

### Алгоритм масштабирования
```go
func (s *ServerManager) checkScaling(region string) error {
    policy, err := s.repository.GetScalingPolicy(context.Background(), region)
    if err != nil || !policy.Enabled {
        return err
    }

    servers, err := s.repository.ListServersByRegion(context.Background(), region)
    if err != nil {
        return err
    }

    activeServers := filterActiveServers(servers)
    
    // Вычисляем средние метрики
    avgCPU := calculateAverageCPU(activeServers)
    avgMemory := calculateAverageMemory(activeServers)
    avgConnections := calculateAverageConnections(activeServers)

    // Проверяем необходимость масштабирования
    if shouldScaleUp(avgCPU, avgMemory, avgConnections, policy) {
        if len(activeServers) < policy.MaxServers {
            return s.scaleUp(region)
        }
    } else if shouldScaleDown(avgCPU, avgMemory, avgConnections, policy) {
        if len(activeServers) > policy.MinServers {
            return s.scaleDown(region)
        }
    }

    return nil
}

func shouldScaleUp(cpu, memory, connections float64, policy *ScalingPolicy) bool {
    return cpu > policy.TargetCPUPercent ||
           memory > policy.TargetMemoryPercent ||
           connections > policy.TargetConnectionsPercent
}

func shouldScaleDown(cpu, memory, connections float64, policy *ScalingPolicy) bool {
    return cpu < policy.TargetCPUPercent*0.5 &&
           memory < policy.TargetMemoryPercent*0.5 &&
           connections < policy.TargetConnectionsPercent*0.5
}
```

## 📊 Мониторинг и метрики

### Основные метрики
```go
var (
    serversTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "silence_servers_total",
            Help: "Total number of servers",
        },
        []string{"region", "status"},
    )

    serverCPUUsage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "silence_server_cpu_usage_percent",
            Help: "Server CPU usage percentage",
        },
        []string{"server_id", "region"},
    )

    serverMemoryUsage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "silence_server_memory_usage_percent",
            Help: "Server memory usage percentage",
        },
        []string{"server_id", "region"},
    )

    serverConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "silence_server_connections",
            Help: "Number of active connections per server",
        },
        []string{"server_id", "region"},
    )

    scalingEvents = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "silence_scaling_events_total",
            Help: "Total number of scaling events",
        },
        []string{"region", "direction"},
    )
)
```

### Health checks
```go
func (s *ServerManager) healthCheck() error {
    // Проверка соединения с базой данных
    if err := s.db.Ping(); err != nil {
        return fmt.Errorf("database connection failed: %w", err)
    }

    // Проверка соединения с Docker
    if _, err := s.dockerClient.Ping(context.Background()); err != nil {
        return fmt.Errorf("docker connection failed: %w", err)
    }

    // Проверка активных серверов
    servers, err := s.repository.ListServers(context.Background(), &ListServersFilter{
        Status: "active",
    })
    if err != nil {
        return fmt.Errorf("failed to list active servers: %w", err)
    }

    unhealthyServers := 0
    for _, server := range servers {
        if !s.isServerHealthy(server) {
            unhealthyServers++
        }
    }

    if len(servers) > 0 && unhealthyServers > len(servers)/2 {
        return fmt.Errorf("too many unhealthy servers: %d/%d", unhealthyServers, len(servers))
    }

    return nil
}
```

## 🔧 Конфигурация

### Конфигурационный файл
```yaml
# server-manager.yaml
server:
  grpc_port: 8085
  log_level: info
  graceful_shutdown_timeout: 30s

database:
  host: postgres
  port: 5432
  user: postgres
  password: password
  name: silence_server_manager
  sslmode: disable
  max_connections: 25
  max_idle_connections: 5
  connection_lifetime: 1h

docker:
  host: unix:///var/run/docker.sock
  api_version: "1.41"
  timeout: 30s
  registry: "silence"
  vpn_server_image: "silence/vpn-server:latest"

scaling:
  enabled: true
  check_interval: 2m
  default_policy:
    min_servers: 1
    max_servers: 5
    target_cpu_percent: 70.0
    target_memory_percent: 80.0
    target_connections_percent: 85.0
    scale_up_cooldown: 10m
    scale_down_cooldown: 30m

monitoring:
  enabled: true
  interval: 30s
  prometheus_port: 9090
  health_check_interval: 1m

logging:
  level: info
  format: json
  output: stdout
```

### Переменные окружения
```bash
# Сервер
SERVER_GRPC_PORT=8085
SERVER_LOG_LEVEL=info

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
DOCKER_REGISTRY=silence
VPN_SERVER_IMAGE=silence/vpn-server:latest

# Масштабирование
SCALING_ENABLED=true
SCALING_CHECK_INTERVAL=2m
DEFAULT_MIN_SERVERS=1
DEFAULT_MAX_SERVERS=5
DEFAULT_TARGET_CPU_PERCENT=70.0
DEFAULT_TARGET_MEMORY_PERCENT=80.0
DEFAULT_TARGET_CONNECTIONS_PERCENT=85.0

# Мониторинг
MONITORING_ENABLED=true
MONITORING_INTERVAL=30s
PROMETHEUS_PORT=9090
HEALTH_CHECK_INTERVAL=1m
```

## 🛠️ Примеры использования

### Создание и запуск сервера
```bash
# 1. Создание сервера
SERVER_ID=$(grpcurl -plaintext -d '{
  "name": "EU-West-1",
  "region": "eu-west",
  "country": "Germany",
  "city": "Frankfurt",
  "protocol": "wireguard",
  "max_connections": 2000,
  "cpu_cores": 8,
  "memory_gb": 16,
  "disk_gb": 200,
  "bandwidth_mbps": 2000
}' localhost:9085 server.ServerManagerService/CreateServer | jq -r '.server.id')

# 2. Запуск сервера
grpcurl -plaintext -d "{\"server_id\": \"$SERVER_ID\"}" \
  localhost:9085 server.ServerManagerService/StartServer

# 3. Проверка статуса
grpcurl -plaintext -d "{\"server_id\": \"$SERVER_ID\"}" \
  localhost:9085 server.ServerManagerService/GetServer
```

### Мониторинг сервера
```bash
# Получение статистики
grpcurl -plaintext -d '{
  "server_id": "'"$SERVER_ID"'",
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-02T00:00:00Z",
  "limit": 100
}' localhost:9085 server.ServerManagerService/GetServerStats

# Получение логов
grpcurl -plaintext -d '{
  "server_id": "'"$SERVER_ID"'",
  "lines": 100,
  "follow": false
}' localhost:9085 server.ServerManagerService/GetServerLogs
```

### Управление масштабированием
```bash
# Создание политики масштабирования
grpcurl -plaintext -d '{
  "region": "eu-west",
  "min_servers": 3,
  "max_servers": 15,
  "target_cpu_percent": 60.0,
  "target_memory_percent": 70.0,
  "target_connections_percent": 80.0,
  "scale_up_cooldown_minutes": 5,
  "scale_down_cooldown_minutes": 15,
  "enabled": true
}' localhost:9085 server.ServerManagerService/CreateScalingPolicy

# Получение политики
grpcurl -plaintext -d '{
  "region": "eu-west"
}' localhost:9085 server.ServerManagerService/GetScalingPolicy
```

## 🚨 Troubleshooting

### Частые проблемы

#### Сервер не запускается
```bash
# Проверка логов
grpcurl -plaintext -d '{
  "server_id": "server-id",
  "lines": 50
}' localhost:9085 server.ServerManagerService/GetServerLogs

# Проверка Docker контейнера
docker logs silence-vpn-server-<server-id>

# Проверка портов
netstat -tlnp | grep 51820
```

#### Высокое потребление ресурсов
```bash
# Получение статистики
grpcurl -plaintext -d '{
  "server_id": "server-id"
}' localhost:9085 server.ServerManagerService/GetServerStats

# Проверка Docker stats
docker stats silence-vpn-server-<server-id>
```

#### Проблемы с автоматическим масштабированием
```bash
# Проверка политики
grpcurl -plaintext -d '{
  "region": "region-name"
}' localhost:9085 server.ServerManagerService/GetScalingPolicy

# Проверка логов сервиса
docker logs silence_server_manager
```

### Диагностические команды
```bash
# Проверка подключения к базе данных
psql -h localhost -U postgres -d silence_server_manager -c "SELECT COUNT(*) FROM servers;"

# Проверка Docker соединения
docker version

# Проверка gRPC сервиса
grpcurl -plaintext localhost:9085 list

# Проверка метрик
curl http://localhost:9090/metrics | grep silence_server
```

## 🔐 Безопасность

### Аутентификация
```go
func (s *ServerManager) validateToken(ctx context.Context) error {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return status.Errorf(codes.Unauthenticated, "missing metadata")
    }

    token := md.Get("authorization")
    if len(token) == 0 {
        return status.Errorf(codes.Unauthenticated, "missing authorization token")
    }

    // Валидация токена
    if !s.auth.ValidateToken(token[0]) {
        return status.Errorf(codes.Unauthenticated, "invalid token")
    }

    return nil
}
```

### Авторизация
```go
func (s *ServerManager) checkPermissions(ctx context.Context, action string, resource string) error {
    userID := getUserIDFromContext(ctx)
    if userID == "" {
        return status.Errorf(codes.Unauthenticated, "user not authenticated")
    }

    allowed, err := s.auth.CheckPermission(userID, action, resource)
    if err != nil {
        return status.Errorf(codes.Internal, "permission check failed: %v", err)
    }

    if !allowed {
        return status.Errorf(codes.PermissionDenied, "access denied")
    }

    return nil
}
```

## 📈 Производительность

### Оптимизация базы данных
```sql
-- Индексы для быстрого поиска
CREATE INDEX idx_servers_region ON servers(region);
CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_country ON servers(country);
CREATE INDEX idx_server_stats_server_id_timestamp ON server_stats(server_id, timestamp);
CREATE INDEX idx_scaling_policies_region ON scaling_policies(region);

-- Партиционирование для server_stats
CREATE TABLE server_stats_y2024m01 PARTITION OF server_stats
FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

### Кэширование
```go
type ServerCache struct {
    cache map[string]*Server
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *ServerCache) Get(serverID string) (*Server, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    server, exists := c.cache[serverID]
    return server, exists
}

func (c *ServerCache) Set(serverID string, server *Server) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.cache[serverID] = server
    
    // Автоматическое удаление через TTL
    time.AfterFunc(c.ttl, func() {
        c.mu.Lock()
        defer c.mu.Unlock()
        delete(c.cache, serverID)
    })
}
```

## 🧪 Тестирование

### Unit тесты
```go
func TestCreateServer(t *testing.T) {
    mockRepo := &MockServerRepository{}
    mockDocker := &MockDockerClient{}
    
    manager := NewServerManager(mockRepo, mockDocker)
    
    request := &CreateServerRequest{
        Name:           "Test Server",
        Region:         "test-region",
        Country:        "Test Country",
        City:           "Test City",
        Protocol:       "wireguard",
        MaxConnections: 1000,
        CpuCores:       2,
        MemoryGb:       4,
        DiskGb:         50,
        BandwidthMbps:  1000,
    }
    
    response, err := manager.CreateServer(context.Background(), request)
    
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.NotEmpty(t, response.Server.Id)
}
```

### Integration тесты
```go
func TestServerLifecycle(t *testing.T) {
    // Создание сервера
    createResp, err := client.CreateServer(ctx, &CreateServerRequest{
        Name:     "Integration Test Server",
        Region:   "test",
        Country:  "Test",
        City:     "Test",
        Protocol: "wireguard",
    })
    require.NoError(t, err)
    
    serverID := createResp.Server.Id
    
    // Запуск сервера
    _, err = client.StartServer(ctx, &StartServerRequest{
        ServerId: serverID,
    })
    require.NoError(t, err)
    
    // Проверка статуса
    getResp, err := client.GetServer(ctx, &GetServerRequest{
        ServerId: serverID,
    })
    require.NoError(t, err)
    assert.Equal(t, "active", getResp.Server.Status)
    
    // Остановка сервера
    _, err = client.StopServer(ctx, &StopServerRequest{
        ServerId: serverID,
    })
    require.NoError(t, err)
    
    // Удаление сервера
    _, err = client.DeleteServer(ctx, &DeleteServerRequest{
        ServerId: serverID,
    })
    require.NoError(t, err)
}
```

## 📚 Дополнительные ресурсы

### Полезные ссылки
- [gRPC Documentation](https://grpc.io/docs/)
- [Docker API Reference](https://docs.docker.com/engine/api/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)

### Примеры кода
- [Server Manager Implementation](../rpc/server-manager/)
- [Proto Definitions](../rpc/server-manager/api/proto/)
- [Database Migrations](../rpc/server-manager/migrations/)

### Мониторинг
- [Grafana Dashboard](../deployments/observability/grafana/dashboards/silence/)
- [Prometheus Alerts](../deployments/observability/alert_rules.yml)
- [Health Check Endpoint](http://localhost:9085/health)

## 🤝 Поддержка

Если у вас есть вопросы или проблемы с Server Manager Service:

1. Проверьте [Troubleshooting](#-troubleshooting) раздел
2. Изучите логи сервиса
3. Создайте issue в репозитории
4. Обратитесь к команде разработки

## 📄 Лицензия

Этот проект лицензирован под MIT License - см. файл [LICENSE](../LICENSE) для деталей.