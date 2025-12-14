package bot

import (
	"fmt"
	"os"
)

// Константы переменных окружения.
const (
	EnvTelegramToken = "TELEGRAM_BOT_TOKEN"
	EnvAPIBaseURL    = "API_BASE_URL"
	EnvBotDebug      = "BOT_DEBUG"
)

// Значения по умолчанию.
const (
	DefaultAPIBaseURL = "http://localhost:8080"
	DefaultDebug      = false
)

// Config содержит настройки бота.
type Config struct {
	TelegramToken string
	APIBaseURL    string
	Debug         bool
}

// LoadConfig загружает конфигурацию из переменных окружения.
func LoadConfig() *Config {
	return &Config{
		TelegramToken: getEnv(EnvTelegramToken, ""),
		APIBaseURL:    getEnv(EnvAPIBaseURL, DefaultAPIBaseURL),
		Debug:         getEnv(EnvBotDebug, "false") == "true",
	}
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.TelegramToken == "" {
		return fmt.Errorf("%s не установлен в переменных окружения", EnvTelegramToken)
	}
	return nil
}

// getEnv получает переменную окружения или возвращает значение по умолчанию.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
