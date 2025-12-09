package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database представляет подключение к базе данных
type Database struct {
	Pool *pgxpool.Pool
}

// New создает новое подключение к базе данных
func New(ctx context.Context, connString string) (*Database, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать строку подключения: %w", err)
	}

	// Настройка пула соединений
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул подключений: %w", err)
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close закрывает подключение к базе данных
func (db *Database) Close() {
	db.Pool.Close()
}

// Health проверяет состояние подключения к базе данных
func (db *Database) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

