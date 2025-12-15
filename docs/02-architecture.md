# Архитектура системы

## Общая архитектура

Система построена по принципу **микросервисной архитектуры** с разделением на три основных компонента:

```
┌─────────────────┐
│  Telegram Users │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      HTTP/REST      ┌─────────────────┐
│  Telegram Bot   │◄───────────────────►│   API Server    │
│   (cmd/bot)     │                      │  (cmd/server)   │
└─────────────────┘                      └────────┬────────┘
                                                  │
                                                  │ PostgreSQL
                                                  ▼
                                         ┌─────────────────┐
                                         │   PostgreSQL    │
                                         │    Database     │
                                         └─────────────────┘
```

## Компоненты системы

### 1. Telegram Bot (`cmd/bot`)

**Назначение**: Обработка взаимодействия с пользователями через Telegram.

**Основные компоненты**:

- **Bot** (`internal/bot/bot.go`) — основной класс бота, управляет состояниями пользователей и маршрутизацией сообщений
- **APIClient** (`internal/bot/client.go`) — HTTP-клиент для взаимодействия с API сервером
- **Commands** (`internal/bot/commands.go`) — обработчики команд бота
- **States** (`internal/bot/states.go`) — управление состояниями диалога
- **Parser** (`internal/bot/parser.go`) — парсинг входящих сообщений
- **Keyboard** (`internal/bot/keyboard.go`) — генерация клавиатур

**Особенности**:
- State machine для управления диалогами
- Поддержка inline и reply клавиатур
- Автоматическая валидация через API перед созданием кэшбэков
- Fuzzy-поиск для исправления опечаток

### 2. HTTP API Server (`cmd/server`)

**Назначение**: Предоставление REST API для управления кэшбэками.

**Стек технологий**:
- **Router**: Chi v5
- **Middleware**: RequestID, Logger, Recoverer, Timeout, CORS
- **Database**: PostgreSQL через pgx/v5

**Слои архитектуры**:

```
HTTP Handlers (handlers/)
    ↓
Service Layer (service/)
    ↓
Repository Layer (database/)
    ↓
PostgreSQL Database
```

**Компоненты**:

- **Handlers** (`internal/handlers/handlers.go`) — HTTP обработчики запросов
- **Service** (`internal/service/service.go`) — бизнес-логика приложения
- **Repository** (`internal/database/repository.go`) — работа с БД
- **Database** (`internal/database/database.go`) — подключение к PostgreSQL

**Особенности**:
- Graceful shutdown
- Таймауты для запросов
- CORS поддержка
- Валидация на уровне сервиса

### 3. PostgreSQL Database

**Назначение**: Хранение всех данных системы.

**Основные таблицы**:

- `cashback_rules` — кэшбэки
- `user_groups` — группы пользователей (из миграции 003)

**Особенности**:
- Расширение `pg_trgm` для fuzzy-поиска
- Триграммные индексы (GIN) для быстрого поиска
- Композитные индексы для оптимизации запросов
- Автоматическое обновление `updated_at` через триггеры

## Потоки данных

### Добавление кэшбэка

```
User → Bot → API Client → API Server → Service → Repository → Database
                                                              ↓
User ← Bot ← API Client ← API Server ← Service ← Repository ←
```

**Детальный поток**:

1. Пользователь отправляет сообщение боту в формате: `Категория, Банк, Процент, Макс.сумма`
2. Бот парсит сообщение и создает `SuggestRequest`
3. Бот отправляет запрос на `/api/v1/cashback/suggest` для валидации
4. API сервер валидирует данные и выполняет fuzzy-поиск
5. Бот показывает предложения пользователю
6. Пользователь подтверждает или исправляет данные
7. Бот отправляет `CreateCashbackRequest` на `/api/v1/cashback`
8. API сервер создает кэшбэк в БД
9. Бот подтверждает создание пользователю

### Поиск лучшего кэшбэка

```
User → Bot → API Client → API Server → Service → Repository → Database
                                                              ↓
User ← Bot ← API Client ← API Server ← Service ← Repository ←
```

**Алгоритм поиска**:

1. Пользователь запрашивает лучший кэшбэк для категории
2. Бот отправляет запрос на `/api/v1/cashback/best?group_name=X&category=Y&month_year=Z`
3. Service ищет кэшбэк с точным совпадением категории
4. Если не найдено, ищет кэшбэк "Все покупки"
5. Возвращает кэшбэк с максимальным процентом

## Управление состояниями (State Machine)

Бот использует state machine для управления диалогами:

```go
StateNone                       // Нет активного состояния
StateAwaitingConfirmation       // Ожидание подтверждения
StateAwaitingBankCorrection     // Ожидание исправления банка
StateAwaitingCategoryCorrection // Ожидание исправления категории
StateAwaitingUpdateData         // Ожидание данных для обновления
StateAwaitingDeleteConfirm      // Ожидание подтверждения удаления
StateAwaitingGroupName          // Ожидание названия группы
StateAwaitingManualInput        // Ожидание ручного ввода
StateAwaitingBestCategory       // Ожидание категории для поиска
StateAwaitingBankInfoName       // Ожидание названия банка для информации
StateAwaitingUpdateID           // Ожидание ID для обновления
StateAwaitingDeleteID           // Ожидание ID для удаления
StateAwaitingJoinGroupName      // Ожидание названия группы для присоединения
StateAwaitingCreateGroupName    // Ожидание названия группы для создания
```

