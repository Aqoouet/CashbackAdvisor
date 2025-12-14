# Разработка

## Настройка окружения разработки

### Требования

- **Go**: версия 1.22 или выше
- **PostgreSQL**: версия 15 или выше
- **Docker**: для запуска БД (опционально)
- **Git**: для версионирования

### Установка Go

**Linux**:
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

**macOS**:
```bash
brew install go
```

**Проверка установки**:
```bash
go version
```

### Настройка рабочего пространства

```bash
# Клонирование репозитория
git clone <repository-url>
cd Bot

# Установка зависимостей
go mod download
go mod tidy
```

---

## Структура проекта

```
Bot/
├── cmd/                    # Точки входа приложений
│   ├── bot/               # Telegram бот
│   │   └── main.go
│   └── server/            # HTTP API сервер
│       └── main.go
├── internal/              # Внутренние пакеты
│   ├── bot/              # Логика Telegram бота
│   │   ├── bot.go        # Основной класс бота
│   │   ├── commands.go   # Обработчики команд
│   │   ├── client.go     # API клиент
│   │   ├── parser.go      # Парсинг сообщений
│   │   └── ...
│   ├── config/           # Конфигурация
│   ├── database/         # Работа с БД
│   │   ├── database.go   # Подключение к БД
│   │   ├── repository.go # Репозиторий
│   │   └── queries.go    # SQL запросы
│   ├── handlers/         # HTTP обработчики
│   ├── models/           # Модели данных
│   ├── service/          # Бизнес-логика
│   └── validator/        # Валидация данных
├── migrations/           # SQL миграции
├── scripts/              # Вспомогательные скрипты
├── docs/                 # Документация
├── examples/             # Примеры использования
├── go.mod                # Зависимости Go
├── go.sum                # Хеши зависимостей
├── Makefile              # Команды для разработки
└── docker-compose.yml    # Docker конфигурация
```

---

## Команды для разработки

### Makefile команды

Проект использует Makefile для упрощения разработки:

```bash
# Показать все доступные команды
make help

# Сборка приложений
make build          # Собрать API сервер
make build-bot      # Собрать бота
make build-all      # Собрать всё

# Запуск приложений
make run            # Запустить API сервер
make run-bot        # Запустить бота

# Тестирование
make test           # Запустить тесты

# Очистка
make clean          # Удалить сгенерированные файлы

# Зависимости
make deps           # Установить/обновить зависимости

# Миграции
make migrate        # Применить миграции
make rollback       # Откатить миграции

# Docker
make docker-up      # Запустить PostgreSQL в Docker
make docker-down    # Остановить Docker контейнеры
make dev-full       # Запустить полный стек (API + Bot)

# Форматирование и линтинг
make fmt            # Форматировать код
make lint           # Проверить код линтером
```

---

## Запуск локальной разработки

### Вариант 1: Только API сервер

```bash
# 1. Запустить PostgreSQL в Docker
make docker-up

# 2. Применить миграции
make migrate

# 3. Настроить переменные окружения
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=cashback_db
export DB_SSLMODE=disable
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# 4. Запустить сервер
make run
```

Или используйте команду `make dev`, которая сделает всё автоматически.

### Вариант 2: Полный стек (API + Bot)

```bash
# 1. Настроить переменные окружения
cp env.example .env
# Отредактируйте .env файл

# 2. Запустить полный стек
make dev-full

# Или через docker-compose напрямую
docker-compose -f docker-compose.full.yml up
```

### Вариант 3: Раздельный запуск

```bash
# Терминал 1: API сервер
make docker-up
make migrate
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=cashback_db DB_SSLMODE=disable
export SERVER_HOST=0.0.0.0 SERVER_PORT=8080
make run

# Терминал 2: Telegram бот
export TELEGRAM_BOT_TOKEN=your_token
export API_BASE_URL=http://localhost:8080
export BOT_DEBUG=true
make run-bot
```

---

## Добавление новых функций

### Добавление новой команды бота

1. **Добавьте команду в `internal/bot/commands.go`**:

```go
// В commandHelpMap
"newcommand": {
    Name:      "/newcommand",
    ShortDesc: "Описание команды",
    LongDesc:  "Подробное описание команды",
    Usage:     "/newcommand [параметры]",
    Examples:  []string{"/newcommand", "/newcommand param"},
},

// В routeCommand
case "newcommand":
    b.handleNewCommand(message)
```

2. **Создайте обработчик в `internal/bot/commands.go`**:

```go
func (b *Bot) handleNewCommand(message *tgbotapi.Message) {
    // Логика обработки команды
    b.sendText(message.Chat.ID, "Ответ пользователю")
}
```

3. **Обновите справку в `handleHelp`** (если нужно)

### Добавление нового API эндпоинта

1. **Добавьте метод в Service** (`internal/service/service.go`):

```go
func (s *Service) NewMethod(ctx context.Context, req *models.NewRequest) (*models.NewResponse, error) {
    // Бизнес-логика
    return response, nil
}
```

2. **Добавьте обработчик в Handler** (`internal/handlers/handlers.go`):

