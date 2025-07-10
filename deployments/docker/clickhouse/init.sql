-- ClickHouse initialization script for Silence VPN Analytics

-- Create database
CREATE DATABASE IF NOT EXISTS silence_analytics;

-- Use the database
USE silence_analytics;

-- Create metrics table for storing various metrics
CREATE TABLE IF NOT EXISTS metrics (
    id UUID DEFAULT generateUUIDv4(),
    service_name String,
    metric_type String,
    metric_name String,
    metric_value Float64,
    labels Map(String, String),
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (service_name, metric_type, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create events table for storing application events
CREATE TABLE IF NOT EXISTS events (
    id UUID DEFAULT generateUUIDv4(),
    service_name String,
    event_type String,
    event_name String,
    user_id String,
    session_id String,
    properties Map(String, String),
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (service_name, event_type, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create performance_metrics table for system performance data
CREATE TABLE IF NOT EXISTS performance_metrics (
    id UUID DEFAULT generateUUIDv4(),
    service_name String,
    instance_id String,
    cpu_usage Float64,
    memory_usage Float64,
    disk_usage Float64,
    network_in Float64,
    network_out Float64,
    response_time Float64,
    error_rate Float64,
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (service_name, instance_id, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create vpn_sessions table for VPN session analytics
CREATE TABLE IF NOT EXISTS vpn_sessions (
    id UUID DEFAULT generateUUIDv4(),
    user_id String,
    session_id String,
    server_id String,
    server_location String,
    start_time DateTime,
    end_time DateTime,
    duration UInt64,
    bytes_sent UInt64,
    bytes_received UInt64,
    protocol String,
    client_ip String,
    server_ip String,
    disconnect_reason String,
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (user_id, session_id, start_time)
PARTITION BY toYYYYMM(start_time);

-- Create user_analytics table for user behavior analytics
CREATE TABLE IF NOT EXISTS user_analytics (
    id UUID DEFAULT generateUUIDv4(),
    user_id String,
    action String,
    resource String,
    ip_address String,
    user_agent String,
    country String,
    city String,
    device_type String,
    os String,
    browser String,
    referrer String,
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (user_id, action, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create alerts table for storing alert data
CREATE TABLE IF NOT EXISTS alerts (
    id UUID DEFAULT generateUUIDv4(),
    alert_type String,
    severity String,
    service_name String,
    message String,
    details Map(String, String),
    resolved Boolean DEFAULT false,
    resolved_at DateTime,
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (alert_type, severity, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create server_stats table for server performance tracking
CREATE TABLE IF NOT EXISTS server_stats (
    id UUID DEFAULT generateUUIDv4(),
    server_id String,
    server_name String,
    server_location String,
    active_connections UInt64,
    total_connections UInt64,
    bandwidth_used UInt64,
    cpu_usage Float64,
    memory_usage Float64,
    disk_usage Float64,
    uptime UInt64,
    status String,
    timestamp DateTime DEFAULT now(),
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (server_id, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Create materialized views for common queries

-- Daily metrics aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_metrics_mv
ENGINE = SummingMergeTree()
ORDER BY (service_name, metric_type, metric_name, date)
AS SELECT
    service_name,
    metric_type,
    metric_name,
    toDate(timestamp) as date,
    avg(metric_value) as avg_value,
    min(metric_value) as min_value,
    max(metric_value) as max_value,
    count() as count
FROM metrics
GROUP BY service_name, metric_type, metric_name, toDate(timestamp);

-- Hourly performance aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_performance_mv
ENGINE = SummingMergeTree()
ORDER BY (service_name, instance_id, hour)
AS SELECT
    service_name,
    instance_id,
    toStartOfHour(timestamp) as hour,
    avg(cpu_usage) as avg_cpu,
    avg(memory_usage) as avg_memory,
    avg(disk_usage) as avg_disk,
    avg(response_time) as avg_response_time,
    avg(error_rate) as avg_error_rate,
    count() as count
FROM performance_metrics
GROUP BY service_name, instance_id, toStartOfHour(timestamp);

-- User session summary
CREATE MATERIALIZED VIEW IF NOT EXISTS user_session_summary_mv
ENGINE = SummingMergeTree()
ORDER BY (user_id, date)
AS SELECT
    user_id,
    toDate(start_time) as date,
    count() as session_count,
    sum(duration) as total_duration,
    sum(bytes_sent + bytes_received) as total_bytes,
    uniq(server_id) as unique_servers
FROM vpn_sessions
GROUP BY user_id, toDate(start_time);

-- Server utilization summary
CREATE MATERIALIZED VIEW IF NOT EXISTS server_utilization_mv
ENGINE = SummingMergeTree()
ORDER BY (server_id, hour)
AS SELECT
    server_id,
    server_name,
    server_location,
    toStartOfHour(timestamp) as hour,
    avg(active_connections) as avg_connections,
    max(active_connections) as max_connections,
    avg(bandwidth_used) as avg_bandwidth,
    avg(cpu_usage) as avg_cpu,
    avg(memory_usage) as avg_memory,
    count() as count
FROM server_stats
GROUP BY server_id, server_name, server_location, toStartOfHour(timestamp);

-- Create indexes for better query performance
-- Note: ClickHouse doesn't use traditional indexes, but we can create bloom filter indexes

-- Add bloom filter indexes for frequently queried string columns
ALTER TABLE metrics ADD INDEX IF NOT EXISTS idx_service_name service_name TYPE bloom_filter GRANULARITY 1;
ALTER TABLE metrics ADD INDEX IF NOT EXISTS idx_metric_type metric_type TYPE bloom_filter GRANULARITY 1;

ALTER TABLE events ADD INDEX IF NOT EXISTS idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE events ADD INDEX IF NOT EXISTS idx_session_id session_id TYPE bloom_filter GRANULARITY 1;

ALTER TABLE vpn_sessions ADD INDEX IF NOT EXISTS idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE vpn_sessions ADD INDEX IF NOT EXISTS idx_server_id server_id TYPE bloom_filter GRANULARITY 1;

ALTER TABLE user_analytics ADD INDEX IF NOT EXISTS idx_user_id user_id TYPE bloom_filter GRANULARITY 1;
ALTER TABLE user_analytics ADD INDEX IF NOT EXISTS idx_action action TYPE bloom_filter GRANULARITY 1;

ALTER TABLE server_stats ADD INDEX IF NOT EXISTS idx_server_id server_id TYPE bloom_filter GRANULARITY 1;

-- Database initialization completed successfully
-- Sample data can be inserted via application API calls
