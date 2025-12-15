#!/usr/bin/env bash

# Локальный скрипт для резервного копирования БД с сервера.
# Бэкап сохраняется в ./backup на вашем ПК, а не на удалённом сервере.
#
# Предполагается:
# - SSH-алиас сервера: cashback-server
# - Проект на сервере: ~/CashbackAdvisor
# - PostgreSQL запущен в контейнере cashback_postgres (docker-compose.full.yml)
# - База данных: cashback_db, пользователь: postgres

set -euo pipefail

REMOTE_HOST="cashback-server"
PROJECT_DIR="~/CashbackAdvisor"
DB_CONTAINER="cashback_postgres"
DB_NAME="cashback_db"
DB_USER="postgres"

BACKUP_DIR="backup"
TIMESTAMP="$(date +'%Y%m%d_%H%M%S')"
BACKUP_FILE="${BACKUP_DIR}/db_backup_${TIMESTAMP}.sql"

mkdir -p "${BACKUP_DIR}"

echo "💾 Создаю резервную копию базы ${DB_NAME} в ${BACKUP_FILE}"

# Делаем pg_dump внутри контейнера на сервере и стримим результат по SSH на локальную машину
ssh "${REMOTE_HOST}" "cd ${PROJECT_DIR} && docker exec -i ${DB_CONTAINER} pg_dump -U ${DB_USER} ${DB_NAME}" > "${BACKUP_FILE}"

echo "✅ Резервная копия сохранена локально: ${BACKUP_FILE}"


