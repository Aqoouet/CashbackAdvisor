#!/usr/bin/env bash

# Скрипт для вывода содержимого таблицы пользователей по SSH.
# Предполагается, что:
# - На сервер можно попасть по алиасу ssh cashback-server
# - Проект расположен в ~/CashbackAdvisor
# - База развернута в контейнере PostgreSQL из docker-compose.full.yml

set -euo pipefail

REMOTE_HOST="cashback-server"
PROJECT_DIR="~/CashbackAdvisor"
DB_CONTAINER="cashback_postgres"

echo "🔐 Подключаюсь к ${REMOTE_HOST} и запрашиваю список пользователей из БД..."

ssh "${REMOTE_HOST}" "cd ${PROJECT_DIR} && docker exec -i ${DB_CONTAINER} psql -U postgres -d cashback_db -c 'SELECT * FROM user_groups ORDER BY user_id;'" || {
  echo "❌ Ошибка при выполнении запроса к базе данных" >&2
  exit 1
}