## Слои приложения

### Presentation Layer (Handlers)

**Ответственность**: Обработка HTTP запросов, валидация входных данных, форматирование ответов.

**Файлы**: `internal/handlers/handlers.go`

**Пример**:
```go
func (h *Handler) CreateCashback(w http.ResponseWriter, r *http.Request) {
    var req models.CreateCashbackRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    rule, err := h.service.CreateCashback(r.Context(), &req)
    // Обработка ответа
}
```

### Business Logic Layer (Service)

**Ответственность**: Бизнес-логика, валидация, координация между компонентами.

**Файлы**: `internal/service/service.go`

**Основные методы**:
- `Suggest()` — анализ и предложения
- `CreateCashback()` — создание кэшбэка
- `GetBestCashback()` — поиск лучшего кэшбэка
- `UpdateCashback()` — обновление кэшбэка
- `DeleteCashback()` — удаление кэшбэка
- `ListCashback()` — список кэшбэков с пагинацией

**Особенности**:
- Fuzzy-поиск с порогами схожести
- Валидация через `validator` пакет
- Fallback на "Все покупки" при поиске лучшего кэшбэка

### Data Access Layer (Repository)

**Ответственность**: Работа с базой данных, выполнение SQL запросов.

**Файлы**: 
- `internal/database/repository.go` — реализация репозитория
- `internal/database/queries.go` — SQL запросы
- `internal/database/interfaces.go` — интерфейсы

**Основные методы**:
- `Create()` — создание записи
- `GetByID()` — получение по ID
- `Update()` — обновление
- `Delete()` — удаление
- `List()` — список с пагинацией
- `FuzzySearch*()` — fuzzy-поиск по полям
- `GetBestCashback()` — поиск лучшего кэшбэка

## Конфигурация

### Переменные окружения

**API Server**:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `SERVER_HOST`, `SERVER_PORT`

**Telegram Bot**:
- `TELEGRAM_BOT_TOKEN` — токен бота
- `API_BASE_URL` — URL API сервера
- `BOT_DEBUG` — режим отладки

### Загрузка конфигурации

Конфигурация загружается из переменных окружения с значениями по умолчанию:

```go
cfg := config.Load()  // Для сервера
cfg := bot.LoadConfig()  // Для бота
```

## База данных

### Схема данных

**Таблица `cashback_rules`**:
```sql
- id (BIGSERIAL PRIMARY KEY)
- group_name (TEXT NOT NULL)
- category (TEXT NOT NULL)
- bank_name (TEXT NOT NULL)
- user_id (TEXT NOT NULL)
- user_display_name (TEXT NOT NULL)
- month_year (DATE NOT NULL)
- cashback_percent (NUMERIC(5,2) CHECK 0-100)
- max_amount (NUMERIC(10,2) CHECK >= 0)
- created_at (TIMESTAMPTZ DEFAULT NOW())
- updated_at (TIMESTAMPTZ DEFAULT NOW())
```

**Таблица `user_groups`** (из миграции 003):
```sql
- user_id (TEXT PRIMARY KEY)
- group_name (TEXT NOT NULL)
- joined_at (TIMESTAMPTZ DEFAULT NOW())
```

### Индексы

- **Триграммные индексы** (GIN) для fuzzy-поиска:
  - `idx_group_trgm` на `group_name`
  - `idx_category_trgm` на `category`
  - `idx_bank_trgm` на `bank_name`
  - `idx_user_name_trgm` на `user_display_name`

- **Композитные индексы**:
  - `idx_group_month_cat` на `(group_name, month_year, category)`
  - `idx_user_id` на `user_id`

### Пул соединений

Настройки пула PostgreSQL:
- `MaxConns`: 25
- `MinConns`: 5
- `MaxConnLifetime`: 1 час
- `MaxConnIdleTime`: 30 минут
- `HealthCheckPeriod`: 1 минута

## Безопасность

### Валидация данных

Все входные данные валидируются через пакет `validator`:
- Проверка текстовых полей (длина, формат)
- Проверка процента кэшбэка (0-100%)
- Проверка максимальной суммы (>= 0)
- Проверка формата даты (YYYY-MM)

### Обработка ошибок

- Все ошибки логируются
- Пользователю возвращаются понятные сообщения
- API возвращает структурированные ошибки в формате JSON

## Масштабируемость

### Горизонтальное масштабирование

- **API Server**: Можно запустить несколько экземпляров за балансировщиком
- **Telegram Bot**: Один экземпляр (Telegram API не поддерживает несколько ботов с одним токеном)
- **Database**: PostgreSQL с репликацией для чтения

### Вертикальное масштабирование

- Настройка пула соединений БД
- Увеличение лимитов пагинации
- Оптимизация индексов

## Мониторинг и логирование

### Логирование

- Стандартный `log` пакет Go
- Логирование всех HTTP запросов через middleware
- Логирование ошибок с контекстом

### Health Check

Эндпоинт `/health` для проверки работоспособности сервера.

## Следующие шаги

- [Установка и настройка](03-installation.md) — как развернуть систему
- [API Справочник](04-api-reference.md) — детальная документация API
- [База данных](06-database.md) — работа с БД и миграциями

