-- Создание таблицы статистики туннелей VPN
CREATE TABLE IF NOT EXISTS tunnel_stats (
    id SERIAL PRIMARY KEY,
    tunnel_id VARCHAR(36) NOT NULL,
    bytes_rx BIGINT DEFAULT 0,
    bytes_tx BIGINT DEFAULT 0,
    peers_count INTEGER DEFAULT 0,
    active_peers INTEGER DEFAULT 0,
    uptime INTERVAL DEFAULT '0 seconds',
    error_count INTEGER DEFAULT 0,
    recovery_count INTEGER DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Внешний ключ к таблице туннелей
    CONSTRAINT fk_tunnel_stats_tunnel_id
        FOREIGN KEY (tunnel_id)
        REFERENCES tunnels(id)
        ON DELETE CASCADE
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_tunnel_id ON tunnel_stats(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_timestamp ON tunnel_stats(timestamp);
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_tunnel_timestamp ON tunnel_stats(tunnel_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_created_at ON tunnel_stats(created_at);

-- Создание композитных индексов для аналитики
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_tunnel_time_desc ON tunnel_stats(tunnel_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_active_peers ON tunnel_stats(active_peers, timestamp);
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_errors ON tunnel_stats(error_count, timestamp) WHERE error_count > 0;

-- Добавление constraints
ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_bytes_rx
    CHECK (bytes_rx >= 0);

ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_bytes_tx
    CHECK (bytes_tx >= 0);

ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_peers_count
    CHECK (peers_count >= 0);

ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_active_peers
    CHECK (active_peers >= 0);

ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_error_count
    CHECK (error_count >= 0);

ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_recovery_count
    CHECK (recovery_count >= 0);

-- Constraint для проверки что активных пиров не больше общего количества
ALTER TABLE tunnel_stats ADD CONSTRAINT chk_tunnel_stats_active_peers_limit
    CHECK (active_peers <= peers_count);

-- Создание частичного индекса для недавних записей (последние 24 часа)
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_recent
    ON tunnel_stats(tunnel_id, timestamp DESC)
    WHERE timestamp >= NOW() - INTERVAL '24 hours';

-- Создание частичного индекса для записей с ошибками
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_with_errors
    ON tunnel_stats(tunnel_id, timestamp, error_count)
    WHERE error_count > 0;

-- Комментарии к таблице и колонкам
COMMENT ON TABLE tunnel_stats IS 'Таблица статистики VPN туннелей';
COMMENT ON COLUMN tunnel_stats.id IS 'Уникальный идентификатор записи статистики';
COMMENT ON COLUMN tunnel_stats.tunnel_id IS 'Идентификатор туннеля';
COMMENT ON COLUMN tunnel_stats.bytes_rx IS 'Количество полученных байт';
COMMENT ON COLUMN tunnel_stats.bytes_tx IS 'Количество отправленных байт';
COMMENT ON COLUMN tunnel_stats.peers_count IS 'Общее количество пиров';
COMMENT ON COLUMN tunnel_stats.active_peers IS 'Количество активных пиров';
COMMENT ON COLUMN tunnel_stats.uptime IS 'Время работы туннеля';
COMMENT ON COLUMN tunnel_stats.error_count IS 'Количество ошибок';
COMMENT ON COLUMN tunnel_stats.recovery_count IS 'Количество восстановлений';
COMMENT ON COLUMN tunnel_stats.timestamp IS 'Время сбора статистики';
COMMENT ON COLUMN tunnel_stats.created_at IS 'Время создания записи';

-- Создание функции для автоматической очистки старых записей (старше 30 дней)
CREATE OR REPLACE FUNCTION cleanup_old_tunnel_stats()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM tunnel_stats
    WHERE timestamp < NOW() - INTERVAL '30 days';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    -- Логируем количество удаленных записей
    RAISE NOTICE 'Cleaned up % old tunnel stats records', deleted_count;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Создание функции для получения последней статистики туннеля
CREATE OR REPLACE FUNCTION get_latest_tunnel_stats(tunnel_uuid VARCHAR(36))
RETURNS TABLE (
    tunnel_id VARCHAR(36),
    bytes_rx BIGINT,
    bytes_tx BIGINT,
    peers_count INTEGER,
    active_peers INTEGER,
    uptime INTERVAL,
    error_count INTEGER,
    recovery_count INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ts.tunnel_id,
        ts.bytes_rx,
        ts.bytes_tx,
        ts.peers_count,
        ts.active_peers,
        ts.uptime,
        ts.error_count,
        ts.recovery_count,
        ts.timestamp
    FROM tunnel_stats ts
    WHERE ts.tunnel_id = tunnel_uuid
    ORDER BY ts.timestamp DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Создание функции для получения агрегированной статистики за период
CREATE OR REPLACE FUNCTION get_tunnel_stats_aggregated(
    tunnel_uuid VARCHAR(36),
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    tunnel_id VARCHAR(36),
    avg_bytes_rx BIGINT,
    avg_bytes_tx BIGINT,
    max_bytes_rx BIGINT,
    max_bytes_tx BIGINT,
    avg_peers_count INTEGER,
    max_peers_count INTEGER,
    avg_active_peers INTEGER,
    max_active_peers INTEGER,
    total_uptime INTERVAL,
    total_errors INTEGER,
    total_recoveries INTEGER,
    records_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        tunnel_uuid,
        AVG(ts.bytes_rx)::BIGINT,
        AVG(ts.bytes_tx)::BIGINT,
        MAX(ts.bytes_rx),
        MAX(ts.bytes_tx),
        AVG(ts.peers_count)::INTEGER,
        MAX(ts.peers_count),
        AVG(ts.active_peers)::INTEGER,
        MAX(ts.active_peers),
        SUM(ts.uptime),
        SUM(ts.error_count),
        SUM(ts.recovery_count),
        COUNT(*)::INTEGER
    FROM tunnel_stats ts
    WHERE ts.tunnel_id = tunnel_uuid
      AND ts.timestamp BETWEEN start_time AND end_time
    GROUP BY ts.tunnel_id;
END;
$$ LANGUAGE plpgsql;

-- Создание представления для последней статистики всех туннелей
CREATE OR REPLACE VIEW latest_tunnel_stats AS
SELECT DISTINCT ON (ts.tunnel_id)
    ts.tunnel_id,
    t.name as tunnel_name,
    t.status as tunnel_status,
    ts.bytes_rx,
    ts.bytes_tx,
    ts.peers_count,
    ts.active_peers,
    ts.uptime,
    ts.error_count,
    ts.recovery_count,
    ts.timestamp
FROM tunnel_stats ts
JOIN tunnels t ON ts.tunnel_id = t.id
ORDER BY ts.tunnel_id, ts.timestamp DESC;

COMMENT ON VIEW latest_tunnel_stats IS 'Представление с последней статистикой всех туннелей';

-- Создание индекса для представления
CREATE INDEX IF NOT EXISTS idx_tunnel_stats_view_support
    ON tunnel_stats(tunnel_id, timestamp DESC);
