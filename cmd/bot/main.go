package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rymax1e/open-cashback-advisor/internal/bot"
)

func main() {
	log.Printf("🚀 Запуск Telegram бота Open Cashback Advisor %s...", bot.BuildInfo())

	// Загрузка конфигурации
	cfg := bot.LoadConfig()

	// Проверка токена
	if cfg.TelegramToken == "" {
		log.Fatal("❌ TELEGRAM_BOT_TOKEN не установлен в переменных окружения")
	}

	// Создание API клиента
	apiClient := bot.NewAPIClient(cfg.APIBaseURL)
	log.Printf("✅ API клиент создан: %s", cfg.APIBaseURL)

	// Создание бота
	telegramBot, err := bot.NewBot(cfg.TelegramToken, apiClient, cfg.Debug)
	if err != nil {
		log.Fatalf("❌ Не удалось создать бота: %v", err)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("\n⚠️  Получен сигнал остановки бота...")
		os.Exit(0)
	}()

	log.Printf("🤖 Бот %s готов к работе!", bot.BuildInfo())
	log.Println("📖 Команды:")
	log.Println("   /start  - Начать работу")
	log.Println("   /help   - Справка")
	log.Println("   /add    - Добавить правило")
	log.Println("   /list   - Мои правила")
	log.Println("   /best   - Лучший кэшбэк")
	log.Println()

	// Запуск бота
	telegramBot.Start()
}

