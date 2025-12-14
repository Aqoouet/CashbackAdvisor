# Справочник API

## Базовый URL

```
http://localhost:8080
```

Для продакшн замените `localhost:8080` на адрес вашего сервера.

## Формат ответов

### Успешный ответ

Все успешные ответы возвращаются со статусом `200 OK` (кроме `201 Created` для создания) и содержат JSON данные.

### Ошибки

Ошибки возвращаются в следующем формате:

```json
{
  "error": "Описание ошибки",
  "details": ["Деталь 1", "Деталь 2"]
}
```

Коды статусов:
- `400 Bad Request` — неверный формат запроса или ошибки валидации
- `404 Not Found` — ресурс не найден
- `500 Internal Server Error` — внутренняя ошибка сервера

## Эндпоинты

### Health Check

Проверка работоспособности сервера.

**Запрос**:
```http
GET /health
```

**Ответ**:
```json
{
  "status": "ok"
}
```

**Пример**:
```bash
curl http://localhost:8080/health
```

---

## Управление правилами кэшбэка

### Анализ и предложения (Suggest)

Анализирует данные и возвращает предложения для исправления опечаток через fuzzy-поиск.

**Запрос**:
```http
POST /api/v1/cashback/suggest
Content-Type: application/json
```

**Тело запроса**:
```json
{
  "group_name": "Транспорт",
  "category": "такси",
  "bank_name": "Тинькоф",
  "user_display_name": "Иван",
  "month_year": "2024-12",
  "cashback_percent": 5.5,
  "max_amount": 3000.0
}
```

**Параметры**:
- `group_name` (string, обязательный) — название группы
- `category` (string, обязательный) — категория покупок
- `bank_name` (string, обязательный) — название банка
- `user_display_name` (string, обязательный) — отображаемое имя пользователя
- `month_year` (string, обязательный) — месяц и год в формате `YYYY-MM`
- `cashback_percent` (float, обязательный) — процент кэшбэка (0-100)
- `max_amount` (float, обязательный) — максимальная сумма кэшбэка (>= 0)

**Ответ**:
```json
{
  "valid": true,
  "can_proceed": true,
  "suggestions": {
    "group_name": [
      {
        "value": "Транспорт",
        "similarity": 0.95
      }
    ],
    "category": [
      {
        "value": "Такси",
        "similarity": 0.92
      }
    ],
    "bank_name": [
      {
        "value": "Тинькофф",
        "similarity": 0.88
      }
    ]
  }
}
```

**Пример с ошибками валидации**:
```json
{
  "valid": false,
  "can_proceed": false,
  "errors": [
    "month_year: неверный формат, ожидается YYYY-MM",
    "cashback_percent: значение должно быть от 0 до 100"
  ],
  "suggestions": {
    "group_name": [...],
    "category": [...]
  }
}
```

**Пример**:
```bash
curl -X POST http://localhost:8080/api/v1/cashback/suggest \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Транспорт",
    "category": "такси",
    "bank_name": "Тинькоф",
    "user_display_name": "Иван",
    "month_year": "2024-12",
    "cashback_percent": 5.5,
    "max_amount": 3000.0
  }'
```

---

### Создание правила кэшбэка

Создает новое правило кэшбэка.

**Запрос**:
```http
POST /api/v1/cashback
Content-Type: application/json
```

**Тело запроса**:
```json
{
  "group_name": "Транспорт",
  "category": "Такси",
  "bank_name": "Тинькофф",
  "user_id": "123456789",
  "user_display_name": "Иван",
  "month_year": "2024-12",
  "cashback_percent": 5.5,
  "max_amount": 3000.0,
  "force": false
}
```

**Параметры**:
- Все параметры обязательные, кроме `force`
- `force` (boolean, опциональный) — принудительное создание без валидации

**Ответ** (`201 Created`):
```json
{
  "id": 1,
  "group_name": "Транспорт",
  "category": "Такси",
  "bank_name": "Тинькофф",
  "user_id": "123456789",
  "user_display_name": "Иван",
  "month_year": "2024-12-01T00:00:00Z",
  "cashback_percent": 5.5,
  "max_amount": 3000.0,
  "created_at": "2024-12-15T10:30:00Z",
  "updated_at": "2024-12-15T10:30:00Z"
}
```

**Пример**:
```bash
curl -X POST http://localhost:8080/api/v1/cashback \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Транспорт",
    "category": "Такси",
    "bank_name": "Тинькофф",
    "user_id": "123456789",
    "user_display_name": "Иван",
    "month_year": "2024-12",
    "cashback_percent": 5.5,
    "max_amount": 3000.0
  }'
```

---

### Получение правила по ID

Получает правило кэшбэка по его идентификатору.

**Запрос**:
```http
GET /api/v1/cashback/{id}
```

**Параметры пути**:
- `id` (integer) — идентификатор правила

