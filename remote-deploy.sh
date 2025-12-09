#!/bin/bash
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVER="82.26.150.98"
USER="cashback"
REPO="git@github.com:Aqoouet/CashbackAdvisor.git"
APP_DIR="CashbackAdvisor"

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}  Удалённый деплой на сервер $SERVER${NC}"
echo -e "${BLUE}==================================================${NC}"

# Проверка токена
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo -e "${RED}❌ Ошибка: TELEGRAM_BOT_TOKEN не установлен!${NC}"
    echo -e "${YELLOW}Установите токен бота:${NC}"
    echo -e "${YELLOW}export TELEGRAM_BOT_TOKEN=your_token_here${NC}"
    exit 1
fi

echo -e "${GREEN}✅ TELEGRAM_BOT_TOKEN установлен${NC}"
echo ""

# Подключение к серверу и выполнение команд
echo -e "${BLUE}🔗 Подключение к серверу $SERVER...${NC}"
ssh -t $USER@$SERVER << EOF
    set -e
    
    echo -e "${GREEN}✅ Подключено к серверу${NC}"
    
    # Проверка наличия Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker не установлен на сервере!${NC}"
        echo -e "${YELLOW}Установите Docker следуя инструкции в DEPLOY.md${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker установлен${NC}"
    
    # Клонирование или обновление репозитория
    if [ -d "$APP_DIR" ]; then
        echo -e "${BLUE}📦 Обновление репозитория...${NC}"
        cd $APP_DIR
        git pull origin main || git pull origin master
    else
        echo -e "${BLUE}📦 Клонирование репозитория...${NC}"
        git clone $REPO $APP_DIR
        cd $APP_DIR
    fi
    
    echo -e "${GREEN}✅ Код обновлён${NC}"
    
    # Создание .env файла
    if [ ! -f .env ]; then
        echo -e "${BLUE}📝 Создание .env файла...${NC}"
        cp env.example .env
        
        # Установка токена
        sed -i "s|your_telegram_bot_token_here|$TELEGRAM_BOT_TOKEN|g" .env
        
        # Генерация случайного пароля для БД
        DB_PASSWORD=\$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
        sed -i "s|your_secure_password_here|\$DB_PASSWORD|g" .env
        
        echo -e "${GREEN}✅ Файл .env создан${NC}"
    else
        echo -e "${YELLOW}⚠️  Файл .env уже существует${NC}"
        # Обновление только токена
        sed -i "s|^TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN|g" .env
        echo -e "${GREEN}✅ Токен обновлён${NC}"
    fi
    
    # Экспорт переменных для docker-compose
    export TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN
    
    # Остановка старых контейнеров
    echo -e "${BLUE}🛑 Остановка старых контейнеров...${NC}"
    docker-compose -f docker-compose.full.yml down 2>/dev/null || true
    
    # Сборка образов
    echo -e "${BLUE}🔨 Сборка Docker образов...${NC}"
    docker-compose -f docker-compose.full.yml build
    
    # Запуск контейнеров
    echo -e "${BLUE}🚀 Запуск контейнеров...${NC}"
    docker-compose -f docker-compose.full.yml up -d
    
    # Ожидание запуска
    echo -e "${BLUE}⏳ Ожидание запуска сервисов...${NC}"
    sleep 15
    
    # Проверка статуса
    echo -e "${BLUE}🔍 Проверка статуса:${NC}"
    docker-compose -f docker-compose.full.yml ps
    
    echo -e "${GREEN}==================================================${NC}"
    echo -e "${GREEN}  ✅ Деплой завершён успешно!${NC}"
    echo -e "${GREEN}==================================================${NC}"
    
    echo -e "${BLUE}📖 Для просмотра логов используйте:${NC}"
    echo -e "  ssh $USER@$SERVER"
    echo -e "  cd $APP_DIR"
    echo -e "  docker-compose -f docker-compose.full.yml logs -f"
EOF

echo ""
echo -e "${GREEN}✨ Готово! Бот развёрнут на сервере.${NC}"

