package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rymax1e/open-cashback-advisor/internal/config"
	"github.com/rymax1e/open-cashback-advisor/internal/database"
	"github.com/rymax1e/open-cashback-advisor/internal/handlers"
	"github.com/rymax1e/open-cashback-advisor/internal/service"
)

// Таймауты сервера.
const (
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	requestTimeout  = 60 * time.Second
	shutdownTimeout = 30 * time.Second
	corsMaxAge      = 300
)

func main() {
	// Загрузка и валидация конфигурации
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Ошибка конфигурации: %v", err)
	}

	log.Println("🚀 Запуск Open Cashback Advisor...")

	// Инициализация базы данных
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Создание зависимостей
	repo := database.NewRepository(db)
	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)

	// Настройка и запуск сервера
	router := setupRouter(handler)
	srv := createServer(cfg.Server.Address(), router)

	// Запуск с graceful shutdown
	runServer(srv)
}

// initDatabase инициализирует подключение к базе данных.
func initDatabase(cfg *config.Config) (*database.Database, error) {
	ctx := context.Background()
	db, err := database.New(ctx, cfg.Database.ConnectionString())
	if err != nil {
		return nil, err
	}
	log.Println("✅ Успешное подключение к базе данных")
	return db, nil
}

// setupRouter настраивает маршрутизатор.
func setupRouter(handler *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           corsMaxAge,
	}))

	// Регистрация маршрутов
	handler.RegisterRoutes(r)

	return r
}

// createServer создаёт HTTP сервер.
func createServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

// runServer запускает сервер с поддержкой graceful shutdown.
func runServer(srv *http.Server) {
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("\n⚠️  Получен сигнал остановки сервера...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("❌ Не удалось корректно остановить сервер: %v", err)
		}
		close(done)
	}()

	logServerInfo(srv.Addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Не удалось запустить сервер: %v", err)
	}

	<-done
	log.Println("✅ Сервер корректно остановлен")
}

// logServerInfo выводит информацию о запуске сервера.
func logServerInfo(addr string) {
	log.Printf("🌐 Сервер запущен на http://%s", addr)
	log.Println("📖 API документация:")
	log.Println("   POST   /api/v1/cashback/suggest  - Анализ и предложения")
	log.Println("   POST   /api/v1/cashback          - Создать правило")
	log.Println("   GET    /api/v1/cashback          - Список правил")
	log.Println("   GET    /api/v1/cashback/best     - Лучший кэшбэк")
	log.Println("   GET    /api/v1/cashback/{id}     - Получить правило")
	log.Println("   PUT    /api/v1/cashback/{id}     - Обновить правило")
	log.Println("   DELETE /api/v1/cashback/{id}     - Удалить правило")
	log.Println("   GET    /health                   - Проверка здоровья")
	log.Println()
}
