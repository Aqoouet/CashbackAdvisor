# 📁 Структура проекта Open Cashback Advisor

```
open-cashback-advisor/
│
├── 📄 README.md                    # Основная документация
├── 📄 QUICKSTART.md                # Быстрый старт
├── 📄 ARCHITECTURE.md              # Подробная архитектура
├── 📄 PROJECT_STRUCTURE.md         # Этот файл
├── 📄 LICENSE                      # MIT лицензия
│
├── 📄 go.mod                       # Go модули
├── 📄 go.sum                       # Контрольные суммы зависимостей
├── 📄 Makefile                     # Команды сборки и управления
├── 📄 Dockerfile                   # Docker образ приложения
├── 📄 .dockerignore                # Исключения для Docker
├── 📄 docker-compose.yml           # Конфигурация Docker Compose
├── 📄 .gitignore                   # Исключения для Git
│
├── 📁 cmd/                         # Точки входа приложения
│   └── 📁 server/
│       └── 📄 main.go              # Главный файл сервера
│
├── 📁 internal/                    # Внутренняя логика приложения
│   │
│   ├── 📁 config/                  # Конфигурация
│   │   └── 📄 config.go            # Загрузка настроек из env
│   │
│   ├── 📁 models/                  # Модели данных
│   │   └── 📄 cashback.go          # Структуры и DTO
│   │
│   ├── 📁 validator/               # Валидация
│   │   ├── 📄 validator.go         # Правила валидации
│   │   └── 📄 validator_test.go    # Unit тесты валидатора
│   │
│   ├── 📁 database/                # Работа с БД
│   │   ├── 📄 database.go          # Подключение и пул
│   │   └── 📄 repository.go        # CRUD + Fuzzy-поиск
│   │
│   ├── 📁 service/                 # Бизнес-логика
│   │   └── 📄 service.go           # Оркестрация операций
│   │
│   └── 📁 handlers/                # HTTP обработчики
│       └── 📄 handlers.go          # REST API endpoints
│
├── 📁 migrations/                  # Миграции базы данных
│   ├── 📄 001_initial_schema.sql   # Создание таблиц и индексов
│   └── 📄 002_down.sql             # Откат миграций
│
├── 📁 scripts/                     # Вспомогательные скрипты
│   ├── 📄 migrate.sh               # Применение миграций
│   └── 📄 rollback.sh              # Откат миграций
│
└── 📁 examples/                    # Примеры использования
    └── 📄 requests.http            # HTTP запросы для тестирования
```

## 📊 Статистика проекта

### Файлы Go

| Компонент | Файл | Строк | Описание |
|-----------|------|-------|----------|
| Main | `cmd/server/main.go` | ~100 | Точка входа, HTTP сервер |
| Models | `internal/models/cashback.go` | ~110 | Структуры данных |
| Validator | `internal/validator/validator.go` | ~200 | Валидация данных |
| Tests | `internal/validator/validator_test.go` | ~250 | Unit тесты |
| Database | `internal/database/database.go` | ~50 | Подключение к БД |
| Repository | `internal/database/repository.go` | ~300 | SQL запросы, fuzzy-поиск |
| Service | `internal/service/service.go` | ~200 | Бизнес-логика |
| Handlers | `internal/handlers/handlers.go` | ~200 | REST API |
| Config | `internal/config/config.go` | ~60 | Конфигурация |

**Всего: ~1470 строк Go кода**

### SQL

| Файл | Строк | Описание |
|------|-------|----------|
| `001_initial_schema.sql` | ~45 | Схема БД, индексы, триггеры |
| `002_down.sql` | ~5 | Откат миграций |

### Документация

| Файл | Размер | Описание |
|------|--------|----------|
| `README.md` | Large | Полная документация API |
| `QUICKSTART.md` | Medium | Быстрый старт |
| `ARCHITECTURE.md` | Large | Архитектура проекта |
| `PROJECT_STRUCTURE.md` | Medium | Структура проекта |

## 🔗 Взаимосвязи компонентов

```
main.go
  ↓
config.Load()
  ↓
database.New()
  ↓
database.NewRepository()
  ↓
service.NewService()
  ↓
handlers.NewHandler()
  ↓
chi.Router + Middleware
  ↓
HTTP Server (0.0.0.0:8080)
```

## 🎯 Ключевые особенности каждого компонента

### cmd/server/main.go
- Инициализация приложения
- Настройка middleware
- Graceful shutdown
- Красивое логирование

