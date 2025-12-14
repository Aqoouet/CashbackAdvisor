# База данных и миграции

## Обзор

Система использует **PostgreSQL 15** в качестве основной базы данных. База данных хранит информацию о правилах кэшбэков, группах пользователей и связях между пользователями и группами.

## Расширения PostgreSQL

### pg_trgm

Система использует расширение `pg_trgm` для реализации fuzzy-поиска (нечеткого поиска) по текстовым полям.

**Установка**:
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

**Использование**:
- Исправление опечаток в названиях банков
- Исправление опечаток в категориях
- Исправление опечаток в названиях групп
- Поиск похожих значений с коэффициентом схожести

---

## Схема базы данных

### Таблица `cashback_rules`

Основная таблица для хранения правил кэшбэков.

**Структура**:

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | BIGSERIAL | Первичный ключ, автоинкремент |
| `group_name` | TEXT | Название группы (legacy, берется из user_groups) |
| `category` | TEXT | Категория покупок |
| `bank_name` | TEXT | Название банка |
| `user_id` | TEXT | ID пользователя Telegram |
| `user_display_name` | TEXT | Отображаемое имя пользователя |
| `month_year` | DATE | Месяц и год действия правила |
| `cashback_percent` | NUMERIC(5,2) | Процент кэшбэка (0-100) |
| `max_amount` | NUMERIC(10,2) | Максимальная сумма кэшбэка |
| `created_at` | TIMESTAMPTZ | Дата создания записи |
| `updated_at` | TIMESTAMPTZ | Дата последнего обновления |

**Ограничения**:
- `cashback_percent`: CHECK (>= 0.00 AND <= 100.00)
- `max_amount`: CHECK (>= 0.00)

