-- Добавление полей role и status в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Создание индексов для быстрого поиска по роли и статусу
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Создание составного индекса для фильтрации
CREATE INDEX IF NOT EXISTS idx_users_role_status ON users(role, status); 