# Установка и настройка

## Требования

### Системные требования

- **Операционная система**: Linux, macOS, или Windows (с WSL2)
- **Docker**: версия 20.10 или выше
- **Docker Compose**: версия 2.0 или выше
- **Git**: для клонирования репозитория

### Для локальной разработки (без Docker)

- **Go**: версия 1.22 или выше
- **PostgreSQL**: версия 15 или выше
- **Telegram Bot Token**: полученный от [@BotFather](https://t.me/BotFather)

## Быстрая установка (Docker)

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd Bot
```

### 2. Настройка переменных окружения

Скопируйте файл примера конфигурации:

```bash
cp env.example .env
```

Отредактируйте `.env` файл и укажите необходимые значения:

```bash
# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_NAME=cashback_db
DB_SSLMODE=disable

# Сервер API
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
API_BASE_URL=http://api:8080
BOT_DEBUG=false
```

**Важно**: 
- Замените `your_secure_password_here` на надежный пароль для PostgreSQL
- Получите токен бота от [@BotFather](https://t.me/BotFather) и укажите его в `TELEGRAM_BOT_TOKEN`

### 3. Запуск системы

```bash
docker-compose -f docker-compose.full.yml up -d
```

Эта команда:
- Создаст и запустит контейнеры PostgreSQL, API сервера и Telegram бота
- Выполнит миграции базы данных (если настроены)
- Запустит все сервисы в фоновом режиме

### 4. Проверка работоспособности

**Проверка статуса контейнеров**:
```bash
docker-compose -f docker-compose.full.yml ps
```

**Проверка логов**:
```bash
# Логи всех сервисов
docker-compose -f docker-compose.full.yml logs

# Логи конкретного сервиса
docker-compose -f docker-compose.full.yml logs bot
docker-compose -f docker-compose.full.yml logs api
docker-compose -f docker-compose.full.yml logs postgres
```

**Проверка API**:
```bash
curl http://localhost:8080/health
```

Должен вернуться ответ:
```json
{"status":"ok"}
```

**Проверка бота**:
Найдите вашего бота в Telegram и отправьте команду `/start`

## Установка для разработки (без Docker)

### 1. Установка зависимостей

**Установка Go**:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# macOS
brew install go

# Проверка версии
go version
```

**Установка PostgreSQL**:
```bash
# Ubuntu/Debian
sudo apt install postgresql-15

# macOS
brew install postgresql@15

# Запуск PostgreSQL
sudo systemctl start postgresql  # Linux
brew services start postgresql@15  # macOS
```

### 2. Настройка базы данных

**Создание базы данных**:
```bash
sudo -u postgres psql
```

В консоли PostgreSQL:
```sql
CREATE DATABASE cashback_db;
CREATE USER cashback_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE cashback_db TO cashback_user;
\q
```

**Применение миграций**:
```bash
# Установите переменные окружения
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=cashback_user
export DB_PASSWORD=your_password
export DB_NAME=cashback_db
export DB_SSLMODE=disable

# Примените миграции
psql -h localhost -U cashback_user -d cashback_db -f migrations/001_initial_schema.sql
psql -h localhost -U cashback_user -d cashback_db -f migrations/003_user_groups.sql
```

### 3. Клонирование и сборка

```bash
# Клонирование
git clone <repository-url>
cd Bot

# Установка зависимостей
go mod download

# Сборка API сервера
go build -o bin/server cmd/server/main.go

# Сборка бота
go build -o bin/bot cmd/bot/main.go
```

### 4. Настройка переменных окружения

Создайте файл `.env` или экспортируйте переменные:

```bash
# Для API сервера
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=cashback_user
export DB_PASSWORD=your_password
export DB_NAME=cashback_db
export DB_SSLMODE=disable
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Для бота
export TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
export API_BASE_URL=http://localhost:8080
export BOT_DEBUG=true
```

### 5. Запуск приложений

**В первом терминале - API сервер**:
```bash
./bin/server
```

**Во втором терминале - Telegram бот**:
```bash
./bin/bot
```

## Получение Telegram Bot Token

1. Откройте Telegram и найдите [@BotFather](https://t.me/BotFather)
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Скопируйте полученный токен (формат: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)
5. Укажите токен в переменной окружения `TELEGRAM_BOT_TOKEN`

## Проверка установки

### Проверка API сервера

```bash
# Health check
curl http://localhost:8080/health

# Список кэшбэков (должен вернуть пустой список)
curl http://localhost:8080/api/v1/cashback?limit=10
```

### Проверка бота

1. Найдите вашего бота в Telegram по имени, указанному при создании
2. Отправьте команду `/start`
3. Бот должен ответить приветственным сообщением

### Проверка базы данных

```bash
# Подключение к БД
psql -h localhost -U cashback_user -d cashback_db

# Проверка таблиц
\dt

# Проверка расширений
\dx
```

Должны быть видны:
- Таблица `cashback_rules`
- Таблица `user_groups`
- Расширение `pg_trgm`

## Первоначальная настройка

### Создание первой группы

После запуска бота:

1. Отправьте команду `/creategroup НазваниеГруппы`
2. Бот создаст группу и автоматически добавит вас в неё

### Добавление первого кэшбэка

Отправьте боту сообщение в формате:
```
Категория, Банк, Процент, Макс.сумма
```

Например:
```
Рестораны, Сбербанк, 5, 1000
```

Бот проверит данные и предложит подтвердить или исправить их.

## Обновление системы

### Обновление через Docker

```bash
# Остановка контейнеров
docker-compose -f docker-compose.full.yml down

# Получение обновлений
git pull

# Пересборка и запуск
docker-compose -f docker-compose.full.yml up -d --build
```

### Обновление для разработки

```bash
# Получение обновлений
git pull

# Обновление зависимостей
go mod download
go mod tidy

# Пересборка
go build -o bin/server cmd/server/main.go
go build -o bin/bot cmd/bot/main.go
```

## Решение проблем при установке

### Проблема: Контейнеры не запускаются

**Решение**:
- Проверьте, что Docker запущен: `docker ps`
- Проверьте логи: `docker-compose -f docker-compose.full.yml logs`
- Убедитесь, что порты 5432 и 8080 свободны

### Проблема: Ошибка подключения к БД

**Решение**:
- Проверьте переменные окружения в `.env`
- Убедитесь, что PostgreSQL запущен
- Проверьте правильность пароля и имени пользователя

### Проблема: Бот не отвечает

**Решение**:
- Проверьте токен бота в `.env`
- Проверьте логи бота: `docker-compose -f docker-compose.full.yml logs bot`
- Убедитесь, что API сервер запущен и доступен

### Проблема: Ошибки миграций

**Решение**:
- Убедитесь, что расширение `pg_trgm` установлено
- Проверьте права доступа пользователя БД
- Выполните миграции вручную (см. раздел "Настройка базы данных")

## Дополнительные настройки

### Настройка логирования

Для включения детального логирования установите:
```bash
export BOT_DEBUG=true
```

### Настройка производительности

В файле `internal/database/database.go` можно настроить пул соединений:
- `MaxConns`: максимальное количество соединений
- `MinConns`: минимальное количество соединений
- `MaxConnLifetime`: время жизни соединения

### Настройка CORS

В файле `cmd/server/main.go` можно настроить CORS для веб-приложений:
```go
AllowedOrigins: []string{"https://yourdomain.com"}
```

## Следующие шаги

- [API Справочник](04-api-reference.md) — изучите возможности API
- [Команды бота](05-bot-commands.md) — узнайте все команды бота
- [Развертывание](07-deployment.md) — развертывание на продакшн сервере

