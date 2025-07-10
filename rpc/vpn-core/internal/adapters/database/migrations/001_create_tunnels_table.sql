-- Создание таблицы туннелей VPN
CREATE TABLE IF NOT EXISTS tunnels (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    interface VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    listen_port INTEGER NOT NULL,
    mtu INTEGER NOT NULL DEFAULT 1420,
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    auto_recovery BOOLEAN DEFAULT false,
    recovery_attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_tunnels_status ON tunnels(status);
CREATE INDEX IF NOT EXISTS idx_tunnels_interface ON tunnels(interface);
CREATE INDEX IF NOT EXISTS idx_tunnels_name ON tunnels(name);
CREATE INDEX IF NOT EXISTS idx_tunnels_created_at ON tunnels(created_at);
CREATE INDEX IF NOT EXISTS idx_tunnels_health_status ON tunnels(health_status);
CREATE INDEX IF NOT EXISTS idx_tunnels_auto_recovery ON tunnels(auto_recovery);

-- Создание композитных индексов
CREATE INDEX IF NOT EXISTS idx_tunnels_status_health ON tunnels(status, health_status);
CREATE INDEX IF NOT EXISTS idx_tunnels_recovery_status ON tunnels(auto_recovery, status);

-- Добавление constraints
ALTER TABLE tunnels ADD CONSTRAINT chk_tunnels_status
    CHECK (status IN ('inactive', 'active', 'error', 'recovering'));

ALTER TABLE tunnels ADD CONSTRAINT chk_tunnels_health_status
    CHECK (health_status IN ('healthy', 'unhealthy', 'unknown', 'recovering'));

ALTER TABLE tunnels ADD CONSTRAINT chk_tunnels_listen_port
    CHECK (listen_port > 0 AND listen_port <= 65535);

ALTER TABLE tunnels ADD CONSTRAINT chk_tunnels_mtu
    CHECK (mtu >= 576 AND mtu <= 9000);

ALTER TABLE tunnels ADD CONSTRAINT chk_tunnels_recovery_attempts
    CHECK (recovery_attempts >= 0);

-- Создание уникального индекса для предотвращения дублирования портов
CREATE UNIQUE INDEX IF NOT EXISTS idx_tunnels_listen_port_unique ON tunnels(listen_port);

-- Создание уникального индекса для предотвращения дублирования интерфейсов
CREATE UNIQUE INDEX IF NOT EXISTS idx_tunnels_interface_unique ON tunnels(interface);

-- Комментарии к таблице и колонкам
COMMENT ON TABLE tunnels IS 'Таблица VPN туннелей';
COMMENT ON COLUMN tunnels.id IS 'Уникальный идентификатор туннеля';
COMMENT ON COLUMN tunnels.name IS 'Имя туннеля';
COMMENT ON COLUMN tunnels.interface IS 'Сетевой интерфейс (например, wg0)';
COMMENT ON COLUMN tunnels.status IS 'Статус туннеля (inactive, active, error, recovering)';
COMMENT ON COLUMN tunnels.public_key IS 'Публичный ключ WireGuard';
COMMENT ON COLUMN tunnels.private_key IS 'Приватный ключ WireGuard';
COMMENT ON COLUMN tunnels.listen_port IS 'Порт для прослушивания';
COMMENT ON COLUMN tunnels.mtu IS 'Максимальный размер пакета';
COMMENT ON COLUMN tunnels.last_health_check IS 'Время последней проверки здоровья';
COMMENT ON COLUMN tunnels.health_status IS 'Статус здоровья туннеля';
COMMENT ON COLUMN tunnels.auto_recovery IS 'Включено ли автоматическое восстановление';
COMMENT ON COLUMN tunnels.recovery_attempts IS 'Количество попыток восстановления';
COMMENT ON COLUMN tunnels.created_at IS 'Время создания записи';
COMMENT ON COLUMN tunnels.updated_at IS 'Время последнего обновления записи';
