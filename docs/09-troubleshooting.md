# Решение проблем

## Общие проблемы

### Проблема: Приложение не запускается

**Симптомы**:
- Ошибки при запуске
- Контейнеры падают сразу после запуска

**Решения**:

1. **Проверьте переменные окружения**:
```bash
# Проверка .env файла
cat .env

# Проверка переменных в контейнере
docker-compose -f docker-compose.full.yml exec bot env
```

2. **Проверьте логи**:
```bash
docker-compose -f docker-compose.full.yml logs bot
docker-compose -f docker-compose.full.yml logs api
docker-compose -f docker-compose.full.yml logs postgres
```

3. **Проверьте доступность портов**:
```bash
# Проверка занятости портов
sudo netstat -tulpn | grep :8080
sudo netstat -tulpn | grep :5432
```

4. **Пересоберите образы**:
```bash
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml build --no-cache
docker-compose -f docker-compose.full.yml up -d
```

---

## Проблемы с базой данных

### Проблема: Ошибка подключения к PostgreSQL

**Симптомы**:
```
❌ Не удалось подключиться к базе данных: connection refused
```

**Решения**:

1. **Проверьте, запущен ли PostgreSQL**:
```bash
docker-compose -f docker-compose.full.yml ps postgres
```

2. **Проверьте переменные окружения**:
```bash
# Должны быть установлены:
# DB_HOST=postgres (для Docker) или localhost (для локальной разработки)
# DB_PORT=5432
# DB_USER=postgres
# DB_PASSWORD=your_password
# DB_NAME=cashback_db
```

3. **Проверьте подключение вручную**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db
```

4. **Проверьте логи PostgreSQL**:
```bash
docker-compose -f docker-compose.full.yml logs postgres
```

5. **Перезапустите PostgreSQL**:
```bash
docker-compose -f docker-compose.full.yml restart postgres
```

---

### Проблема: Ошибка "extension pg_trgm does not exist"

**Симптомы**:
```
ERROR: extension "pg_trgm" does not exist
```

**Решения**:

1. **Создайте расширение вручную**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -c "CREATE EXTENSION IF NOT EXISTS pg_trgm;"
```

2. **Проверьте, что расширение установлено**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -c "\dx"
```

---

### Проблема: Ошибка миграций

**Симптомы**:
```
ERROR: relation "cashback_rules" already exists
```

**Решения**:

1. **Проверьте существующие таблицы**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -c "\dt"
```

2. **Если таблицы уже существуют, пропустите миграцию или откатите её**:
```bash
# Откат миграций
./scripts/rollback.sh

# Повторное применение
./scripts/migrate.sh
```

3. **Полный сброс БД (ОСТОРОЖНО: удалит все данные)**:
```bash
docker-compose -f docker-compose.full.yml down -v
docker-compose -f docker-compose.full.yml up -d postgres
sleep 10
./scripts/migrate.sh
```

---

### Проблема: Медленные запросы

**Симптомы**:
- Запросы выполняются долго
- Высокая нагрузка на БД

**Решения**:

1. **Проверьте использование индексов**:
```sql
EXPLAIN ANALYZE SELECT * FROM cashback_rules WHERE group_name = 'Семья';
```

2. **Обновите статистику**:
```sql
ANALYZE cashback_rules;
ANALYZE user_groups;
```

3. **Проверьте размер таблиц**:
```sql
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public';
```

4. **Оптимизируйте запросы** - проверьте `internal/database/queries.go`

---

## Проблемы с Telegram ботом

### Проблема: Бот не отвечает на сообщения

**Симптомы**:
- Бот запущен, но не отвечает
- Сообщения не доставляются

**Решения**:

1. **Проверьте токен бота**:
```bash
# Проверка переменной окружения
echo $TELEGRAM_BOT_TOKEN

# Проверка в контейнере
docker-compose -f docker-compose.full.yml exec bot env | grep TELEGRAM_BOT_TOKEN
```

2. **Проверьте логи бота**:
```bash
docker-compose -f docker-compose.full.yml logs bot
```

