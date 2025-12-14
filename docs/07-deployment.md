# Развертывание на продакшн

## Подготовка сервера

### Требования к серверу

- **ОС**: Linux (Ubuntu 20.04+ или аналогичная)
- **RAM**: минимум 2GB (рекомендуется 4GB+)
- **CPU**: минимум 2 ядра
- **Диск**: минимум 20GB свободного места
- **Docker**: версия 20.10+
- **Docker Compose**: версия 2.0+

### Установка Docker и Docker Compose

**Ubuntu/Debian**:
```bash
# Обновление пакетов
sudo apt update
sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Проверка установки
docker --version
docker-compose --version
```

---

## Настройка сервера

### Создание пользователя для приложения

```bash
# Создание пользователя
sudo adduser cashback

# Добавление в группу docker
sudo usermod -aG docker cashback

# Переключение на пользователя
su - cashback
```

### Настройка SSH ключей

Для безопасного доступа к серверу:

```bash
# На локальной машине
ssh-keygen -t ed25519 -C "your_email@example.com"

# Копирование ключа на сервер
ssh-copy-id cashback@your-server-ip
```

### Настройка firewall

```bash
# Установка UFW (если не установлен)
sudo apt install ufw

# Разрешение SSH
sudo ufw allow 22/tcp

# Разрешение HTTP/HTTPS (если нужен внешний доступ к API)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Включение firewall
sudo ufw enable
```

---

## Развертывание приложения

### 1. Клонирование репозитория

```bash
cd ~
git clone <repository-url> CashbackAdvisor
cd CashbackAdvisor
```

### 2. Настройка переменных окружения

```bash
# Копирование примера
cp env.example .env

# Редактирование конфигурации
nano .env
```

**Важные настройки для продакшн**:

```bash
# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_very_secure_password_here
DB_NAME=cashback_db
DB_SSLMODE=disable  # Для продакшн используйте 'require'

# Сервер API
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
API_BASE_URL=http://api:8080
BOT_DEBUG=false
```

**Безопасность**:
- Используйте надежные пароли (минимум 16 символов)
- Для продакшн включите SSL для БД: `DB_SSLMODE=require`
- Не коммитьте `.env` файл в Git

### 3. Применение миграций

```bash
# Запуск контейнеров
docker-compose -f docker-compose.full.yml up -d postgres

# Ожидание запуска БД
sleep 10

# Применение миграций
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -f /path/to/migrations/001_initial_schema.sql
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -f /path/to/migrations/003_user_groups.sql
```

Или используйте скрипт миграции:

```bash
./scripts/migrate.sh
```

### 4. Запуск приложения

```bash
# Запуск всех сервисов
docker-compose -f docker-compose.full.yml up -d

# Проверка статуса
docker-compose -f docker-compose.full.yml ps

# Просмотр логов
docker-compose -f docker-compose.full.yml logs -f
```

---

## Настройка Nginx (опционально)

Если нужен внешний доступ к API через домен:

### Установка Nginx

```bash
sudo apt install nginx
```

### Конфигурация Nginx

Создайте файл `/etc/nginx/sites-available/cashback-api`:

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Активация конфигурации

```bash
sudo ln -s /etc/nginx/sites-available/cashback-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Настройка SSL (Let's Encrypt)

```bash
# Установка Certbot
sudo apt install certbot python3-certbot-nginx

# Получение сертификата
sudo certbot --nginx -d api.yourdomain.com

# Автоматическое обновление
sudo certbot renew --dry-run
```

---

## Обновление приложения

### Автоматическое обновление

Используйте скрипт `update.sh`:

```bash
# С локальной машины
ssh cashback-server "cd ~/CashbackAdvisor && bash update.sh"
```

Скрипт автоматически:
1. Получает последние изменения из Git
2. Останавливает контейнеры
3. Пересобирает Docker образы
4. Запускает контейнеры заново

### Ручное обновление

```bash
# 1. Подключение к серверу
ssh cashback-server

# 2. Переход в директорию проекта
cd ~/CashbackAdvisor

# 3. Получение обновлений
git pull origin main

# 4. Остановка контейнеров
docker-compose -f docker-compose.full.yml down

# 5. Пересборка (с кешем)
docker-compose -f docker-compose.full.yml build

# Или без кеша (полная пересборка)
docker-compose -f docker-compose.full.yml build --no-cache

# 6. Запуск контейнеров
docker-compose -f docker-compose.full.yml up -d

# 7. Проверка статуса
docker-compose -f docker-compose.full.yml ps
```

---

## Мониторинг

### Просмотр логов

**Все сервисы**:
```bash
docker-compose -f docker-compose.full.yml logs -f
```

**Конкретный сервис**:
```bash
# Бот
docker-compose -f docker-compose.full.yml logs -f bot