**Ответ** (`200 OK`):
```json
{
  "id": 1,
  "group_name": "Транспорт",
  "category": "Такси",
  "bank_name": "Тинькофф",
  "user_id": "123456789",
  "user_display_name": "Иван",
  "month_year": "2024-12-01T00:00:00Z",
  "cashback_percent": 5.5,
  "max_amount": 3000.0,
  "created_at": "2024-12-15T10:30:00Z",
  "updated_at": "2024-12-15T10:30:00Z"
}
```

**Пример**:
```bash
curl http://localhost:8080/api/v1/cashback/1
```

---

### Обновление правила

Обновляет существующее правило кэшбэка. Можно обновить только указанные поля.

**Запрос**:
```http
PUT /api/v1/cashback/{id}
Content-Type: application/json
```

**Тело запроса** (все поля опциональные):
```json
{
  "group_name": "Транспорт",
  "category": "Такси",
  "bank_name": "Сбербанк",
  "month_year": "2025-01",
  "cashback_percent": 6.0,
  "max_amount": 3500.0
}
```

**Параметры пути**:
- `id` (integer) — идентификатор правила

**Ответ** (`200 OK`):
```json
{
  "message": "Правило успешно обновлено"
}
```

**Пример**:
```bash
curl -X PUT http://localhost:8080/api/v1/cashback/1 \
  -H "Content-Type: application/json" \
  -d '{
    "cashback_percent": 6.0,
    "max_amount": 3500.0
  }'
```

---

### Удаление правила

Удаляет правило кэшбэка.

**Запрос**:
```http
DELETE /api/v1/cashback/{id}
```

**Параметры пути**:
- `id` (integer) — идентификатор правила

**Ответ** (`200 OK`):
```json
{
  "message": "Правило успешно удалено"
}
```

**Пример**:
```bash
curl -X DELETE http://localhost:8080/api/v1/cashback/1
```

---

### Список правил

Получает список правил кэшбэка с поддержкой пагинации.

**Запрос**:
```http
GET /api/v1/cashback?limit=20&offset=0&group_name=Транспорт
```

**Query параметры**:
- `limit` (integer, опциональный) — количество записей (по умолчанию 20, максимум 1000)
- `offset` (integer, опциональный) — смещение для пагинации (по умолчанию 0)
- `group_name` (string, опциональный) — фильтр по группе
- `user_id` (string, опциональный, устаревший) — фильтр по пользователю (legacy)

**Ответ** (`200 OK`):
```json
{
  "rules": [
    {
      "id": 1,
      "group_name": "Транспорт",
      "category": "Такси",
      "bank_name": "Тинькофф",
      "user_id": "123456789",
      "user_display_name": "Иван",
      "month_year": "2024-12-01T00:00:00Z",
      "cashback_percent": 5.5,
      "max_amount": 3000.0,
      "created_at": "2024-12-15T10:30:00Z",
      "updated_at": "2024-12-15T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

**Пример**:
```bash
curl "http://localhost:8080/api/v1/cashback?limit=10&offset=0&group_name=Транспорт"
```

---

### Лучший кэшбэк

Находит правило с лучшим процентом кэшбэка для указанной категории. Если точной категории нет, ищет правило "Все покупки".

**Запрос**:
```http
GET /api/v1/cashback/best?group_name=Транспорт&category=Такси&month_year=2024-12
```

**Query параметры**:
- `group_name` (string, обязательный) — название группы
- `category` (string, обязательный) — категория покупок
- `month_year` (string, обязательный) — месяц и год в формате `YYYY-MM`

**Ответ** (`200 OK`):
```json
{
  "id": 1,
  "group_name": "Транспорт",
  "category": "Такси",
  "bank_name": "Тинькофф",
  "user_id": "123456789",
  "user_display_name": "Иван",
  "month_year": "2024-12-01T00:00:00Z",
  "cashback_percent": 5.5,
  "max_amount": 3000.0,
  "created_at": "2024-12-15T10:30:00Z",
  "updated_at": "2024-12-15T10:30:00Z"
}
```

**Пример**:
```bash
curl "http://localhost:8080/api/v1/cashback/best?group_name=Транспорт&category=Такси&month_year=2024-12"
```

---

## Управление группами

### Создание группы

Создает новую группу пользователей.

**Запрос**:
```http
POST /api/v1/groups
Content-Type: application/json
```

**Тело запроса**:
```json
{
  "group_name": "Семья",
  "creator_id": "123456789"
}
```

**Параметры**:
- `group_name` (string, обязательный) — название группы
- `creator_id` (string, обязательный) — ID создателя группы

**Ответ** (`201 Created`):
```json
{
  "message": "Группа создана",
  "group_name": "Семья"
}
```

**Пример**:
```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Семья",
    "creator_id": "123456789"
  }'