3. **Проверьте подключение к API серверу**:
```bash
# Проверка переменной API_BASE_URL
docker-compose -f docker-compose.full.yml exec bot env | grep API_BASE_URL

# Проверка доступности API
docker-compose -f docker-compose.full.yml exec bot wget -O- http://api:8080/health
```

4. **Перезапустите бота**:
```bash
docker-compose -f docker-compose.full.yml restart bot
```

---

### Проблема: Ошибка "bot not found" или "Unauthorized"

**Симптомы**:
```
❌ Не удалось создать бота: Unauthorized
```

**Решения**:

1. **Проверьте правильность токена**:
   - Получите новый токен от [@BotFather](https://t.me/BotFather)
   - Убедитесь, что токен скопирован полностью без пробелов

2. **Обновите токен в .env**:
```bash
nano .env
# Обновите TELEGRAM_BOT_TOKEN
```

3. **Перезапустите бота**:
```bash
docker-compose -f docker-compose.full.yml restart bot
```

---

### Проблема: Бот не может подключиться к API

**Симптомы**:
```
❌ Ошибка запроса: connection refused
```

**Решения**:

1. **Проверьте, запущен ли API сервер**:
```bash
docker-compose -f docker-compose.full.yml ps api
```

2. **Проверьте переменную API_BASE_URL**:
   - Для Docker: `http://api:8080`
   - Для локальной разработки: `http://localhost:8080`

3. **Проверьте сеть Docker**:
```bash
docker network ls
docker network inspect cashback-network
```

4. **Проверьте доступность API из контейнера бота**:
```bash
docker-compose -f docker-compose.full.yml exec bot wget -O- http://api:8080/health
```

---

## Проблемы с API сервером

### Проблема: API сервер не отвечает

**Симптомы**:
- Запросы к API возвращают ошибки
- Сервер не запускается

**Решения**:

1. **Проверьте логи API**:
```bash
docker-compose -f docker-compose.full.yml logs api
```

2. **Проверьте health check**:
```bash
curl http://localhost:8080/health
```

3. **Проверьте порт**:
```bash
# Проверка занятости порта
sudo lsof -i :8080

# Или
sudo netstat -tulpn | grep :8080
```

4. **Перезапустите API**:
```bash
docker-compose -f docker-compose.full.yml restart api
```

---

### Проблема: Ошибка CORS

**Симптомы**:
```
Access to fetch at 'http://localhost:8080/api/v1/...' from origin 'http://localhost:3000' has been blocked by CORS policy
```

**Решения**:

1. **Настройте CORS в `cmd/server/main.go`**:
```go
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"http://localhost:3000", "https://yourdomain.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: false,
    MaxAge:           300,
}))
```

2. **Пересоберите и перезапустите**:
```bash
docker-compose -f docker-compose.full.yml build api
docker-compose -f docker-compose.full.yml restart api
```

---

## Проблемы с Docker

### Проблема: Контейнеры не запускаются

**Симптомы**:
- Контейнеры падают сразу после запуска
- Ошибки при сборке образов

**Решения**:

1. **Проверьте логи**:
```bash
docker-compose -f docker-compose.full.yml logs
```

2. **Проверьте доступность ресурсов**:
```bash
# Проверка места на диске
df -h

# Проверка памяти
free -h

# Проверка Docker
docker info
```

3. **Очистите Docker**:
```bash
# Остановка всех контейнеров
docker-compose -f docker-compose.full.yml down

# Очистка неиспользуемых ресурсов
docker system prune -a

# Пересборка
docker-compose -f docker-compose.full.yml build --no-cache
docker-compose -f docker-compose.full.yml up -d
```

---

### Проблема: Ошибка "port is already allocated"

**Симптомы**:
```
Error: bind: address already in use
```

**Решения**:

1. **Найдите процесс, использующий порт**:
```bash
sudo lsof -i :8080
sudo lsof -i :5432
```

2. **Остановите процесс или измените порт**:
```bash
# Остановка процесса
sudo kill -9 <PID>

# Или измените порт в docker-compose.yml
```

---

### Проблема: Ошибка "network not found"

**Симптомы**:
```
Error: network cashback-network not found
```

**Решения**:

1. **Создайте сеть вручную**:
```bash
docker network create cashback-network
```

2. **Или пересоздайте контейнеры**:
```bash
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml up -d
```

---

## Проблемы с валидацией

### Проблема: Ошибки валидации данных

**Симптомы**:
- API возвращает ошибки валидации
- Бот не принимает данные

**Решения**:

1. **Проверьте формат данных**:
   - `month_year`: должен быть в формате `YYYY-MM`
   - `cashback_percent`: должен быть от 0 до 100
   - `max_amount`: должен быть >= 0

2. **Проверьте валидатор** (`internal/validator/validator.go`):
```bash
# Запуск тестов валидатора
go test ./internal/validator/...
```

3. **Проверьте логи**:
```bash
docker-compose -f docker-compose.full.yml logs api | grep validation
```

---

## Проблемы с производительностью

### Проблема: Медленная работа приложения

**Симптомы**:
- Запросы выполняются долго
- Высокая нагрузка на сервер

**Решения**:

1. **Проверьте использование ресурсов**:
```bash
docker stats
htop
```

2. **Оптимизируйте запросы к БД**:
   - Проверьте использование индексов
   - Оптимизируйте SQL запросы
   - Обновите статистику БД

3. **Настройте пул соединений**:
   - Увеличьте `MaxConns` в `internal/database/database.go`
   - Настройте таймауты

4. **Добавьте кэширование** (если нужно)

---

## Проблемы с миграциями

### Проблема: Миграции не применяются

**Решения**:

1. **Проверьте подключение к БД**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db
```

2. **Примените миграции вручную**:
```bash
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -f /path/to/migrations/001_initial_schema.sql
```

3. **Проверьте права доступа**:
```sql
GRANT ALL PRIVILEGES ON DATABASE cashback_db TO postgres;
```

---

## Получение помощи

### Логи и диагностика

Соберите следующую информацию для диагностики:

1. **Версии**:
```bash
go version
docker --version
docker-compose --version
```

2. **Логи**:
```bash
docker-compose -f docker-compose.full.yml logs > logs.txt
```

3. **Статус контейнеров**:
```bash
docker-compose -f docker-compose.full.yml ps > status.txt
```

4. **Конфигурация**:
```bash
cat .env > config.txt
```

### Полезные команды для диагностики

```bash
# Проверка всех сервисов
docker-compose -f docker-compose.full.yml ps

# Логи всех сервисов
docker-compose -f docker-compose.full.yml logs --tail=100

# Использование ресурсов
docker stats --no-stream

# Проверка сети
docker network inspect cashback-network

# Проверка томов
docker volume ls
```

---

## Часто задаваемые вопросы

### Q: Как полностью переустановить приложение?

```bash
# Остановка и удаление контейнеров
docker-compose -f docker-compose.full.yml down -v

# Удаление образов
docker rmi $(docker images | grep cashback | awk '{print $3}')

# Очистка
docker system prune -a

# Повторная установка
git pull
cp env.example .env
# Отредактируйте .env
docker-compose -f docker-compose.full.yml up -d
./scripts/migrate.sh
```

### Q: Как откатить изменения?

```bash
# Откат через Git
git reset --hard HEAD~1
git pull

# Пересборка
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml build
docker-compose -f docker-compose.full.yml up -d
```

### Q: Как проверить, что всё работает?

```bash
# 1. Проверка контейнеров
docker-compose -f docker-compose.full.yml ps

# 2. Проверка API
curl http://localhost:8080/health

# 3. Проверка БД
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -c "SELECT COUNT(*) FROM cashback_rules;"

# 4. Проверка бота
# Отправьте /start боту в Telegram
```

---

## Следующие шаги

- [Установка](03-installation.md) — повторная установка
- [Развертывание](07-deployment.md) — настройка продакшн
- [Разработка](08-development.md) — разработка новых функций

