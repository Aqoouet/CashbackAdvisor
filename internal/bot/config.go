package bot

import (
	"os"
)

// Config содержит настройки бота
type Config struct {
	TelegramToken string
	APIBaseURL    string
	Debug         bool
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	return &Config{
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		APIBaseURL:    getEnv("API_BASE_URL", "http://localhost:8080"),
		Debug:         getEnv("BOT_DEBUG", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