# API сервер
docker-compose -f docker-compose.full.yml logs -f api

# База данных
docker-compose -f docker-compose.full.yml logs -f postgres
```

**Последние N строк**:
```bash
docker-compose -f docker-compose.full.yml logs --tail=100 bot
```

### Проверка статуса

```bash
# Статус контейнеров
docker-compose -f docker-compose.full.yml ps

# Использование ресурсов
docker stats

# Проверка здоровья
curl http://localhost:8080/health
```

### Мониторинг базы данных

```bash
# Подключение к БД
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db

# Проверка размера таблиц
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## Резервное копирование

### Автоматическое резервное копирование БД

Создайте скрипт `backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/home/cashback/backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Создание резервной копии
docker-compose -f ~/CashbackAdvisor/docker-compose.full.yml exec -T postgres pg_dump -U postgres cashback_db > $BACKUP_DIR/backup_$DATE.sql

# Удаление старых резервных копий (старше 7 дней)
find $BACKUP_DIR -name "backup_*.sql" -mtime +7 -delete

echo "Резервная копия создана: backup_$DATE.sql"
```

**Настройка cron для автоматического резервного копирования**:

```bash
# Редактирование crontab
crontab -e

# Добавление задачи (ежедневно в 2:00)
0 2 * * * /home/cashback/backup.sh >> /home/cashback/backup.log 2>&1
```

### Восстановление из резервной копии

```bash
# Остановка приложения
docker-compose -f docker-compose.full.yml down

# Восстановление БД
docker-compose -f docker-compose.full.yml up -d postgres
sleep 10
docker-compose -f docker-compose.full.yml exec -T postgres psql -U postgres -d cashback_db < backup_20241215_120000.sql

# Запуск приложения
docker-compose -f docker-compose.full.yml up -d
```

---

## Безопасность

### Рекомендации

1. **Пароли**:
   - Используйте надежные пароли (минимум 16 символов)
   - Не храните пароли в открытом виде
   - Регулярно меняйте пароли

2. **SSL/TLS**:
   - Включите SSL для PostgreSQL: `DB_SSLMODE=require`
   - Используйте HTTPS для внешнего доступа к API
   - Настройте SSL сертификаты через Let's Encrypt

3. **Firewall**:
   - Откройте только необходимые порты
   - Закройте прямой доступ к PostgreSQL (только через Docker сеть)

4. **Обновления**:
   - Регулярно обновляйте систему
   - Обновляйте Docker образы
   - Следите за уязвимостями

5. **Логирование**:
   - Настройте ротацию логов
   - Мониторьте логи на подозрительную активность

### Настройка ротации логов

Создайте файл `/etc/logrotate.d/docker-containers`:

```
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    size=10M
    missingok
    delaycompress
    copytruncate
}
```

---

## Масштабирование

### Горизонтальное масштабирование API

Для масштабирования API сервера:

1. Запустите несколько экземпляров API за балансировщиком нагрузки
2. Используйте Nginx или другой балансировщик
3. Настройте sticky sessions (если нужно)

**Пример конфигурации Nginx для балансировки**:

```nginx
upstream api_backend {
    least_conn;
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://api_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Вертикальное масштабирование

Для увеличения производительности:

1. Увеличьте ресурсы сервера (RAM, CPU)
2. Настройте пул соединений БД (в `internal/database/database.go`)
3. Оптимизируйте индексы БД

---

## Откат изменений

### Откат через Git

```bash
# Просмотр истории коммитов
git log --oneline

# Откат к предыдущему коммиту
git reset --hard HEAD~1

# Пересборка и перезапуск
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml build
docker-compose -f docker-compose.full.yml up -d
```

### Откат миграций БД

```bash
# Применение миграции отката
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d cashback_db -f migrations/002_down.sql
```

---

## Полезные команды

### Управление контейнерами

```bash
# Перезапуск сервиса
docker-compose -f docker-compose.full.yml restart bot

# Остановка сервиса
docker-compose -f docker-compose.full.yml stop api

# Просмотр использования ресурсов
docker stats

# Очистка неиспользуемых ресурсов
docker system prune -a
```

### Отладка

```bash
# Подключение к контейнеру
docker-compose -f docker-compose.full.yml exec bot sh
docker-compose -f docker-compose.full.yml exec api sh

# Проверка переменных окружения
docker-compose -f docker-compose.full.yml exec bot env
```

---

## Следующие шаги

- [Решение проблем](09-troubleshooting.md) — решение проблем при развертывании
- [Разработка](08-development.md) — разработка новых функций
- [База данных](06-database.md) — работа с БД

