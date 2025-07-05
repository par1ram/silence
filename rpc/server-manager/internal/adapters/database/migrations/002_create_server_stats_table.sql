-- Создание таблицы статистики серверов
CREATE TABLE IF NOT EXISTS server_stats (
    id SERIAL PRIMARY KEY,
    server_id VARCHAR(36) NOT NULL,
    cpu DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    memory DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    disk DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    network DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    connections INTEGER NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

-- Индексы для оптимизации запросов статистики
CREATE INDEX IF NOT EXISTS idx_server_stats_server_id ON server_stats(server_id);
CREATE INDEX IF NOT EXISTS idx_server_stats_timestamp ON server_stats(timestamp);
CREATE INDEX IF NOT EXISTS idx_server_stats_server_timestamp ON server_stats(server_id, timestamp);

-- Партиционирование по времени (для больших объемов данных)
-- CREATE TABLE server_stats_y2024m01 PARTITION OF server_stats
-- FOR VALUES FROM ('2024-01-01') TO ('2024-02-01'); 