### internal/models/cashback.go
- Все структуры данных в одном месте
- JSON теги для сериализации
- Request/Response типы
- Типобезопасность

### internal/validator/validator.go
- Строгая валидация по ТЗ
- Понятные сообщения об ошибках
- Округление чисел
- 100% покрытие тестами

### internal/database/repository.go
- Параметризованные запросы (защита от SQL injection)
- Fuzzy-поиск через pg_trgm
- Динамическое построение UPDATE
- Пагинация

### internal/service/service.go
- Бизнес-логика
- Оркестрация операций
- Применение валидации
- Координация между слоями

### internal/handlers/handlers.go
- REST API endpoints
- JSON сериализация
- HTTP коды статусов
- Обработка ошибок

## 🧪 Тестирование

### Текущие тесты
- ✅ `validator_test.go` - 7 test suites, 24 test cases

### Планируемые тесты
- ⏳ `repository_test.go` - интеграционные тесты БД
- ⏳ `service_test.go` - unit тесты сервисного слоя
- ⏳ `handlers_test.go` - E2E тесты API

## 📦 Зависимости

### Прямые зависимости

```go
require (
    github.com/jackc/pgx/v5 v5.5.1      // PostgreSQL драйвер
    github.com/go-chi/chi/v5 v5.0.11    // HTTP роутер
    github.com/go-chi/cors v1.2.1       // CORS middleware
)
```

### Транзитивные зависимости
- `jackc/pgpassfile`
- `jackc/pgservicefile`
- `jackc/puddle/v2` (connection pooling)
- `golang.org/x/crypto`
- `golang.org/x/sync`
- `golang.org/x/text`

## 🐳 Docker

### Образы
- **Builder**: `golang:1.22-alpine` (~300MB)
- **Runtime**: `alpine:latest` (~5MB base)
- **Final Image**: ~15-20MB (multi-stage build)

### Volumes
- `postgres_data` - хранение данных PostgreSQL

## 🚀 Команды Make

```bash
make help          # Помощь
make build         # Сборка
make run           # Запуск
make test          # Тесты
make clean         # Очистка
make migrate       # Миграции
make rollback      # Откат
make deps          # Зависимости
make docker-up     # PostgreSQL up
make docker-down   # PostgreSQL down
make dev           # Полный dev stack
make fmt           # Форматирование
make lint          # Линтер
```

## 📈 Масштабируемость

### Текущая конфигурация
- Пул: 5-25 соединений
- Таймаут запроса: 60 секунд
- Пагинация: 20 записей по умолчанию

### Рекомендации для продакшена
- Увеличить пул до 50-100 соединений
- Добавить Redis для кэширования
- Настроить rate limiting
- Добавить метрики (Prometheus)

## 🔒 Безопасность

### Реализовано
- ✅ Параметризованные SQL запросы
- ✅ Валидация всех входных данных
- ✅ Таймауты на запросы
- ✅ CORS настроен

### Рекомендуется добавить
- ⏳ JWT аутентификацию
- ⏳ Rate limiting
- ⏳ HTTPS/TLS
- ⏳ Логирование безопасности

## 📝 Соответствие ТЗ

- ✅ Строгая валидация числовых и временных полей
- ✅ Fuzzy-поиск для исправления опечаток (pg_trgm)
- ✅ 4 текстовых поля с разными порогами
- ✅ Редактирование и удаление записей
- ✅ Конкурентная работа (транзакции)
- ✅ REST API со всеми эндпоинтами
- ✅ Интеграция с Telegram (через API)
- ✅ PostgreSQL 14+ с pg_trgm
- ✅ Go 1.22+
- ✅ Тесты (unit)

## 🎓 Демонстрация навыков

Этот проект демонстрирует:

1. **Go разработка**
   - Чистая архитектура
   - Правильная работа с ошибками
   - Тестирование
   - Параллелизм и конкурентность

2. **PostgreSQL**
   - Сложные запросы
   - Fuzzy-поиск (pg_trgm)
   - Индексы и оптимизация
   - Триггеры

3. **REST API**
   - RESTful дизайн
   - Правильные HTTP коды
   - Валидация
   - Документация

4. **DevOps**
   - Docker
   - Миграции
   - Makefile
   - CI/CD ready

5. **Документация**
   - README
   - Архитектура
   - Примеры
   - Комментарии

---

**Автор:** rymax1e  
**Версия:** 4.0  
**Дата:** Декабрь 2024

