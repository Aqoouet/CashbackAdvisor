#!/bin/bash

# Скрипт для запуска всего стека (PostgreSQL + API + Bot)

set -e

echo "🚀 Запуск Open Cashback Advisor (полный стек)..."

# Проверка переменных окружения
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ TELEGRAM_BOT_TOKEN не установлен!"
    echo "Установите токен: export TELEGRAM_BOT_TOKEN=your_token_here"
    exit 1
fi

# Запуск всех сервисов
echo "🐳 Запуск контейнеров..."
docker-compose -f docker-compose.full.yml up -d

echo "⏳ Ожидание готовности PostgreSQL..."
sleep 5

echo "📊 Применение миграций..."
docker-compose -f docker-compose.full.yml exec -T postgres psql -U postgres -d cashback_db < migrations/001_initial_schema.sql || true

echo ""
echo "✅ Все сервисы запущены!"
echo ""
echo "📡 Доступные сервисы:"
echo "   API:        http://localhost:8080"
echo "   Health:     http://localhost:8080/health"
echo "   Telegram:   Бот готов к работе"
echo ""
echo "📋 Логи:"
echo "   API:  docker-compose -f docker-compose.full.yml logs -f api"
echo "   Bot:  docker-compose -f docker-compose.full.yml logs -f bot"
echo "   All:  docker-compose -f docker-compose.full.yml logs -f"
echo ""
echo "🛑 Остановка: docker-compose -f docker-compose.full.yml down"

