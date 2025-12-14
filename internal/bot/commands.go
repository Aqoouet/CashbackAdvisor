package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStart обрабатывает команду /start.
func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := fmt.Sprintf(`👋 Привет! Я помогаю не упустить выгодный кэшбэк.

🎯 Что я умею:
• Запоминаю условия кэшбэка от разных банков
• Подсказываю, где сейчас самый выгодный кэшбэк
• Исправляю опечатки в названиях банков и категорий
• Работаю с группами - делитесь с друзьями!

👥 Группы (обязательно!):
/creategroup Название - создать
/joingroup Название - присоединиться
/groupinfo - инфо о вашей группе

✍️ Добавить %% кешбек (месяц опционален):
"Тинькофф, Такси, 5%%, 3000"
"Сбер, Перекресток доставка, 12%%, 30000, январь"

🔍 Найти лучший кэшбэк (без запятых):
"Такси" - покажет для текущего месяца
"Перекресток доставка"

📋 Команды:
/list - все %% кешбека группы
/best - найти лучший кэшбэк
/update ID - обновить свой %% кешбек
/delete ID - удалить свой %% кешбек
/help - подробная справка

Я пойму, проверю и сохраню! 😊

ℹ️ Версия: %s`, BuildInfo())

	b.sendText(message.Chat.ID, text)
}

// handleHelp обрабатывает команду /help.
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := fmt.Sprintf(`📖 Подробная справка (Версия: %s)

👥 Группы:
🔹 /creategroup Название - Создать группу
🔹 /joingroup Название - Присоединиться
🔹 /groupinfo [Название] - Информация

💳 Кэшбэк:
🔹 /add - Добавить новый %% кешбек
🔹 /list - Показать все %% кешбека группы
🔹 /best - Найти лучший кэшбэк
🔹 /update ID - Обновить свой %% кешбек
🔹 /delete ID - Удалить свой %% кешбек
🔹 /cancel - Отменить текущую операцию

💡 Формат добавления (с запятыми):
Банк, Категория, Процент, Сумма[, Месяц]

📝 Примеры добавления:
• "Тинькофф, Такси, 5%%, 3000" (месяц = текущий)
• "Сбер, Супермаркеты, 10, 5000, январь"
• "Альфа, Рестораны, 7.5, 4000"
• "ВТБ, Перекресток доставка, 12, 30000, март"

🔍 Поиск лучшего кэшбэка (БЕЗ запятых):
Бот найдёт лучшее предложение среди участников группы!
• "Такси" (покажет для текущего месяца)
• "Перекресток доставка"
• "Рестораны"

💡 Все операции в рамках вашей группы!
Делитесь кэшбэком с друзьями! 🤝

✨ Бот умеет исправлять опечатки! 😊`, BuildInfo())

	b.sendText(message.Chat.ID, text)
}

// handleAddCommand обрабатывает команду /add.
func (b *Bot) handleAddCommand(message *tgbotapi.Message) {
	text := `📝 Отправьте данные о кэшбэке.

Пример: "Тинькофф такси 5% 3000р декабрь"

Или используйте /cancel для отмены.`

	b.sendText(message.Chat.ID, text)
}

// handleBestCommand обрабатывает команду /best.
func (b *Bot) handleBestCommand(message *tgbotapi.Message) {
	text := `🔍 Для поиска лучшего кэшбэка отправьте:

📝 Просто напишите категорию и месяц:

Примеры:
• "Лучший кэшбэк такси декабрь"
• "Где выгоднее рестораны январь"
• "Такси декабрь"
• "Супермаркеты февраль"`

	b.sendText(message.Chat.ID, text)
}

// handleList обрабатывает команду /list.
func (b *Bot) handleList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	list, err := b.client.ListCashback(groupName, 100, 0)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка: %s", err))
		return
	}

	b.sendText(message.Chat.ID, formatCashbackList(list.Rules, list.Total))
}

// handleUpdateCommand обрабатывает команду /update ID.
func (b *Bot) handleUpdateCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendText(message.Chat.ID, "❌ Укажите ID %% кешбека.\n\nПример: /update 5")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %% кешбек с ID %d не найден.", id))
		return
	}

	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете обновлять только свой %% кешбек.")
		return
	}

	b.sendText(message.Chat.ID, formatUpdatePrompt(rule))
	b.setState(message.From.ID, StateAwaitingUpdateData, nil, nil, id)
}

// handleDeleteCommand обрабатывает команду /delete ID.
func (b *Bot) handleDeleteCommand(message *tgbotapi.Message) {
	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendText(message.Chat.ID, "❌ Укажите ID %% кешбека.\n\nПример: /delete 5")
		return
	}

	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Неверный формат ID. Используйте число.")
		return
	}

	rule, err := b.client.GetCashbackByID(id)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ %% кешбек с ID %d не найден.", id))
		return
	}

	// Проверяем владельца
	if rule.UserID != strconv.FormatInt(message.From.ID, 10) {
		b.sendText(message.Chat.ID, "❌ Вы можете удалять только свой %% кешбек.")
		return
	}

	b.sendWithButtons(message.Chat.ID, formatDeletePrompt(rule), ButtonsDelete)
	b.setState(message.From.ID, StateAwaitingDeleteConfirm, nil, nil, id)
}

// handleCancel обрабатывает команду /cancel.
func (b *Bot) handleCancel(message *tgbotapi.Message) {
	b.clearState(message.From.ID)
	b.sendText(message.Chat.ID, "🚫 Операция отменена")
}

