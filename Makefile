.PHONY: help build run test clean migrate rollback docker-up docker-down build-bot run-bot

help: ## Показать помощь
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение (сервер)
	@echo "🔨 Сборка сервера..."
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Сборка завершена!"

build-bot: ## Собрать бота
	@echo "🔨 Сборка бота..."
	@go build -o bin/bot cmd/bot/main.go
	@echo "✅ Сборка бота завершена!"

build-all: ## Собрать сервер и бота
	@$(MAKE) build
	@$(MAKE) build-bot

run: ## Запустить сервер
	@echo "🚀 Запуск сервера..."
	@go run cmd/server/main.go

run-bot: ## Запустить бота
	@echo "🤖 Запуск Telegram бота..."
	@go run cmd/bot/main.go

test: ## Запустить тесты
	@echo "🧪 Запуск тестов..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Тесты пройдены!"

clean: ## Очистить сгенерированные файлы
	@echo "🧹 Очистка..."
	@rm -rf bin/
	@rm -f coverage.out
	@echo "✅ Очистка завершена!"

migrate: ## Применить миграции
	@chmod +x scripts/migrate.sh
	@./scripts/migrate.sh

rollback: ## Откатить миграции
	@chmod +x scripts/rollback.sh
	@./scripts/rollback.sh

deps: ## Установить зависимости
	@echo "📦 Установка зависимостей..."
	@go mod download
	@go mod tidy
	@echo "✅ Зависимости установлены!"

docker-up: ## Запустить PostgreSQL в Docker
	@echo "🐳 Запуск PostgreSQL в Docker..."
	@docker-compose up -d
	@echo "✅ PostgreSQL запущен!"

docker-down: ## Остановить Docker контейнеры
	@echo "🛑 Остановка Docker контейнеров..."
	@docker-compose down
	@echo "✅ Контейнеры остановлены!"

dev: docker-up ## Запустить среду разработки (API)
	@sleep 3
	@$(MAKE) migrate
	@$(MAKE) run

dev-full: ## Запустить полный стек (API + Bot)
	@echo "🚀 Запуск полного стека..."
	@if [ -z "$$TELEGRAM_BOT_TOKEN" ]; then \
		echo "❌ TELEGRAM_BOT_TOKEN не установлен!"; \
		echo "Установите: export TELEGRAM_BOT_TOKEN=your_token"; \
		exit 1; \
	fi
	@docker-compose -f docker-compose.full.yml up -d
	@echo "✅ Полный стек запущен!"

stop-full: ## Остановить полный стек
	@docker-compose -f docker-compose.full.yml down

fmt: ## Форматирование кода
	@echo "🎨 Форматирование кода..."
	@go fmt ./...
	@echo "✅ Форматирование завершено!"

lint: ## Проверка кода линтером
	@echo "🔍 Проверка кода линтером..."
	@golangci-lint run || true
	@echo "✅ Проверка завершена!"

.DEFAULT_GOAL := help