```go
func (h *Handler) NewHandler(w http.ResponseWriter, r *http.Request) {
    var req models.NewRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Неверный формат запроса", err.Error())
        return
    }
    
    response, err := h.service.NewMethod(r.Context(), &req)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Ошибка", err.Error())
        return
    }
    
    respondJSON(w, http.StatusOK, response)
}
```

3. **Зарегистрируйте маршрут** (`internal/handlers/handlers.go`):

```go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1", func(r chi.Router) {
        r.Post("/new-endpoint", h.NewHandler)
        // ...
    })
}
```

4. **Добавьте модель** (`internal/models/cashback.go` или новый файл):

```go
type NewRequest struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}

type NewResponse struct {
    Result string `json:"result"`
}
```

### Добавление новой таблицы в БД

1. **Создайте миграцию** (`migrations/004_new_table.sql`):

```sql
CREATE TABLE IF NOT EXISTS new_table (
    id BIGSERIAL PRIMARY KEY,
    field1 TEXT NOT NULL,
    field2 INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_new_table_field1 ON new_table(field1);
```

2. **Примените миграцию**:

```bash
make migrate
```

3. **Добавьте методы в Repository** (`internal/database/repository.go`):

```go
func (r *Repository) CreateNewRecord(ctx context.Context, field1 string, field2 int) error {
    _, err := r.db.Pool.Exec(ctx, 
        "INSERT INTO new_table (field1, field2) VALUES ($1, $2)",
        field1, field2)
    return err
}
```

---

## Тестирование

### Запуск тестов

```bash
# Все тесты
make test

# Конкретный пакет
go test ./internal/validator/...

# С покрытием
go test -cover ./...

# С детальным выводом
go test -v ./...
```

### Написание тестов

Создайте файл `*_test.go` в том же пакете:

```go
package validator

import (
    "testing"
)

func TestValidateTextField(t *testing.T) {
    tests := []struct {
        name    string
        field   string
        wantErr bool
    }{
        {"valid", "Test", false},
        {"empty", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateTextField("test", tt.field, true)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateTextField() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Форматирование и линтинг

### Форматирование кода

```bash
# Автоматическое форматирование
make fmt

# Или напрямую
go fmt ./...
```

### Линтинг

Установите `golangci-lint`:

```bash
# Linux/macOS
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

Запуск линтера:

```bash
make lint
```

---

## Отладка

### Отладка API сервера

Используйте встроенный отладчик Go или Delve:

```bash
# Установка Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Запуск с отладчиком
dlv debug cmd/server/main.go
```

В VS Code:
1. Установите расширение Go
2. Создайте `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch API Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/server",
            "env": {
                "DB_HOST": "localhost",
                "DB_PORT": "5432",
                "DB_USER": "postgres",
                "DB_PASSWORD": "postgres",
                "DB_NAME": "cashback_db"
            }
        }
    ]
}
```

### Отладка бота

Для отладки бота включите режим отладки:

```bash
export BOT_DEBUG=true
make run-bot
```

### Логирование

Используйте стандартный пакет `log`:

```go
import "log"

log.Printf("Debug: %v", value)
log.Printf("Error: %v", err)
```

---

## Работа с Git

### Workflow

1. **Создайте ветку для новой функции**:

```bash
git checkout -b feature/new-feature
```

2. **Внесите изменения и закоммитьте**:

```bash
git add .
git commit -m "feat: добавлена новая функция"
```

3. **Отправьте изменения**:

```bash
git push origin feature/new-feature
```

4. **Создайте Pull Request**

### Соглашения о коммитах

Используйте префиксы:
- `feat:` — новая функция
- `fix:` — исправление бага
- `docs:` — изменения в документации
- `refactor:` — рефакторинг
- `test:` — добавление тестов
- `chore:` — обновление зависимостей и т.д.

Примеры:
```
feat: добавлена команда /bankinfo
fix: исправлена ошибка валидации даты
docs: обновлена документация API
```

---

## Производительность

### Профилирование

Используйте встроенные инструменты Go:

```bash
# CPU профилирование
go test -cpuprofile=cpu.prof ./...

# Memory профилирование
go test -memprofile=mem.prof ./...

# Анализ профиля
go tool pprof cpu.prof
```

### Оптимизация запросов

1. Проверьте использование индексов:

```sql
EXPLAIN ANALYZE SELECT * FROM cashback_rules WHERE group_name = 'Семья';
```

2. Обновите статистику:

```sql
ANALYZE cashback_rules;
```

3. Оптимизируйте запросы в `internal/database/queries.go`

---

## Полезные инструменты

### Генерация документации

```bash
# Генерация Go документации
go doc ./internal/service

# Запуск локального сервера документации
godoc -http=:6060
```

### Проверка зависимостей

```bash
# Обновление зависимостей
go get -u ./...

# Проверка уязвимостей
go list -json -m all | nancy sleuth
```

### Генерация моков

Используйте `mockgen` для генерации моков:

```bash
go install github.com/golang/mock/mockgen@latest
mockgen -source=internal/service/interfaces.go -destination=internal/service/mocks.go
```

---

## Следующие шаги

- [Решение проблем](09-troubleshooting.md) — решение проблем при разработке
- [Вклад в проект](10-contributing.md) — как внести вклад
- [Архитектура](02-architecture.md) — понимание архитектуры

