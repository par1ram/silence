-- Создание таблицы серверов
CREATE TABLE IF NOT EXISTS servers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'creating',
    region VARCHAR(100) NOT NULL,
    ip VARCHAR(45),
    port INTEGER,
    cpu DOUBLE PRECISION DEFAULT 0.0,
    memory DOUBLE PRECISION DEFAULT 0.0,
    disk DOUBLE PRECISION DEFAULT 0.0,
    network DOUBLE PRECISION DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_servers_type ON servers(type);
CREATE INDEX IF NOT EXISTS idx_servers_status ON servers(status);
CREATE INDEX IF NOT EXISTS idx_servers_region ON servers(region);
CREATE INDEX IF NOT EXISTS idx_servers_created_at ON servers(created_at);
CREATE INDEX IF NOT EXISTS idx_servers_deleted_at ON servers(deleted_at);

-- Композитный индекс для фильтрации
CREATE INDEX IF NOT EXISTS idx_servers_type_status ON servers(type, status);
CREATE INDEX IF NOT EXISTS idx_servers_region_status ON servers(region, status); 