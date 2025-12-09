# Инструкция по деплою CashbackAdvisor

## Предварительные требования

На сервере должны быть установлены:
- Docker (версия 20.10+)
- Docker Compose (версия 2.0+)
- Git

## Шаг 1: Подключение к серверу

```bash
ssh cashback@82.26.150.98
```

## Шаг 2: Установка Docker (если не установлен)

```bash
# Обновление системы
sudo apt-get update
sudo apt-get upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo apt-get install docker-compose-plugin -y

# Выход и повторный вход для применения изменений группы
exit
# Снова подключитесь: ssh cashback@82.26.150.98

# Проверка установки
docker --version
docker compose version
```

## Шаг 3: Клонирование репозитория

```bash
# Клонирование репозитория
git clone git@github.com:Aqoouet/CashbackAdvisor.git
cd CashbackAdvisor

# Или если используете HTTPS:
git clone https://github.com/Aqoouet/CashbackAdvisor.git
cd CashbackAdvisor
```

## Шаг 4: Настройка переменных окружения

### Вариант 1: Создание .env файла вручную

```bash
# Копирование примера конфигурации
cp env.example .env

# Редактирование файла .env
nano .env
```

В файле `.env` укажите:
```env
# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=УКАЖИТЕ_НАДЕЖНЫЙ_ПАРОЛЬ
DB_NAME=cashback_db
DB_SSLMODE=disable

# Сервер API
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Telegram Bot
TELEGRAM_BOT_TOKEN=ВАШ_ТОКЕН_БОТА_ЗДЕСЬ
API_BASE_URL=http://api:8080
BOT_DEBUG=false
```

### Вариант 2: Использование переменной окружения

```bash
# Экспорт токена в переменную окружения
export TELEGRAM_BOT_TOKEN="ваш_токен_бота_от_BotFather"

# Проверка
echo $TELEGRAM_BOT_TOKEN
```

## Шаг 5: Запуск деплоя

```bash
# Сделать скрипт исполняемым
chmod +x deploy.sh

# Запустить деплой
./deploy.sh
```

Скрипт автоматически:
- Проверит наличие TELEGRAM_BOT_TOKEN
- Создаст/обновит файл .env
- Соберёт Docker образы
- Запустит все сервисы (PostgreSQL, API, Bot)
- Применит миграции базы данных
- Проверит статус всех контейнеров

## Шаг 6: Проверка работоспособности

```bash
# Проверка статуса контейнеров
docker-compose -f docker-compose.full.yml ps

# Проверка логов API
docker-compose -f docker-compose.full.yml logs -f api

# Проверка логов бота
docker-compose -f docker-compose.full.yml logs -f bot

# Проверка логов базы данных
docker-compose -f docker-compose.full.yml logs -f postgres

# Проверка здоровья API
curl http://localhost:8080/health
```

## Управление сервисами

### Просмотр логов

```bash
# Все сервисы
docker-compose -f docker-compose.full.yml logs -f

# Только API
docker-compose -f docker-compose.full.yml logs -f api

# Только бот
docker-compose -f docker-compose.full.yml logs -f bot

# Последние 100 строк
docker-compose -f docker-compose.full.yml logs --tail=100 api
```

### Перезапуск сервисов

```bash
# Перезапуск всех сервисов
docker-compose -f docker-compose.full.yml restart

# Перезапуск только бота
docker-compose -f docker-compose.full.yml restart bot

# Перезапуск только API
docker-compose -f docker-compose.full.yml restart api
```

### Остановка и удаление

```bash
# Остановка всех сервисов
docker-compose -f docker-compose.full.yml stop

# Остановка и удаление контейнеров
docker-compose -f docker-compose.full.yml down

# Остановка и удаление контейнеров + volumes (УДАЛИТ ВСЕ ДАННЫЕ!)
docker-compose -f docker-compose.full.yml down -v
```

### Обновление кода

```bash
# Получение последних изменений
git pull origin main

# Пересборка и перезапуск
docker-compose -f docker-compose.full.yml down
docker-compose -f docker-compose.full.yml build --no-cache
docker-compose -f docker-compose.full.yml up -d
```

## Открытие портов (если используется firewall)

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 8080/tcp
sudo ufw allow 22/tcp  # SSH
sudo ufw enable

# Проверка статуса
sudo ufw status
```

## Настройка автозапуска

Docker Compose контейнеры настроены с `restart: unless-stopped`, поэтому они автоматически перезапустятся при перезагрузке сервера.

## Резервное копирование базы данных

```bash
# Создание бэкапа
docker exec cashback_postgres pg_dump -U postgres cashback_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановление из бэкапа
cat backup_20241209_120000.sql | docker exec -i cashback_postgres psql -U postgres -d cashback_db
```

## Мониторинг ресурсов

```bash
# Использование ресурсов контейнерами
docker stats

# Размер контейнеров и образов
docker system df
```

## Troubleshooting

### Бот не запускается

```bash
# Проверьте логи
docker-compose -f docker-compose.full.yml logs bot

# Проверьте токен
docker exec cashback_bot env | grep TELEGRAM_BOT_TOKEN
```

### API не отвечает

```bash
# Проверьте логи
docker-compose -f docker-compose.full.yml logs api

# Проверьте подключение к базе данных
docker exec cashback_api env | grep DB_
```

### База данных не запускается

```bash
# Проверьте логи PostgreSQL
docker-compose -f docker-compose.full.yml logs postgres

# Проверьте наличие порта
sudo netstat -tulpn | grep 5432
```

### Очистка Docker (освобождение места)

```bash
# Удаление неиспользуемых образов, контейнеров, сетей
docker system prune -a

# Удаление неиспользуемых volumes (ОСТОРОЖНО!)
docker volume prune
```

## Получение токена Telegram бота

1. Откройте Telegram
2. Найдите бота @BotFather
3. Отправьте команду `/newbot`
4. Следуйте инструкциям
5. Получите токен вида: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`
6. Сохраните токен в безопасном месте

## Полезные ссылки

- [Документация Docker](https://docs.docker.com/)
- [Документация Docker Compose](https://docs.docker.com/compose/)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [PostgreSQL Docker](https://hub.docker.com/_/postgres)

## Безопасность

1. **Никогда не коммитьте файл .env в git!**
2. Используйте сильные пароли для базы данных
3. Регулярно обновляйте систему и Docker
4. Настройте firewall
5. Используйте SSH ключи вместо паролей
6. Регулярно делайте резервные копии базы данных

