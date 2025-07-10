-- Создание таблицы алертов для системы аналитики в ClickHouse
CREATE TABLE IF NOT EXISTS alerts (
    id String,
    type Enum8(
        'system_alert' = 1,
        'server_down' = 2,
        'server_up' = 3,
        'high_load' = 4,
        'low_disk_space' = 5,
        'backup_failed' = 6,
        'backup_success' = 7,
        'update_failed' = 8,
        'update_success' = 9,
        'user_login' = 10,
        'user_logout' = 11,
        'user_registered' = 12,
        'user_blocked' = 13,
        'user_unblocked' = 14,
        'password_reset' = 15,
        'subscription_expired' = 16,
        'subscription_renewed' = 17,
        'vpn_connected' = 18,
        'vpn_disconnected' = 19,
        'vpn_error' = 20,
        'bypass_blocked' = 21,
        'bypass_success' = 22,
        'metrics_alert' = 23,
        'anomaly_detected' = 24,
        'threshold_exceeded' = 25
    ),
    severity Enum8('low' = 1, 'normal' = 2, 'high' = 3, 'urgent' = 4),
    title String,
    message String,
    data Map(String, String),
    source String,
    source_id String,
    acknowledged UInt8 DEFAULT 0,
    acknowledged_by String,
    acknowledged_at DateTime64(3),
    resolved UInt8 DEFAULT 0,
    resolved_by String,
    resolved_at DateTime64(3),
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (severity, created_at)
PARTITION BY toYYYYMM(created_at)
TTL created_at + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для истории изменений алертов
CREATE TABLE IF NOT EXISTS alert_history (
    alert_id String,
    action Enum8('created' = 1, 'acknowledged' = 2, 'resolved' = 3, 'updated' = 4),
    performed_by String,
    old_data Map(String, String),
    new_data Map(String, String),
    timestamp DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (alert_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для статистики алертов по дням
CREATE MATERIALIZED VIEW IF NOT EXISTS alerts_daily_stats_mv
TO alerts_daily_stats AS
SELECT
    toDate(created_at) as date,
    type,
    severity,
    source,
    count() as total_count,
    countIf(acknowledged = 1) as acknowledged_count,
    countIf(resolved = 1) as resolved_count,
    avg(if(resolved = 1, dateDiff('second', created_at, resolved_at), 0)) as avg_resolution_time_seconds
FROM alerts
GROUP BY date, type, severity, source;

-- Создание таблицы для ежедневной статистики алертов
CREATE TABLE IF NOT EXISTS alerts_daily_stats (
    date Date,
    type Enum8(
        'system_alert' = 1,
        'server_down' = 2,
        'server_up' = 3,
        'high_load' = 4,
        'low_disk_space' = 5,
        'backup_failed' = 6,
        'backup_success' = 7,
        'update_failed' = 8,
        'update_success' = 9,
        'user_login' = 10,
        'user_logout' = 11,
        'user_registered' = 12,
        'user_blocked' = 13,
        'user_unblocked' = 14,
        'password_reset' = 15,
        'subscription_expired' = 16,
        'subscription_renewed' = 17,
        'vpn_connected' = 18,
        'vpn_disconnected' = 19,
        'vpn_error' = 20,
        'bypass_blocked' = 21,
        'bypass_success' = 22,
        'metrics_alert' = 23,
        'anomaly_detected' = 24,
        'threshold_exceeded' = 25
    ),
    severity Enum8('low' = 1, 'normal' = 2, 'high' = 3, 'urgent' = 4),
    source String,
    total_count UInt64,
    acknowledged_count UInt64,
    resolved_count UInt64,
    avg_resolution_time_seconds Float64
) ENGINE = SummingMergeTree()
ORDER BY (date, type, severity, source)
PARTITION BY toYYYYMM(date)
TTL date + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для hourly статистики алертов
CREATE MATERIALIZED VIEW IF NOT EXISTS alerts_hourly_stats_mv
TO alerts_hourly_stats AS
SELECT
    toStartOfHour(created_at) as hour,
    type,
    severity,
    source,
    count() as total_count,
    countIf(acknowledged = 1) as acknowledged_count,
    countIf(resolved = 1) as resolved_count,
    countIf(severity IN ('high', 'urgent')) as critical_count
FROM alerts
GROUP BY hour, type, severity, source;

-- Создание таблицы для hourly статистики алертов
CREATE TABLE IF NOT EXISTS alerts_hourly_stats (
    hour DateTime,
    type Enum8(
        'system_alert' = 1,
        'server_down' = 2,
        'server_up' = 3,
        'high_load' = 4,
        'low_disk_space' = 5,
        'backup_failed' = 6,
        'backup_success' = 7,
        'update_failed' = 8,
        'update_success' = 9,
        'user_login' = 10,
        'user_logout' = 11,
        'user_registered' = 12,
        'user_blocked' = 13,
        'user_unblocked' = 14,
        'password_reset' = 15,
        'subscription_expired' = 16,
        'subscription_renewed' = 17,
        'vpn_connected' = 18,
        'vpn_disconnected' = 19,
        'vpn_error' = 20,
        'bypass_blocked' = 21,
        'bypass_success' = 22,
        'metrics_alert' = 23,
        'anomaly_detected' = 24,
        'threshold_exceeded' = 25
    ),
    severity Enum8('low' = 1, 'normal' = 2, 'high' = 3, 'urgent' = 4),
    source String,
    total_count UInt64,
    acknowledged_count UInt64,
    resolved_count UInt64,
    critical_count UInt64
) ENGINE = SummingMergeTree()
ORDER BY (hour, type, severity, source)
PARTITION BY toYYYYMM(hour)
TTL hour + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для real-time алертов (для dashboard)
CREATE TABLE IF NOT EXISTS alerts_realtime (
    id String,
    type Enum8(
        'system_alert' = 1,
        'server_down' = 2,
        'server_up' = 3,
        'high_load' = 4,
        'low_disk_space' = 5,
        'backup_failed' = 6,
        'backup_success' = 7,
        'update_failed' = 8,
        'update_success' = 9,
        'user_login' = 10,
        'user_logout' = 11,
        'user_registered' = 12,
        'user_blocked' = 13,
        'user_unblocked' = 14,
        'password_reset' = 15,
        'subscription_expired' = 16,
        'subscription_renewed' = 17,
        'vpn_connected' = 18,
        'vpn_disconnected' = 19,
        'vpn_error' = 20,
        'bypass_blocked' = 21,
        'bypass_success' = 22,
        'metrics_alert' = 23,
        'anomaly_detected' = 24,
        'threshold_exceeded' = 25
    ),
    severity Enum8('low' = 1, 'normal' = 2, 'high' = 3, 'urgent' = 4),
    title String,
    message String,
    data Map(String, String),
    source String,
    source_id String,
    acknowledged UInt8 DEFAULT 0,
    resolved UInt8 DEFAULT 0,
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(created_at)
ORDER BY (id)
PARTITION BY toYYYYMM(created_at)
TTL created_at + INTERVAL 7 DAY
SETTINGS index_granularity = 8192;

-- Создание таблицы для escalation правил
CREATE TABLE IF NOT EXISTS alert_escalation_rules (
    id String,
    alert_type Enum8(
        'system_alert' = 1,
        'server_down' = 2,
        'server_up' = 3,
        'high_load' = 4,
        'low_disk_space' = 5,
        'backup_failed' = 6,
        'backup_success' = 7,
        'update_failed' = 8,
        'update_success' = 9,
        'user_login' = 10,
        'user_logout' = 11,
        'user_registered' = 12,
        'user_blocked' = 13,
        'user_unblocked' = 14,
        'password_reset' = 15,
        'subscription_expired' = 16,
        'subscription_renewed' = 17,
        'vpn_connected' = 18,
        'vpn_disconnected' = 19,
        'vpn_error' = 20,
        'bypass_blocked' = 21,
        'bypass_success' = 22,
        'metrics_alert' = 23,
        'anomaly_detected' = 24,
        'threshold_exceeded' = 25
    ),
    severity Enum8('low' = 1, 'normal' = 2, 'high' = 3, 'urgent' = 4),
    escalation_timeout_minutes UInt32,
    escalation_target String,
    enabled UInt8 DEFAULT 1,
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(created_at)
ORDER BY (alert_type, severity)
SETTINGS index_granularity = 8192;

-- Создание распределенных таблиц для кластерной конфигурации
-- (раскомментировать при использовании ClickHouse кластера)

-- CREATE TABLE IF NOT EXISTS alerts_distributed AS alerts
-- ENGINE = Distributed(cluster, default, alerts, rand());

-- CREATE TABLE IF NOT EXISTS alert_history_distributed AS alert_history
-- ENGINE = Distributed(cluster, default, alert_history, rand());

-- CREATE TABLE IF NOT EXISTS alerts_realtime_distributed AS alerts_realtime
-- ENGINE = Distributed(cluster, default, alerts_realtime, rand());
