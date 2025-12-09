#!/bin/bash

# Скрипт для отката миграций базы данных

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

echo "⚠️  Откат миграций базы данных $DB_NAME..."

# Откат миграций
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/002_down.sql

echo "✅ Миграции успешно откачены!"

