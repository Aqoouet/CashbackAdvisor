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

⚠️ ВАЖНО: 
Бот НЕ находит информацию в интернете!
Он показывает только кэшбэк, добавленный участниками вашей группы.
Без группы бот работать не будет!

👥 Группы (обязательно!):
/creategroup Название - создать
/joingroup Название - присоединиться
/groupinfo - инфо о вашей группе

✍️ Добавить кешбек (дата опциональна):
"Тинькофф, Такси, 5%%, 3000"
"Сбер, Перекресток доставка, 12%%, 30000, 31.01.2025"

🔍 Найти лучший кэшбэк (без запятых):
"Такси" - покажет для текущего месяца
"Перекресток доставка"

📋 Команды:
/list - все кешбека группы
/best - найти лучший кэшбэк
/update ID - обновить свой кешбек
/delete ID - удалить свой кешбек
/help - подробная справка

Я пойму, проверю и сохраню! 😊

ℹ️ Версия: %s`, BuildInfo())

	b.sendText(message.Chat.ID, text)
}

// handleHelp обрабатывает команду /help.
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := fmt.Sprintf(`📖 Подробная справка (Версия: %s)

⚠️ ВАЖНАЯ ИНФОРМАЦИЯ:
Бот НЕ ищет информацию в интернете!
Он показывает только кэшбэк, который добавили участники вашей группы.

💡 Назначение бота:
• Сохранять информацию о кэшбэке ваших карт
• Делиться кэшбэком с друзьями в группе
• Быстро находить лучшее предложение

🔐 Группы нужны для:
• Разделения кэшбэка разных коллективов
• Совместного использования информации
• Без группы бот НЕ РАБОТАЕТ!

👥 Группы:
🔹 /creategroup Название - Создать группу
🔹 /joingroup Название - Присоединиться
🔹 /groupinfo [Название] - Информация

💳 Кэшбэк:
🔹 /add - Добавить новый кешбек
🔹 /list - Показать все кешбека группы
🔹 /best - Найти лучший кэшбэк
🔹 /update ID - Обновить свой кешбек
🔹 /delete ID - Удалить свой кешбек
🔹 /cancel - Отменить текущую операцию

💡 Формат добавления (с запятыми):
Банк, Категория, Процент, Сумма[, Дата окончания]

📝 Примеры добавления:
• "Тинькофф, Такси, 5%%, 3000" (дата = конец текущего месяца)
• "Сбер, Супермаркеты, 10, 5000, 31.01.2025"
• "Альфа, Рестораны, 7.5, 4000, 28.02.2025"
• "ВТБ, Перекресток доставка, 12, 30000, 31.03.2025"

📅 Формат даты окончания: дд.мм.гггг
Например: 31.12.2024, 28.02.2025

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

Формат: Банк, Категория, Процент, Сумма[, Дата окончания]

Примеры:
• "Тинькофф, Такси, 5%, 3000"
• "Сбер, Супермаркеты, 10, 5000, 31.01.2025"

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

// handleList обрабатывает команду /list с поддержкой пагинации.
// Форматы:
// /list - последние 5 строк
// /list all - все строки
// /list 1-10 - строки с 1 по 10
// /list 1-5,8,10 - строки с 1 по 5, а также 8 и 10
func (b *Bot) handleList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Парсим аргументы команды
	args := strings.TrimPrefix(message.Text, "/list")
	args = strings.TrimSpace(args)
	
	indices, showAll, err := ParseListArguments(args)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Неверный формат: %s\n\n"+
			"Примеры:\n"+
			"• /list - последние 5\n"+
			"• /list all - все\n"+
			"• /list 1-10 - с 1 по 10\n"+
			"• /list 1-5,8,10 - с 1 по 5, а также 8 и 10", err))
		return
	}

	// Получаем все записи группы
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Ошибка: %s", err))
		return
	}

	// Фильтруем записи по индексам
	var filtered []models.CashbackRule
	if showAll {
		filtered = list.Rules
	} else if indices == nil {
		// По умолчанию - последние 5
		start := 0
		if len(list.Rules) > 5 {
			start = len(list.Rules) - 5
		}
		filtered = list.Rules[start:]
	} else {
		// Выбираем по индексам
		for _, idx := range indices {
			if idx > 0 && idx <= len(list.Rules) {
				filtered = append(filtered, list.Rules[idx-1])
			}
		}
	}

	if len(filtered) == 0 {
		b.sendText(message.Chat.ID, "📝 Нет записей для отображения.")
		return
	}

	b.sendText(message.Chat.ID, formatCashbackListTable(filtered, list.Total, showAll, indices))
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

