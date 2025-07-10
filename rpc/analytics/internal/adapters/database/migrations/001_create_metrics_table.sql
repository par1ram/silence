-- Создание таблицы метрик для аналитики в ClickHouse
CREATE TABLE IF NOT EXISTS metrics (
    id String,
    name String,
    type Enum8('counter' = 1, 'gauge' = 2, 'histogram' = 3, 'summary' = 4),
    value Float64,
    unit String,
    tags Map(String, String),
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (name, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для агрегации метрик по минутам
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1min_mv
TO metrics_1min AS
SELECT
    name,
    type,
    unit,
    tags,
    toStartOfMinute(timestamp) as timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as count_value,
    sum(value) as sum_value
FROM metrics
GROUP BY name, type, unit, tags, toStartOfMinute(timestamp);

-- Создание таблицы для агрегированных метрик по минутам
CREATE TABLE IF NOT EXISTS metrics_1min (
    name String,
    type Enum8('counter' = 1, 'gauge' = 2, 'histogram' = 3, 'summary' = 4),
    unit String,
    tags Map(String, String),
    timestamp DateTime,
    avg_value Float64,
    min_value Float64,
    max_value Float64,
    count_value UInt64,
    sum_value Float64
) ENGINE = SummingMergeTree()
ORDER BY (name, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для агрегации метрик по часам
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1hour_mv
TO metrics_1hour AS
SELECT
    name,
    type,
    unit,
    tags,
    toStartOfHour(timestamp) as timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as count_value,
    sum(value) as sum_value
FROM metrics
GROUP BY name, type, unit, tags, toStartOfHour(timestamp);

-- Создание таблицы для агрегированных метрик по часам
CREATE TABLE IF NOT EXISTS metrics_1hour (
    name String,
    type Enum8('counter' = 1, 'gauge' = 2, 'histogram' = 3, 'summary' = 4),
    unit String,
    tags Map(String, String),
    timestamp DateTime,
    avg_value Float64,
    min_value Float64,
    max_value Float64,
    count_value UInt64,
    sum_value Float64
) ENGINE = SummingMergeTree()
ORDER BY (name, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для агрегации метрик по дням
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics_1day_mv
TO metrics_1day AS
SELECT
    name,
    type,
    unit,
    tags,
    toStartOfDay(timestamp) as timestamp,
    avg(value) as avg_value,
    min(value) as min_value,
    max(value) as max_value,
    count() as count_value,
    sum(value) as sum_value
FROM metrics
GROUP BY name, type, unit, tags, toStartOfDay(timestamp);

-- Создание таблицы для агрегированных метрик по дням
CREATE TABLE IF NOT EXISTS metrics_1day (
    name String,
    type Enum8('counter' = 1, 'gauge' = 2, 'histogram' = 3, 'summary' = 4),
    unit String,
    tags Map(String, String),
    timestamp DateTime,
    avg_value Float64,
    min_value Float64,
    max_value Float64,
    count_value UInt64,
    sum_value Float64
) ENGINE = SummingMergeTree()
ORDER BY (name, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 5 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для специализированных метрик подключений
CREATE TABLE IF NOT EXISTS connection_metrics (
    id String,
    user_id String,
    server_id String,
    protocol String,
    bypass_type String,
    region String,
    duration_ms UInt64,
    bytes_in UInt64,
    bytes_out UInt64,
    success UInt8,
    error_code String,
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (user_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 6 MONTH
SETTINGS index_granularity = 8192;

-- Создание таблицы для метрик эффективности обхода DPI
CREATE TABLE IF NOT EXISTS bypass_effectiveness_metrics (
    id String,
    bypass_type String,
    success_rate Float64,
    latency_ms UInt64,
    throughput_mbps Float64,
    blocked_count UInt64,
    total_attempts UInt64,
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (bypass_type, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для метрик активности пользователей
CREATE TABLE IF NOT EXISTS user_activity_metrics (
    id String,
    user_id String,
    session_count UInt64,
    total_time_minutes UInt64,
    data_usage_mb UInt64,
    login_count UInt64,
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (user_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для метрик нагрузки на серверы
CREATE TABLE IF NOT EXISTS server_load_metrics (
    id String,
    server_id String,
    region String,
    cpu_usage_percent Float64,
    memory_usage_percent Float64,
    network_in_mbps Float64,
    network_out_mbps Float64,
    active_connections UInt64,
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (server_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для метрик ошибок
CREATE TABLE IF NOT EXISTS error_metrics (
    id String,
    error_type String,
    service String,
    user_id String,
    server_id String,
    status_code UInt16,
    description String,
    timestamp DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (service, error_type, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание распределенных таблиц для кластерной конфигурации
-- (раскомментировать при использовании ClickHouse кластера)

-- CREATE TABLE IF NOT EXISTS metrics_distributed AS metrics
-- ENGINE = Distributed(cluster, default, metrics, rand());

-- CREATE TABLE IF NOT EXISTS connection_metrics_distributed AS connection_metrics
-- ENGINE = Distributed(cluster, default, connection_metrics, rand());

-- CREATE TABLE IF NOT EXISTS bypass_effectiveness_metrics_distributed AS bypass_effectiveness_metrics
-- ENGINE = Distributed(cluster, default, bypass_effectiveness_metrics, rand());

-- CREATE TABLE IF NOT EXISTS user_activity_metrics_distributed AS user_activity_metrics
-- ENGINE = Distributed(cluster, default, user_activity_metrics, rand());

-- CREATE TABLE IF NOT EXISTS server_load_metrics_distributed AS server_load_metrics
-- ENGINE = Distributed(cluster, default, server_load_metrics, rand());

-- CREATE TABLE IF NOT EXISTS error_metrics_distributed AS error_metrics
-- ENGINE = Distributed(cluster, default, error_metrics, rand());