**SQL создания**:
```sql
CREATE TABLE IF NOT EXISTS cashback_rules (
    id BIGSERIAL PRIMARY KEY,
    group_name TEXT NOT NULL,
    category TEXT NOT NULL,
    bank_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    user_display_name TEXT NOT NULL,
    month_year DATE NOT NULL,
    cashback_percent NUMERIC(5,2) NOT NULL 
        CHECK (cashback_percent >= 0.00 AND cashback_percent <= 100.00),
    max_amount NUMERIC(10,2) NOT NULL CHECK (max_amount >= 0.00),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### Таблица `user_groups`

Связь пользователей с группами.

**Структура**:

| Поле | Тип | Описание |
|------|-----|----------|
| `user_id` | VARCHAR(50) | ID пользователя Telegram (первичный ключ) |
| `group_name` | VARCHAR(100) | Название группы |
| `created_at` | TIMESTAMPTZ | Дата присоединения к группе |
| `updated_at` | TIMESTAMPTZ | Дата последнего обновления |

**Ограничения**:
- Один пользователь может состоять только в одной группе (PRIMARY KEY на `user_id`)

**SQL создания**:
```sql
CREATE TABLE IF NOT EXISTS user_groups (
    user_id VARCHAR(50) PRIMARY KEY,
    group_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

### Таблица `groups`

Метаданные о группах.

**Структура**:

| Поле | Тип | Описание |
|------|-----|----------|
| `group_name` | VARCHAR(100) | Название группы (первичный ключ) |
| `created_by` | VARCHAR(50) | ID пользователя, создавшего группу |
| `created_at` | TIMESTAMPTZ | Дата создания группы |
| `description` | TEXT | Описание группы (опционально) |

**SQL создания**:
```sql
CREATE TABLE IF NOT EXISTS groups (
    group_name VARCHAR(100) PRIMARY KEY,
    created_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);
```

---

## Индексы

### Триграммные индексы (GIN)

Для быстрого fuzzy-поиска используются триграммные индексы:

```sql
-- Индекс для поиска по названию группы
CREATE INDEX idx_group_trgm ON cashback_rules 
    USING GIN (group_name gin_trgm_ops);

-- Индекс для поиска по категории
CREATE INDEX idx_category_trgm ON cashback_rules 
    USING GIN (category gin_trgm_ops);

-- Индекс для поиска по названию банка
CREATE INDEX idx_bank_trgm ON cashback_rules 
    USING GIN (bank_name gin_trgm_ops);

-- Индекс для поиска по имени пользователя
CREATE INDEX idx_user_name_trgm ON cashback_rules 
    USING GIN (user_display_name gin_trgm_ops);
```

### Композитные индексы

Для оптимизации частых запросов:

```sql
-- Индекс для поиска лучшего кэшбэка
CREATE INDEX idx_group_month_cat ON cashback_rules 
    (group_name, month_year, category);

-- Индекс для поиска по пользователю
CREATE INDEX idx_user_id ON cashback_rules (user_id);
```

### Индексы для групп

```sql
-- Индекс для быстрого поиска по группе
CREATE INDEX idx_user_groups_group_name ON user_groups(group_name);
```

---

## Триггеры

### Автоматическое обновление `updated_at`

Триггер автоматически обновляет поле `updated_at` при изменении записи:

```sql
-- Функция триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для таблицы cashback_rules
CREATE TRIGGER update_cashback_rules_updated_at
    BEFORE UPDATE ON cashback_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## Миграции

### Миграция 001: Начальная схема

**Файл**: `migrations/001_initial_schema.sql`

**Содержимое**:
- Создание расширения `pg_trgm`
- Создание таблицы `cashback_rules`
- Создание триграммных индексов
- Создание композитных индексов
- Создание триггера для `updated_at`

**Применение**:
```bash
psql -h localhost -U cashback_user -d cashback_db -f migrations/001_initial_schema.sql
```

---

### Миграция 002: Откат (Down)

**Файл**: `migrations/002_down.sql`

**Содержимое**:
- Удаление триггеров
- Удаление индексов
- Удаление таблиц
- Удаление расширений

**Применение** (для отката):
```bash
psql -h localhost -U cashback_user -d cashback_db -f migrations/002_down.sql
```

---

### Миграция 003: Группы пользователей

**Файл**: `migrations/003_user_groups.sql`

**Содержимое**:
- Создание таблицы `user_groups`
- Создание таблицы `groups`
- Создание индексов для групп

**Применение**:
```bash
psql -h localhost -U cashback_user -d cashback_db -f migrations/003_user_groups.sql
```

---

## Основные SQL запросы

### Создание правила кэшбэка

```sql
INSERT INTO cashback_rules (
    group_name, category, bank_name, user_id, user_display_name,
    month_year, cashback_percent, max_amount
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at;
```

### Получение правила по ID

```sql
SELECT id, group_name, category, bank_name, user_id, user_display_name,
       month_year, cashback_percent, max_amount, created_at, updated_at
FROM cashback_rules
WHERE id = $1;
```

### Список правил группы с пагинацией

```sql
SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, 
       cr.user_display_name, cr.month_year, cr.cashback_percent, 
       cr.max_amount, cr.created_at, cr.updated_at
FROM cashback_rules cr
INNER JOIN user_groups ug ON cr.user_id = ug.user_id
WHERE ug.group_name = $1
ORDER BY cr.created_at DESC 
LIMIT $2 OFFSET $3;
```

### Поиск лучшего кэшбэка

```sql
SELECT cr.id, cr.group_name, cr.category, cr.bank_name, cr.user_id, 
       cr.user_display_name, cr.month_year, cr.cashback_percent, 
       cr.max_amount, cr.created_at, cr.updated_at
FROM cashback_rules cr
INNER JOIN user_groups ug ON cr.user_id = ug.user_id
WHERE ug.group_name = $1 AND cr.category = $2 AND cr.month_year >= $3
ORDER BY cr.cashback_percent DESC, cr.max_amount DESC
LIMIT 1;
```

### Fuzzy-поиск по полю

```sql
SELECT DISTINCT field_name, similarity(field_name, $1) as sim
FROM cashback_rules
WHERE similarity(field_name, $1) >= $2
ORDER BY sim DESC
LIMIT $3;
```

**Примеры**:
- Поиск похожих банков: `similarity(bank_name, 'Тинькоф') >= 0.65`
- Поиск похожих категорий: `similarity(category, 'такси') >= 0.6`

### Создание группы

```sql
INSERT INTO groups (group_name, created_by)
VALUES ($1, $2)
ON CONFLICT (group_name) DO NOTHING;
```

### Присоединение пользователя к группе

```sql
INSERT INTO user_groups (user_id, group_name, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (user_id) 
DO UPDATE SET group_name = $2, updated_at = CURRENT_TIMESTAMP;
```

### Получение группы пользователя

```sql
SELECT group_name FROM user_groups WHERE user_id = $1;
```

---

## Пул соединений

Настройки пула соединений PostgreSQL (в `internal/database/database.go`):

- **MaxConns**: 25 — максимальное количество соединений
- **MinConns**: 5 — минимальное количество соединений
- **MaxConnLifetime**: 1 час — максимальное время жизни соединения
- **MaxConnIdleTime**: 30 минут — максимальное время простоя соединения
- **HealthCheckPeriod**: 1 минута — период проверки здоровья соединений

---

## Резервное копирование

### Создание резервной копии

```bash
pg_dump -h localhost -U cashback_user -d cashback_db > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Восстановление из резервной копии

```bash
psql -h localhost -U cashback_user -d cashback_db < backup_20241215_120000.sql
```

---

## Мониторинг и оптимизация

### Проверка размера таблиц

```sql
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Проверка использования индексов

```sql
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

### Анализ таблиц

Для обновления статистики и оптимизации запросов:

```sql
ANALYZE cashback_rules;
ANALYZE user_groups;
ANALYZE groups;
```

### Проверка медленных запросов

Включите логирование медленных запросов в `postgresql.conf`:

```conf
log_min_duration_statement = 1000  # Логировать запросы > 1 секунды
```

---

## Безопасность

### Рекомендации

1. **Пароли**: Используйте надежные пароли для пользователей БД
2. **SSL**: Включите SSL для продакшн окружения (`DB_SSLMODE=require`)
3. **Права доступа**: Ограничьте права пользователя БД только необходимыми операциями
4. **Резервное копирование**: Регулярно создавайте резервные копии
5. **Обновления**: Регулярно обновляйте PostgreSQL для безопасности

### Создание пользователя с ограниченными правами

```sql
-- Создание пользователя
CREATE USER cashback_user WITH PASSWORD 'secure_password';

-- Предоставление прав
GRANT CONNECT ON DATABASE cashback_db TO cashback_user;
GRANT USAGE ON SCHEMA public TO cashback_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO cashback_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cashback_user;

-- Права на будущие таблицы
ALTER DEFAULT PRIVILEGES IN SCHEMA public 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO cashback_user;
```

---

## Миграции в Docker

При использовании Docker Compose миграции можно применить через скрипт:

```bash
# Через контейнер
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -f /path/to/migrations/001_initial_schema.sql
```

Или использовать скрипт миграции:

```bash
./scripts/migrate.sh
```

---

## Следующие шаги

- [Развертывание](07-deployment.md) — развертывание на продакшн
- [Разработка](08-development.md) — разработка новых функций
- [Решение проблем](09-troubleshooting.md) — решение проблем с БД