// handleBankInfo обрабатывает команду /bankinfo bank_name.
func (b *Bot) handleBankInfo(message *tgbotapi.Message) {
	args := strings.TrimPrefix(message.Text, "/bankinfo")
	args = strings.TrimSpace(args)

	if args == "" {
		b.sendText(message.Chat.ID, "❌ Укажите название банка.\n\nПример: /bankinfo Тинькофф")
		return
	}

	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Попытка найти похожий банк
	correctedBank, found := FindSimilarBank(args)
	bankToSearch := args
	if found && correctedBank != args {
		bankToSearch = correctedBank
	}

	rules, err := b.client.GetCashbackByBank(groupName, bankToSearch)
	if err != nil {
		b.sendText(message.Chat.ID, fmt.Sprintf("❌ Кэшбэки для банка \"%s\" не найдены.\n\n"+
			"💡 Используйте /banklist для просмотра доступных банков.", args))
		return
	}

	b.sendText(message.Chat.ID, formatBankInfo(bankToSearch, rules))
}

// handleCategoryList обрабатывает команду /categorylist.
func (b *Bot) handleCategoryList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	categories, err := b.client.GetActiveCategories(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения категорий")
		return
	}

	if len(categories) == 0 {
		b.sendText(message.Chat.ID, "📝 Пока нет активных категорий в группе.")
		return
	}

	b.sendText(message.Chat.ID, formatCategoryList(categories))
}

// handleBankList обрабатывает команду /banklist.
func (b *Bot) handleBankList(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	banks, err := b.client.GetActiveBanks(groupName)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения банков")
		return
	}

	if len(banks) == 0 {
		b.sendText(message.Chat.ID, "📝 Пока нет активных банков в группе.")
		return
	}

	b.sendText(message.Chat.ID, formatBankList(banks))
}

// handleUserInfo обрабатывает команду /userinfo [ID].
func (b *Bot) handleUserInfo(message *tgbotapi.Message) {
	userIDStr := strconv.FormatInt(message.From.ID, 10)
	groupName, err := b.client.GetUserGroup(userIDStr)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Вы должны быть в группе. Используйте /creategroup или /joingroup")
		return
	}

	// Парсим аргументы
	args := strings.TrimPrefix(message.Text, "/userinfo")
	args = strings.TrimSpace(args)

	targetUserID := userIDStr
	if args != "" {
		// Указан ID другого пользователя
		targetUserID = args
	}

	// Получаем все кэшбэки группы
	list, err := b.client.ListCashback(groupName, 1000, 0)
	if err != nil {
		b.sendText(message.Chat.ID, "❌ Ошибка получения данных")
		return
	}

	// Фильтруем по пользователю
	var userRules []models.CashbackRule
	for _, rule := range list.Rules {
		if rule.UserID == targetUserID {
			userRules = append(userRules, rule)
		}
	}

	if len(userRules) == 0 {
		if targetUserID == userIDStr {
			b.sendText(message.Chat.ID, "📝 У вас пока нет кэшбэков.")
		} else {
			b.sendText(message.Chat.ID, fmt.Sprintf("📝 У пользователя %s пока нет кэшбэков.", targetUserID))
		}
		return
	}

	b.sendText(message.Chat.ID, formatUserInfo(userRules))
}

