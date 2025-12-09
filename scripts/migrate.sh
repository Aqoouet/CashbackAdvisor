#!/bin/bash

# Скрипт для применения миграций базы данных

set -e

# Загрузка переменных окружения
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Значения по умолчанию
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-cashback_db}

echo "🔄 Применение миграций к базе данных $DB_NAME..."

# Применение миграций по порядку
for migration in migrations/*.sql; do
    # Пропускаем down-миграции
    if [[ "$migration" == *"_down.sql" ]]; then
        continue
    fi
    
    echo "  Применение $migration..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration" 2>&1 | grep -v "already exists" || true
done

echo "✅ Миграции успешно применены!"