```

---

### Список всех групп

Получает список всех существующих групп.

**Запрос**:
```http
GET /api/v1/groups
```

**Ответ** (`200 OK`):
```json
{
  "groups": ["Семья", "Друзья", "Коллеги"]
}
```

**Пример**:
```bash
curl http://localhost:8080/api/v1/groups
```

---

### Проверка существования группы

Проверяет, существует ли группа с указанным именем.

**Запрос**:
```http
GET /api/v1/groups/check?name=Семья
```

**Query параметры**:
- `name` (string, обязательный) — название группы

**Ответ** (`200 OK`):
```json
{
  "group_name": "Семья",
  "exists": true
}
```

**Пример**:
```bash
curl "http://localhost:8080/api/v1/groups/check?name=Семья"
```

---

### Участники группы

Получает список участников группы.

**Запрос**:
```http
GET /api/v1/groups/members?name=Семья
```

**Query параметры**:
- `name` (string, обязательный) — название группы

**Ответ** (`200 OK`):
```json
{
  "members": ["123456789", "987654321"]
}
```

**Пример**:
```bash
curl "http://localhost:8080/api/v1/groups/members?name=Семья"
```

---

## Управление пользователями и группами

### Получение группы пользователя

Получает название группы, к которой принадлежит пользователь.

**Запрос**:
```http
GET /api/v1/users/{userID}/group
```

**Параметры пути**:
- `userID` (string) — идентификатор пользователя

**Ответ** (`200 OK`):
```json
{
  "group_name": "Семья"
}
```

**Ошибка** (`404 Not Found`):
```json
{
  "error": "Пользователь не в группе"
}
```

**Пример**:
```bash
curl http://localhost:8080/api/v1/users/123456789/group
```

---

### Установка группы пользователя

Присоединяет пользователя к группе или меняет его группу.

**Запрос**:
```http
PUT /api/v1/users/{userID}/group
Content-Type: application/json
```

**Тело запроса**:
```json
{
  "group_name": "Семья"
}
```

**Параметры пути**:
- `userID` (string) — идентификатор пользователя

**Параметры тела**:
- `group_name` (string, обязательный) — название группы

**Ответ** (`200 OK`):
```json
{
  "message": "Группа установлена",
  "group_name": "Семья"
}
```

**Пример**:
```bash
curl -X PUT http://localhost:8080/api/v1/users/123456789/group \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Семья"
  }'
```

---

## Валидация данных

### Правила валидации

**Текстовые поля** (`group_name`, `category`, `bank_name`, `user_display_name`):
- Не могут быть пустыми
- Максимальная длина: 255 символов
- Должны содержать хотя бы один не пробельный символ

**month_year**:
- Формат: `YYYY-MM` (например, `2024-12`)
- Год: от 2000 до 2099
- Месяц: от 01 до 12

**cashback_percent**:
- Тип: число с плавающей точкой
- Диапазон: от 0.00 до 100.00
- Округляется до 2 знаков после запятой

**max_amount**:
- Тип: число с плавающей точкой
- Диапазон: >= 0.00
- Округляется до 2 знаков после запятой

**user_id**:
- Не может быть пустым
- Строковое значение

---

## Fuzzy-поиск

API использует триграммный поиск PostgreSQL для исправления опечаток. При вызове `/suggest` система ищет похожие значения в базе данных и возвращает предложения с коэффициентом схожести.

**Пороги схожести**:
- `group_name`: 0.6
- `category`: 0.6
- `bank_name`: 0.65
- `user_display_name`: 0.7

**Лимит предложений**: до 5 для каждого поля

---

## Примеры использования

### Полный цикл работы с правилом

1. **Анализ данных перед созданием**:
```bash
curl -X POST http://localhost:8080/api/v1/cashback/suggest \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Транспорт",
    "category": "такси",
    "bank_name": "Тинькоф",
    "user_display_name": "Иван",
    "month_year": "2024-12",
    "cashback_percent": 5.5,
    "max_amount": 3000.0
  }'
```

2. **Создание правила**:
```bash
curl -X POST http://localhost:8080/api/v1/cashback \
  -H "Content-Type: application/json" \
  -d '{
    "group_name": "Транспорт",
    "category": "Такси",
    "bank_name": "Тинькофф",
    "user_id": "123456789",
    "user_display_name": "Иван",
    "month_year": "2024-12",
    "cashback_percent": 5.5,
    "max_amount": 3000.0
  }'
```

3. **Получение лучшего кэшбэка**:
```bash
curl "http://localhost:8080/api/v1/cashback/best?group_name=Транспорт&category=Такси&month_year=2024-12"
```

4. **Обновление правила**:
```bash
curl -X PUT http://localhost:8080/api/v1/cashback/1 \
  -H "Content-Type: application/json" \
  -d '{
    "cashback_percent": 6.0
  }'
```

5. **Удаление правила**:
```bash
curl -X DELETE http://localhost:8080/api/v1/cashback/1
```

---

## Ограничения

- Максимальный `limit` для списка правил: 1000
- По умолчанию `limit`: 20
- Минимальный `offset`: 0
- Таймаут запроса: 60 секунд

## Следующие шаги

- [Команды бота](05-bot-commands.md) — работа с ботом через Telegram
- [База данных](06-database.md) — структура БД и миграции
- [Развертывание](07-deployment.md) — развертывание на продакшн

