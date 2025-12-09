#!/bin/bash
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}  Деплой CashbackAdvisor на сервер${NC}"
echo -e "${BLUE}==================================================${NC}"

# Проверка переменной окружения TELEGRAM_BOT_TOKEN
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo -e "${RED}❌ Ошибка: TELEGRAM_BOT_TOKEN не установлен!${NC}"
    echo -e "${YELLOW}Установите токен бота:${NC}"
    echo -e "${YELLOW}export TELEGRAM_BOT_TOKEN=your_token_here${NC}"
    exit 1
fi

echo -e "${GREEN}✅ TELEGRAM_BOT_TOKEN установлен${NC}"

# Проверка наличия docker и docker-compose
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен!${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose не установлен!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker и Docker Compose доступны${NC}"

# Создание файла .env если он не существует
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  Файл .env не найден, создаю из env.example...${NC}"
    if [ -f env.example ]; then
        cp env.example .env
        echo -e "${GREEN}✅ Файл .env создан. Отредактируйте его перед продолжением!${NC}"
        echo -e "${YELLOW}Нажмите Enter после редактирования .env...${NC}"
        read
    else
        echo -e "${RED}❌ Файл env.example не найден!${NC}"
        exit 1
    fi
fi

# Проверка/создание .env файла с токеном
echo -e "${BLUE}📝 Обновление .env файла...${NC}"
if ! grep -q "TELEGRAM_BOT_TOKEN=" .env; then
    echo "TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN" >> .env
else
    # Обновляем токен в существующем файле
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|^TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN|" .env
    else
        # Linux
        sed -i "s|^TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN|" .env
    fi
fi
echo -e "${GREEN}✅ Файл .env обновлён${NC}"

# Остановка старых контейнеров (если есть)
echo -e "${BLUE}🛑 Остановка старых контейнеров...${NC}"
docker-compose -f docker-compose.full.yml down 2>/dev/null || true

# Очистка старых образов (опционально)
echo -e "${BLUE}🧹 Очистка старых образов...${NC}"
docker-compose -f docker-compose.full.yml down --rmi local 2>/dev/null || true

# Сборка образов
echo -e "${BLUE}🔨 Сборка Docker образов...${NC}"
docker-compose -f docker-compose.full.yml build --no-cache

# Запуск контейнеров
echo -e "${BLUE}🚀 Запуск контейнеров...${NC}"
docker-compose -f docker-compose.full.yml up -d

# Ожидание запуска PostgreSQL
echo -e "${BLUE}⏳ Ожидание готовности базы данных...${NC}"
sleep 10

# Применение миграций
echo -e "${BLUE}📊 Применение миграций базы данных...${NC}"
docker exec cashback_api /bin/sh -c "cd /app && ./scripts/migrate.sh" 2>/dev/null || echo -e "${YELLOW}⚠️  Миграции не применены (возможно уже применены)${NC}"

# Проверка статуса контейнеров
echo -e "${BLUE}🔍 Проверка статуса контейнеров...${NC}"
docker-compose -f docker-compose.full.yml ps

# Проверка здоровья API
echo -e "${BLUE}🏥 Проверка здоровья API...${NC}"
sleep 5
if curl -f http://localhost:8080/health &> /dev/null; then
    echo -e "${GREEN}✅ API работает!${NC}"
else
    echo -e "${YELLOW}⚠️  API может быть ещё не готов, проверьте логи${NC}"
fi

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}  ✅ Деплой завершён успешно!${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""
echo -e "${BLUE}📖 Полезные команды:${NC}"
echo -e "  ${YELLOW}Логи всех сервисов:${NC} docker-compose -f docker-compose.full.yml logs -f"
echo -e "  ${YELLOW}Логи API:${NC} docker-compose -f docker-compose.full.yml logs -f api"
echo -e "  ${YELLOW}Логи бота:${NC} docker-compose -f docker-compose.full.yml logs -f bot"
echo -e "  ${YELLOW}Остановка:${NC} docker-compose -f docker-compose.full.yml down"
echo -e "  ${YELLOW}Перезапуск:${NC} docker-compose -f docker-compose.full.yml restart"
echo ""
echo -e "${BLUE}🌐 Сервисы:${NC}"
echo -e "  ${YELLOW}API:${NC} http://localhost:8080"
echo -e "  ${YELLOW}Health:${NC} http://localhost:8080/health"
echo -e "  ${YELLOW}Bot:${NC} Telegram @your_bot_username"
echo ""

