-- Создание таблицы уведомлений
CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    channels TEXT[] NOT NULL,
    recipients TEXT[] NOT NULL,
    source VARCHAR(100),
    source_id VARCHAR(36),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_priority ON notifications(priority);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_source ON notifications(source);
CREATE INDEX IF NOT EXISTS idx_notifications_source_id ON notifications(source_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_updated_at ON notifications(updated_at);
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled_at ON notifications(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_notifications_sent_at ON notifications(sent_at);

-- Создание композитных индексов
CREATE INDEX IF NOT EXISTS idx_notifications_type_priority ON notifications(type, priority);
CREATE INDEX IF NOT EXISTS idx_notifications_status_priority ON notifications(status, priority);
CREATE INDEX IF NOT EXISTS idx_notifications_source_type ON notifications(source, type);
CREATE INDEX IF NOT EXISTS idx_notifications_status_created ON notifications(status, created_at DESC);

-- Создание GIN индексов для массивов и JSONB
CREATE INDEX IF NOT EXISTS idx_notifications_channels ON notifications USING GIN(channels);
CREATE INDEX IF NOT EXISTS idx_notifications_recipients ON notifications USING GIN(recipients);
CREATE INDEX IF NOT EXISTS idx_notifications_data ON notifications USING GIN(data);

-- Создание частичных индексов для активных уведомлений
CREATE INDEX IF NOT EXISTS idx_notifications_pending ON notifications(priority, created_at DESC)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_notifications_sending ON notifications(attempts, created_at DESC)
    WHERE status = 'sending';

CREATE INDEX IF NOT EXISTS idx_notifications_failed ON notifications(attempts, created_at DESC)
    WHERE status = 'failed';

CREATE INDEX IF NOT EXISTS idx_notifications_scheduled ON notifications(scheduled_at)
    WHERE status = 'pending' AND scheduled_at IS NOT NULL;

-- Создание индекса для приоритетных уведомлений
CREATE INDEX IF NOT EXISTS idx_notifications_high_priority ON notifications(created_at DESC)
    WHERE priority IN ('high', 'urgent');

-- Добавление constraints
ALTER TABLE notifications ADD CONSTRAINT chk_notifications_type
    CHECK (type IN ('system_alert', 'server_down', 'server_up', 'high_load', 'low_disk_space',
                    'backup_failed', 'backup_success', 'update_failed', 'update_success',
                    'user_login', 'user_logout', 'user_registered', 'user_blocked', 'user_unblocked',
                    'password_reset', 'subscription_expired', 'subscription_renewed',
                    'vpn_connected', 'vpn_disconnected', 'vpn_error', 'bypass_blocked', 'bypass_success',
                    'metrics_alert', 'anomaly_detected', 'threshold_exceeded'));

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_priority
    CHECK (priority IN ('low', 'normal', 'high', 'urgent'));

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_status
    CHECK (status IN ('pending', 'sending', 'sent', 'delivered', 'failed', 'cancelled'));

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_attempts
    CHECK (attempts >= 0);

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_max_attempts
    CHECK (max_attempts > 0);

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_title_not_empty
    CHECK (LENGTH(TRIM(title)) > 0);

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_message_not_empty
    CHECK (LENGTH(TRIM(message)) > 0);

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_channels_not_empty
    CHECK (array_length(channels, 1) > 0);

ALTER TABLE notifications ADD CONSTRAINT chk_notifications_recipients_not_empty
    CHECK (array_length(recipients, 1) > 0);

-- Constraint для проверки попыток
ALTER TABLE notifications ADD CONSTRAINT chk_notifications_attempts_limit
    CHECK (attempts <= max_attempts);

-- Комментарии к таблице и колонкам
COMMENT ON TABLE notifications IS 'Таблица уведомлений';
COMMENT ON COLUMN notifications.id IS 'Уникальный идентификатор уведомления';
COMMENT ON COLUMN notifications.type IS 'Тип уведомления';
COMMENT ON COLUMN notifications.priority IS 'Приоритет уведомления (low, normal, high, urgent)';
COMMENT ON COLUMN notifications.title IS 'Заголовок уведомления';
COMMENT ON COLUMN notifications.message IS 'Текст уведомления';
COMMENT ON COLUMN notifications.data IS 'Дополнительные данные в формате JSON';
COMMENT ON COLUMN notifications.channels IS 'Каналы отправки (email, sms, push, telegram, webhook, slack)';
COMMENT ON COLUMN notifications.recipients IS 'Получатели уведомления';
COMMENT ON COLUMN notifications.source IS 'Источник уведомления (сервис)';
COMMENT ON COLUMN notifications.source_id IS 'ID источника уведомления';
COMMENT ON COLUMN notifications.status IS 'Статус уведомления (pending, sending, sent, delivered, failed, cancelled)';
COMMENT ON COLUMN notifications.attempts IS 'Количество попыток отправки';
COMMENT ON COLUMN notifications.max_attempts IS 'Максимальное количество попыток';
COMMENT ON COLUMN notifications.scheduled_at IS 'Время запланированной отправки';
COMMENT ON COLUMN notifications.sent_at IS 'Время отправки';
COMMENT ON COLUMN notifications.error IS 'Текст ошибки при неудачной отправке';
COMMENT ON COLUMN notifications.created_at IS 'Время создания записи';
COMMENT ON COLUMN notifications.updated_at IS 'Время последнего обновления записи';

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_notifications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создание триггера для автоматического обновления updated_at
CREATE TRIGGER trigger_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_notifications_updated_at();

-- Создание функции для получения уведомлений к отправке
CREATE OR REPLACE FUNCTION get_notifications_to_send(
    batch
