#!/bin/bash
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}  Обновление CashbackAdvisor${NC}"
echo -e "${BLUE}==================================================${NC}"

# Проверка docker-compose
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен!${NC}"
    exit 1
fi

echo -e "${BLUE}📦 Получение последних изменений из Git...${NC}"
git pull origin main

echo -e "${BLUE}🛑 Остановка контейнеров...${NC}"
docker-compose -f docker-compose.full.yml down

echo -e "${BLUE}🔨 Пересборка образов без кеша...${NC}"
docker-compose -f docker-compose.full.yml build --no-cache

echo -e "${BLUE}🚀 Запуск контейнеров...${NC}"
docker-compose -f docker-compose.full.yml up -d

echo -e "${BLUE}⏳ Ожидание запуска сервисов...${NC}"
sleep 10

echo -e "${BLUE}🔍 Проверка статуса:${NC}"
docker-compose -f docker-compose.full.yml ps

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}  ✅ Обновление завершено успешно!${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""
echo -e "${BLUE}📖 Проверьте логи:${NC}"
echo -e "  ${YELLOW}docker-compose -f docker-compose.full.yml logs -f bot${NC}"
echo ""
echo -e "${BLUE}📊 Версия бота:${NC}"
docker-compose -f docker-compose.full.yml logs bot 2>&1 | grep "Запуск Telegram бота" | tail -1 || echo -e "${YELLOW}  Запустите бота и проверьте логи${NC}"
echo ""

