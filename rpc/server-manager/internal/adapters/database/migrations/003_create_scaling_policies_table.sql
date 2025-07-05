-- Создание таблицы политик масштабирования
CREATE TABLE IF NOT EXISTS scaling_policies (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    min_servers INTEGER NOT NULL DEFAULT 1,
    max_servers INTEGER NOT NULL DEFAULT 10,
    cpu_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.8,
    memory_threshold DOUBLE PRECISION NOT NULL DEFAULT 0.8,
    scale_up_cooldown BIGINT NOT NULL DEFAULT 300, -- секунды
    scale_down_cooldown BIGINT NOT NULL DEFAULT 600, -- секунды
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Индексы для политик масштабирования
CREATE INDEX IF NOT EXISTS idx_scaling_policies_enabled ON scaling_policies(enabled);
CREATE INDEX IF NOT EXISTS idx_scaling_policies_name ON scaling_policies(name);

-- Создание таблицы истории масштабирования
CREATE TABLE IF NOT EXISTS scaling_history (
    id SERIAL PRIMARY KEY,
    policy_id VARCHAR(36) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'scale_up', 'scale_down'
    reason TEXT,
    servers_before INTEGER NOT NULL,
    servers_after INTEGER NOT NULL,
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    FOREIGN KEY (policy_id) REFERENCES scaling_policies(id) ON DELETE CASCADE
);

-- Индексы для истории масштабирования
CREATE INDEX IF NOT EXISTS idx_scaling_history_policy_id ON scaling_history(policy_id);
CREATE INDEX IF NOT EXISTS idx_scaling_history_triggered_at ON scaling_history(triggered_at); 