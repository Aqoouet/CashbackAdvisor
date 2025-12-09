-- Создание таблицы для связи пользователь-группа
CREATE TABLE IF NOT EXISTS user_groups (
    user_id VARCHAR(50) PRIMARY KEY,
    group_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по группе
CREATE INDEX IF NOT EXISTS idx_user_groups_group_name ON user_groups(group_name);

-- Таблица групп для хранения метаданных
CREATE TABLE IF NOT EXISTS groups (
    group_name VARCHAR(100) PRIMARY KEY,
    created_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Комментарии
COMMENT ON TABLE user_groups IS 'Связь пользователь-группа';
COMMENT ON TABLE groups IS 'Информация о группах';
COMMENT ON COLUMN cashback_rules.group_name IS 'Legacy поле - не используется, группа берется из user_groups';

