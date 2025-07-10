-- Создание таблицы конфигураций дашбордов для системы аналитики в ClickHouse
CREATE TABLE IF NOT EXISTS dashboard_configs (
    id String,
    name String,
    description String,
    owner_id String,
    is_public UInt8 DEFAULT 0,
    is_favorite UInt8 DEFAULT 0,
    tags Array(String),
    config String, -- JSON конфигурация дашборда
    layout String, -- JSON конфигурация layout
    refresh_interval UInt32 DEFAULT 300, -- в секундах
    time_range_start String DEFAULT '-1h', -- относительное время
    time_range_end String DEFAULT 'now',
    version UInt32 DEFAULT 1,
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (id)
SETTINGS index_granularity = 8192;

-- Создание таблицы для виджетов дашборда
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id String,
    dashboard_id String,
    widget_type Enum8(
        'chart' = 1,
        'metric' = 2,
        'table' = 3,
        'gauge' = 4,
        'counter' = 5,
        'heatmap' = 6,
        'pie' = 7,
        'bar' = 8,
        'line' = 9,
        'area' = 10,
        'scatter' = 11,
        'histogram' = 12,
        'text' = 13,
        'alert_list' = 14,
        'log_panel' = 15
    ),
    title String,
    description String,
    position_x UInt16,
    position_y UInt16,
    width UInt16,
    height UInt16,
    query String, -- SQL запрос для данных
    query_type Enum8('sql' = 1, 'prometheus' = 2, 'custom' = 3),
    data_source String,
    config String, -- JSON конфигурация виджета
    refresh_interval UInt32 DEFAULT 300,
    is_visible UInt8 DEFAULT 1,
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (dashboard_id, id)
SETTINGS index_granularity = 8192;

-- Создание таблицы для пользовательских настроек дашбордов
CREATE TABLE IF NOT EXISTS dashboard_user_preferences (
    user_id String,
    dashboard_id String,
    is_favorite UInt8 DEFAULT 0,
    custom_time_range String,
    custom_refresh_interval UInt32,
    custom_layout String, -- JSON кастомного layout
    last_viewed_at DateTime64(3) DEFAULT now64(),
    view_count UInt64 DEFAULT 1,
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (user_id, dashboard_id)
SETTINGS index_granularity = 8192;

-- Создание таблицы для истории просмотров дашбордов
CREATE TABLE IF NOT EXISTS dashboard_view_history (
    id String,
    dashboard_id String,
    user_id String,
    view_duration_seconds UInt32,
    widgets_interacted Array(String),
    filters_applied String, -- JSON фильтров
    timestamp DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (dashboard_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
TTL timestamp + INTERVAL 6 MONTH
SETTINGS index_granularity = 8192;

-- Создание материализованного представления для статистики использования дашбордов
CREATE MATERIALIZED VIEW IF NOT EXISTS dashboard_usage_stats_mv
TO dashboard_usage_stats AS
SELECT
    dashboard_id,
    toDate(timestamp) as date,
    count() as total_views,
    uniq(user_id) as unique_users,
    avg(view_duration_seconds) as avg_view_duration,
    sum(view_duration_seconds) as total_view_duration
FROM dashboard_view_history
GROUP BY dashboard_id, date;

-- Создание таблицы для статистики использования дашбордов
CREATE TABLE IF NOT EXISTS dashboard_usage_stats (
    dashboard_id String,
    date Date,
    total_views UInt64,
    unique_users UInt64,
    avg_view_duration Float64,
    total_view_duration UInt64
) ENGINE = SummingMergeTree()
ORDER BY (dashboard_id, date)
PARTITION BY toYYYYMM(date)
TTL date + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для шаблонов дашбордов
CREATE TABLE IF NOT EXISTS dashboard_templates (
    id String,
    name String,
    description String,
    category Enum8(
        'system' = 1,
        'network' = 2,
        'security' = 3,
        'performance' = 4,
        'user_activity' = 5,
        'server_monitoring' = 6,
        'vpn_analytics' = 7,
        'custom' = 8
    ),
    template_config String, -- JSON конфигурация шаблона
    preview_image String, -- URL превью изображения
    is_builtin UInt8 DEFAULT 0,
    usage_count UInt64 DEFAULT 0,
    rating Float32 DEFAULT 0.0,
    created_by String,
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (id)
SETTINGS index_granularity = 8192;

-- Создание таблицы для шаринга дашбордов
CREATE TABLE IF NOT EXISTS dashboard_shares (
    id String,
    dashboard_id String,
    share_token String,
    shared_by String,
    shared_with String, -- может быть пустым для публичных ссылок
    permissions Enum8('read' = 1, 'write' = 2, 'admin' = 3),
    expires_at DateTime64(3),
    is_active UInt8 DEFAULT 1,
    access_count UInt64 DEFAULT 0,
    last_accessed_at DateTime64(3),
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(created_at)
ORDER BY (share_token)
SETTINGS index_granularity = 8192;

-- Создание таблицы для снэпшотов дашбордов
CREATE TABLE IF NOT EXISTS dashboard_snapshots (
    id String,
    dashboard_id String,
    snapshot_name String,
    snapshot_data String, -- JSON данных снэпшота
    created_by String,
    expires_at DateTime64(3),
    is_public UInt8 DEFAULT 0,
    created_at DateTime64(3) DEFAULT now64()
) ENGINE = MergeTree()
ORDER BY (dashboard_id, created_at)
PARTITION BY toYYYYMM(created_at)
TTL created_at + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для комментариев и аннотаций дашбордов
CREATE TABLE IF NOT EXISTS dashboard_annotations (
    id String,
    dashboard_id String,
    widget_id String, -- может быть пустым для аннотаций всего дашборда
    annotation_type Enum8(
        'comment' = 1,
        'alert' = 2,
        'deployment' = 3,
        'incident' = 4,
        'maintenance' = 5,
        'custom' = 6
    ),
    title String,
    description String,
    tags Array(String),
    time_start DateTime64(3),
    time_end DateTime64(3),
    created_by String,
    is_visible UInt8 DEFAULT 1,
    created_at DateTime64(3) DEFAULT now64(),
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (dashboard_id, time_start)
PARTITION BY toYYYYMM(time_start)
TTL time_start + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Создание таблицы для системных настроек дашбордов
CREATE TABLE IF NOT EXISTS dashboard_system_settings (
    key String,
    value String,
    description String,
    category Enum8(
        'general' = 1,
        'security' = 2,
        'performance' = 3,
        'ui' = 4,
        'notifications' = 5
    ),
    is_readonly UInt8 DEFAULT 0,
    updated_by String,
    updated_at DateTime64(3) DEFAULT now64()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (key)
SETTINGS index_granularity = 8192;

-- Создание распределенных таблиц для кластерной конфигурации
-- (раскомментировать при использовании ClickHouse кластера)

-- CREATE TABLE IF NOT EXISTS dashboard_configs_distributed AS dashboard_configs
-- ENGINE = Distributed(cluster, default, dashboard_configs, rand());

-- CREATE TABLE IF NOT EXISTS dashboard_widgets_distributed AS dashboard_widgets
-- ENGINE = Distributed(cluster, default, dashboard_widgets, rand());

-- CREATE TABLE IF NOT EXISTS dashboard_view_history_distributed AS dashboard_view_history
-- ENGINE = Distributed(cluster, default, dashboard_view_history, rand());

-- CREATE TABLE IF NOT EXISTS dashboard_templates_distributed AS dashboard_templates
-- ENGINE = Distributed(cluster, default, dashboard_templates, rand());

-- Создание индексов для оптимизации поиска
-- (ClickHouse использует минимальные и максимальные значения для каждого блока)

-- Вставка базовых системных настроек
INSERT INTO dashboard_system_settings (key, value, description, category, is_readonly) VALUES
('max_dashboards_per_user', '50', 'Максимальное количество дашбордов на пользователя', 'general', 0),
('default_refresh_interval', '300', 'Интервал обновления по умолчанию (секунды)', 'general', 0),
('max_widgets_per_dashboard', '100', 'Максимальное количество виджетов на дашборд', 'general', 0),
('enable_public_dashboards', '1', 'Разрешить публичные дашборды', 'security', 0),
('snapshot_retention_days', '365', 'Время хранения снэпшотов (дни)', 'general', 0),
('max_query_timeout', '30', 'Максимальное время выполнения запроса (секунды)', 'performance', 0),
('enable_dashboard_comments', '1', 'Разрешить комментарии к дашбордам', 'ui', 0),
('max_dashboard_history', '100', 'Максимальное количество записей в истории дашборда', 'general', 0);

-- Вставка базовых шаблонов дашбордов
INSERT INTO dashboard_templates (id, name, description, category, template_config, is_builtin, created_by) VALUES
('tpl_system_overview', 'System Overview', 'Общий обзор системы', 'system', '{"widgets": [{"type": "metric", "title": "CPU Usage"}, {"type": "metric", "title": "Memory Usage"}, {"type": "chart", "title": "Network Traffic"}]}', 1, 'system'),
('tpl_vpn_analytics', 'VPN Analytics', 'Аналитика VPN подключений', 'vpn_analytics', '{"widgets": [{"type": "chart", "title": "Active Connections"}, {"type": "pie", "title": "Traffic by Region"}, {"type": "table", "title": "Top Users"}]}', 1, 'system'),
('tpl_security_monitor', 'Security Monitor', 'Мониторинг безопасности', 'security', '{"widgets": [{"type": "alert_list", "title": "Security Alerts"}, {"type": "chart", "title": "Failed Logins"}, {"type": "heatmap", "title": "Attack Sources"}]}', 1, 'system'),
('tpl_performance', 'Performance Dashboard', 'Мониторинг производительности', 'performance', '{"widgets": [{"type": "gauge", "title": "Response Time"}, {"type": "chart", "title": "Throughput"}, {"type": "histogram", "title": "Latency Distribution"}]}', 1, 'system');
