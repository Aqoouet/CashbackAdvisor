#!/bin/bash
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}⚠️  ВНИМАНИЕ: Вы собираетесь удалить ВСЕ данные из базы!${NC}"
echo -e "${RED}Это действие необратимо!${NC}"
echo ""
read -p "Продолжить? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo -e "${GREEN}Операция отменена.${NC}"
    exit 0
fi

echo -e "${BLUE}🗑️  Удаление всех записей из базы данных...${NC}"

docker-compose -f docker-compose.full.yml exec -T postgres psql -U postgres -d cashback_db -c "DELETE FROM cashback_rules; ALTER SEQUENCE cashback_rules_id_seq RESTART WITH 1;"

echo ""
echo -e "${GREEN}✅ База данных очищена!${NC}"
echo -e "${GREEN}Счетчик ID сброшен на 1.${NC}"
echo ""
echo -e "${YELLOW}Проверка:${NC}"
docker-compose -f docker-compose.full.yml exec -T postgres psql -U postgres -d cashback_db -c "SELECT COUNT(*) as total_rows FROM cashback_rules;"

