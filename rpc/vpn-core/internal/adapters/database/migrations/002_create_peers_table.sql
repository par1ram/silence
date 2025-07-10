-- Создание таблицы пиров VPN туннелей
CREATE TABLE IF NOT EXISTS peers (
    id VARCHAR(36) PRIMARY KEY,
    tunnel_id VARCHAR(36) NOT NULL,
    name VARCHAR(255),
    public_key TEXT NOT NULL,
    allowed_ips TEXT[] NOT NULL,
    endpoint VARCHAR(255),
    persistent_keepalive INTEGER DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    last_handshake TIMESTAMP WITH TIME ZONE,
    transfer_rx BIGINT DEFAULT 0,
    transfer_tx BIGINT DEFAULT 0,
    last_seen TIMESTAMP WITH TIME ZONE,
    connection_quality DOUBLE PRECISION DEFAULT 0,
    latency INTERVAL DEFAULT '0 milliseconds',
    packet_loss DOUBLE PRECISION DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Внешний ключ к таблице туннелей
    CONSTRAINT fk_peers_tunnel_id
        FOREIGN KEY (tunnel_id)
        REFERENCES tunnels(id)
        ON DELETE CASCADE
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_peers_tunnel_id ON peers(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_peers_status ON peers(status);
CREATE INDEX IF NOT EXISTS idx_peers_public_key ON peers(public_key);
CREATE INDEX IF NOT EXISTS idx_peers_endpoint ON peers(endpoint);
CREATE INDEX IF NOT EXISTS idx_peers_created_at ON peers(created_at);
CREATE INDEX IF NOT EXISTS idx_peers_last_handshake ON peers(last_handshake);
CREATE INDEX IF NOT EXISTS idx_peers_last_seen ON peers(last_seen);
CREATE INDEX IF NOT EXISTS idx_peers_connection_quality ON peers(connection_quality);

-- Создание композитных индексов
CREATE INDEX IF NOT EXISTS idx_peers_tunnel_status ON peers(tunnel_id, status);
CREATE INDEX IF NOT EXISTS idx_peers_tunnel_last_seen ON peers(tunnel_id, last_seen);
CREATE INDEX IF NOT EXISTS idx_peers_status_quality ON peers(status, connection_quality);

-- Добавление constraints
ALTER TABLE peers ADD CONSTRAINT chk_peers_status
    CHECK (status IN ('inactive', 'active', 'error', 'offline'));

ALTER TABLE peers ADD CONSTRAINT chk_peers_persistent_keepalive
    CHECK (persistent_keepalive >= 0 AND persistent_keepalive <= 65535);

ALTER TABLE peers ADD CONSTRAINT chk_peers_transfer_rx
    CHECK (transfer_rx >= 0);

ALTER TABLE peers ADD CONSTRAINT chk_peers_transfer_tx
    CHECK (transfer_tx >= 0);

ALTER TABLE peers ADD CONSTRAINT chk_peers_connection_quality
    CHECK (connection_quality >= 0 AND connection_quality <= 1);

ALTER TABLE peers ADD CONSTRAINT chk_peers_packet_loss
    CHECK (packet_loss >= 0 AND packet_loss <= 1);

-- Создание уникального индекса для предотвращения дублирования публичных ключей в рамках туннеля
CREATE UNIQUE INDEX IF NOT EXISTS idx_peers_tunnel_public_key_unique
    ON peers(tunnel_id, public_key);

-- Создание частичного индекса для активных пиров
CREATE INDEX IF NOT EXISTS idx_peers_active_quality
    ON peers(connection_quality, last_seen)
    WHERE status = 'active';

-- Создание частичного индекса для пиров с проблемами
CREATE INDEX IF NOT EXISTS idx_peers_problematic
    ON peers(tunnel_id, last_seen, packet_loss)
    WHERE status IN ('error', 'offline') OR packet_loss > 0.1;

-- Комментарии к таблице и колонкам
COMMENT ON TABLE peers IS 'Таблица пиров VPN туннелей';
COMMENT ON COLUMN peers.id IS 'Уникальный идентификатор пира';
COMMENT ON COLUMN peers.tunnel_id IS 'Идентификатор туннеля, к которому принадлежит пир';
COMMENT ON COLUMN peers.name IS 'Имя пира (опционально)';
COMMENT ON COLUMN peers.public_key IS 'Публичный ключ пира WireGuard';
COMMENT ON COLUMN peers.allowed_ips IS 'Массив разрешенных IP адресов';
COMMENT ON COLUMN peers.endpoint IS 'Конечная точка пира (IP:port)';
COMMENT ON COLUMN peers.persistent_keepalive IS 'Интервал keepalive в секундах';
COMMENT ON COLUMN peers.status IS 'Статус пира (inactive, active, error, offline)';
COMMENT ON COLUMN peers.last_handshake IS 'Время последнего handshake';
COMMENT ON COLUMN peers.transfer_rx IS 'Количество полученных байт';
COMMENT ON COLUMN peers.transfer_tx IS 'Количество отправленных байт';
COMMENT ON COLUMN peers.last_seen IS 'Время последней активности';
COMMENT ON COLUMN peers.connection_quality IS 'Качество соединения (0.0 - 1.0)';
COMMENT ON COLUMN peers.latency IS 'Задержка соединения';
COMMENT ON COLUMN peers.packet_loss IS 'Процент потерянных пакетов (0.0 - 1.0)';
COMMENT ON COLUMN peers.created_at IS 'Время создания записи';
COMMENT ON COLUMN peers.updated_at IS 'Время последнего обновления записи';

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_peers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создание триггера для автоматического обновления updated_at
CREATE TRIGGER trigger_peers_updated_at
    BEFORE UPDATE ON peers
    FOR EACH ROW
    EXECUTE FUNCTION update_peers_updated_at();
