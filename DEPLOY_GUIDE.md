# Руководство по обновлению бота на сервере

## Быстрый способ (рекомендуется)

1. **Закоммитьте и запушьте изменения локально:**
   ```bash
   git add .
   git commit -m "Описание изменений"
   git push origin main
   ```

2. **Подключитесь к серверу и запустите скрипт обновления:**
   ```bash
   ssh cashback-server "cd ~/CashbackAdvisor && bash update.sh"
   ```

   Скрипт автоматически:
   - Получит последние изменения из Git (`git pull`)
   - Остановит контейнеры
   - Пересоберёт Docker образы
   - Запустит контейнеры заново

## Альтернативный способ (ручной)

Если нужно больше контроля:

```bash
# 1. Подключиться к серверу
ssh cashback-server

# 2. Перейти в директорию проекта
cd ~/CashbackAdvisor

# 3. Получить последние изменения
git pull origin main

# 4. Остановить контейнеры
docker-compose -f docker-compose.full.yml down

# 5. Пересобрать образы (с кешем)
docker-compose -f docker-compose.full.yml build

# Или без кеша (если нужна полная пересборка)
docker-compose -f docker-compose.full.yml build --no-cache

# 6. Запустить контейнеры
docker-compose -f docker-compose.full.yml up -d

# 7. Проверить статус
docker-compose -f docker-compose.full.yml ps

# 8. Посмотреть логи (опционально)
docker-compose -f docker-compose.full.yml logs -f bot
```

## Полезные команды

**Просмотр логов бота:**
```bash
ssh cashback-server "docker-compose -f ~/CashbackAdvisor/docker-compose.full.yml logs --tail=100 bot"
```

**Просмотр логов в реальном времени:**
```bash
ssh cashback-server "docker-compose -f ~/CashbackAdvisor/docker-compose.full.yml logs -f bot"
```

**Перезапуск только бота (без пересборки):**
```bash
ssh cashback-server "docker-compose -f ~/CashbackAdvisor/docker-compose.full.yml restart bot"
```

**Проверка статуса всех сервисов:**
```bash
ssh cashback-server "docker-compose -f ~/CashbackAdvisor/docker-compose.full.yml ps"
```

## Примечания

- Скрипт `update.sh` использует кеш Docker для ускорения сборки
- Для полной пересборки используйте: `bash update.sh --no-cache`
- После обновления проверьте логи, чтобы убедиться, что бот запустился корректно
- Если что-то пошло не так, можно откатить изменения через `git revert` или `git reset`

