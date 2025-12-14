// Package config предоставляет конфигурацию для сервера.
package config

import (
	"fmt"
	"os"
)

// Константы переменных окружения.
const (
	EnvDBHost     = "DB_HOST"
	EnvDBPort     = "DB_PORT"
	EnvDBUser     = "DB_USER"
	EnvDBPassword = "DB_PASSWORD"
	EnvDBName     = "DB_NAME"
	EnvDBSSLMode  = "DB_SSLMODE"
	EnvServerHost = "SERVER_HOST"
	EnvServerPort = "SERVER_PORT"
)

// Значения по умолчанию.
const (
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "5432"
	DefaultDBUser     = "postgres"
	DefaultDBPassword = "postgres"
	DefaultDBName     = "cashback_db"
	DefaultDBSSLMode  = "disable"
	DefaultServerHost = "0.0.0.0"
	DefaultServerPort = "8080"
)

// Config представляет конфигурацию приложения.
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

// DatabaseConfig содержит настройки базы данных.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServerConfig содержит настройки сервера.
type ServerConfig struct {
	Host string
	Port string
}

// ConnectionString возвращает строку подключения к PostgreSQL.
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Address возвращает адрес сервера.
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Load загружает конфигурацию из переменных окружения.
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv(EnvDBHost, DefaultDBHost),
			Port:     getEnv(EnvDBPort, DefaultDBPort),
			User:     getEnv(EnvDBUser, DefaultDBUser),
			Password: getEnv(EnvDBPassword, DefaultDBPassword),
			DBName:   getEnv(EnvDBName, DefaultDBName),
			SSLMode:  getEnv(EnvDBSSLMode, DefaultDBSSLMode),
		},
		Server: ServerConfig{
			Host: getEnv(EnvServerHost, DefaultServerHost),
			Port: getEnv(EnvServerPort, DefaultServerPort),
		},
	}
}

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST не может быть пустым")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME не может быть пустым")
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
